package configscan

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func ScanFile(path, format string) (*ScanResult, error) {
	resolvedFormat, err := resolveFormat(path, format)
	if err != nil {
		return nil, err
	}

	switch resolvedFormat {
	case "yaml":
		return ScanYAML(path)
	case "props":
		return ScanProperties(path)
	default:
		return nil, fmt.Errorf("unsupported format %q", resolvedFormat)
	}
}

func resolveFormat(path, format string) (string, error) {
	if format != "" && format != "auto" {
		return format, nil
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yml", ".yaml":
		return "yaml", nil
	case ".properties":
		return "props", nil
	default:
		return "", fmt.Errorf("could not detect config format for %s; use --format=yaml or --format=props", path)
	}
}

func appendPlaceholders(result *ScanResult, placeholders []Placeholder) {
	for _, placeholder := range placeholders {
		if placeholder.Required {
			result.Required = append(result.Required, placeholder)
		} else {
			result.Optional = append(result.Optional, placeholder)
		}
	}
	sortPlaceholders(result.Required)
	sortPlaceholders(result.Optional)
}

func sortPlaceholders(placeholders []Placeholder) {
	sort.Slice(placeholders, func(i, j int) bool {
		if placeholders[i].Name != placeholders[j].Name {
			return placeholders[i].Name < placeholders[j].Name
		}
		if placeholders[i].SourceFile != placeholders[j].SourceFile {
			return placeholders[i].SourceFile < placeholders[j].SourceFile
		}
		return placeholders[i].SourcePath < placeholders[j].SourcePath
	})
}
