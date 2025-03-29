package run

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func RunContainer(command string, args []string) error {
	// Look up the absolute path of the command
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("command not found: %v", err)
	}

	cmd := exec.Command(cmdPath, args...)
	applyNamespaces(cmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container exited with error: %v", err)
	}

	return nil
}

func applyNamespaces(cmd *exec.Cmd) {

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUTS,
	}

}
