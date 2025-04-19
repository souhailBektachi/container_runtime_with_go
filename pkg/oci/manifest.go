package oci

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func GetImageManifestDigest(imagePath string) (string, error) {
	indexPath := filepath.Join(imagePath, "index.json")
	indexBytes, err := os.ReadFile(indexPath)
	defer os.Remove(indexPath)
	if err != nil {
		return "", fmt.Errorf("failed to read index file '%s': %w", indexPath, err)
	}

	var index OciIndex
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		return "", fmt.Errorf("failed to unmarshal index file '%s': %w", indexPath, err)
	}

	if len(index.Manifests) > 0 {
		return index.Manifests[0].Digest.String(), nil

	}

	return "", fmt.Errorf("no manifests found in index file '%s'", indexPath)

}

func ReadManifest(imagePath, manifestDigest string) (*OciManifest, error) {

	manifestFilename := DigestToFilename(manifestDigest)
	manifestPath := filepath.Join(imagePath, "blobs", "sha256", manifestFilename)
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file '%s': %w", manifestPath, err)
	}
	var manifest OciManifest

	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest file '%s': %w", manifestPath, err)
	}

	return &manifest, nil
}
func ReadConfig(imagePath, configDigest string) (*OciConfig, error) {
	configFilename := DigestToFilename(configDigest)
	configPath := filepath.Join(imagePath, "blobs", "sha256", configFilename)
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", configPath, err)
	}
	var config OciConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file '%s': %w", configPath, err)
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
