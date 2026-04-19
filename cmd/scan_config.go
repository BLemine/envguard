package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BLemine/envguard/internal/configscan"
	"github.com/BLemine/envguard/internal/parser"
	"github.com/spf13/cobra"
)

var (
	scanFiles   string
	scanFormat  string
	scanJSON    bool
	scanQuiet   bool
	scanStrict  bool
	scanLocal   string
	scanExample string
)

var scanConfigCmd = &cobra.Command{
	Use:   "scan-config",
	Short: "Scan Spring Boot / Quarkus config files for env variable placeholders",
	Long: `Extracts ${VAR} and ${VAR:default} placeholders from YAML and .properties config files.

Supports Spring Boot application.yml / Quarkus application.properties and similar formats.
Use --strict with --local / --example to verify all required variables are defined.`,
	Run: func(cmd *cobra.Command, args []string) {
		if scanFiles == "" {
			fmt.Fprintln(os.Stderr, "error: --files is required")
			os.Exit(1)
		}

		files := strings.Split(scanFiles, ",")
		merged := &configscan.ScanResult{}

		for _, f := range files {
			f = strings.TrimSpace(f)
			format, err := resolveFormat(f, scanFormat)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}

			var result *configscan.ScanResult
			switch format {
			case "yaml":
				result, err = configscan.ScanYAML(f)
			case "props":
				result, err = configscan.ScanProps(f)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}

			merged.Required = append(merged.Required, result.Required...)
			merged.Optional = append(merged.Optional, result.Optional...)
		}

		var missingLocal, missingExample []string

		if scanLocal != "" {
			localEnv, err := parser.Parse(scanLocal)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error reading", scanLocal+":", err)
				os.Exit(1)
			}
			localKeys := parser.KeySet(localEnv)
			for _, p := range merged.Required {
				if _, ok := localKeys[p.Name]; !ok {
					missingLocal = append(missingLocal, p.Name)
				}
			}
		}

		if scanExample != "" {
			exampleEnv, err := parser.Parse(scanExample)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error reading", scanExample+":", err)
				os.Exit(1)
			}
			exampleKeys := parser.KeySet(exampleEnv)
			for _, p := range merged.Required {
				if _, ok := exampleKeys[p.Name]; !ok {
					missingExample = append(missingExample, p.Name)
				}
			}
		}

		switch {
		case scanJSON:
			printScanJSON(merged, missingLocal, missingExample)
		case scanQuiet:
			printScanQuiet(merged)
		default:
			printScanHuman(files, merged, missingLocal, missingExample)
		}

		if scanStrict && (len(missingLocal) > 0 || len(missingExample) > 0) {
			os.Exit(1)
		}
	},
}

func resolveFormat(filePath, formatFlag string) (string, error) {
	if formatFlag != "auto" && formatFlag != "" {
		switch formatFlag {
		case "yaml", "props":
			return formatFlag, nil
		default:
			return "", fmt.Errorf("unknown format %q: use yaml or props", formatFlag)
		}
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yml", ".yaml":
		return "yaml", nil
	case ".properties":
		return "props", nil
	default:
		return "", fmt.Errorf("cannot auto-detect format for %q: use --format yaml or --format props", filePath)
	}
}

func printScanHuman(files []string, result *configscan.ScanResult, missingLocal, missingExample []string) {
	hasComparison := scanLocal != "" || scanExample != ""

	if hasComparison {
		fmt.Printf("Required variables found in config: %d\n", len(result.Required))
		fmt.Printf("Optional variables found in config: %d\n", len(result.Optional))
	} else {
		for _, f := range files {
			fmt.Printf("Scanned: %s\n", strings.TrimSpace(f))
		}
	}
	fmt.Println()

	if len(result.Required) > 0 {
		fmt.Println("Required variables:")
		for _, p := range result.Required {
			fmt.Printf("  - %-28s (%s)\n", p.Name, p.SourcePath)
		}
		fmt.Println()
	}

	if len(result.Optional) > 0 {
		fmt.Println("Optional variables with defaults:")
		for _, p := range result.Optional {
			dv := ""
			if p.DefaultValue != nil {
				dv = *p.DefaultValue
			}
			fmt.Printf("  - %s=%s    (%s)\n", p.Name, dv, p.SourcePath)
		}
		fmt.Println()
	}

	if scanLocal != "" && len(missingLocal) > 0 {
		fmt.Printf("Missing in %s:\n", scanLocal)
		for _, name := range missingLocal {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()
	}

	if scanExample != "" && len(missingExample) > 0 {
		fmt.Printf("Missing in %s:\n", scanExample)
		for _, name := range missingExample {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Println()
	}
}

func printScanQuiet(result *configscan.ScanResult) {
	for _, p := range result.Required {
		fmt.Println(p.Name)
	}
	for _, p := range result.Optional {
		fmt.Println(p.Name)
	}
}

type scanJSONOutput struct {
	Required       []jsonPlaceholder `json:"required"`
	Optional       []jsonPlaceholder `json:"optional"`
	MissingLocal   []string          `json:"missing_local,omitempty"`
	MissingExample []string          `json:"missing_example,omitempty"`
}

type jsonPlaceholder struct {
	Name         string  `json:"name"`
	DefaultValue *string `json:"default,omitempty"`
	SourceFile   string  `json:"source_file"`
	SourcePath   string  `json:"source_path"`
}

func printScanJSON(result *configscan.ScanResult, missingLocal, missingExample []string) {
	out := scanJSONOutput{
		Required: make([]jsonPlaceholder, 0, len(result.Required)),
		Optional: make([]jsonPlaceholder, 0, len(result.Optional)),
	}
	for _, p := range result.Required {
		out.Required = append(out.Required, jsonPlaceholder{
			Name:       p.Name,
			SourceFile: p.SourceFile,
			SourcePath: p.SourcePath,
		})
	}
	for _, p := range result.Optional {
		out.Optional = append(out.Optional, jsonPlaceholder{
			Name:         p.Name,
			DefaultValue: p.DefaultValue,
			SourceFile:   p.SourceFile,
			SourcePath:   p.SourcePath,
		})
	}
	if len(missingLocal) > 0 {
		out.MissingLocal = missingLocal
	}
	if len(missingExample) > 0 {
		out.MissingExample = missingExample
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func init() {
	scanConfigCmd.Flags().StringVar(&scanFiles, "files", "", "Comma-separated list of config files to scan")
	scanConfigCmd.Flags().StringVar(&scanFormat, "format", "auto", "Format override: auto|yaml|props")
	scanConfigCmd.Flags().BoolVar(&scanJSON, "json", false, "Output as JSON")
	scanConfigCmd.Flags().BoolVar(&scanQuiet, "quiet", false, "Output only variable names")
	scanConfigCmd.Flags().BoolVar(&scanStrict, "strict", false, "Exit non-zero if required placeholders are not satisfied by --local or --example")
	scanConfigCmd.Flags().StringVar(&scanLocal, "local", "", "Path to local .env file for comparison")
	scanConfigCmd.Flags().StringVar(&scanExample, "example", "", "Path to .env.example file for comparison")
}
