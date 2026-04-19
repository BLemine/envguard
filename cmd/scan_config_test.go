package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunScanConfigStrictSatisfied(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "application.yml")
	localPath := filepath.Join(tmpDir, ".env")
	examplePath := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(configPath, []byte("database:\n  url: ${DATABASE_URL}\nserver:\n  port: ${SERVER_PORT:8080}\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(localPath, []byte("DATABASE_URL=postgres://db\n"), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}
	if err := os.WriteFile(examplePath, []byte("DATABASE_URL=\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}

	result, err := runScanConfig(scanConfigOptions{
		Files:   []string{configPath},
		Format:  "auto",
		Strict:  true,
		Local:   localPath,
		Example: examplePath,
	})
	if err != nil {
		t.Fatalf("runScanConfig returned error: %v", err)
	}

	if len(result.MissingLocal) != 0 {
		t.Fatalf("expected no missing local vars, got %v", result.MissingLocal)
	}
	if len(result.MissingExample) != 0 {
		t.Fatalf("expected no missing example vars, got %v", result.MissingExample)
	}
}

func TestRunScanConfigStrictMissingRequiredVars(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "application.yml")
	localPath := filepath.Join(tmpDir, ".env")
	examplePath := filepath.Join(tmpDir, ".env.example")

	if err := os.WriteFile(configPath, []byte("database:\n  url: ${DATABASE_URL}\n  password: ${DATABASE_PASSWORD}\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(localPath, []byte("DATABASE_URL=postgres://db\n"), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}
	if err := os.WriteFile(examplePath, []byte("DATABASE_PASSWORD=\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}

	result, err := runScanConfig(scanConfigOptions{
		Files:   []string{configPath},
		Format:  "auto",
		Strict:  true,
		Local:   localPath,
		Example: examplePath,
	})
	if err != nil {
		t.Fatalf("runScanConfig returned error: %v", err)
	}

	if len(result.MissingLocal) != 1 || result.MissingLocal[0] != "DATABASE_PASSWORD" {
		t.Fatalf("unexpected missing local vars: %v", result.MissingLocal)
	}
	if len(result.MissingExample) != 1 || result.MissingExample[0] != "DATABASE_URL" {
		t.Fatalf("unexpected missing example vars: %v", result.MissingExample)
	}
	if !shouldFailScanConfig(result, true) {
		t.Fatalf("expected strict scan config result to fail")
	}
}

func TestRunScanConfigIgnoresOptionalDefaultsForStrictChecks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "application.properties")
	localPath := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(configPath, []byte("server.port=${SERVER_PORT:8080}\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(localPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}

	result, err := runScanConfig(scanConfigOptions{
		Files:  []string{configPath},
		Format: "auto",
		Strict: true,
		Local:  localPath,
	})
	if err != nil {
		t.Fatalf("runScanConfig returned error: %v", err)
	}

	if len(result.Required) != 0 {
		t.Fatalf("expected no required placeholders, got %#v", result.Required)
	}
	if len(result.MissingLocal) != 0 {
		t.Fatalf("expected no missing local vars, got %v", result.MissingLocal)
	}
	if shouldFailScanConfig(result, true) {
		t.Fatalf("expected strict scan config result to pass with only optional placeholders")
	}
}
