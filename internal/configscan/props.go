package configscan

import (
	"bufio"
	"os"
	"strings"
)

func ScanProperties(path string) (*ScanResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := &ScanResult{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		appendPlaceholders(result, ExtractPlaceholders(strings.TrimSpace(value), path, strings.TrimSpace(key)))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
