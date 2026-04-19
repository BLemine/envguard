package configscan

import (
	"testing"
)

func TestExtractPlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        string
		wantNames    []string
		wantRequired []bool
		wantDefaults []*string
	}{
		{
			name:         "single required",
			value:        "${DATABASE_URL}",
			wantNames:    []string{"DATABASE_URL"},
			wantRequired: []bool{true},
			wantDefaults: []*string{nil},
		},
		{
			name:         "optional with default",
			value:        "${PORT:8080}",
			wantNames:    []string{"PORT"},
			wantRequired: []bool{false},
			wantDefaults: []*string{strPtr("8080")},
		},
		{
			name:         "multiple placeholders in one string",
			value:        "jdbc:postgresql://${DB_HOST}:${DB_PORT}/mydb",
			wantNames:    []string{"DB_HOST", "DB_PORT"},
			wantRequired: []bool{true, true},
			wantDefaults: []*string{nil, nil},
		},
		{
			name:         "mixed required and optional",
			value:        "http://${HOST:localhost}:${PORT:8080}",
			wantNames:    []string{"HOST", "PORT"},
			wantRequired: []bool{false, false},
			wantDefaults: []*string{strPtr("localhost"), strPtr("8080")},
		},
		{
			name:         "malformed placeholders ignored",
			value:        "${lowercase} ${123INVALID} ${VALID}",
			wantNames:    []string{"VALID"},
			wantRequired: []bool{true},
			wantDefaults: []*string{nil},
		},
		{
			name:         "duplicate within same value deduplicated",
			value:        "${VAR}/${VAR}",
			wantNames:    []string{"VAR"},
			wantRequired: []bool{true},
			wantDefaults: []*string{nil},
		},
		{
			name:      "empty value returns no placeholders",
			value:     "",
			wantNames: []string{},
		},
		{
			name:         "no placeholder in plain text",
			value:        "just a plain string",
			wantNames:    []string{},
		},
		{
			name:         "optional with empty default",
			value:        "${VAR:}",
			wantNames:    []string{"VAR"},
			wantRequired: []bool{false},
			wantDefaults: []*string{strPtr("")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractPlaceholders(tt.value, "test.yml", "test.path")

			if len(got) != len(tt.wantNames) {
				t.Fatalf("got %d placeholders, want %d; names: %v", len(got), len(tt.wantNames), placeholderNames(got))
			}
			for i, p := range got {
				if p.Name != tt.wantNames[i] {
					t.Errorf("[%d] Name = %q, want %q", i, p.Name, tt.wantNames[i])
				}
				if tt.wantRequired != nil && p.Required != tt.wantRequired[i] {
					t.Errorf("[%d] Required = %v, want %v", i, p.Required, tt.wantRequired[i])
				}
				if tt.wantDefaults != nil {
					wantDV := tt.wantDefaults[i]
					if wantDV == nil && p.DefaultValue != nil {
						t.Errorf("[%d] DefaultValue = %q, want nil", i, *p.DefaultValue)
					} else if wantDV != nil {
						if p.DefaultValue == nil {
							t.Errorf("[%d] DefaultValue = nil, want %q", i, *wantDV)
						} else if *p.DefaultValue != *wantDV {
							t.Errorf("[%d] DefaultValue = %q, want %q", i, *p.DefaultValue, *wantDV)
						}
					}
				}
			}
		})
	}
}

func placeholderNames(ps []Placeholder) []string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}
	return names
}
