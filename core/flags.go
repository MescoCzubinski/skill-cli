package core

import (
	"fmt"
	"strings"
)

// ParseFlags splits args into positional arguments and the set of flags that
// are present. Every token starting with "-" must appear in allowed, otherwise
// it is reported as an unknown flag. The returned map holds only the flags that
// were passed (all drawn from allowed); an absent flag reads back as false via
// the map's zero value.
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
