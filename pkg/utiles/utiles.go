package utiles

import "strings"

func ParseImageName(name string) (string, string) {
	split := strings.Split(name, ":")
	if len(split) == 1 {
		return split[0], "latest"
	}
	return split[0], split[1]

}
