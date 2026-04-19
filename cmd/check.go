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
	checkFormat  string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Compare your .env against .env.example",
	Long:  `Diffs your local .env file against .env.example and reports missing, undocumented, or empty keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		format, err := resolveFormat(checkFormat, checkExample, checkLocal)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		result, err := runCheck(checkExample, checkLocal, format)
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

func runCheck(examplePath, localPath string, format parser.Format) (*differ.DiffResult, error) {
	example, err := parser.ParseWithFormat(examplePath, format)
	if err != nil {
		return nil, formatErrorContext("error reading "+examplePath, format, err)
	}

	local, err := parser.ParseWithFormat(localPath, format)
	if err != nil {
		return nil, formatErrorContext("error reading "+localPath, format, err)
	}

	return differ.Diff(example, local), nil
}

func shouldFailCheck(result *differ.DiffResult) bool {
	return result.Missing > 0 || result.Empty > 0
}

func init() {
	checkCmd.Flags().StringVar(&checkExample, "example", ".env.example", "Path to your .env.example file")
	checkCmd.Flags().StringVar(&checkLocal, "local", ".env", "Path to your local .env file")
	checkCmd.Flags().StringVar(&checkFormat, "format", "", "File format: env, yaml, or props (auto-detected when omitted)")
}
