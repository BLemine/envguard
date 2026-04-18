package reporter

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/BLemine/envguard/internal/auditor"
	"github.com/BLemine/envguard/internal/differ"
	"github.com/fatih/color"
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
			red.Printf("  ✗ %-40s missing from %s\n", entry.Key, filepath.Base(localPath))
		case differ.StatusUndocumented:
			yellow.Printf("  ⚠ %-40s in %s but not in %s\n", entry.Key, filepath.Base(localPath), filepath.Base(examplePath))
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

	if result.Missing > 0 || result.Empty > 0 {
		red.Println("✗ Check failed — fix missing or empty keys before continuing")
	} else {
		green.Println("✓ All example keys are present and non-empty")
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

func PrintAuditResult(result *auditor.Result) {
	bold.Printf("\nAudit: scanning git history for secrets and .env files\n\n")

	if len(result.EnvFiles) > 0 {
		bold.Println(".env Files Committed to History")
		for _, hit := range result.EnvFiles {
			yellow.Printf("  ⚠ %s  %s\n", sha7(hit.CommitSHA), hit.File)
		}
		fmt.Println()
	}

	if len(result.Secrets) > 0 {
		bold.Println("Secret Patterns Detected")
		for _, hit := range result.Secrets {
			red.Printf("  ✗ %s  %-35s %s\n", sha7(hit.CommitSHA), hit.File, hit.Pattern)
			faint.Printf("         %s\n", hit.Line)
		}
		fmt.Println()
	}

	bold.Println("Summary")
	if len(result.EnvFiles) == 0 && len(result.Secrets) == 0 {
		green.Println("  ✓ No .env files or secret patterns found in git history")
		fmt.Println()
		green.Println("✓ Audit passed")
		fmt.Println()
		return
	}
	if len(result.EnvFiles) > 0 {
		yellow.Printf("  ⚠ %d .env file(s) found in history\n", len(result.EnvFiles))
	}
	if len(result.Secrets) > 0 {
		red.Printf("  ✗ %d secret pattern(s) detected\n", len(result.Secrets))
	}
	fmt.Println()
	red.Println("✗ Audit failed — sensitive data may be exposed in git history")
	faint.Println("  Tip: use `git filter-repo` or BFG Repo Cleaner to scrub history")
	fmt.Println()
}

func sha7(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
