package parser

import (
	"github.com/joho/godotenv"
)

type EnvFile struct {
	Path string
	Keys map[string]string
}

func Parse(path string) (*EnvFile, error) {
	keys, err := godotenv.Read(path)
	if err != nil {
		return nil, err
	}
	return &EnvFile{Path: path, Keys: keys}, nil
}

func KeySet(env *EnvFile) map[string]struct{} {
	set := make(map[string]struct{}, len(env.Keys))
	for k := range env.Keys {
		set[k] = struct{}{}
	}
	return set
}
