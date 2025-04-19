package commands

import "github.com/spf13/cobra"

func RegisterCommands(root *cobra.Command) {
	root.AddCommand(runCmd)
	root.AddCommand(rmCmd)
	root.AddCommand(listCmd)
	root.AddCommand(pullCmd)
}
