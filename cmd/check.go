package cmd

import (
	"fmt"
	"os"

	"github.com/BLemine/envguard/internal/differ"
	"github.com/BLemine/envguard/internal/parser"
	"github.com/BLemine/envguard/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	checkExample string
	checkLocal   string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Compare your .env against .env.example",
	Long:  `Diffs your local .env file against .env.example and reports missing, undocumented, or empty keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := runCheck(checkExample, checkLocal)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		reporter.PrintDiff(result, checkExample, checkLocal)

		if shouldFailCheck(result) {
			os.Exit(1)
		}
	},
}

func runCheck(examplePath, localPath string) (*differ.DiffResult, error) {
	example, err := parser.Parse(examplePath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", examplePath, err)
	}

	local, err := parser.Parse(localPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", localPath, err)
	}

	return differ.Diff(example, local), nil
}

func shouldFailCheck(result *differ.DiffResult) bool {
	return result.Missing > 0 || result.Empty > 0
}

func init() {
	checkCmd.Flags().StringVar(&checkExample, "example", ".env.example", "Path to your .env.example file")
	checkCmd.Flags().StringVar(&checkLocal, "local", ".env", "Path to your local .env file")
}
