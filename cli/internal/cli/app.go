package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/replace"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
)

func (a *app) appPaths(name string) replace.Paths {
	return replace.Paths{AppsDir: a.eng.Settings.AppsDir, App: name}
}

func appCmd(a *app) *cobra.Command {
	c := &cobra.Command{
		Use:   "app",
		Short: "Start JavaScript apps, back up projects, recover failed bootstraps",
		Long: `Bootstraps take a guard before changing a project and release it when they
succeed. If they fail, ` + "`devarch app recover <app>`" + ` moves the partial project to
<apps_dir>/.devarch-failed, drops the databases created for it, restores any
database it replaced, and moves the previous project back from
<apps_dir>/.devarch-backups. A recovery that stops part-way can be rerun.`,
	}
	var withReplace bool
	guard := &cobra.Command{
		Use:   "guard <app>",
		Short: "Take the provisioning guard (with --replace, move an existing project to backups)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			p := a.appPaths(args[0])
			if a.dryRun {
				fmt.Fprintf(a.out, "would guard %s in %s\n", args[0], p.RecordPath())
				return nil
			}
			r, err := replace.Guard(p, withReplace)
			if err != nil {
				return err
			}
			if r.Backup != "" {
				fmt.Fprintf(a.err, "devarch: moved %s to %s\n", r.Target, r.Backup)
			}
			fmt.Fprintf(a.out, "BACKUP_PATH=%s\n", r.Backup)
			return nil
		},
	}
	guard.Flags().BoolVar(&withReplace, "replace", false, "move an existing project to the backups directory instead of refusing")

	release := &cobra.Command{
		Use:   "release <app>",
		Short: "Release the guard after a successful run (backups are kept)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			r, err := replace.Load(a.appPaths(args[0]))
			if err != nil || a.dryRun {
				return err
			}
			if r.Backup != "" {
				fmt.Fprintf(a.err, "devarch: the previous %s is kept at %s\n", r.App, r.Backup)
			}
			return r.Release()
		},
	}

	recover := &cobra.Command{
		Use:   "recover <app>",
		Short: "Undo a failed run: quarantine, drop new databases, restore the previous project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			r, err := replace.Load(a.appPaths(args[0]))
			if err != nil {
				return err
			}
			if a.dryRun {
				fmt.Fprintf(a.out, "would recover %s: target %s, backup %s, %d databases\n", r.App, r.Target, orDash(r.Backup), len(r.Databases))
				return nil
			}
			err = r.Recover(contextOrBackground(cmd), a.eng, time.Now())
			if r.Failed != "" {
				fmt.Fprintf(a.err, "devarch: partial project quarantined at %s\n", r.Failed)
			}
			if r.Restored && r.Backup != "" {
				fmt.Fprintf(a.err, "devarch: previous project restored to %s\n", r.Target)
			}
			for _, p := range r.Problems {
				fmt.Fprintf(a.err, "devarch: recovery problem: %s\n", p)
			}
			return err
		},
	}

	backup := &cobra.Command{
		Use:   "backup <app>",
		Short: "Copy a project to the backups directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			p := a.appPaths(args[0])
			if _, err := os.Stat(p.Target()); err != nil {
				return fmt.Errorf("no project at %s", p.Target())
			}
			dest := replace.BackupPath(p)
			if !a.dryRun {
				if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
					return err
				}
			}
			if err := a.eng.Runner.Run(contextOrBackground(cmd), runner.Cmd{Name: "cp", Args: []string{"-a", p.Target(), dest}}); err != nil {
				return err
			}
			if !a.dryRun {
				fmt.Fprintln(a.out, dest)
			}
			return nil
		},
	}
	var startOpts engine.AppOptions
	start := &cobra.Command{
		Use:   "start <app>",
		Short: "Run a JavaScript app in its own Node container behind https://<app>.test",
		Long: `Run apps/<app> in its own node-<app> container. The app's package.json script
(default devarch) must serve HTTP on 0.0.0.0:3000. start brings up the shared
node router and the proxy, recreates the app container so changed options
apply, reloads the proxy, and registers <app>.test.`,
		Example: "  devarch app start storefront\n  devarch app start api --script dev --package-manager pnpm",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			plan, err := a.eng.PlanApp(args[0], startOpts)
			if err != nil {
				return err
			}
			if err := a.eng.StartApp(contextOrBackground(cmd), plan, startOpts); err != nil {
				return err
			}
			if !a.dryRun {
				fmt.Fprintf(a.out, "%s: https://%s.test\n", plan.Container, plan.App)
			}
			return nil
		},
	}
	start.Flags().StringVar(&startOpts.Script, "script", "devarch", "package.json script to run")
	start.Flags().StringVar(&startOpts.PackageManager, "package-manager", "auto", "auto (from the lock file), npm, pnpm or yarn")
	start.Flags().BoolVar(&startOpts.NoHosts, "no-hosts", false, "do not register <app>.test")

	var stopVolumes bool
	stop := &cobra.Command{
		Use:   "stop <app>",
		Short: "Remove a JavaScript app's Node container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			return a.eng.StopApp(contextOrBackground(cmd), args[0], stopVolumes)
		},
	}
	stop.Flags().BoolVar(&stopVolumes, "volumes", false, "also delete the app's cached node_modules volume")
	c.AddCommand(start, stop, guard, release, recover, backup)
	return c
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// guardFor loads the guard that `db create --app` records into.
func (a *app) guardFor(name string) (*replace.Record, error) {
	r, err := replace.Load(a.appPaths(name))
	if err != nil {
		return nil, errors.New(err.Error() + "; run `devarch app guard " + name + "` first")
	}
	return r, nil
}
