package configscan

import (
	"regexp"
	"sort"
)

var placeholderPattern = regexp.MustCompile(`\$\{([A-Z0-9_]+)(?::([^}]*))?\}`)

func ExtractPlaceholders(value, sourceFile, sourcePath string) []Placeholder {
	matches := placeholderPattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	placeholders := make([]Placeholder, 0, len(matches))
	for _, match := range matches {
		name := match[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		placeholder := Placeholder{
			Name:       name,
			Required:   true,
			SourceFile: sourceFile,
			SourcePath: sourcePath,
		}
		if len(match) > 2 && match[2] != "" {
			defaultValue := match[2]
			placeholder.DefaultValue = &defaultValue
			placeholder.Required = false
		}

		placeholders = append(placeholders, placeholder)
	}

	sort.Slice(placeholders, func(i, j int) bool {
		return placeholders[i].Name < placeholders[j].Name
	})

	return placeholders
}
