package main

import (
	"fmt"
	"os"

	"github.com/souhailBektachi/container_runtime_with_go/cmd"
	"github.com/souhailBektachi/container_runtime_with_go/cmd/commands"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "child-init" {
		if err := commands.HandleChildInit(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Child init error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Child process failed to exec.\n")
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
