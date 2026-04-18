package differ

import "github.com/yourusername/envguard/internal/parser"

type Status int

const (
	StatusOK           Status = iota
	StatusMissing             // in example but not in local .env
	StatusUndocumented        // in local .env but not in example
	StatusEmpty               // key exists but has no value
)

type DiffEntry struct {
	Key    string
	Status Status
}

type DiffResult struct {
	Entries []DiffEntry
	OK      int
	Missing int
	Undocumented int
	Empty   int
}

func Diff(example, local *parser.EnvFile) *DiffResult {
	exampleKeys := parser.KeySet(example)
	localKeys := parser.KeySet(local)

	result := &DiffResult{}

	for key := range exampleKeys {
		if _, exists := localKeys[key]; !exists {
			result.Entries = append(result.Entries, DiffEntry{Key: key, Status: StatusMissing})
			result.Missing++
		} else if local.Keys[key] == "" {
			result.Entries = append(result.Entries, DiffEntry{Key: key, Status: StatusEmpty})
			result.Empty++
		} else {
			result.Entries = append(result.Entries, DiffEntry{Key: key, Status: StatusOK})
			result.OK++
		}
	}

	for key := range localKeys {
		if _, exists := exampleKeys[key]; !exists {
			result.Entries = append(result.Entries, DiffEntry{Key: key, Status: StatusUndocumented})
			result.Undocumented++
		}
	}

	return result
}
