package parser

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func parseYAML(path string) (*EnvFile, error) {
	var root any
	data, err := osReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	keys := make(map[string]string)
	flattenYAML("", root, keys)
	return &EnvFile{Path: path, Keys: keys}, nil
}

func flattenYAML(prefix string, value any, out map[string]string) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			flattenYAML(joinKey(prefix, key), nested, out)
		}
	case map[any]any:
		for key, nested := range typed {
			flattenYAML(joinKey(prefix, fmt.Sprint(key)), nested, out)
		}
	case []any:
		for index, nested := range typed {
			flattenYAML(joinKey(prefix, strconv.Itoa(index)), nested, out)
		}
	case nil:
		if prefix != "" {
			out[prefix] = ""
		}
	default:
		if prefix != "" {
			out[prefix] = fmt.Sprint(typed)
		}
	}
}

func serializeYAML(keys map[string]string) ([]byte, error) {
	root := map[string]any{}

	sortedKeys := make([]string, 0, len(keys))
	for key := range keys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		insertYAMLValue(root, strings.Split(key, "."), keys[key])
	}

	data, err := yaml.Marshal(root)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func insertYAMLValue(root map[string]any, parts []string, value string) {
	if len(parts) == 0 {
		return
	}

	current := root
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if i == len(parts)-1 {
			current[part] = value
			return
		}

		next := parts[i+1]
		existing, ok := current[part]
		if !ok {
			if isIndex(next) {
				current[part] = []any{}
			} else {
				current[part] = map[string]any{}
			}
			existing = current[part]
		}

		switch typed := existing.(type) {
		case map[string]any:
			current = typed
		case []any:
			index, _ := strconv.Atoi(next)
			typed = ensureListSize(typed, index+1)
			current[part] = typed
			if typed[index] == nil {
				typed[index] = map[string]any{}
			}
			nested, _ := typed[index].(map[string]any)
			current = nested
			i++
		}
	}
}

func ensureListSize(list []any, size int) []any {
	for len(list) < size {
		list = append(list, nil)
	}
	return list
}

func isIndex(part string) bool {
	_, err := strconv.Atoi(part)
	return err == nil
}

func joinKey(prefix, next string) string {
	if prefix == "" {
		return next
	}
	return prefix + "." + next
}
