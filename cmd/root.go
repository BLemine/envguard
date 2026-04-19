package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envguard",
	Short: "envguard — keep your .env files honest",
	Long: `envguard helps you manage, validate, and audit .env files
across your projects. Never ship with missing or undocumented keys again.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(scanConfigCmd)
}
