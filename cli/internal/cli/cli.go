// Package cli is the cobra transport over the engine. It holds no workflow
// logic of its own beyond argument handling and output formatting.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/root"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
)

// Version is set at build time with -ldflags "-X .../cli.Version=...".
var Version = "dev"

type app struct {
	out, err  io.Writer
	in        io.Reader
	dryRun    bool
	configDir string
	rootHow   root.How
	eng       *engine.Engine
	// base is the runner without the --dry-run wrapper; recipes get --dry-run
	// forwarded so the script prints its own plan.
	base     runner.Runner
	problems []catalog.Problem
	// newRunner is replaceable in tests.
	newRunner func(log io.Writer) runner.Runner
}

func (a *app) load() error {
	if a.eng != nil {
		return nil
	}
	dir, err := state.Dir()
	if err != nil {
		return err
	}
	a.configDir = dir
	cfg, err := state.LoadConfig(dir)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	rootDir, how, err := root.Find(cfg.Root, wd)
	if err != nil {
		return err
	}
	a.rootHow = how
	if how == root.FromWalk && cfg.Root == "" && !a.dryRun {
		cfg.Root = rootDir
		if err := state.SaveConfig(dir, cfg); err == nil {
			fmt.Fprintf(a.err, "devarch: remembered checkout %s in %s\n", rootDir, filepath.Join(dir, "config.yml"))
		}
	}
	cat, problems := catalog.Load([]catalog.Path{
		{Dir: filepath.Join(dir, "services"), Source: "user"},
		{Dir: filepath.Join(rootDir, "services-library"), Source: "builtin"},
	})
	versions, err := state.LoadVersions(dir)
	if err != nil {
		return err
	}
	var r runner.Runner
	if a.newRunner != nil {
		r = a.newRunner(a.err)
	} else {
		r = runner.System{Log: a.err}
	}
	a.base = r
	if a.dryRun {
		r = runner.DryRun{Inner: r, Log: a.out}
	}
	a.problems = problems
	a.eng = &engine.Engine{
		Root: rootDir, Catalog: cat, Runner: r, Versions: versions, Log: a.err, DryRun: a.dryRun,
		SaveVersions: func(v state.Versions) error { return state.SaveVersions(dir, v) },
		Hosts:        hosts.NewManager(r, filepath.Join(dir, "windows")),
	}
	return nil
}

// NewRoot builds the command tree.
func NewRoot(in io.Reader, out, errOut io.Writer) *cobra.Command {
	a := &app{in: in, out: out, err: errOut}
	return newRoot(a)
}

func newRoot(a *app) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "devarch",
		Short: "Run the DevArch service library by name",
		Long: `DevArch runs services from services-library/ by name with native podman compose.
Every action prints the command it runs; --dry-run prints without running.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.SetIn(a.in)
	rootCmd.SetOut(a.out)
	rootCmd.SetErr(a.err)
	rootCmd.PersistentFlags().BoolVar(&a.dryRun, "dry-run", false, "print native commands instead of running them")
	rootCmd.AddGroup(&cobra.Group{ID: "services", Title: "Services:"}, &cobra.Group{ID: "projects", Title: "Projects:"}, &cobra.Group{ID: "env", Title: "Environment:"})
	for _, c := range []*cobra.Command{lsCmd(a), upCmd(a), downCmd(a), restartCmd(a), psCmd(a), logsCmd(a), useCmd(a), composeCmd(a)} {
		c.GroupID = "services"
		rootCmd.AddCommand(c)
	}
	newC := newCmd(a)
	newC.GroupID = "projects"
	rootCmd.AddCommand(newC)
	for _, c := range []*cobra.Command{doctorCmd(a), hostsCmd(a), lintCmd(a)} {
		c.GroupID = "env"
		rootCmd.AddCommand(c)
	}
	return rootCmd
}

// Main runs the CLI and returns the process exit code.
func Main(args []string) int {
	rootCmd := NewRoot(os.Stdin, os.Stdout, os.Stderr)
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		if _, _, err := rootCmd.Find(args); err != nil && !isBuiltin(rootCmd, args[0]) {
			if path, lookErr := exec.LookPath("devarch-" + args[0]); lookErr == nil {
				return runExternal(path, args[1:])
			}
		}
	}
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err == nil {
		return 0
	}
	if errors.Is(err, errSilent) {
		return 1
	}
	var exit *runner.ExitError
	if errors.As(err, &exit) && exit.Stderr == "" {
		// The native command already reported its own failure.
		fmt.Fprintf(os.Stderr, "devarch: %s exited with status %d\n", firstWord(exit.Cmd), exit.Code)
		return exit.Code
	}
	fmt.Fprintln(os.Stderr, "devarch: "+err.Error())
	return 1
}

func firstWord(s string) string {
	s = strings.TrimPrefix(s, "(")
	if i := strings.Index(s, "&& "); i >= 0 {
		s = s[i+3:]
	}
	f := strings.Fields(s)
	if len(f) >= 2 {
		return f[0] + " " + f[1]
	}
	return s
}

func isBuiltin(rootCmd *cobra.Command, name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return true
		}
	}
	return name == "help" || name == "completion"
}

func runExternal(path string, args []string) int {
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = os.Environ()
	if dir, err := state.Dir(); err == nil {
		if cfg, err := state.LoadConfig(dir); err == nil {
			wd, _ := os.Getwd()
			if r, _, err := root.Find(cfg.Root, wd); err == nil {
				cmd.Env = append(cmd.Env, "DEVARCH_ROOT="+r)
			}
		}
	}
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "devarch: "+err.Error())
		return 1
	}
	return 0
}

// selectServices resolves positional names plus an optional --tag.
func (a *app) selectServices(names []string, tag string) ([]catalog.Service, error) {
	var out []catalog.Service
	seen := map[string]bool{}
	add := func(s catalog.Service) {
		if !seen[s.ID] {
			seen[s.ID] = true
			out = append(out, s)
		}
	}
	for _, n := range names {
		s, err := a.eng.Catalog.Resolve(n)
		if err != nil {
			return nil, err
		}
		add(s)
	}
	if tag != "" {
		tagged := a.eng.Catalog.Search("", tag)
		if len(tagged) == 0 {
			return nil, fmt.Errorf("no services are tagged %q", tag)
		}
		for _, s := range tagged {
			add(s)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("name at least one service, or use --tag")
	}
	return out, nil
}

func upCmd(a *app) *cobra.Command {
	var tag string
	var opts engine.UpOptions
	c := &cobra.Command{
		Use:   "up [service...]",
		Short: "Start services (and their requires) with podman compose up -d",
		Example: `  devarch up php mariadb nginx-proxy-manager --wait
  devarch up --tag project-management`,
		ValidArgsFunction: a.completeServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			svcs, err := a.selectServices(args, tag)
			if err != nil {
				return err
			}
			if err := a.eng.Up(cmd.Context(), svcs, opts); err != nil {
				return err
			}
			if !a.dryRun {
				a.printURLs(svcs)
			}
			return nil
		},
	}
	c.Flags().StringVar(&tag, "tag", "", "also start every service with this tag")
	c.Flags().BoolVarP(&opts.Wait, "wait", "w", false, "wait until containers are healthy")
	c.Flags().DurationVar(&opts.Timeout, "timeout", 120*time.Second, "how long --wait waits per service")
	c.Flags().BoolVar(&opts.NoRequires, "no-requires", false, "do not start x-devarch requires first")
	c.Flags().BoolVar(&opts.NoHosts, "no-hosts", false, "do not register missing .test hostnames")
	c.Flags().BoolVar(&opts.Build, "build", false, "rebuild Dockerfile-based services before starting")
	return c
}

func (a *app) printURLs(svcs []catalog.Service) {
	for _, s := range svcs {
		for _, u := range s.Meta.URLs {
			fmt.Fprintf(a.out, "%-28s %s\n", s.Title()+":", u)
		}
		if s.Meta.Notes != "" {
			fmt.Fprintf(a.out, "%-28s %s\n", "", s.Meta.Notes)
		}
	}
}

func downCmd(a *app) *cobra.Command {
	var tag string
	var volumes, yes bool
	c := &cobra.Command{
		Use:               "down [service...]",
		Short:             "Stop and remove service containers (volumes are kept)",
		ValidArgsFunction: a.completeServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			svcs, err := a.selectServices(args, tag)
			if err != nil {
				return err
			}
			if volumes && !a.dryRun {
				var names []string
				for _, s := range svcs {
					names = append(names, s.Volumes...)
				}
				if len(names) > 0 && !yes {
					ok, err := a.confirm(fmt.Sprintf("Permanently delete volumes %s?", strings.Join(names, ", ")))
					if err != nil {
						return err
					}
					if !ok {
						return errors.New("aborted; no volumes were deleted")
					}
				}
			}
			return a.eng.Down(cmd.Context(), svcs, volumes)
		},
	}
	c.Flags().StringVar(&tag, "tag", "", "also stop every service with this tag")
	c.Flags().BoolVar(&volumes, "volumes", false, "also delete the services' named volumes (asks first)")
	c.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask before deleting volumes")
	return c
}

func (a *app) confirm(question string) (bool, error) {
	if f, ok := a.in.(*os.File); !ok || !isTerminal(f) {
		return false, errors.New(question + " Re-run with --yes to confirm non-interactively")
	}
	fmt.Fprint(a.err, question+" [y/N] ")
	var answer string
	fmt.Fscanln(a.in, &answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func restartCmd(a *app) *cobra.Command {
	var tag string
	c := &cobra.Command{
		Use:               "restart [service...]",
		Short:             "Restart service containers",
		ValidArgsFunction: a.completeServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			svcs, err := a.selectServices(args, tag)
			if err != nil {
				return err
			}
			return a.eng.Restart(cmd.Context(), svcs)
		},
	}
	c.Flags().StringVar(&tag, "tag", "", "also restart every service with this tag")
	return c
}

func logsCmd(a *app) *cobra.Command {
	var follow bool
	var tail string
	c := &cobra.Command{
		Use:               "logs <service> [-- compose-logs-args...]",
		Short:             "Show service logs with podman compose logs",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: a.completeServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			svc, err := a.eng.Catalog.Resolve(args[0])
			if err != nil {
				return err
			}
			return a.eng.Runner.Exec(a.eng.LogsCmd(svc, follow, tail, args[1:]...))
		},
	}
	c.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	c.Flags().StringVarP(&tail, "tail", "n", "", "number of lines to show from the end")
	return c
}

func composeCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "compose <service> [--] <podman-compose-args...>",
		Short: "Run any podman compose command in a service directory",
		Long: `Run any podman compose command from the service's directory with its selected
version applied. Arguments after the service name are passed through unchanged.`,
		Example:            "  devarch compose postgres -- exec postgres psql -U postgres",
		DisableFlagParsing: true,
		ValidArgsFunction:  a.completeServices,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
				return cmd.Help()
			}
			if err := a.load(); err != nil {
				return err
			}
			svc, err := a.eng.Catalog.Resolve(args[0])
			if err != nil {
				return err
			}
			rest := args[1:]
			if len(rest) > 0 && rest[0] == "--" {
				rest = rest[1:]
			}
			if len(rest) == 0 {
				return errors.New("no podman compose arguments given")
			}
			return a.eng.Runner.Exec(a.eng.Compose(svc, rest...))
		},
	}
}

func (a *app) completeServices(_ *cobra.Command, _ []string, prefix string) ([]string, cobra.ShellCompDirective) {
	if err := a.load(); err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	counts := map[string]int{}
	for _, s := range a.eng.Catalog.Services {
		counts[s.Name]++
	}
	for _, s := range a.eng.Catalog.Services {
		name := s.Name
		if counts[name] > 1 {
			name = s.ID
		}
		if strings.HasPrefix(name, prefix) {
			out = append(out, name+"\t"+s.Title())
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func contextOrBackground(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}
