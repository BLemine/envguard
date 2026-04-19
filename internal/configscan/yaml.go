package configscan

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

func ScanYAML(path string) (*ScanResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("parse yaml %s: %w", path, err)
	}

	result := &ScanResult{}
	if len(node.Content) == 0 {
		return result, nil
	}

	walkYAMLNode(result, path, "", node.Content[0])
	return result, nil
}

func walkYAMLNode(result *ScanResult, sourceFile, sourcePath string, node *yaml.Node) {
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			walkYAMLNode(result, sourceFile, sourcePath, child)
		}
	case yaml.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			childPath := joinPath(sourcePath, key)
			walkYAMLNode(result, sourceFile, childPath, node.Content[i+1])
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			childPath := joinPath(sourcePath, strconv.Itoa(i))
			walkYAMLNode(result, sourceFile, childPath, child)
		}
	case yaml.ScalarNode:
		if node.Tag != "!!str" {
			return
		}
		appendPlaceholders(result, ExtractPlaceholders(node.Value, sourceFile, sourcePath))
	}
}

func joinPath(base, next string) string {
	if base == "" {
		return next
	}
	return base + "." + next
}
