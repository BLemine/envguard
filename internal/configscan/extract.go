package configscan

import (
	"regexp"
	"strings"
)

var variableName = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// ExtractPlaceholders preserves nested defaults as a tree. Only top-level
// references are returned: fallback children are conditional on their parent.
func ExtractPlaceholders(value, sourceFile, sourcePath string) []Placeholder {
	result := make([]Placeholder, 0)
	seen := make(map[string]bool)
	for i := 0; i < len(value); i++ {
		if !strings.HasPrefix(value[i:], "${") {
			continue
		}
		start, colon, depth, end := i+2, -1, 1, -1
		for j := start; j < len(value); j++ {
			if strings.HasPrefix(value[j:], "${") {
				depth++
				j++
				continue
			}
			if value[j] == '}' {
				depth--
				if depth == 0 {
					end = j
					break
				}
			}
			if value[j] == ':' && depth == 1 && colon < 0 {
				colon = j
			}
		}
		if end < 0 {
			continue
		}
		nameEnd := end
		if colon >= 0 {
			nameEnd = colon
		}
		name := value[start:nameEnd]
		if !variableName.MatchString(name) {
			i = end
			continue
		}
		expression := value[i : end+1]
		i = end
		// Different requirements for the same variable must not hide one another.
		if seen[expression] {
			continue
		}
		seen[expression] = true
		p := Placeholder{Name: name, Required: colon < 0, SourceFile: sourceFile, SourcePath: sourcePath}
		if colon >= 0 {
			dv := value[colon+1 : end]
			p.DefaultValue = &dv
			p.Fallback = ExtractPlaceholders(dv, sourceFile, sourcePath)
		}
		result = append(result, p)
	}
	return result
}

// Missing returns unresolved top-level requirements. A fallback containing
// several references resolves only when every reference in it resolves.
// Presence (including an empty value) matches scan-config's existing contract.
func Missing(result *ScanResult, keys map[string]struct{}) []string {
	var missing []string
	seen := make(map[string]bool)
	for _, group := range [][]Placeholder{result.Required, result.Optional} {
		for _, p := range group {
			if !p.satisfied(keys) {
				requirement := p.Name
				if len(p.Fallback) > 0 {
					requirement = "${" + p.Name + ":" + *p.DefaultValue + "}"
				}
				if !seen[requirement] {
					missing = append(missing, requirement)
					seen[requirement] = true
				}
			}
		}
	}
	return missing
}

func (p Placeholder) satisfied(keys map[string]struct{}) bool {
	if _, ok := keys[p.Name]; ok {
		return true
	}
	if p.DefaultValue == nil {
		return false
	}
	for _, child := range p.Fallback {
		if !child.satisfied(keys) {
			return false
		}
	}
	return true
}
