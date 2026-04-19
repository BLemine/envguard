package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Format string

const (
	FormatEnv   Format = "env"
	FormatYAML  Format = "yaml"
	FormatProps Format = "props"
)

func ParseFormat(raw string) (Format, error) {
	switch Format(strings.ToLower(strings.TrimSpace(raw))) {
	case "", FormatEnv:
		return FormatEnv, nil
	case FormatYAML:
		return FormatYAML, nil
	case FormatProps:
		return FormatProps, nil
	default:
		return "", fmt.Errorf("unsupported format %q (expected env, yaml, or props)", raw)
	}
}

func DetectFormat(path string) Format {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yml", ".yaml":
		return FormatYAML
	case ".properties":
		return FormatProps
	default:
		return FormatEnv
	}
}

func ParseWithFormat(path string, format Format) (*EnvFile, error) {
	switch format {
	case FormatEnv:
		return parseEnv(path)
	case FormatYAML:
		return parseYAML(path)
	case FormatProps:
		return parseProps(path)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func SerializeWithFormat(format Format, keys map[string]string) ([]byte, error) {
	switch format {
	case FormatEnv:
		return serializeFlat(keys), nil
	case FormatYAML:
		return serializeYAML(keys)
	case FormatProps:
		return serializeFlat(keys), nil
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}
