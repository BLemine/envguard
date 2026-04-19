package parser

import (
	"os"
	"strings"
)

// ParseProps parses a .properties file into a flat map.
// Format is key=value one per line; lines starting with # and empty lines are skipped.
func ParseProps(path string) (*EnvFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := ""
		if len(parts) == 2 {
			value = strings.TrimSpace(parts[1])
		}
		keys[key] = value
	}

	return &EnvFile{Path: path, Keys: keys}, nil
}
