package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses a YAML file into a flat map using dot notation for nested keys.
// Arrays are indexed: server.hosts.0, server.hosts.1, etc.
func ParseYAML(path string) (*EnvFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	keys := make(map[string]string)
	flattenYAML("", raw, keys)

	return &EnvFile{Path: path, Keys: keys}, nil
}

func flattenYAML(prefix string, value interface{}, result map[string]string) {
	switch v := value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenYAML(newKey, val, result)
		}
	case []interface{}:
		for i, val := range v {
			flattenYAML(fmt.Sprintf("%s.%d", prefix, i), val, result)
		}
	case nil:
		if prefix != "" {
			result[prefix] = ""
		}
	default:
		if prefix != "" {
			result[prefix] = fmt.Sprintf("%v", v)
		}
	}
}
