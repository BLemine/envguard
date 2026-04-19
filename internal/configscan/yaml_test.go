package configscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanYAMLExtractsNestedAndEmbeddedPlaceholders(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "application.yml")
	content := `spring:
  datasource:
    url: jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb
    password: ${DATABASE_PASSWORD}
server:
  port: ${SERVER_PORT:8080}
`

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	result, err := ScanYAML(path)
	if err != nil {
		t.Fatalf("ScanYAML returned error: %v", err)
	}

	if len(result.Required) != 3 {
		t.Fatalf("expected 3 required placeholders, got %#v", result.Required)
	}
	if len(result.Optional) != 1 {
		t.Fatalf("expected 1 optional placeholder, got %#v", result.Optional)
	}
	if !hasPlaceholder(result.Required, "DB_HOST", "spring.datasource.url") {
		t.Fatalf("expected DB_HOST at spring.datasource.url, got %#v", result.Required)
	}
	if !hasPlaceholder(result.Required, "DB_PORT", "spring.datasource.url") {
		t.Fatalf("expected DB_PORT at spring.datasource.url, got %#v", result.Required)
	}
	if !hasPlaceholder(result.Required, "DATABASE_PASSWORD", "spring.datasource.password") {
		t.Fatalf("expected DATABASE_PASSWORD at spring.datasource.password, got %#v", result.Required)
	}
	if result.Optional[0].Name != "SERVER_PORT" {
		t.Fatalf("unexpected optional placeholder: %#v", result.Optional[0])
	}
	if result.Optional[0].DefaultValue == nil || *result.Optional[0].DefaultValue != "8080" {
		t.Fatalf("expected SERVER_PORT default 8080, got %#v", result.Optional[0].DefaultValue)
	}
}

func TestScanYAMLSupportsArraysInPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "application.yml")
	content := `servers:
  - host: ${REDIS_HOST}
  - host: ${CACHE_HOST:localhost}
`

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	result, err := ScanYAML(path)
	if err != nil {
		t.Fatalf("ScanYAML returned error: %v", err)
	}

	if len(result.Required) != 1 {
		t.Fatalf("expected 1 required placeholder, got %#v", result.Required)
	}
	if result.Required[0].SourcePath != "servers.0.host" {
		t.Fatalf("unexpected required source path: %q", result.Required[0].SourcePath)
	}
	if len(result.Optional) != 1 {
		t.Fatalf("expected 1 optional placeholder, got %#v", result.Optional)
	}
	if result.Optional[0].SourcePath != "servers.1.host" {
		t.Fatalf("unexpected optional source path: %q", result.Optional[0].SourcePath)
	}
}

func hasPlaceholder(placeholders []Placeholder, name, sourcePath string) bool {
	for _, placeholder := range placeholders {
		if placeholder.Name == name && placeholder.SourcePath == sourcePath {
			return true
		}
	}
	return false
}
