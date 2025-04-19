package oci

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnpackImageLayers(layerTarPaths []string, rootfsPath string) error {

	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", rootfsPath, err)
	}

	for _, LayerPath := range layerTarPaths {

		if err := unpackLayer(LayerPath, rootfsPath); err != nil {
			return fmt.Errorf("failed to unpack layer %s: %w", LayerPath, err)
		}
	}

	return nil

}

func unpackLayer(tarGzPath, destPath string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to open layer file '%s': %w", tarGzPath, err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for '%s': %w", tarGzPath, err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Join(destPath, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar header name (path traversal): %s", header.Name)
		}

		baseName := filepath.Base(header.Name)
		if strings.HasPrefix(baseName, ".wh.") {
			originalName := strings.TrimPrefix(baseName, ".wh.")
			pathToDelete := filepath.Join(filepath.Dir(target), originalName)
			if err := os.RemoveAll(pathToDelete); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to remove whiteout path '%s': %v\n", pathToDelete, err)
			}
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", target, err)
			}
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to chmod directory '%s': %v\n", target, err)
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for file '%s': %w", target, err)
			}

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file '%s': %w", target, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file content for '%s': %w", target, err)
			}
			outFile.Close()

			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to chmod file '%s': %v\n", target, err)
			}

		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for symlink '%s': %w", target, err)
			}
			if _, err := os.Lstat(target); err == nil {
				os.Remove(target)
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink '%s' -> '%s': %w", target, header.Linkname, err)
			}

		case tar.TypeLink:
			linkTarget := filepath.Join(destPath, header.Linkname)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for hardlink '%s': %w", target, err)
			}
			if _, err := os.Lstat(target); err == nil {
				os.Remove(target)
			}
			if err := os.Link(linkTarget, target); err != nil {
				return fmt.Errorf("failed to create hard link '%s' -> '%s': %w", target, linkTarget, err)
			}

		default:
			fmt.Fprintf(os.Stderr, "warning: unsupported file type '%c' for '%s'\n", header.Typeflag, header.Name)
		}

		os.Chtimes(target, header.AccessTime, header.ModTime)
	}

	return nil
}
