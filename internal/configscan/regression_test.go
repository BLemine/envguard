package configscan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScannerRegressions(t *testing.T) {
	for _, tt := range []struct {
		name, content string
		names         []string
		wantErr       bool
	}{
		{"multi.yml", "a: ${FIRST}\n---\nb: ${SECOND}\n", []string{"FIRST", "SECOND"}, false},
		{"broken.yml", "a: ${FIRST}\n---\nb: [\n", nil, true},
		{"nested.yml", "a: ${FIRST:${SECOND}}\n", nil, false},
		{"nested.properties", "a=${FIRST:${SECOND}}\n", nil, false},
		{"separators.properties", "a: ${FIRST}\nb ${SECOND}\nc = ${THIRD}\n", []string{"FIRST", "SECOND", "THIRD"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.name)
			if err := os.WriteFile(path, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}
			var result *ScanResult
			var err error
			if strings.HasSuffix(path, ".yml") {
				result, err = ScanYAML(path)
			} else {
				result, err = ScanProps(path)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected rejection instead of partial success")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assertPlaceholderNames(t, "required", result.Required, tt.names)
		})
	}
}
