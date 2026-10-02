package cli

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
)

func configCmd(a *app) *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "Show the effective settings and where each came from",
		Long: `Show the effective config.yml settings and where each value came from.

  runtime        podman or docker (default podman)
  apps_dir       where projects live; relative to the checkout (default apps)
  network        shared external network (default microservices-net)
  hosts.manage   whether devarch edits the hosts file (default true)
  hosts.address  address .test names map to (default 127.0.0.1)
  editor         command that opens a project folder (default code)

The .test suffix and Nginx Proxy Manager are fixed: the local certificate,
the proxy's routing rules and the bootstraps all assume them.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			s := a.eng.Settings
			tw := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "KEY\tVALUE\tSOURCE")
			fmt.Fprintf(tw, "root\t%s\t%s\n", a.eng.Root, a.rootHow)
			for _, k := range state.Keys {
				v, _ := s.Get(k)
				fmt.Fprintf(tw, "%s\t%s\t%s\n", k, v, s.Sources[k])
			}
			tw.Flush()
			fmt.Fprintf(a.err, "devarch: settings file %s\n", filepath.Join(a.configDir, "config.yml"))
			return nil
		},
	}
	get := &cobra.Command{
		Use:       "get <key>",
		Short:     "Print one effective setting",
		Args:      cobra.ExactArgs(1),
		ValidArgs: state.Keys,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			v, err := a.eng.Settings.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(a.out, v)
			return nil
		},
	}
	set := &cobra.Command{
		Use:       "set <key> <value>",
		Short:     "Write one setting to config.yml",
		Example:   "  devarch config set editor 'code -n'\n  devarch config set hosts.manage false",
		Args:      cobra.ExactArgs(2),
		ValidArgs: state.Keys,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := state.Dir()
			if err != nil {
				return err
			}
			path := filepath.Join(dir, "config.yml")
			if a.dryRun {
				fmt.Fprintf(a.out, "would set %s: %s in %s\n", args[0], args[1], path)
				return nil
			}
			if err := state.SetConfig(dir, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(a.out, "set %s in %s\n", args[0], path)
			return nil
		},
	}
	c.AddCommand(get, set)
	return c
}
