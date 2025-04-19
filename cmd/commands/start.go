package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/souhailBektachi/container_runtime_with_go/pkg/run"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [containerID]",
	Short: "Start (re-run) an existing container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerID := args[0]
		containerBasePath := filepath.Join("_containers", containerID)
		configFilePath := filepath.Join(containerBasePath, "config.json")

		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			return fmt.Errorf("container '%s' not found or missing config.json", containerID)
		} else if err != nil {
			return fmt.Errorf("failed to stat container config '%s': %w", configFilePath, err)
		}

		configBytes, err := os.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read container config '%s': %w", configFilePath, err)
		}
		var runConfig run.ImageConfig
		if err := json.Unmarshal(configBytes, &runConfig); err != nil {
			return fmt.Errorf("failed to parse container config '%s': %w", configFilePath, err)
		}

		fmt.Printf("Starting container %s...\n", containerID)

		childArgs := []string{"child-init", containerID}
		childCmd := exec.Command("/proc/self/exe", childArgs...)

		childCmd.Stdin = os.Stdin
		childCmd.Stdout = os.Stdout
		childCmd.Stderr = os.Stderr

		hostUID := os.Getuid()
		hostGID := os.Getgid()
		run.ApplyNamespaces(childCmd, hostUID, hostGID)

		fmt.Printf("Starting container process (ID: %s)...\n", containerID)
		if err := childCmd.Start(); err != nil {
			return fmt.Errorf("failed to start container process: %w", err)
		}

		err = childCmd.Wait()

		if err != nil {
			fmt.Printf("Container process exited with error: %v\n", err)
		} else {
			fmt.Printf("Container %s finished successfully.\n", containerID)
		}

		return err
	},
}
