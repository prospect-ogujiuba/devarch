package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/recipe"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

func (a *app) recipes() ([]recipe.Recipe, error) {
	if err := a.load(); err != nil {
		return nil, err
	}
	return recipe.Load(a.eng.Root, filepath.Join(a.configDir, "recipes"), filepath.Join(a.eng.Root, "recipes"))
}

// recipeEnv is the environment every recipe script receives.
func (a *app) recipeEnv() []string {
	env := []string{"DEVARCH_ROOT=" + a.eng.Root, "DEVARCH_APPS_DIR=" + a.eng.Settings.AppsDir, "DEVARCH_NETWORK=" + a.eng.Settings.Network}
	if self, err := os.Executable(); err == nil {
		env = append(env, "DEVARCH_BIN="+self)
	}
	return env
}

func newCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "new [recipe] [name] [flags...]",
		Short: "Create a project from a recipe (wordpress, laravel, javascript)",
		Long: `Create a project by running a recipe's bootstrap script. Flags are passed to the
script unchanged; ` + "`devarch new <recipe> --help`" + ` shows them. Without arguments, list recipes.`,
		Example: `  devarch new wordpress shop --profile clean
  devarch new laravel api --database sqlite
  devarch new javascript store --framework next --profile fullstack --start
  devarch --dry-run new wordpress shop`,
		DisableFlagParsing: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveDefault
			}
			rs, _ := a.recipes()
			var out []string
			for _, r := range rs {
				out = append(out, r.Name+"\t"+r.Title)
			}
			return out, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Flag parsing is disabled so script flags pass through. A leading
			// --dry-run is ours; anywhere else it reaches the script unchanged.
			if len(args) > 0 && args[0] == "--dry-run" {
				a.dryRun, args = true, args[1:]
			}
			if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
				return cmd.Help()
			}
			rs, err := a.recipes()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				tw := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
				fmt.Fprintln(tw, "RECIPE\tTITLE\tSTARTS")
				for _, r := range rs {
					fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Name, r.Title, joinOrDash(r.Requires))
				}
				tw.Flush()
				fmt.Fprintln(a.err, "devarch: run `devarch new <recipe> --help` for a recipe's options")
				return nil
			}
			r, err := recipe.Find(rs, args[0])
			if err != nil {
				return err
			}
			argv := args[1:]
			if err := r.Validate(r.Parse(argv)); err != nil {
				return err
			}
			if a.dryRun && !slices.Contains(argv, "--dry-run") {
				argv = append(argv, "--dry-run")
			}
			return a.base.Exec(runner.Cmd{Name: r.EntryPath(), Args: argv, Dir: a.eng.Root, Env: a.recipeEnv()})
		},
	}
}

func joinOrDash(list []string) string {
	if len(list) == 0 {
		return "-"
	}
	out := list[0]
	for _, s := range list[1:] {
		out += ", " + s
	}
	return out
}
