package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/BLemine/envguard/internal/parser"
	"github.com/BLemine/envguard/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	validateLocal    string
	validateRequired string
	validateFormat   string
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Assert that required keys are present and non-empty",
	Long:    `Checks that a specific set of keys exist in your .env and have non-empty values. Useful in CI pipelines.`,
	Example: `  envguard validate --required=DATABASE_URL,API_KEY,JWT_SECRET`,
	Run: func(cmd *cobra.Command, args []string) {
		if validateRequired == "" {
			fmt.Fprintln(os.Stderr, "Error: --required flag is required")
			os.Exit(1)
		}

		format, err := resolveFormat(validateFormat, validateLocal)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		requiredKeys := strings.Split(validateRequired, ",")
		failed, err := validateFile(validateLocal, format, requiredKeys)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if failed > 0 {
			fmt.Fprintf(os.Stderr, "✗ Validation failed — %d required key(s) missing or empty\n\n", failed)
			os.Exit(1)
		}

		fmt.Println("✓ All required keys are set")
	},
}

func validateFile(localPath string, format parser.Format, requiredKeys []string) (int, error) {
	local, err := parser.ParseWithFormat(localPath, format)
	if err != nil {
		return 0, formatErrorContext("error reading "+localPath, format, err)
	}

	failed := 0
	fmt.Println()
	for _, key := range requiredKeys {
		key = strings.TrimSpace(key)
		value, exists := local.Keys[key]
		ok := exists && value != ""
		reporter.PrintValidation(key, value, ok)
		if !ok {
			failed++
		}
	}
	fmt.Println()
	return failed, nil
}

func init() {
	validateCmd.Flags().StringVar(&validateLocal, "local", ".env", "Path to your local .env file")
	validateCmd.Flags().StringVar(&validateRequired, "required", "", "Comma-separated list of required keys")
	validateCmd.Flags().StringVar(&validateFormat, "format", "", "File format: env, yaml, or props (auto-detected when omitted)")
}
