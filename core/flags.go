package core

import (
	"fmt"
	"strings"
)

func ParseFlags(args []string, allowed map[string]bool) ([]string, map[string]bool, error) {
	positional := []string{}
	present := map[string]bool{}
	for _, arg := range args {
		isFlag := strings.HasPrefix(arg, "-")
		if !isFlag {
			positional = append(positional, arg)
			continue
		}
		ok := allowed[arg]
		if !ok {
			return nil, nil, fmt.Errorf("unknown flag: %s", arg)
		}
		present[arg] = true
	}

	return positional, present, nil
}
