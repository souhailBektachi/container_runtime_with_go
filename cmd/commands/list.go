package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var listImagesFlag bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers or images",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listImagesFlag {
			images := ListImages("_images")
			if images == nil {
				fmt.Println("No images found or error listing images.")
				return nil
			}
			fmt.Println("IMAGES:")
			for _, img := range images {
				fmt.Println(img)
			}
		} else {
			containers := ListContainers("_containers")
			if containers == nil {
				fmt.Println("No containers found or error listing containers.")
				return nil
			}
			fmt.Println("CONTAINER IDS:")
			for _, id := range containers {
				fmt.Println(id)
			}
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listImagesFlag, "images", false, "List images instead of containers")
}

func ListContainers(containerdir string) []string {
	entries, err := os.ReadDir(containerdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading container directory %s: %v\n", containerdir, err)
		return nil
	}
	var containerNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			containerNames = append(containerNames, entry.Name())
		}
	}
	return containerNames
}

func ListImages(imagedir string) []string {
	entries, err := os.ReadDir(imagedir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading image directory %s: %v\n", imagedir, err)
		return nil
	}
	var imageNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) == 2 {
				imageName := fmt.Sprintf("%s:%s", parts[0], parts[1])
				imageNames = append(imageNames, imageName)
			} else {
				imageNames = append(imageNames, entry.Name()+" (malformed?)")
			}
		}
	}
	return imageNames
}
