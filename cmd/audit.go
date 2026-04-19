package cmd

import (
	"fmt"
	"os"

	"github.com/BLemine/envguard/internal/auditor"
	"github.com/BLemine/envguard/internal/reporter"
	"github.com/spf13/cobra"
)

var auditRepo string
var auditFormat string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Scan git history for committed .env files or secrets",
	Long:  `Walks the full git history and flags any .env files that were committed or lines matching common secret patterns (API keys, tokens, passwords, private keys).`,
	Example: `  envguard audit
  envguard audit --repo /path/to/other/repo
  envguard audit --format=yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		format, err := resolveFormat(auditFormat)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		result, err := auditor.Run(auditRepo, format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		reporter.PrintAuditResult(result)
		if len(result.EnvFiles) > 0 || len(result.Secrets) > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	auditCmd.Flags().StringVar(&auditRepo, "repo", ".", "Path to the git repository to audit")
	auditCmd.Flags().StringVar(&auditFormat, "format", "", "Config format to audit: env, yaml, or props")
}
