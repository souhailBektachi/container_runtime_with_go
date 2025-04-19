# Simple Go Container Runtime

This project is a basic container runtime implemented in Go, inspired by tools like Docker but significantly simplified. It demonstrates core containerization concepts like namespaces, chroot, and image layer management using OCI standards where possible.

## Features

*   Pull container images (currently relies on local Docker daemon via `docker save`).
*   Run commands inside isolated container environments.
*   List pulled images and created containers.
*   Remove containers.
*   Start existing containers.

## Project Structure

*   `_images/`: Stores pulled and extracted container images.
    *   Images are stored in directories named `<image_name>_<tag>`.
    *   Supports both OCI layout (index.json, blobs/, oci-layout) and Docker `save` format (manifest.json, layer tarballs).
*   `_containers/`: Stores container instances.
    *   Each container has a directory named by its ID.
    *   Contains the container's root filesystem (`rootfs/`) and configuration (`config.json`).
*   `cmd/`: Contains the command-line interface logic using Cobra.
*   `pkg/`: Contains the core runtime logic.
    *   `oci/`: Handles image pulling, manifest parsing, and layer unpacking.
    *   `run/`: Manages container execution, namespaces, and filesystem setup.
    *   `utiles/`: Utility functions.

## Prerequisites

*   Go (version 1.23 or later recommended)
*   Linux environment (for namespace and chroot functionality)
*   `docker` CLI installed and running (for the `pull` command)
*   `tar` utility

## Building

```bash
go build -o container .
```

## Usage

The main executable is `container` (or run directly using `go run .`).

### Pulling Images

Pulls an image using `docker save` and extracts it to the `_images` directory.

```bash
go run . pull <image_name>:<tag>
# Example:
go run . pull alpine:latest
```

### Running Containers

Creates and runs a command in a new container. If the image isn't local, it attempts to pull it first.

```bash
go run . run <image_name>:<tag> [command...]
# Example (run default shell):
go run . run alpine:latest
# Example (run specific command):
go run . run alpine:latest echo "Hello from container!"
```

### Listing Images

Lists images available in the `_images` directory.

```bash
go run . list --images
```

### Listing Containers

Lists container instances created in the `_containers` directory.

```bash
go run . list
```

### Starting Containers

Starts (re-runs) an existing container using its saved configuration.

```bash
go run . start <container_id>
# Example:
go run . start fb25dc9f
```

### Removing Containers

Removes one or more container directories from `_containers`.

```bash
go run . rm <container_id...>
# Example:
go run . rm fb25dc9f abc123ef
```

## Author

*   Souhail Bektachi

## Disclaimer

This is a learning project and is **not** suitable for production use. It's primarily for educational purposes to understand containerization concepts. It lacks many security features, robustness checks, and advanced functionalities found in mature container runtimes like Docker or containerd.
