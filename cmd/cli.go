package cmd

import (
	"fmt"
	"os"

	"github.com/souhailBektachi/container_runtime_with_go/cmd/commands"
)

func RunCli() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: container <command> [args]")
		return
	}

	switch os.Args[1] {
	case "run":
		commands.RunCommand()
	default:
		fmt.Println("Unknown command:", os.Args[1])

	}

}
