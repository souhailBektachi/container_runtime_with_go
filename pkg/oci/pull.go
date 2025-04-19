package oci

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func PullImage(imageDir, image string) ([]byte, error) {
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", imageDir, err)
	}

	_, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("Docker CLI is required but not found: %w", err)
	}

	fmt.Printf("Pulling image %s using Docker CLI...\n", image)

	pullCmd := exec.Command("docker", "pull", image)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to pull image using Docker CLI: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "container-with-go-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tarPath := filepath.Join(tempDir, "image.tar")
	saveCmd := exec.Command("docker", "save", "-o", tarPath, image)
	saveCmd.Stdout = os.Stdout
	saveCmd.Stderr = os.Stderr
	if err := saveCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to save image to tar: %w", err)
	}

	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	extractCmd := exec.Command("tar", "-xf", tarPath, "-C", imageDir)
	extractCmd.Stdout = os.Stdout
	extractCmd.Stderr = os.Stderr
	if err := extractCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract image tar: %w", err)
	}

	fmt.Printf("Image %s successfully pulled and extracted to %s\n", image, imageDir)

	return []byte{}, nil
}
