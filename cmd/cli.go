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
	rootCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run a command in a container",
		Long:  "Run a command in a container",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("Usage: container run <command> [args]")
				return
			}

			if err := commands.RunCommand(); err != nil {
				fmt.Printf("Error running command: %v\n", err)
			}
		},
	})

}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("error executing command: %w", err)

	}

	return nil
}
