package configscan

import "testing"

func TestExtractPlaceholdersRequiredVariable(t *testing.T) {
	t.Parallel()

	placeholders := ExtractPlaceholders("${DATABASE_URL}", "application.yml", "spring.datasource.url")

	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	got := placeholders[0]
	if got.Name != "DATABASE_URL" {
		t.Fatalf("expected DATABASE_URL, got %q", got.Name)
	}
	if !got.Required {
		t.Fatalf("expected placeholder to be required")
	}
	if got.DefaultValue != nil {
		t.Fatalf("expected no default value, got %q", *got.DefaultValue)
	}
	if got.SourceFile != "application.yml" {
		t.Fatalf("unexpected source file: %q", got.SourceFile)
	}
	if got.SourcePath != "spring.datasource.url" {
		t.Fatalf("unexpected source path: %q", got.SourcePath)
	}
}

func TestExtractPlaceholdersOptionalVariableWithDefault(t *testing.T) {
	t.Parallel()

	placeholders := ExtractPlaceholders("${PORT:8080}", "application.yml", "server.port")

	if len(placeholders) != 1 {
		t.Fatalf("expected 1 placeholder, got %d", len(placeholders))
	}

	got := placeholders[0]
	if got.Name != "PORT" {
		t.Fatalf("expected PORT, got %q", got.Name)
	}
	if got.Required {
		t.Fatalf("expected placeholder to be optional")
	}
	if got.DefaultValue == nil || *got.DefaultValue != "8080" {
		t.Fatalf("expected default 8080, got %#v", got.DefaultValue)
	}
}

func TestExtractPlaceholdersMultipleInSingleValue(t *testing.T) {
	t.Parallel()

	placeholders := ExtractPlaceholders("jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb", "application.yml", "spring.datasource.url")

	if len(placeholders) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(placeholders))
	}

	if placeholders[0].Name != "DB_HOST" || placeholders[1].Name != "DB_PORT" {
		t.Fatalf("unexpected placeholders: %#v", placeholders)
	}
}

func TestExtractPlaceholdersIgnoresMalformedPatterns(t *testing.T) {
	t.Parallel()

	placeholders := ExtractPlaceholders("${lowercase}${INVALID-NAME}${MISSING", "application.yml", "broken.value")

	if len(placeholders) != 0 {
		t.Fatalf("expected malformed placeholders to be ignored, got %#v", placeholders)
	}
}

func TestExtractPlaceholdersDeduplicatesWithinSameSource(t *testing.T) {
	t.Parallel()

	placeholders := ExtractPlaceholders("${DB_HOST}:${DB_HOST}", "application.yml", "spring.datasource.url")

	if len(placeholders) != 1 {
		t.Fatalf("expected duplicate placeholders to collapse, got %#v", placeholders)
	}
}
