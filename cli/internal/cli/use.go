package cli

import (
	"fmt"
	"slices"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
)

func useCmd(a *app) *cobra.Command {
	var opts engine.UseOptions
	c := &cobra.Command{
		Use:   "use <service> [version]",
		Short: "Show or switch a service's version",
		Long: `Without a version, list the service's version choices. With one, record it in
~/.config/devarch/versions.env, rebuild services built from a Dockerfile, and
recreate the service if it is running.

Database versions are protected: a downgrade of an upgrade-only service, or a
major-version change of a same-major service, is refused while its data volume
exists.`,
		Example: `  devarch use php
  devarch use php 8.3
  devarch use postgres 17 --force`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: a.completeUse,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			svc, err := a.eng.Catalog.Resolve(args[0])
			if err != nil {
				return err
			}
			v := svc.Meta.Version
			if len(args) == 1 {
				if v == nil {
					return fmt.Errorf("%s has no switchable version (add x-devarch.version to its compose.yml)", svc.ID)
				}
				current := a.eng.SelectedVersion(svc)
				choices := v.Choices
				if len(choices) == 0 {
					choices = []string{v.Default}
				}
				tw := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
				fmt.Fprintf(tw, "%s versions (%s):\n", svc.Title(), v.Var)
				for _, ch := range choices {
					mark, note := " ", ""
					if ch == current {
						mark = "*"
					}
					if ch == v.Default {
						note = "default"
					}
					fmt.Fprintf(tw, "  %s %s\t%s\n", mark, ch, note)
				}
				if !slices.Contains(choices, current) {
					fmt.Fprintf(tw, "  * %s\t(unlisted)\n", current)
				}
				return tw.Flush()
			}
			res, err := a.eng.Use(contextOrBackground(cmd), svc, args[1], opts)
			if err != nil {
				return err
			}
			if a.dryRun {
				return nil
			}
			switch {
			case res.Recreated:
				fmt.Fprintf(a.out, "%s now runs %s (was %s)\n", svc.Title(), res.To, res.From)
			case opts.NoRecreate:
				fmt.Fprintf(a.out, "%s set to %s; not rebuilt or recreated\n", svc.Title(), res.To)
			default:
				fmt.Fprintf(a.out, "%s set to %s; it takes effect on `devarch up %s`\n", svc.Title(), res.To, svc.Name)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&opts.Force, "force", false, "allow unlisted versions and skip the data-volume check")
	c.Flags().BoolVar(&opts.NoRecreate, "no-recreate", false, "only record the version")
	return c
}

func (a *app) completeUse(cmd *cobra.Command, args []string, prefix string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		if err := a.load(); err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var out []string
		for _, s := range a.eng.Catalog.Services {
			if s.Meta.Version != nil {
				out = append(out, s.Name+"\t"+s.Title())
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
	if len(args) == 1 && a.load() == nil {
		if s, err := a.eng.Catalog.Resolve(args[0]); err == nil && s.Meta.Version != nil {
			return s.Meta.Version.Choices, cobra.ShellCompDirectiveNoFileComp
		}
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
