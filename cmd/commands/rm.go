package commands

import (
	"fmt"

	"github.com/souhailBektachi/container_runtime_with_go/pkg/run"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [containerID...]",
	Short: "Remove one or more containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var finalErr error
		for _, containerID := range args {
			fmt.Printf("Attempting to remove container %s...\n", containerID)
			if err := run.DeleteContainer(containerID); err != nil {
				fmt.Printf("Error removing container %s: %v\n", containerID, err)
				if finalErr == nil {
					finalErr = fmt.Errorf("failed to remove container(s)")
				}
			}
		}
		return finalErr
	},
}
