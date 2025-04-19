package oci

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/souhailBektachi/container_runtime_with_go/pkg/run"
)

type OciIndex struct {
	SchemaVersion int                `json:"schemaVersion"`
	MediaType     string             `json:"mediaType,omitempty"`
	ArtifactType  string             `json:"artifactType,omitempty"`
	Manifests     []specs.Descriptor `json:"manifests"`
	Subject       *specs.Descriptor  `json:"subject,omitempty"`
	Annotations   map[string]string  `json:"annotations,omitempty"`
}

type OciManifest struct {
	SchemaVersion int                `json:"schemaVersion"`
	MediaType     string             `json:"mediaType,omitempty"`
	ArtifactType  string             `json:"artifactType,omitempty"`
	Config        specs.Descriptor   `json:"config"`
	Layers        []specs.Descriptor `json:"layers"`
	Subject       *specs.Descriptor  `json:"subject,omitempty"`
	Annotations   map[string]string  `json:"annotations,omitempty"`
}

type OciConfig struct {
	Created      *string  `json:"created,omitempty"`
	Author       string   `json:"author,omitempty"`
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	Variant      string   `json:"variant,omitempty"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Config       struct {
		Hostname     string            `json:"Hostname,omitempty"`
		Domainname   string            `json:"Domainname,omitempty"`
		User         string            `json:"User,omitempty"`
		AttachStdin  bool              `json:"AttachStdin,omitempty"`
		AttachStdout bool              `json:"AttachStdout,omitempty"`
		AttachStderr bool              `json:"AttachStderr,omitempty"`
		Tty          bool              `json:"Tty,omitempty"`
		OpenStdin    bool              `json:"OpenStdin,omitempty"`
		StdinOnce    bool              `json:"StdinOnce,omitempty"`
		Env          []string          `json:"Env,omitempty"`
		Cmd          []string          `json:"Cmd,omitempty"`
		Entrypoint   []string          `json:"Entrypoint,omitempty"`
		Labels       map[string]string `json:"Labels,omitempty"`
		WorkingDir   string            `json:"WorkingDir,omitempty"`
		StopSignal   string            `json:"StopSignal,omitempty"`
	} `json:"config,omitempty"`
	RootFS struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
	History []struct {
		Created    *string `json:"created,omitempty"`
		CreatedBy  string  `json:"created_by,omitempty"`
		Author     string  `json:"author,omitempty"`
		Comment    string  `json:"comment,omitempty"`
		EmptyLayer bool    `json:"empty_layer,omitempty"`
	} `json:"history,omitempty"`
}

type DockerManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

func GetImageManifestDigest(imagePath string) (string, error) {
	manifestPath := filepath.Join(imagePath, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		manifestBytes, err := os.ReadFile(manifestPath)
		if err != nil {
			return "", fmt.Errorf("failed to read Docker manifest at '%s': %w", manifestPath, err)
		}

		var dockerManifests []DockerManifest
		if err := json.Unmarshal(manifestBytes, &dockerManifests); err != nil {
			return "", fmt.Errorf("failed to unmarshal Docker manifest at '%s': %w", manifestPath, err)
		}

		if len(dockerManifests) == 0 {
			return "", fmt.Errorf("no image manifests found in Docker manifest at '%s'", manifestPath)
		}

		return dockerManifests[0].Config, nil
	}

	indexPath := filepath.Join(imagePath, "index.json")
	if _, err := os.Stat(indexPath); err == nil {
		indexBytes, readErr := os.ReadFile(indexPath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				return "", fmt.Errorf("could not find manifest.json or index.json in '%s'", imagePath)
			}
			return "", fmt.Errorf("failed to read index file '%s': %w", indexPath, readErr)
		}

		var index OciIndex
		if err := json.Unmarshal(indexBytes, &index); err != nil {
			return "", fmt.Errorf("failed to unmarshal index file '%s': %w", indexPath, err)
		}

		if len(index.Manifests) == 0 {
			return "", fmt.Errorf("no manifests found in index file '%s'", indexPath)
		}
		return index.Manifests[0].Digest.String(), nil
	}

	entries, err := os.ReadDir(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image directory '%s': %w", imagePath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "manifest.json" {
			return entry.Name(), nil
		}
	}

	layerDirs := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			layerTarPath := filepath.Join(imagePath, entry.Name(), "layer.tar")
			if _, err := os.Stat(layerTarPath); err == nil {
				layerDirs = append(layerDirs, entry.Name())
			}
		}
	}

	if len(layerDirs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: No manifest.json or index.json found, but found %d potential layer directories\n", len(layerDirs))
		return "__DOCKER_LAYERS_ONLY__", nil
	}

	return "", fmt.Errorf("could not find manifest.json, index.json, or layer directories in '%s'", imagePath)
}

func ReadManifest(imagePath, manifestDigestOrFilename string) (*OciManifest, error) {
	if manifestDigestOrFilename == "__DOCKER_LAYERS_ONLY__" {
		fmt.Println("DEBUG: Creating synthetic manifest for layer-only Docker image")

		entries, err := os.ReadDir(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read image directory for layer scanning: %w", err)
		}

		var layerPaths []string
		for _, entry := range entries {
			if entry.IsDir() {
				layerTarPath := filepath.Join(imagePath, entry.Name(), "layer.tar")
				if _, err := os.Stat(layerTarPath); err == nil {
					layerPaths = append(layerPaths, filepath.Join(entry.Name(), "layer.tar"))
				}
			}
		}

		if len(layerPaths) == 0 {
			return nil, fmt.Errorf("no layer.tar files found in Docker image directories")
		}

		syntheticManifest := &OciManifest{
			SchemaVersion: 2,
			Layers:        make([]specs.Descriptor, len(layerPaths)),
		}

		for i, layerPath := range layerPaths {
			layerDir := filepath.Dir(layerPath)
			syntheticManifest.Layers[i] = specs.Descriptor{
				MediaType: specs.MediaTypeImageLayerGzip,
				Digest:    digest.Digest("sha256:" + layerDir),
			}
		}

		return syntheticManifest, nil
	}

	manifestPath := filepath.Join(imagePath, "manifest.json")
	if _, err := os.Stat(manifestPath); err == nil {
		manifestBytes, readErr := os.ReadFile(manifestPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read Docker manifest '%s': %w", manifestPath, readErr)
		}

		var dockerManifests []DockerManifest
		if err := json.Unmarshal(manifestBytes, &dockerManifests); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Docker manifest '%s': %w", manifestPath, err)
		}

		if len(dockerManifests) == 0 {
			return nil, fmt.Errorf("no manifests found in Docker manifest file '%s'", manifestPath)
		}

		var dockerManifest DockerManifest
		found := false
		for _, m := range dockerManifests {
			if m.Config == manifestDigestOrFilename {
				dockerManifest = m
				found = true
				break
			}
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Warning: Could not find exact Docker manifest for config %s, using first entry.\n", manifestDigestOrFilename)
			dockerManifest = dockerManifests[0]
			manifestDigestOrFilename = dockerManifest.Config
		}

		ociManifest := &OciManifest{
			SchemaVersion: 2,
			MediaType:     specs.MediaTypeImageManifest,
			Config: specs.Descriptor{
				MediaType: specs.MediaTypeImageConfig,
				Digest:    digest.Digest("sha256:" + strings.TrimSuffix(manifestDigestOrFilename, ".json")),
			},
			Layers: make([]specs.Descriptor, len(dockerManifest.Layers)),
		}

		for i, layerPath := range dockerManifest.Layers {
			var layerDigestHex string
			base := filepath.Base(layerPath)
			if base == "layer.tar" {
				layerDir := filepath.Dir(layerPath)
				layerDigestHex = layerDir
			} else if strings.HasSuffix(base, ".tar") {
				layerDigestHex = strings.TrimSuffix(base, ".tar")
			} else {
				layerDigestHex = base
			}
			layerDigestHex = strings.Trim(layerDigestHex, "/")
			ociManifest.Layers[i] = specs.Descriptor{
				MediaType: specs.MediaTypeImageLayerGzip,
				Digest:    digest.Digest("sha256:" + layerDigestHex),
			}
		}
		return ociManifest, nil

	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat docker manifest '%s': %w", manifestPath, err)
	}

	ociManifestFilename := DigestToFilename(manifestDigestOrFilename)
	ociManifestPath := filepath.Join(imagePath, "blobs", "sha256", ociManifestFilename)
	ociManifestBytes, readErr := os.ReadFile(ociManifestPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, fmt.Errorf("could not find manifest.json or OCI manifest blob '%s'", ociManifestPath)
		}
		return nil, fmt.Errorf("failed to read OCI manifest file '%s': %w", ociManifestPath, readErr)
	}

	var manifest OciManifest
	if err := json.Unmarshal(ociManifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OCI manifest file '%s': %w", ociManifestPath, err)
	}
	return &manifest, nil
}

func ReadConfig(imagePath, configDigestOrFilename string) (*OciConfig, error) {
	if configDigestOrFilename == "__DOCKER_LAYERS_ONLY__" {
		fmt.Println("DEBUG: Creating synthetic config for layer-only Docker image")

		syntheticConfig := &OciConfig{
			Architecture: "amd64",
			OS:           "linux",
		}

		syntheticConfig.Config.Cmd = []string{"/bin/sh"}

		return syntheticConfig, nil
	}

	directConfigPath := filepath.Join(imagePath, configDigestOrFilename)
	if _, err := os.Stat(directConfigPath); err == nil {
		configBytes, readErr := os.ReadFile(directConfigPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read Docker config file '%s': %w", directConfigPath, readErr)
		}

		var config OciConfig
		if err := json.Unmarshal(configBytes, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal Docker config file '%s': %w", directConfigPath, err)
		}
		return &config, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat potential Docker config file '%s': %w", directConfigPath, err)
	}

	ociConfigFilename := DigestToFilename(configDigestOrFilename)
	ociConfigPath := filepath.Join(imagePath, "blobs", "sha256", ociConfigFilename)
	ociConfigBytes, readErr := os.ReadFile(ociConfigPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, fmt.Errorf("config file not found at direct path '%s' or OCI path '%s'", directConfigPath, ociConfigPath)
		}
		return nil, fmt.Errorf("failed to read OCI config file '%s': %w", ociConfigPath, readErr)
	}

	var config OciConfig
	if err := json.Unmarshal(ociConfigBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OCI config file '%s': %w", ociConfigPath, err)
	}
	return &config, nil
}

func MapOciConfigToRunConfig(ociCgg *OciConfig, rootfsPath string) (*run.ImageConfig, error) {
	var args []string
	if len(ociCgg.Config.Entrypoint) > 0 {
		args = append(args, ociCgg.Config.Entrypoint...)
		if len(ociCgg.Config.Cmd) > 0 {
			args = append(args, ociCgg.Config.Cmd...)
		}
	} else {
		args = ociCgg.Config.Cmd
	}

	uid, gid := 0, 0

	if ociCgg.Config.User != "" {
		parts := strings.SplitN(ociCgg.Config.User, ":", 2)
		if len(parts) > 0 {
			var err error
			uid, err = parseIntOrDefault(parts[0], 0)
			if err != nil {
				return nil, fmt.Errorf("invalid UID: %w", err)
			}
		}
		if len(parts) > 1 {
			var err error
			gid, err = parseIntOrDefault(parts[1], 0)
			if err != nil {
				return nil, fmt.Errorf("invalid GID: %w", err)
			}
		} else {
			gid = uid
		}
	}

	runCfg := &run.ImageConfig{
		Hostname: ociCgg.Config.Hostname,
		Root:     run.RootConfig{Path: rootfsPath},
		ProcessConfig: run.ProcessConfig{
			Env:  ociCgg.Config.Env,
			Args: args,
			Cwd:  ociCgg.Config.WorkingDir,
			User: map[string]int{
				"uid": uid,
				"gid": gid,
			},
		},
	}
	if runCfg.ProcessConfig.Cwd == "" {
		runCfg.ProcessConfig.Cwd = "/"
	}
	if len(runCfg.ProcessConfig.Args) == 0 {
		return nil, fmt.Errorf("Image does not have entrypoint or command")
	}

	return runCfg, nil
}

func parseIntOrDefault(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	val := 0
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		return defaultVal, err
	}
	return val, nil
}

func DigestToFilename(digest string) string {
	parts := strings.SplitN(digest, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}

	if !strings.Contains(digest, ":") && len(digest) == 64 {
		return digest
	}

	return digest
}
