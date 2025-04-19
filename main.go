package main

import (
	"fmt"
	"os"

	"github.com/souhailBektachi/container_runtime_with_go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {

		fmt.Errorf("error: %v", err)
		os.Exit(1)
	}

}
