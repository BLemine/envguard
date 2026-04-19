package configscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanPropertiesExtractsPlaceholdersAndSkipsComments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "application.properties")
	content := `# comment
! another comment
spring.datasource.url=jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb
spring.datasource.password=${DATABASE_PASSWORD}
server.port=${SERVER_PORT:8080}
`

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write properties: %v", err)
	}

	result, err := ScanProperties(path)
	if err != nil {
		t.Fatalf("ScanProperties returned error: %v", err)
	}

	if len(result.Required) != 3 {
		t.Fatalf("expected 3 required placeholders, got %#v", result.Required)
	}
	if len(result.Optional) != 1 {
		t.Fatalf("expected 1 optional placeholder, got %#v", result.Optional)
	}
	if result.Optional[0].SourcePath != "server.port" {
		t.Fatalf("unexpected optional source path: %q", result.Optional[0].SourcePath)
	}
}
