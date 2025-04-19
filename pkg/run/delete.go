package run

import (
	"fmt"
	"os"
	"path/filepath"
)

func DeleteContainer(containerID string) error {
	containerBasePath := filepath.Join("_containers", containerID)

	if _, err := os.Stat(containerBasePath); os.IsNotExist(err) {
		return fmt.Errorf("container '%s' not found", containerID)
	} else if err != nil {
		return fmt.Errorf("failed to stat container directory '%s': %w", containerBasePath, err)
	}

	fmt.Printf("Removing container directory: %s\n", containerBasePath)
	if err := os.RemoveAll(containerBasePath); err != nil {
		return fmt.Errorf("failed to remove container directory '%s': %w", containerBasePath, err)
	}

	fmt.Printf("Successfully removed container '%s'\n", containerID)
	return nil
}
