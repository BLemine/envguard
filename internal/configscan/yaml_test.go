package configscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanYAML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantRequired []string
		wantOptional []string
		wantPaths    map[string]string // varname -> expected SourcePath
	}{
		{
			name: "spring boot style nested keys",
			content: `
spring:
  datasource:
    url: ${DATABASE_URL}
    password: ${DATABASE_PASSWORD}
server:
  port: ${SERVER_PORT:8080}
`,
			wantRequired: []string{"DATABASE_URL", "DATABASE_PASSWORD"},
			wantOptional: []string{"SERVER_PORT"},
			wantPaths: map[string]string{
				"DATABASE_URL":      "spring.datasource.url",
				"DATABASE_PASSWORD": "spring.datasource.password",
				"SERVER_PORT":       "server.port",
			},
		},
		{
			name: "embedded placeholders in string value",
			content: `
datasource:
  url: jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb
`,
			wantRequired: []string{"DB_HOST", "DB_PORT"},
			wantOptional: []string{},
		},
		{
			name: "optional placeholders with defaults",
			content: `
server:
  host: ${HOST:localhost}
  port: ${PORT:8080}
`,
			wantRequired: []string{},
			wantOptional: []string{"HOST", "PORT"},
			wantPaths: map[string]string{
				"HOST": "server.host",
				"PORT": "server.port",
			},
		},
		{
			name:         "empty file",
			content:      "",
			wantRequired: []string{},
			wantOptional: []string{},
		},
		{
			name: "no placeholders",
			content: `
app:
  name: myapp
  version: 1.0.0
`,
			wantRequired: []string{},
			wantOptional: []string{},
		},
		{
			name: "array elements with paths",
			content: `
servers:
  - host: ${SERVER1_HOST}
  - host: ${SERVER2_HOST:backup}
`,
			wantRequired: []string{"SERVER1_HOST"},
			wantOptional: []string{"SERVER2_HOST"},
			wantPaths: map[string]string{
				"SERVER1_HOST": "servers.0.host",
				"SERVER2_HOST": "servers.1.host",
			},
		},
		{
			name: "mixed embedded and default",
			content: `
redis:
  url: redis://${REDIS_HOST:localhost}:${REDIS_PORT:6379}
`,
			wantRequired: []string{},
			wantOptional: []string{"REDIS_HOST", "REDIS_PORT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "application.yml")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result, err := ScanYAML(path)
			if err != nil {
				t.Fatalf("ScanYAML error: %v", err)
			}

			assertPlaceholderNames(t, "required", result.Required, tt.wantRequired)
			assertPlaceholderNames(t, "optional", result.Optional, tt.wantOptional)

			if tt.wantPaths != nil {
				all := append(result.Required, result.Optional...)
				for _, p := range all {
					if wantPath, ok := tt.wantPaths[p.Name]; ok {
						if p.SourcePath != wantPath {
							t.Errorf("var %s: SourcePath = %q, want %q", p.Name, p.SourcePath, wantPath)
						}
					}
				}
			}
		})
	}
}

func TestScanYAML_DefaultValues(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yml")
	content := `server:
  port: ${SERVER_PORT:8080}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ScanYAML(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Optional) != 1 {
		t.Fatalf("expected 1 optional, got %d", len(result.Optional))
	}
	p := result.Optional[0]
	if p.Name != "SERVER_PORT" {
		t.Errorf("Name = %q, want SERVER_PORT", p.Name)
	}
	if p.DefaultValue == nil || *p.DefaultValue != "8080" {
		dv := "<nil>"
		if p.DefaultValue != nil {
			dv = *p.DefaultValue
		}
		t.Errorf("DefaultValue = %q, want 8080", dv)
	}
}

// assertPlaceholderNames checks that the placeholder list contains exactly the expected names (order-independent).
func assertPlaceholderNames(t *testing.T, label string, got []Placeholder, want []string) {
	t.Helper()
	gotNames := placeholderNames(got)

	if len(gotNames) != len(want) {
		t.Errorf("%s: got names %v, want %v", label, gotNames, want)
		return
	}
	wantSet := make(map[string]int, len(want))
	for _, n := range want {
		wantSet[n]++
	}
	gotSet := make(map[string]int, len(gotNames))
	for _, n := range gotNames {
		gotSet[n]++
	}
	for k, v := range wantSet {
		if gotSet[k] != v {
			t.Errorf("%s: want %q (x%d) but got %v", label, k, v, gotNames)
		}
	}
}
