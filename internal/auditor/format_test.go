package auditor

import (
	"testing"

	"github.com/BLemine/envguard/internal/parser"
)

func TestMatchesConfigFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		format parser.Format
		want   bool
	}{
		{name: "env file", path: ".env", format: parser.FormatEnv, want: true},
		{name: "env example excluded", path: ".env.example", format: parser.FormatEnv, want: false},
		{name: "application yaml", path: "application.yml", format: parser.FormatYAML, want: true},
		{name: "docker compose yaml", path: "docker-compose.yml", format: parser.FormatYAML, want: true},
		{name: "github workflow yaml excluded", path: ".github/workflows/ci.yml", format: parser.FormatYAML, want: false},
		{name: "application properties", path: "application.properties", format: parser.FormatProps, want: true},
		{name: "other properties excluded", path: "gradle.properties", format: parser.FormatProps, want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := matchesConfigFile(tc.path, tc.format); got != tc.want {
				t.Fatalf("matchesConfigFile(%q, %q) = %v, want %v", tc.path, tc.format, got, tc.want)
			}
		})
	}
}
