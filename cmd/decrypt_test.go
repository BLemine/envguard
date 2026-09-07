package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteDecryptedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := writeDecryptedFile(path, []byte("original long content"), false); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe new file mode: %v", info.Mode())
	}
	if err := writeDecryptedFile(filepath.Join(t.TempDir(), ".env"), []byte("new"), true); err != nil {
		t.Fatal(err)
	}
	if err := writeDecryptedFile(path, []byte("replacement"), false); err == nil {
		t.Fatal("overwrote existing file without force")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "original long content" {
		t.Fatalf("existing content changed: %q, %v", data, err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeDecryptedFile(path, []byte("secret"), true); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(path)
	if err != nil || string(data) != "secret" {
		t.Fatalf("replacement: %q, %v", data, err)
	}
	info, err = os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe mode: %v", info.Mode())
	}
}
