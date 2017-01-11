package cmd

import "github.com/spf13/cobra"

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new component",
	Long: `Scaffold a new component. Example:

bp new api`,
	Run: func(cmd *cobra.Command, args []string) {
		er("must specify component type")
	},
}

func init() {
	RootCmd.AddCommand(newCmd)
}
