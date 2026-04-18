package reporter

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/yourusername/envguard/internal/differ"
)

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
	faint  = color.New(color.Faint)
)

func PrintDiff(result *differ.DiffResult, examplePath, localPath string) {
	bold.Printf("\nComparing %s → %s\n\n", faint.Sprint(examplePath), faint.Sprint(localPath))

	entries := result.Entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	for _, entry := range entries {
		switch entry.Status {
		case differ.StatusOK:
			green.Printf("  ✓ %-40s ok\n", entry.Key)
		case differ.StatusMissing:
			red.Printf("  ✗ %-40s missing from your .env\n", entry.Key)
		case differ.StatusUndocumented:
			yellow.Printf("  ⚠ %-40s in .env but not in .env.example\n", entry.Key)
		case differ.StatusEmpty:
			yellow.Printf("  ⚠ %-40s present but empty\n", entry.Key)
		}
	}

	fmt.Println()
	printSummary(result)
}

func printSummary(result *differ.DiffResult) {
	bold.Println("Summary")
	green.Printf("  ✓ %d ok\n", result.OK)

	if result.Missing > 0 {
		red.Printf("  ✗ %d missing\n", result.Missing)
	}
	if result.Undocumented > 0 {
		yellow.Printf("  ⚠ %d undocumented\n", result.Undocumented)
	}
	if result.Empty > 0 {
		yellow.Printf("  ⚠ %d empty\n", result.Empty)
	}
	fmt.Println()

	if result.Missing > 0 {
		red.Println("✗ Check failed — run `envguard sync` to fill missing keys")
	} else {
		green.Println("✓ All required keys are present")
	}
	fmt.Println()
}

func PrintValidation(key, value string, ok bool) {
	if ok {
		green.Printf("  ✓ %-40s set\n", key)
	} else {
		red.Printf("  ✗ %-40s missing or empty\n", key)
	}
}

func PrintSyncResult(added []string, skipped []string) {
	bold.Println("\nSync result")
	for _, k := range added {
		green.Printf("  + %-40s added (empty value)\n", k)
	}
	for _, k := range skipped {
		faint.Printf("  ~ %-40s already exists, skipped\n", k)
	}
	fmt.Println()
}
