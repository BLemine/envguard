package configscan

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ScanYAML parses filePath as YAML and extracts all env variable placeholders.
func ScanYAML(filePath string) (*ScanResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", filePath, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}

	result := &ScanResult{}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		return result, nil
	}

	walkYAML(doc.Content[0], "", filePath, result)
	return result, nil
}

func walkYAML(node *yaml.Node, path, filePath string, result *ScanResult) {
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			childPath := keyNode.Value
			if path != "" {
				childPath = path + "." + keyNode.Value
			}
			walkYAML(valNode, childPath, filePath, result)
		}
	case yaml.SequenceNode:
		for i, item := range node.Content {
			childPath := fmt.Sprintf("%s.%d", path, i)
			walkYAML(item, childPath, filePath, result)
		}
	case yaml.ScalarNode:
		for _, p := range ExtractPlaceholders(node.Value, filePath, path) {
			if p.Required {
				result.Required = append(result.Required, p)
			} else {
				result.Optional = append(result.Optional, p)
			}
		}
	}
}
