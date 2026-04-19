package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BLemine/envguard/internal/parser"
)

func TestResolveFormatUsesExplicitFlag(t *testing.T) {
	t.Parallel()

	got, err := resolveFormat("yaml", ".env", "application.properties")
	if err != nil {
		t.Fatalf("resolveFormat returned error: %v", err)
	}
	if got != parser.FormatYAML {
		t.Fatalf("expected yaml format, got %q", got)
	}
}

func TestResolveFormatAutodetectsFromPaths(t *testing.T) {
	t.Parallel()

	got, err := resolveFormat("", "application.yml.example", "application.yml")
	if err != nil {
		t.Fatalf("resolveFormat returned error: %v", err)
	}
	if got != parser.FormatYAML {
		t.Fatalf("expected yaml format, got %q", got)
	}
}

func TestSyncFileSupportsYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, "application.yml.example")
	localPath := filepath.Join(tmpDir, "application.yml")

	example := "" +
		"spring:\n" +
		"  datasource:\n" +
		"    url:\n" +
		"    password:\n" +
		"server:\n" +
		"  port:\n"
	local := "" +
		"spring:\n" +
		"  datasource:\n" +
		"    url: jdbc:postgresql://localhost:5432/mydb\n"

	if err := os.WriteFile(examplePath, []byte(example), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}
	if err := os.WriteFile(localPath, []byte(local), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}

	added, skipped, err := syncFile(examplePath, localPath, parser.FormatYAML)
	if err != nil {
		t.Fatalf("syncFile returned error: %v", err)
	}

	if len(added) != 2 {
		t.Fatalf("expected 2 added keys, got %v", added)
	}
	if len(skipped) != 1 || skipped[0] != "spring.datasource.url" {
		t.Fatalf("unexpected skipped keys: %v", skipped)
	}

	got, err := parser.ParseWithFormat(localPath, parser.FormatYAML)
	if err != nil {
		t.Fatalf("ParseWithFormat returned error: %v", err)
	}

	if got.Keys["spring.datasource.password"] != "" {
		t.Fatalf("expected empty password key, got %#v", got.Keys["spring.datasource.password"])
	}
	if got.Keys["server.port"] != "" {
		t.Fatalf("expected empty server.port key, got %#v", got.Keys["server.port"])
	}
}

func TestRunCheckAutodetectsYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, "application.yml.example")
	localPath := filepath.Join(tmpDir, "application.yml")

	if err := os.WriteFile(examplePath, []byte("server:\n  port:\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}
	if err := os.WriteFile(localPath, []byte("server:\n  port: 8080\n"), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}

	format, err := resolveFormat("", examplePath, localPath)
	if err != nil {
		t.Fatalf("resolveFormat returned error: %v", err)
	}

	result, err := runCheck(examplePath, localPath, format)
	if err != nil {
		t.Fatalf("runCheck returned error: %v", err)
	}

	if result.OK != 1 {
		t.Fatalf("expected one ok key, got %#v", result)
	}
}

func TestValidateFileSupportsProperties(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	localPath := filepath.Join(tmpDir, "application.properties")

	content := "" +
		"spring.datasource.url=jdbc:postgresql://localhost:5432/mydb\n" +
		"spring.datasource.password=secret123\n"
	if err := os.WriteFile(localPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write properties: %v", err)
	}

	failed, err := validateFile(localPath, parser.FormatProps, []string{
		"spring.datasource.url",
		"spring.datasource.password",
	})
	if err != nil {
		t.Fatalf("validateFile returned error: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected zero failed keys, got %d", failed)
	}
}
