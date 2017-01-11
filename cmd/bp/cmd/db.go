package cmd

import (
	"github.com/nstogner/blueprint"
	"github.com/spf13/cobra"
)

var db blueprint.DB

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Create a new database instance",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			er("missing required argument: name")
		}

		db.Kind = blueprint.KindDB
		db.Name = args[0]

		if err := runInDocker("new", db); err != nil {
			er(err)
		}
	},
}

func init() {
	newCmd.AddCommand(dbCmd)
	dbCmd.Flags().StringVar(&db.Variation, "database", blueprint.DBTypeMySQL, "Database type")
}
