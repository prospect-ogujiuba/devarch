package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/replace"
)

func dbCmd(a *app) *cobra.Command {
	c := &cobra.Command{
		Use:   "db",
		Short: "Create and drop app databases on the shared MariaDB or PostgreSQL",
		Long: `Create and drop a database plus a user that owns it on the shared mariadb or
postgres service. SQL is piped to one visible exec command; passwords never
appear in arguments. The service must be running (devarch up mariadb).`,
	}
	var spec engine.DBSpec
	var asEnv, asJSON bool
	var forApp string
	create := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a database and its user with a generated password",
		Example: `  devarch db create shop
  devarch db create api --engine postgres --env
  devarch db create wp_blog --existing replace     # saves the old database first`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			spec.Name = args[0]
			var guard *replace.Record
			if forApp != "" && !a.dryRun {
				var err error
				if guard, err = a.guardFor(forApp); err != nil {
					return err
				}
				if spec.BackupDir == "" {
					spec.BackupDir = guard.DumpDir()
				}
			}
			if spec.Existing == engine.ExistingReplace && spec.BackupDir == "" {
				spec.BackupDir = filepath.Join(a.eng.Settings.AppsDir, ".devarch-backups",
					"db-"+spec.Engine+"-"+spec.Name+"-"+time.Now().Format("20060102-150405"))
			}
			res, err := a.eng.CreateDB(contextOrBackground(cmd), spec)
			if err != nil {
				return err
			}
			if guard != nil {
				if err := guard.AddDatabase(res); err != nil {
					return fmt.Errorf("created %s but could not record it for recovery: %w", res.Name, err)
				}
			}
			switch {
			case asJSON:
				return writeJSON(a.out, res)
			case asEnv:
				fmt.Fprintln(a.out, strings.Join(res.Env(), "\n"))
			default:
				fmt.Fprintf(a.out, "database  %s\nuser      %s\npassword  %s\nhost      %s:%d (from containers on the shared network)\n",
					res.Name, res.User, res.Password, res.Host, res.Port)
				if res.Dump != "" {
					fmt.Fprintf(a.out, "previous  %s\n", res.Dump)
				}
			}
			return nil
		},
	}
	create.Flags().StringVar(&spec.Engine, "engine", "mariadb", "mariadb or postgres")
	create.Flags().StringVar(&spec.User, "user", "", "user name (default: the database name)")
	create.Flags().StringVar(&spec.Existing, "existing", engine.ExistingFail, "when the database or user exists: fail, reuse (reset the password) or replace (dump, drop, recreate)")
	create.Flags().StringVar(&spec.BackupDir, "backup-dir", "", "where --existing replace saves the old database (default under <apps_dir>/.devarch-backups)")
	create.Flags().StringVar(&forApp, "app", "", "record the database in this app's guard so `devarch app recover` can undo it")
	create.Flags().BoolVar(&asEnv, "env", false, "print DB_* NAME=value lines for scripts")
	create.Flags().BoolVar(&asJSON, "json", false, "print JSON")

	var dropEngine, dropUser string
	var yes bool
	drop := &cobra.Command{
		Use:   "drop <name>",
		Short: "Drop a database (and its user with --user); asks first",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.load(); err != nil {
				return err
			}
			if !yes && !a.dryRun {
				what := "database " + args[0]
				if dropUser != "" {
					what += " and user " + dropUser
				}
				ok, err := a.confirm(fmt.Sprintf("Permanently drop %s %s?", dropEngine, what))
				if err != nil {
					return err
				}
				if !ok {
					return errors.New("aborted; nothing was dropped")
				}
			}
			return a.eng.DropDB(contextOrBackground(cmd), dropEngine, args[0], dropUser)
		},
	}
	drop.Flags().StringVar(&dropEngine, "engine", "mariadb", "mariadb or postgres")
	drop.Flags().StringVar(&dropUser, "user", "", "also drop this user")
	drop.Flags().BoolVarP(&yes, "yes", "y", false, "do not ask first")
	c.AddCommand(create, drop)
	return c
}
