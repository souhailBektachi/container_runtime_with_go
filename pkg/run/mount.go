package run

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func setupMounts(rootfsPath string, mounts []MountsConfig) error {
	standardMounts := []MountsConfig{
		{Destination: "/proc", Type: "proc", Source: "proc"},
		{Destination: "/sys", Type: "sysfs", Source: "sysfs"},
		{Destination: "/dev", Type: "tmpfs", Source: "tmpfs", Options: []string{"nosuid", "strictatime", "mode=755", "size=65536k"}},
		{Destination: "/dev/pts", Type: "devpts", Source: "devpts", Options: []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620"}},
	}

	allMounts := append(standardMounts, mounts...)

	for _, m := range allMounts {
		dest := filepath.Join(rootfsPath, m.Destination)

		if err := os.MkdirAll(dest, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create mount destination '%s': %w", dest, err)
		}

		var flags uintptr
		var data []string

		flags = syscall.MS_NOSUID | syscall.MS_NODEV

		for _, opt := range m.Options {
			switch opt {
			case "ro":
				flags |= syscall.MS_RDONLY
			case "nosuid":
				flags |= syscall.MS_NOSUID
			case "suid":
				flags &^= syscall.MS_NOSUID
			case "nodev":
				flags |= syscall.MS_NODEV
			case "dev":
				flags &^= syscall.MS_NODEV
			case "noexec":
				flags |= syscall.MS_NOEXEC
			case "exec":
				flags &^= syscall.MS_NOEXEC
			case "sync":
				flags |= syscall.MS_SYNCHRONOUS
			case "async":
				flags &^= syscall.MS_SYNCHRONOUS
			case "dirsync":
				flags |= syscall.MS_DIRSYNC
			case "remount":
				flags |= syscall.MS_REMOUNT
			case "mand":
				flags |= syscall.MS_MANDLOCK
			case "nomand":
				flags &^= syscall.MS_MANDLOCK
			case "atime":
				flags &^= syscall.MS_NOATIME
			case "noatime":
				flags |= syscall.MS_NOATIME
			case "relatime":
				flags &^= syscall.MS_NOATIME
				flags &^= syscall.MS_STRICTATIME
			case "norelatime":
				flags &^= syscall.MS_NOATIME
			case "strictatime":
				flags |= syscall.MS_STRICTATIME
			case "nostrictatime":
				flags &^= syscall.MS_STRICTATIME
			case "bind":
				flags |= syscall.MS_BIND
			case "rbind":
				flags |= syscall.MS_BIND | syscall.MS_REC
			case "private":
				flags |= syscall.MS_PRIVATE
			case "rprivate":
				flags |= syscall.MS_PRIVATE | syscall.MS_REC
			case "slave":
				flags |= syscall.MS_SLAVE
			case "rslave":
				flags |= syscall.MS_SLAVE | syscall.MS_REC
			case "shared":
				flags |= syscall.MS_SHARED
			case "rshared":
				flags |= syscall.MS_SHARED | syscall.MS_REC
			case "unbindable":
				flags |= syscall.MS_UNBINDABLE
			case "runbindable":
				flags |= syscall.MS_UNBINDABLE | syscall.MS_REC
			default:
				data = append(data, opt)
			}
		}

		if flags&syscall.MS_BIND != 0 {
			if _, err := os.Stat(m.Source); os.IsNotExist(err) {
				return fmt.Errorf("bind mount source '%s' does not exist", m.Source)
			}
			srcInfo, err := os.Stat(m.Source)
			if err != nil {
				return fmt.Errorf("could not stat bind mount source '%s': %w", m.Source, err)
			}
			if srcInfo.IsDir() {
				if err := os.MkdirAll(dest, 0755); err != nil && !os.IsExist(err) {
					return fmt.Errorf("failed to create directory for bind mount destination '%s': %w", dest, err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil && !os.IsExist(err) {
					return fmt.Errorf("failed to create parent directory for bind mount file destination '%s': %w", dest, err)
				}
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					f, err := os.Create(dest)
					if err != nil {
						return fmt.Errorf("failed to create placeholder file for bind mount destination '%s': %w", dest, err)
					}
					f.Close()
				}
			}
		}

		dataStr := strings.Join(data, ",")
		fmt.Printf("Mounting '%s' to '%s' (type: %s, flags: 0x%x, data: %s)\n", m.Source, dest, m.Type, flags, dataStr)
		if err := syscall.Mount(m.Source, dest, m.Type, flags, dataStr); err != nil {
			return fmt.Errorf("failed to mount '%s' to '%s' (type: %s): %w", m.Source, dest, m.Type, err)
		}

		if (flags&syscall.MS_BIND != 0) && (flags&syscall.MS_RDONLY != 0) {
			remountFlags := syscall.MS_BIND | syscall.MS_REMOUNT | syscall.MS_RDONLY
			remountFlags |= (flags & (syscall.MS_NOEXEC | syscall.MS_NODEV | syscall.MS_NOSUID))

			fmt.Printf("Remounting '%s' as read-only (flags: 0x%x)\n", dest, remountFlags)
			if err := syscall.Mount("", dest, "", remountFlags, ""); err != nil {
				return fmt.Errorf("failed to remount '%s' as read-only: %w", dest, err)
			}
		}
	}

	fmt.Println("All mounts completed successfully.")
	return nil
}
