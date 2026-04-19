package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func parseProps(path string) (*EnvFile, error) {
	data, err := osReadFile(path)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("invalid properties line %q", line)
		}
		keys[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &EnvFile{Path: path, Keys: keys}, nil
}
