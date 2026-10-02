package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
)

// unmanaged reports, and skips, hosts changes when config.yml turns them off.
// Bootstraps call `devarch hosts add`, so this is how they honor the setting.
func (a *app) unmanaged() bool {
	if a.eng.Settings.HostsManage {
		return false
	}
	fmt.Fprintln(a.err, "devarch: hosts.manage is false in config.yml; the hosts file was not changed")
	return true
}

func hostsCmd(a *app) *cobra.Command {
	c := &cobra.Command{
		Use:   "hosts",
		Short: "Manage the DevArch block in the hosts file",
		Long: `Manage .test hostnames. On WSL the Windows hosts file is updated, because that is
the file the browser uses; elsewhere /etc/hosts (override with HOSTS_FILE).
Set hosts.manage to false in config.yml to make sync, add and remove do nothing.
The address defaults to hosts.address (127.0.0.1).`,
	}
	var address string
	sync := &cobra.Command{
		Use:   "sync",
		Short: "Write every service container_name, x-devarch URL host and routable app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			if address == "" {
				address = a.eng.Settings.HostsAddress
			}
			if err := hosts.ValidateAddress(address); err != nil {
				return err
			}
			if a.unmanaged() {
				return nil
			}
			block, d, err := a.eng.HostsBlock(address)
			if err != nil {
				return err
			}
			if a.dryRun {
				fmt.Fprint(a.out, block)
				fmt.Fprintf(a.err, "devarch: discovered %d service and %d application domains (%d total)\n", len(d.Services), len(d.Apps), d.Count())
				return nil
			}
			res, err := a.eng.Hosts.Sync(contextOrBackground(cmd), block)
			if err != nil {
				return err
			}
			if res.Changed {
				fmt.Fprintf(a.out, "synchronized %d domains in %s\n", d.Count(), a.eng.Hosts.Path())
			} else {
				fmt.Fprintf(a.out, "managed block already current in %s (%d domains)\n", a.eng.Hosts.Path(), d.Count())
			}
			return nil
		},
	}
	sync.Flags().StringVar(&address, "address", "", "IPv4 address to map (default hosts.address)")

	var addAddress string
	add := &cobra.Command{
		Use:   "add <hostname>",
		Short: "Map one hostname (idempotent)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hosts.ValidateHostname(args[0]); err != nil {
				return err
			}
			if err := a.load(); err != nil {
				return err
			}
			if addAddress == "" {
				addAddress = a.eng.Settings.HostsAddress
			}
			if err := hosts.ValidateAddress(addAddress); err != nil {
				return err
			}
			if a.unmanaged() {
				return nil
			}
			if a.dryRun {
				fmt.Fprintf(a.out, "would register %s %s in %s\n", addAddress, args[0], a.eng.Hosts.Path())
				return nil
			}
			res, err := a.eng.Hosts.Add(contextOrBackground(cmd), args[0], addAddress)
			if err != nil {
				return err
			}
			report(a, res.Changed, fmt.Sprintf("registered %s %s", addAddress, args[0]), fmt.Sprintf("already registered %s %s", addAddress, args[0]), res.Path)
			return nil
		},
	}
	add.Flags().StringVar(&addAddress, "address", "", "IPv4 address to map (default hosts.address)")

	remove := &cobra.Command{
		Use:   "remove <hostname>",
		Short: "Unmap one hostname (Unix hosts files only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hosts.ValidateHostname(args[0]); err != nil {
				return err
			}
			if err := a.load(); err != nil {
				return err
			}
			if a.unmanaged() {
				return nil
			}
			if a.dryRun {
				fmt.Fprintf(a.out, "would remove %s from %s\n", args[0], a.eng.Hosts.Path())
				return nil
			}
			res, err := a.eng.Hosts.Remove(contextOrBackground(cmd), args[0])
			if err != nil {
				return err
			}
			report(a, res.Changed, "removed "+args[0], args[0]+" was not mapped", res.Path)
			return nil
		},
	}

	list := &cobra.Command{
		Use:   "list",
		Short: "Show the managed block and whether it is current",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			content, err := a.eng.Hosts.Read()
			if err != nil {
				return err
			}
			current := hosts.CurrentBlock(content)
			if current == "" {
				return errors.New("no DevArch block in " + a.eng.Hosts.Path() + "; run `devarch hosts sync`")
			}
			fmt.Fprint(a.out, current)
			want, _, err := a.eng.HostsBlock(a.eng.Settings.HostsAddress)
			switch {
			case err != nil:
				fmt.Fprintf(a.err, "devarch: cannot compare with the catalog: %v\n", err)
			case strings.TrimSpace(want) != strings.TrimSpace(current):
				fmt.Fprintln(a.err, "devarch: the block is out of date; run `devarch hosts sync`")
			}
			return nil
		},
	}
	c.AddCommand(sync, add, remove, list)
	return c
}

func report(a *app, changed bool, did, already, path string) {
	if changed {
		fmt.Fprintf(a.out, "%s in %s\n", did, path)
	} else {
		fmt.Fprintf(a.out, "%s in %s\n", already, path)
	}
}
