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
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

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
