package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want Format
	}{
		{name: "explicit env", path: ".env", want: FormatEnv},
		{name: "yaml extension", path: "application.yml", want: FormatYAML},
		{name: "yaml long extension", path: "application.yaml", want: FormatYAML},
		{name: "properties extension", path: "application.properties", want: FormatProps},
		{name: "unknown defaults to env", path: "config.txt", want: FormatEnv},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := DetectFormat(tc.path); got != tc.want {
				t.Fatalf("DetectFormat(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestParseYAMLFlattensNestedMapsAndArrays(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "application.yml")
	content := "" +
		"spring:\n" +
		"  datasource:\n" +
		"    url: jdbc:postgresql://localhost/mydb\n" +
		"    password: \"\"\n" +
		"server:\n" +
		"  hosts:\n" +
		"    - localhost\n" +
		"    - prod.example.com\n"

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	got, err := ParseWithFormat(path, FormatYAML)
	if err != nil {
		t.Fatalf("ParseWithFormat returned error: %v", err)
	}

	want := map[string]string{
		"spring.datasource.url":      "jdbc:postgresql://localhost/mydb",
		"spring.datasource.password": "",
		"server.hosts.0":             "localhost",
		"server.hosts.1":             "prod.example.com",
	}

	if !reflect.DeepEqual(got.Keys, want) {
		t.Fatalf("unexpected yaml keys: got %#v want %#v", got.Keys, want)
	}
}

func TestParsePropsSkipsCommentsAndBlankLines(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "application.properties")
	content := "" +
		"# comment\n" +
		"spring.datasource.url=jdbc:postgresql://localhost/mydb\n" +
		"\n" +
		"server.port=8080\n"

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write properties: %v", err)
	}

	got, err := ParseWithFormat(path, FormatProps)
	if err != nil {
		t.Fatalf("ParseWithFormat returned error: %v", err)
	}

	want := map[string]string{
		"spring.datasource.url": "jdbc:postgresql://localhost/mydb",
		"server.port":           "8080",
	}

	if !reflect.DeepEqual(got.Keys, want) {
		t.Fatalf("unexpected props keys: got %#v want %#v", got.Keys, want)
	}
}
