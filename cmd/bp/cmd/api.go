package cmd

import (
	"fmt"

	"github.com/nstogner/blueprint"
	"github.com/spf13/cobra"
)

var api blueprint.API
var deps string

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Create a new api",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			er("missing required argument: name")
		}

		api.Kind = blueprint.KindAPI
		api.Name = args[0]

		if deps != "" {
			var err error
			api.DB, err = blueprint.ParseIdentifier(deps)
			if err != nil {
				er(fmt.Errorf("unable to parse dependencies: %s", err))
			}
		}

		if err := runInDocker("new", api); err != nil {
			er(err)
		}
	},
}

func init() {
	newCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringVar(&api.Variation, "lang", blueprint.LangGo, "Programming language")
	apiCmd.Flags().StringVar(&deps, "db", "", "Database dependency")
}
