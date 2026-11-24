package boi

import (
	"fmt"
	"strings"
)

func FormatURI(format string, args ...string) (string, error) {
	if strings.TrimSpace(format) == "" {
		return "", fmt.Errorf("uri format is empty")
	}
	out := format
	used := false
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		if strings.Contains(out, placeholder) {
			out = strings.ReplaceAll(out, placeholder, arg)
			used = true
		}
	}
	if !used && len(args) > 0 && strings.Contains(out, "%s") {
		for _, arg := range args {
			if !strings.Contains(out, "%s") {
				break
			}
			out = strings.Replace(out, "%s", arg, 1)
			used = true
		}
	}
	if len(args) > 0 && !used {
		return "", fmt.Errorf("uri format %q has no supported placeholders", format)
	}
	return out, nil
}
