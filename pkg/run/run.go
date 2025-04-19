package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func applyNamespaces(cmd *exec.Cmd) {

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
	}

}

func pivotRoot(newroot string) error {
	newrootAbs, err := filepath.Abs(newroot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for newroot '%s': %w", newroot, err)
	}

	if err := syscall.Mount(newrootAbs, newrootAbs, "", syscall.MS_BIND|syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to bind mount newroot '%s' onto itself: %w", newrootAbs, err)
	}

	putold := filepath.Join(newrootAbs, ".pivot_root")
	if err := os.MkdirAll(putold, 0700); err != nil {
		syscall.Unmount(newrootAbs, syscall.MNT_DETACH)
		return fmt.Errorf("failed to create putold directory '%s': %w", putold, err)
	}

	if err := syscall.PivotRoot(newrootAbs, putold); err != nil {
		os.RemoveAll(putold)
		syscall.Unmount(newrootAbs, syscall.MNT_DETACH)
		return fmt.Errorf("pivot_root('%s', '%s') failed: %w", newrootAbs, putold, err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir to new root '/' failed after pivot_root: %w", err)
	}

	putoldRelative := "/.pivot_root"
	if err := syscall.Unmount(putoldRelative, syscall.MNT_DETACH); err != nil {
		fmt.Fprintf(os.Stderr, "warning: unmount old root '%s' failed: %v\n", putoldRelative, err)
	}

	if err := os.RemoveAll(putoldRelative); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove '%s': %v\n", putoldRelative, err)
	}

	return nil
}
func applyChroot(imageconf ImageConfig) error {
	if err := pivotRoot(imageconf.Rootfs); err != nil {
		return fmt.Errorf("Error applying pivot_root: %v", err)
	}

	if err := syscall.Chdir(imageconf.ProcessConfig.Cwd); err != nil {
		return fmt.Errorf("Error changing directory: %v", err)
	}
	return nil
}
