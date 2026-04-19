package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCheckFailsWhenRequiredKeyIsEmpty(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, ".env.example")
	localPath := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(examplePath, []byte("REQUIRED=\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}
	if err := os.WriteFile(localPath, []byte("REQUIRED=\n"), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}

	result, err := runCheck(examplePath, localPath, "")
	if err != nil {
		t.Fatalf("runCheck returned error: %v", err)
	}

	if result.Empty != 1 {
		t.Fatalf("expected 1 empty key, got %#v", result)
	}
	if !shouldFailCheck(result) {
		t.Fatalf("expected check to fail for empty values")
	}
}
