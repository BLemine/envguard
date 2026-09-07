package configscan

import "testing"

func TestNestedFallbacks(t *testing.T) {
	for _, tt := range []struct {
		expression string
		keys       []string
		missing    bool
	}{
		{"${PRIMARY:${SECONDARY}}", nil, true},
		{"${PRIMARY:${SECONDARY}}", []string{"PRIMARY"}, false},
		{"${PRIMARY:${SECONDARY}}", []string{"SECONDARY"}, false},
		{"${PRIMARY:${SECONDARY:${THIRD}}}", []string{"THIRD"}, false},
		{"${PRIMARY:${SECONDARY:${THIRD}}}", nil, true},
		{"${PRIMARY:${SECONDARY:localhost}}", nil, false},
		{"${PRIMARY:${SECONDARY:}}", nil, false},
		{"${URL:jdbc://${HOST}:${PORT}}", []string{"HOST"}, true},
		{"${URL:jdbc://${HOST}:${PORT}}", []string{"HOST", "PORT"}, false},
		{"${URL:jdbc://${HOST}:${PORT}}", []string{"URL"}, false},
		{"${PRIMARY:default}/${PRIMARY}", nil, true},
	} {
		t.Run(tt.expression, func(t *testing.T) {
			result := &ScanResult{}
			for _, p := range ExtractPlaceholders(tt.expression, "test.yml", "url") {
				if p.Required {
					result.Required = append(result.Required, p)
				} else {
					result.Optional = append(result.Optional, p)
				}
			}
			keys := make(map[string]struct{})
			for _, key := range tt.keys {
				keys[key] = struct{}{}
			}
			if got := Missing(result, keys); (len(got) > 0) != tt.missing {
				t.Fatalf("keys %v: missing=%v", tt.keys, got)
			}
		})
	}
	ps := ExtractPlaceholders("${PRIMARY:${SECONDARY:${THIRD:8080}}}", "test.yml", "port")
	if len(ps) != 1 || ps[0].DefaultValue == nil || *ps[0].DefaultValue != "${SECONDARY:${THIRD:8080}}" {
		t.Fatalf("truncated default: %#v", ps)
	}
	if len(ps[0].Fallback) != 1 || len(ps[0].Fallback[0].Fallback) != 1 || ps[0].Fallback[0].Fallback[0].Name != "THIRD" {
		t.Fatalf("missing fallback tree: %#v", ps)
	}
}
