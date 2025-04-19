package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/souhailBektachi/container_runtime_with_go/pkg/oci"
	"github.com/souhailBektachi/container_runtime_with_go/pkg/run"
	"github.com/souhailBektachi/container_runtime_with_go/pkg/utiles"
)

var runCmd = &cobra.Command{
	Use:   "run [image] [command...]",
	Short: "Run a command in a new container",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageName := args[0]
		containerCmd := args[1:]

		imgBase, imgTag := utiles.ParseImageName(imageName)
		normalizedImageName := fmt.Sprintf("%s_%s", imgBase, imgTag)
		imageStorePath := filepath.Join("_images", normalizedImageName)

		if _, err := os.Stat(imageStorePath); os.IsNotExist(err) {
			fmt.Printf("Image '%s' not found locally, pulling...\n", imageName)
			_, err := oci.PullImage(imageStorePath, imageName)
			if err != nil {
				return fmt.Errorf("failed to pull image '%s': %w", imageName, err)
			}
			fmt.Printf("Image '%s' pulled successfully.\n", imageName)
		} else {
			fmt.Printf("Using local image '%s' from %s\n", imageName, imageStorePath)
		}

		containerID := uuid.New().String()[:8]
		containerBasePath := filepath.Join("_containers", containerID)
		rootfsPath := filepath.Join(containerBasePath, "rootfs")
		configFilePath := filepath.Join(containerBasePath, "config.json")

		fmt.Printf("Setting up container %s...\n", containerID)
		if err := os.MkdirAll(containerBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create container directory '%s': %w", containerBasePath, err)
		}

		manifestDigest, err := oci.GetImageManifestDigest(imageStorePath)
		if err != nil {
			return fmt.Errorf("failed to get manifest digest for image '%s': %w", imageName, err)
		}
		manifest, err := oci.ReadManifest(imageStorePath, manifestDigest)
		if err != nil {
			return fmt.Errorf("failed to read manifest for image '%s': %w", imageName, err)
		}
		ociConfig, err := oci.ReadConfig(imageStorePath, manifest.Config.Digest.String())
		if err != nil {
			return fmt.Errorf("failed to read config for image '%s': %w", imageName, err)
		}

		layerPaths := make([]string, len(manifest.Layers))
		for i, layer := range manifest.Layers {
			layerPaths[i] = filepath.Join(imageStorePath, "blobs", "sha256", oci.DigestToFilename(layer.Digest.String()))
		}

		if err := oci.UnpackImageLayers(layerPaths, rootfsPath); err != nil {
			os.RemoveAll(containerBasePath)
			return fmt.Errorf("failed to unpack layers for container '%s': %w", containerID, err)
		}

		runConfig, err := oci.MapOciConfigToRunConfig(ociConfig, rootfsPath)
		if err != nil {
			os.RemoveAll(containerBasePath)
			return fmt.Errorf("failed to map OCI config for container '%s': %w", containerID, err)
		}

		if len(containerCmd) > 0 {
			runConfig.ProcessConfig.Args = containerCmd
		}

		configBytes, err := json.MarshalIndent(runConfig, "", "  ")
		if err != nil {
			os.RemoveAll(containerBasePath)
			return fmt.Errorf("failed to marshal runtime config: %w", err)
		}
		if err := os.WriteFile(configFilePath, configBytes, 0644); err != nil {
			os.RemoveAll(containerBasePath)
			return fmt.Errorf("failed to save runtime config to '%s': %w", configFilePath, err)
		}

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
			os.RemoveAll(containerBasePath)
			return fmt.Errorf("failed to start container process: %w", err)
		}

		err = childCmd.Wait()

		if err != nil {
			fmt.Printf("Container process exited with error: %v\n", err)
		} else {
			fmt.Printf("Container %s finished successfully.\n", containerID)
		}

		return nil
	},
}

func HandleChildInit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("child-init requires container ID argument")
	}
	containerID := args[0]

	fmt.Printf("[Child %d] Initializing container %s...\n", os.Getpid(), containerID)

	containerBasePath := filepath.Join("_containers", containerID)
	configFilePath := filepath.Join(containerBasePath, "config.json")
	configBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("[Child] failed to read config file '%s': %w", configFilePath, err)
	}

	var runConfig run.ImageConfig
	if err := json.Unmarshal(configBytes, &runConfig); err != nil {
		return fmt.Errorf("[Child] failed to unmarshal config file '%s': %w", configFilePath, err)
	}

	containerCmd := runConfig.ProcessConfig.Args

	hostname := runConfig.Hostname
	if hostname == "" {
		hostname = containerID
	}
	if err := run.SetHostname(hostname); err != nil {
		return fmt.Errorf("[Child] failed to set hostname: %w", err)
	}

	if err := run.ApplyChroot(runConfig); err != nil {
		return fmt.Errorf("[Child] failed to apply chroot/mounts: %w", err)
	}

	fmt.Printf("[Child %d] Executing command: %v in %s\n", os.Getpid(), containerCmd, runConfig.ProcessConfig.Cwd)

	executable := containerCmd[0]
	if !filepath.IsAbs(executable) {
		foundPath, err := exec.LookPath(executable)
		if err != nil {
			absPath := filepath.Join("/", executable)
			if _, statErr := os.Stat(absPath); statErr == nil {
				executable = absPath
			} else {
				return fmt.Errorf("[Child] command '%s' not found in PATH: %w", containerCmd[0], err)
			}
		} else {
			executable = foundPath
		}
	}

	finalEnv := append(os.Environ(), runConfig.ProcessConfig.Env...)

	if err := syscall.Exec(executable, containerCmd, finalEnv); err != nil {
		return fmt.Errorf("[Child] failed to exec command '%s': %w", executable, err)
	}

	return fmt.Errorf("[Child] syscall.Exec returned unexpectedly")
}
