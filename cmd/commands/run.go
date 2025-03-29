package commands

import (
	"fmt"
	"os"

	"github.com/souhailBektachi/container_runtime_with_go/pkg/run"
)

func RunCommand() error {
	if len(os.Args) < 3 {
		fmt.Println("Usage: mini-container run <command>")
		os.Exit(1)
	}

	command := os.Args[2]
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	if err := run.RunContainer(command, args); err != nil {
		return fmt.Errorf("error running container: %w", err)
	}

	return nil
}
