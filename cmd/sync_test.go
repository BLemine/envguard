package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncFileCreatesMissingLocalFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, ".env.example")
	localPath := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(examplePath, []byte("API_KEY=\nDATABASE_URL=\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}

	added, skipped, err := syncFile(examplePath, localPath, "")
	if err != nil {
		t.Fatalf("syncFile returned error: %v", err)
	}

	if len(skipped) != 0 {
		t.Fatalf("expected no skipped keys, got %v", skipped)
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 added keys, got %v", added)
	}

	got, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read local: %v", err)
	}

	gotStr := string(got)
	if gotStr != "API_KEY=\nDATABASE_URL=\n" && gotStr != "DATABASE_URL=\nAPI_KEY=\n" {
		t.Fatalf("unexpected local contents: %q", gotStr)
	}
}

func TestSyncFileSeparatesAppendedKeysWithNewline(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, ".env.example")
	localPath := filepath.Join(tmpDir, ".env")

	if err := os.WriteFile(examplePath, []byte("EXISTING=1\nNEW_KEY=\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}
	if err := os.WriteFile(localPath, []byte("EXISTING=1"), 0o600); err != nil {
		t.Fatalf("write local: %v", err)
	}

	added, skipped, err := syncFile(examplePath, localPath, "")
	if err != nil {
		t.Fatalf("syncFile returned error: %v", err)
	}

	if len(added) != 1 || added[0] != "NEW_KEY" {
		t.Fatalf("unexpected added keys: %v", added)
	}
	if len(skipped) != 1 || skipped[0] != "EXISTING" {
		t.Fatalf("unexpected skipped keys: %v", skipped)
	}

	got, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read local: %v", err)
	}

	if string(got) != "EXISTING=1\nNEW_KEY=\n" {
		t.Fatalf("unexpected local contents: %q", string(got))
	}
}
