package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/BLemine/envguard/internal/configscan"
	"github.com/BLemine/envguard/internal/parser"
	"github.com/spf13/cobra"
)

type scanConfigOptions struct {
	Files   []string
	Format  string
	JSON    bool
	Quiet   bool
	Strict  bool
	Local   string
	Example string
}

type scanConfigResult struct {
	Files          []string                  `json:"files"`
	Required       []configscan.Placeholder  `json:"required"`
	Optional       []configscan.Placeholder  `json:"optional"`
	MissingLocal   []string                  `json:"missingLocal,omitempty"`
	MissingExample []string                  `json:"missingExample,omitempty"`
	UnusedLocal    []string                  `json:"unusedLocal,omitempty"`
	UnusedExample  []string                  `json:"unusedExample,omitempty"`
}

var (
	scanConfigFiles   string
	scanConfigFormat  string
	scanConfigJSON    bool
	scanConfigQuiet   bool
	scanConfigStrict  bool
	scanConfigLocal   string
	scanConfigExample string
)

var scanConfigCmd = &cobra.Command{
	Use:   "scan-config",
	Short: "Scan Spring Boot or Quarkus config files for env placeholders",
	Long:  `Scans YAML and .properties config files for ${VAR} and ${VAR:default} placeholders without treating those files as .env sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(scanConfigFiles) == "" {
			fmt.Fprintln(os.Stderr, "Error: --files flag is required")
			os.Exit(1)
		}

		result, err := runScanConfig(scanConfigOptions{
			Files:   splitCommaSeparated(scanConfigFiles),
			Format:  scanConfigFormat,
			JSON:    scanConfigJSON,
			Quiet:   scanConfigQuiet,
			Strict:  scanConfigStrict,
			Local:   scanConfigLocal,
			Example: scanConfigExample,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := printScanConfigResult(result, scanConfigOptions{
			JSON:  scanConfigJSON,
			Quiet: scanConfigQuiet,
			Local: scanConfigLocal,
			Example: scanConfigExample,
		}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if shouldFailScanConfig(result, scanConfigStrict) {
			os.Exit(1)
		}
	},
}

func runScanConfig(opts scanConfigOptions) (*scanConfigResult, error) {
	if len(opts.Files) == 0 {
		return nil, fmt.Errorf("at least one config file must be provided")
	}

	results := make([]*configscan.ScanResult, 0, len(opts.Files))
	files := make([]string, 0, len(opts.Files))
	for _, file := range opts.Files {
		if file == "" {
			continue
		}
		scanResult, err := configscan.ScanFile(file, opts.Format)
		if err != nil {
			return nil, err
		}
		results = append(results, scanResult)
		files = append(files, file)
	}

	merged := configscan.Merge(results...)
	out := &scanConfigResult{
		Files:    files,
		Required: merged.Required,
		Optional: merged.Optional,
	}

	requiredNames := placeholderNames(merged.Required)
	if opts.Local != "" {
		local, err := parser.Parse(opts.Local)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", opts.Local, err)
		}
		out.MissingLocal = missingRequired(requiredNames, local)
		out.UnusedLocal = unusedEnv(requiredNames, local)
	}
	if opts.Example != "" {
		example, err := parser.Parse(opts.Example)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", opts.Example, err)
		}
		out.MissingExample = missingRequired(requiredNames, example)
		out.UnusedExample = unusedEnv(requiredNames, example)
	}

	return out, nil
}

func placeholderNames(placeholders []configscan.Placeholder) []string {
	seen := make(map[string]struct{}, len(placeholders))
	names := make([]string, 0, len(placeholders))
	for _, placeholder := range placeholders {
		if _, ok := seen[placeholder.Name]; ok {
			continue
		}
		seen[placeholder.Name] = struct{}{}
		names = append(names, placeholder.Name)
	}
	sort.Strings(names)
	return names
}

func missingRequired(requiredNames []string, env *parser.EnvFile) []string {
	missing := make([]string, 0)
	for _, name := range requiredNames {
		if _, ok := env.Keys[name]; ok {
			continue
		}
		missing = append(missing, name)
	}
	return missing
}

func unusedEnv(requiredNames []string, env *parser.EnvFile) []string {
	requiredSet := make(map[string]struct{}, len(requiredNames))
	for _, name := range requiredNames {
		requiredSet[name] = struct{}{}
	}

	unused := make([]string, 0)
	for key := range env.Keys {
		if _, ok := requiredSet[key]; ok {
			continue
		}
		unused = append(unused, key)
	}
	sort.Strings(unused)
	return unused
}

func printScanConfigResult(result *scanConfigResult, opts scanConfigOptions) error {
	if opts.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	if opts.Quiet {
		printQuietNames(result)
		return nil
	}

	fmt.Printf("Scanned: %s\n\n", strings.Join(result.Files, ", "))
	fmt.Printf("Required variables found in config: %d\n", len(result.Required))
	fmt.Printf("Optional variables found in config: %d\n\n", len(result.Optional))

	printPlaceholderSection("Required variables", result.Required)
	fmt.Println()
	printPlaceholderSection("Optional variables with defaults", result.Optional)

	if opts.Local != "" {
		fmt.Println()
		printNameSection("Missing in "+opts.Local, result.MissingLocal)
		if len(result.UnusedLocal) > 0 {
			fmt.Println()
			printNameSection("Unused in "+opts.Local, result.UnusedLocal)
		}
	}

	if opts.Example != "" {
		fmt.Println()
		printNameSection("Missing in "+opts.Example, result.MissingExample)
		if len(result.UnusedExample) > 0 {
			fmt.Println()
			printNameSection("Unused in "+opts.Example, result.UnusedExample)
		}
	}

	return nil
}

func printQuietNames(result *scanConfigResult) {
	names := append(placeholderNames(result.Required), placeholderNames(result.Optional)...)
	sort.Strings(names)

	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		fmt.Println(name)
	}
}

func printPlaceholderSection(title string, placeholders []configscan.Placeholder) {
	fmt.Printf("%s:\n", title)
	if len(placeholders) == 0 {
		fmt.Println("  (none)")
		return
	}

	for _, placeholder := range placeholders {
		if placeholder.DefaultValue != nil {
			fmt.Printf("  - %s=%s", placeholder.Name, *placeholder.DefaultValue)
		} else {
			fmt.Printf("  - %s", placeholder.Name)
		}
		if placeholder.SourcePath != "" {
			fmt.Printf(" (%s)", placeholder.SourcePath)
		}
		fmt.Println()
	}
}

func printNameSection(title string, names []string) {
	fmt.Printf("%s:\n", title)
	if len(names) == 0 {
		fmt.Println("  (none)")
		return
	}
	for _, name := range names {
		fmt.Printf("  - %s\n", name)
	}
}

func shouldFailScanConfig(result *scanConfigResult, strict bool) bool {
	if !strict {
		return false
	}
	return len(result.MissingLocal) > 0 || len(result.MissingExample) > 0
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func init() {
	rootCmd.AddCommand(scanConfigCmd)

	scanConfigCmd.Flags().StringVar(&scanConfigFiles, "files", "", "Comma-separated list of config files to scan")
	scanConfigCmd.Flags().StringVar(&scanConfigFormat, "format", "auto", "Config format: auto|yaml|props")
	scanConfigCmd.Flags().BoolVar(&scanConfigJSON, "json", false, "Output results as JSON")
	scanConfigCmd.Flags().BoolVar(&scanConfigQuiet, "quiet", false, "Output only variable names")
	scanConfigCmd.Flags().BoolVar(&scanConfigStrict, "strict", false, "Exit non-zero if required placeholders are missing from provided env files")
	scanConfigCmd.Flags().StringVar(&scanConfigLocal, "local", "", "Optional path to a local .env file for strict comparison")
	scanConfigCmd.Flags().StringVar(&scanConfigExample, "example", "", "Optional path to a .env.example file for strict comparison")
}
