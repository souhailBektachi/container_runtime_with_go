package commands

import "os"

func ListContainers(containerdir string) []string {

	containers, err := os.ReadDir(containerdir)
	if err != nil {
		return nil
	}
	var containerNames []string

	for _, containre := range containers {
		if containre.IsDir() {

			containerNames = append(containerNames, containre.Name())
		}

	}

	return containerNames
}

func ListImages(imagedir string) []string {
	return nil
}
