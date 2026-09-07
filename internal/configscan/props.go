package configscan

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ScanProps parses filePath as a Java .properties file and extracts env variable placeholders.
func ScanProps(filePath string) (*ScanResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}
	defer f.Close()

	result := &ScanResult{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		idx := -1
		escaped := false
		for i := 0; i < len(line); i++ {
			if escaped {
				escaped = false
				continue
			}
			if line[i] == '\\' {
				escaped = true
				continue
			}
			if strings.ContainsRune("=: \t\f", rune(line[i])) {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimLeft(line[idx:], " \t\f")
		if len(value) > 0 && (value[0] == '=' || value[0] == ':') {
			value = value[1:]
		}
		value = strings.TrimSpace(value)

		for _, p := range ExtractPlaceholders(value, filePath, key) {
			if p.Required {
				result.Required = append(result.Required, p)
			} else {
				result.Optional = append(result.Optional, p)
			}
		}
	}

	return result, scanner.Err()
}
