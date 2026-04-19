package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BLemine/envguard/internal/parser"
	"github.com/BLemine/envguard/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	syncExample string
	syncLocal   string
	syncFormat  string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Add missing keys from .env.example into your .env",
	Long:  `Reads .env.example and adds any missing keys into your .env with empty values. Existing keys are never overwritten.`,
	Run: func(cmd *cobra.Command, args []string) {
		format, err := resolveFormat(syncFormat, syncExample, syncLocal)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		added, skipped, err := syncFile(syncExample, syncLocal, format)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		reporter.PrintSyncResult(added, skipped)
	},
}

func syncFile(examplePath, localPath string, format parser.Format) ([]string, []string, error) {
	example, err := parser.ParseWithFormat(examplePath, format)
	if err != nil {
		return nil, nil, formatErrorContext("error reading "+examplePath, format, err)
	}

	local, err := parseLocalOrEmpty(localPath, format)
	if err != nil {
		return nil, nil, err
	}

	localKeys := parser.KeySet(local)
	keys := make([]string, 0, len(example.Keys))
	for key := range example.Keys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var added, skipped []string
	toAppend := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, exists := localKeys[key]; exists {
			skipped = append(skipped, key)
			continue
		}
		added = append(added, key)
		toAppend = append(toAppend, key+"=\n")
	}

	if len(toAppend) == 0 {
		return added, skipped, nil
	}

	if format == parser.FormatYAML {
		for _, key := range added {
			local.Keys[key] = ""
		}
		return added, skipped, writeSerializedFile(localPath, format, local.Keys)
	}

	if err := ensureParentDir(localPath); err != nil {
		return nil, nil, err
	}

	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening %s for writing: %w", localPath, err)
	}
	defer f.Close()

	needsNewline, err := fileNeedsLeadingNewline(localPath)
	if err != nil {
		return nil, nil, err
	}
	if needsNewline {
		if _, err := f.WriteString("\n"); err != nil {
			return nil, nil, fmt.Errorf("error writing newline to %s: %w", localPath, err)
		}
	}

	for _, line := range toAppend {
		if _, err := f.WriteString(line); err != nil {
			return nil, nil, fmt.Errorf("error writing %s: %w", localPath, err)
		}
	}

	return added, skipped, nil
}

func parseLocalOrEmpty(path string, format parser.Format) (*parser.EnvFile, error) {
	local, err := parser.ParseWithFormat(path, format)
	if err == nil {
		return local, nil
	}
	if os.IsNotExist(err) {
		return &parser.EnvFile{Path: path, Keys: map[string]string{}}, nil
	}
	return nil, formatErrorContext("error reading "+path, format, err)
}

func writeSerializedFile(path string, format parser.Format, keys map[string]string) error {
	if err := ensureParentDir(path); err != nil {
		return err
	}

	data, err := parser.SerializeWithFormat(format, keys)
	if err != nil {
		return formatErrorContext("error serializing "+path, format, err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return formatErrorContext("error writing "+path, format, err)
	}
	return nil
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("error creating parent directory for %s: %w", path, err)
	}
	return nil
}

func fileNeedsLeadingNewline(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error stating %s: %w", path, err)
	}
	if info.Size() == 0 {
		return false, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("error opening %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.Seek(-1, 2); err != nil {
		return false, fmt.Errorf("error seeking %s: %w", path, err)
	}

	buf := []byte{0}
	if _, err := f.Read(buf); err != nil {
		return false, fmt.Errorf("error reading %s: %w", path, err)
	}

	return buf[0] != '\n', nil
}

func init() {
	syncCmd.Flags().StringVar(&syncExample, "example", ".env.example", "Path to your .env.example file")
	syncCmd.Flags().StringVar(&syncLocal, "local", ".env", "Path to your local .env file")
	syncCmd.Flags().StringVar(&syncFormat, "format", "", "File format: env, yaml, or props (auto-detected when omitted)")
}
