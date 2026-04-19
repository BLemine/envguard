package configscan

import (
	"testing"
)

// TestStrictComparison tests the logic of comparing required placeholders against env key sets.
// This mirrors what cmd/scan_config.go does when --local or --example are provided.
func TestStrictComparison(t *testing.T) {
	t.Parallel()

	required := []Placeholder{
		{Name: "DATABASE_URL", Required: true},
		{Name: "DATABASE_PASSWORD", Required: true},
		{Name: "API_KEY", Required: true},
	}
	optional := []Placeholder{
		{Name: "SERVER_PORT", Required: false, DefaultValue: strPtr("8080")},
	}

	tests := []struct {
		name         string
		envKeys      map[string]struct{}
		wantMissing  []string
	}{
		{
			name: "all required vars present — no missing",
			envKeys: map[string]struct{}{
				"DATABASE_URL":      {},
				"DATABASE_PASSWORD": {},
				"API_KEY":           {},
			},
			wantMissing: []string{},
		},
		{
			name: "some required vars missing",
			envKeys: map[string]struct{}{
				"DATABASE_URL": {},
			},
			wantMissing: []string{"DATABASE_PASSWORD", "API_KEY"},
		},
		{
			name:        "all required vars missing",
			envKeys:     map[string]struct{}{},
			wantMissing: []string{"DATABASE_URL", "DATABASE_PASSWORD", "API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var missing []string
			for _, p := range required {
				if _, ok := tt.envKeys[p.Name]; !ok {
					missing = append(missing, p.Name)
				}
			}
			// optional vars should never appear in missing
			for _, p := range optional {
				for _, m := range missing {
					if m == p.Name {
						t.Errorf("optional var %q should not appear in missing list", p.Name)
					}
				}
			}

			if len(missing) != len(tt.wantMissing) {
				t.Fatalf("missing = %v, want %v", missing, tt.wantMissing)
			}
			wantSet := make(map[string]struct{}, len(tt.wantMissing))
			for _, n := range tt.wantMissing {
				wantSet[n] = struct{}{}
			}
			for _, n := range missing {
				if _, ok := wantSet[n]; !ok {
					t.Errorf("unexpected missing var %q", n)
				}
			}
		})
	}
}
