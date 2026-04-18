package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/envguard/internal/differ"
	"github.com/yourusername/envguard/internal/parser"
	"github.com/yourusername/envguard/internal/reporter"
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
		example, err := parser.Parse(checkExample)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", checkExample, err)
			os.Exit(1)
		}

		local, err := parser.Parse(checkLocal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", checkLocal, err)
			os.Exit(1)
		}

		result := differ.Diff(example, local)
		reporter.PrintDiff(result, checkExample, checkLocal)

		if result.Missing > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkExample, "example", ".env.example", "Path to your .env.example file")
	checkCmd.Flags().StringVar(&checkLocal, "local", ".env", "Path to your local .env file")
}
