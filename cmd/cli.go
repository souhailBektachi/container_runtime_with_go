package cmd

import (
	"fmt"

	"github.com/souhailBektachi/container_runtime_with_go/cmd/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "container",
	Short: "A simple container runtime",
	Long:  "A simple container runtime implementation in Go",
}

func init() {
	commands.RegisterCommands(rootCmd)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
