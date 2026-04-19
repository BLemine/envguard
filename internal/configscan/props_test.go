package configscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanProps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantRequired []string
		wantOptional []string
		wantPaths    map[string]string
	}{
		{
			name: "basic key=value placeholders",
			content: `
database.url=${DATABASE_URL}
database.password=${DATABASE_PASSWORD}
`,
			wantRequired: []string{"DATABASE_URL", "DATABASE_PASSWORD"},
			wantOptional: []string{},
			wantPaths: map[string]string{
				"DATABASE_URL":      "database.url",
				"DATABASE_PASSWORD": "database.password",
			},
		},
		{
			name: "optional with default",
			content: `server.port=${SERVER_PORT:8080}
redis.host=${REDIS_HOST:localhost}
`,
			wantRequired: []string{},
			wantOptional: []string{"SERVER_PORT", "REDIS_HOST"},
		},
		{
			name: "comments skipped",
			content: `# this is a comment
! also a comment
database.url=${DATABASE_URL}
`,
			wantRequired: []string{"DATABASE_URL"},
			wantOptional: []string{},
		},
		{
			name: "empty lines skipped",
			content: `

database.url=${DATABASE_URL}

server.port=${PORT:9090}
`,
			wantRequired: []string{"DATABASE_URL"},
			wantOptional: []string{"PORT"},
		},
		{
			name: "embedded placeholder in jdbc url",
			content: `spring.datasource.url=jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb
`,
			wantRequired: []string{"DB_HOST", "DB_PORT"},
			wantOptional: []string{},
		},
		{
			name:         "empty file",
			content:      "",
			wantRequired: []string{},
			wantOptional: []string{},
		},
		{
			name: "no placeholders",
			content: `app.name=myapp
app.version=1.0.0
`,
			wantRequired: []string{},
			wantOptional: []string{},
		},
		{
			name: "lines without equals sign skipped",
			content: `this line has no equals
DATABASE_URL=${DATABASE_URL}
`,
			wantRequired: []string{"DATABASE_URL"},
			wantOptional: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, "application.properties")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result, err := ScanProps(path)
			if err != nil {
				t.Fatalf("ScanProps error: %v", err)
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

func TestScanProps_DefaultValue(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "app.properties")
	if err := os.WriteFile(path, []byte("server.port=${SERVER_PORT:8080}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ScanProps(path)
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
	if p.SourcePath != "server.port" {
		t.Errorf("SourcePath = %q, want server.port", p.SourcePath)
	}
}
