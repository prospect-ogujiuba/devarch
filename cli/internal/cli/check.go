package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/doctor"
	"github.com/prospect-ogujiuba/devarch/cli/internal/lint"
)

func lintCmd(a *app) *cobra.Command {
	var native, strict, asJSON bool
	c := &cobra.Command{
		Use:   "lint",
		Short: "Validate the catalog: compose files, x-devarch metadata, ports, images",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			findings := lint.Run(contextOrBackground(cmd), a.eng.Catalog, a.problems, lint.Options{Native: native, Runner: a.eng.Runner})
			errs, warns := lint.Count(findings)
			if asJSON {
				if findings == nil {
					findings = []lint.Finding{}
				}
				if err := writeJSON(a.out, findings); err != nil {
					return err
				}
			} else {
				for _, f := range findings {
					fmt.Fprintf(a.out, "%-7s %s: %s\n", f.Severity, f.Service, f.Message)
				}
				fmt.Fprintf(a.err, "devarch: %d services, %d errors, %d warnings\n", len(a.eng.Catalog.Services), errs, warns)
			}
			if errs > 0 || strict && warns > 0 {
				return errSilent
			}
			return nil
		},
	}
	c.Flags().BoolVar(&native, "native", false, "also run podman compose config for every service (slow)")
	c.Flags().BoolVar(&strict, "strict", false, "fail on warnings too")
	c.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return c
}

func doctorCmd(a *app) *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "doctor",
		Short: "Check Podman, network, ports, certificate, hosts and lingering",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			checks := doctor.Run(contextOrBackground(cmd), doctor.Env{Engine: a.eng, RootHow: string(a.rootHow), Problems: len(a.problems)})
			failed := false
			for _, c := range checks {
				failed = failed || c.Status == doctor.Fail
			}
			if asJSON {
				if err := writeJSON(a.out, checks); err != nil {
					return err
				}
			} else {
				for _, c := range checks {
					mark := map[doctor.Status]string{doctor.OK: "✓", doctor.Warn: "!", doctor.Fail: "✗"}[c.Status]
					fmt.Fprintf(a.out, "%s %-17s %s\n", mark, c.Name, c.Detail)
					if c.Fix != "" && c.Status != doctor.OK {
						fmt.Fprintf(a.out, "  %-17s → %s\n", "", c.Fix)
					}
				}
			}
			if failed {
				return errSilent
			}
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return c
}

// errSilent fails the command without another message: output already explains.
var errSilent = errors.New("")
