package parser

import (
	"os"
	"sort"

	"github.com/joho/godotenv"
)

type EnvFile struct {
	Path string
	Keys map[string]string
}

func Parse(path string) (*EnvFile, error) {
	return parseEnv(path)
}

func parseEnv(path string) (*EnvFile, error) {
	keys, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}
	return &EnvFile{Path: path, Keys: keys}, nil
}

func serializeFlat(keys map[string]string) []byte {
	sortedKeys := make([]string, 0, len(keys))
	for key := range keys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	var out []byte
	for _, key := range sortedKeys {
		out = append(out, []byte(key+"="+keys[key]+"\n")...)
	}
	return out
}

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func KeySet(env *EnvFile) map[string]struct{} {
	set := make(map[string]struct{}, len(env.Keys))
	for k := range env.Keys {
		set[k] = struct{}{}
	}
	return set
}
