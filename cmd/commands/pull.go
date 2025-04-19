package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/souhailBektachi/container_runtime_with_go/pkg/oci"
	"github.com/souhailBektachi/container_runtime_with_go/pkg/utiles"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [image]",
	Short: "Pull an image from a registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageName := args[0]

		imgBase, imgTag := utiles.ParseImageName(imageName)
		normalizedImageName := fmt.Sprintf("%s_%s", imgBase, imgTag)
		imageStorePath := filepath.Join("_images", normalizedImageName)

		if _, err := os.Stat(imageStorePath); err == nil {
			fmt.Printf("Image '%s' already exists locally at %s. Skipping pull.\n", imageName, imageStorePath)
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check image directory '%s': %w", imageStorePath, err)
		}

		fmt.Printf("Pulling image '%s'...\n", imageName)
		_, err := oci.PullImage(imageStorePath, imageName)
		if err != nil {
			os.RemoveAll(imageStorePath)
			return fmt.Errorf("failed to pull image '%s': %w", imageName, err)
		}

		fmt.Printf("Successfully pulled image '%s' to %s\n", imageName, imageStorePath)
		return nil
	},
}
