package configscan

import "regexp"

// matches ${VAR} or ${VAR:default}
// group 1 = variable name ([A-Z0-9_]+)
// group 2 = ":default" portion if present (includes the colon)
// group 3 = default value (without the colon)
var placeholderRe = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)(:([^}]*))?\}`)

// ExtractPlaceholders finds all ${VAR} and ${VAR:default} patterns in value.
// Duplicates (same name+sourceFile+sourcePath) within a single call are deduplicated.
func ExtractPlaceholders(value, sourceFile, sourcePath string) []Placeholder {
	matches := placeholderRe.FindAllStringSubmatch(value, -1)
	seen := make(map[string]struct{}, len(matches))
	result := make([]Placeholder, 0, len(matches))

	for _, m := range matches {
		name := m[1]
		dedupKey := name + "\x00" + sourceFile + "\x00" + sourcePath
		if _, exists := seen[dedupKey]; exists {
			continue
		}
		seen[dedupKey] = struct{}{}

		hasColon := m[2] != ""
		p := Placeholder{
			Name:       name,
			Required:   !hasColon,
			SourceFile: sourceFile,
			SourcePath: sourcePath,
		}
		if hasColon {
			dv := m[3]
			p.DefaultValue = &dv
		}
		result = append(result, p)
	}
	return result
}
