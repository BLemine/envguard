package parser

import (
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

const (
	FormatEnv   = "env"
	FormatYAML  = "yaml"
	FormatProps = "props"
)

type EnvFile struct {
	Path string
	Keys map[string]string
}

// Parse parses an .env file using godotenv.
func Parse(path string) (*EnvFile, error) {
	keys, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}
	return &EnvFile{Path: path, Keys: keys}, nil
}

// DetectFormat infers the file format from the file extension.
func DetectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yml", ".yaml":
		return FormatYAML
	case ".properties":
		return FormatProps
	default:
		return FormatEnv
	}
}

// ParseAs parses a file using the given format. If format is empty, format is
// auto-detected from the file extension.
func ParseAs(path, format string) (*EnvFile, error) {
	if format == "" {
		format = DetectFormat(path)
	}
	switch format {
	case FormatYAML:
		return ParseYAML(path)
	case FormatProps:
		return ParseProps(path)
	default:
		return Parse(path)
	}
}

func KeySet(env *EnvFile) map[string]struct{} {
	set := make(map[string]struct{}, len(env.Keys))
	for k := range env.Keys {
		set[k] = struct{}{}
	}
	return set
}
