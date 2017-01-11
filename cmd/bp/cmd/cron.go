package cmd

import (
	"github.com/nstogner/blueprint"
	"github.com/spf13/cobra"
)

var cron blueprint.Cron

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Create a new cron job",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			er("missing required argument: name")
		}

		cron.Kind = blueprint.KindCron
		cron.Name = args[0]

		if err := runInDocker("new", cron); err != nil {
			er(err)
		}
	},
}

func init() {
	newCmd.AddCommand(cronCmd)
	cronCmd.Flags().StringVar(&cron.Variation, "lang", blueprint.LangGo, "Programming language")
	cronCmd.Flags().StringVar(&cron.Schedule, "schedule", "* * * * *", "Cron schedule")
}
