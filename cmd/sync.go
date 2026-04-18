package cmd

import (
	"fmt"
	"os"

	"github.com/BLemine/envguard/internal/parser"
	"github.com/BLemine/envguard/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	syncExample string
	syncLocal   string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Add missing keys from .env.example into your .env",
	Long:  `Reads .env.example and adds any missing keys into your .env with empty values. Existing keys are never overwritten.`,
	Run: func(cmd *cobra.Command, args []string) {
		example, err := parser.Parse(syncExample)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", syncExample, err)
			os.Exit(1)
		}

		local, err := parser.Parse(syncLocal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", syncLocal, err)
			os.Exit(1)
		}

		localKeys := parser.KeySet(local)
		var added, skipped []string

		f, err := os.OpenFile(syncLocal, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s for writing: %v\n", syncLocal, err)
			os.Exit(1)
		}
		defer f.Close()

		for key := range example.Keys {
			if _, exists := localKeys[key]; exists {
				skipped = append(skipped, key)
			} else {
				if _, err := fmt.Fprintf(f, "%s=\n", key); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing key %s: %v\n", key, err)
					os.Exit(1)
				}
				added = append(added, key)
			}
		}

		reporter.PrintSyncResult(added, skipped)
	},
}

func init() {
	syncCmd.Flags().StringVar(&syncExample, "example", ".env.example", "Path to your .env.example file")
	syncCmd.Flags().StringVar(&syncLocal, "local", ".env", "Path to your local .env file")
}
