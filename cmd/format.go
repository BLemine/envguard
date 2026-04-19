package cmd

import (
	"fmt"

	"github.com/BLemine/envguard/internal/parser"
)

func resolveFormat(raw string, paths ...string) (parser.Format, error) {
	if raw != "" {
		return parser.ParseFormat(raw)
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		format := parser.DetectFormat(path)
		if format != parser.FormatEnv {
			return format, nil
		}
	}
	return parser.FormatEnv, nil
}

func formatErrorContext(action string, format parser.Format, err error) error {
	return fmt.Errorf("%s (%s): %w", action, format, err)
}
