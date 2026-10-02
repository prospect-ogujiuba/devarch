package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
)

type lsRow struct {
	catalog.Service
	Version string `json:"version,omitempty"`
	State   string `json:"state"`
	Health  string `json:"health,omitempty"`
}

func lsCmd(a *app) *cobra.Command {
	var tag string
	var asJSON, running bool
	c := &cobra.Command{
		Use:     "ls [query]",
		Aliases: []string{"list", "search"},
		Short:   "List and search the service catalog",
		Example: `  devarch ls
  devarch ls queue
  devarch ls --tag project-management`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			query := ""
			if len(args) == 1 {
				query = args[0]
			}
			svcs := a.eng.Catalog.Search(query, tag)
			containers, psErr := a.eng.Containers(contextOrBackground(cmd))
			states := engine.States(svcs, containers)
			rows := make([]lsRow, 0, len(svcs))
			for _, s := range svcs {
				r := lsRow{Service: s, Version: a.eng.SelectedVersion(s), State: "unknown"}
				if psErr == nil {
					r.State, r.Health = states[s.ID].State, states[s.ID].Health
				}
				if running && r.State != "running" && r.State != "partial" {
					continue
				}
				rows = append(rows, r)
			}
			if asJSON {
				return writeJSON(a.out, rows)
			}
			writeServiceTable(a.out, rows)
			if psErr != nil {
				fmt.Fprintf(a.err, "devarch: container state unavailable: %v\n", psErr)
			}
			for _, p := range a.problems {
				fmt.Fprintf(a.err, "devarch: skipped %s\n", p.Error())
			}
			return nil
		},
	}
	c.Flags().StringVar(&tag, "tag", "", "only services with this tag")
	c.Flags().BoolVar(&running, "running", false, "only services with running containers")
	c.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return c
}

func stateLabel(state, health string) string {
	switch {
	case state == "absent":
		return "-"
	case health != "" && health != "healthy":
		return state + " (" + health + ")"
	}
	return state
}

func writeServiceTable(w io.Writer, rows []lsRow) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SERVICE\tVERSION\tSTATE\tTITLE\tURL")
	for _, r := range rows {
		url := ""
		if len(r.Meta.URLs) > 0 {
			url = r.Meta.URLs[0]
		}
		title := ""
		if r.Meta.Title != "" {
			title = r.Meta.Title
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.ID, dash(r.Version), stateLabel(r.State, r.Health), dash(title), url)
	}
	tw.Flush()
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func psCmd(a *app) *cobra.Command {
	var all, asJSON bool
	c := &cobra.Command{
		Use:   "ps",
		Short: "Show containers, matched to catalog services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			cs, err := a.eng.Containers(contextOrBackground(cmd))
			if err != nil {
				return err
			}
			var rows []engine.Container
			for _, c := range cs {
				if all || c.State == "running" {
					rows = append(rows, c)
				}
			}
			if asJSON {
				if rows == nil {
					rows = []engine.Container{}
				}
				return writeJSON(a.out, rows)
			}
			sort.SliceStable(rows, func(i, j int) bool {
				if (rows[i].Service == "") != (rows[j].Service == "") {
					return rows[i].Service != ""
				}
				return rows[i].Service < rows[j].Service
			})
			tw := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "CONTAINER\tSERVICE\tSTATUS\tPORTS")
			for _, c := range rows {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", c.Name, dash(c.Service), c.Status, formatPorts(c.Ports))
			}
			return tw.Flush()
		},
	}
	c.Flags().BoolVarP(&all, "all", "a", false, "include stopped containers")
	c.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	return c
}

func formatPorts(ps []engine.Port) string {
	parts := make([]string, 0, len(ps))
	for _, p := range ps {
		if p.HostPort == 0 {
			continue
		}
		host := p.HostIP
		if host == "" {
			host = "0.0.0.0"
		}
		parts = append(parts, fmt.Sprintf("%s:%d->%d", host, p.HostPort, p.ContainerPort))
	}
	return strings.Join(parts, ", ")
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
