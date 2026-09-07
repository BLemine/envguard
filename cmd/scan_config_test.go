package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanConfigSubprocess(t *testing.T) {
	if os.Getenv("ENVGUARD_SCAN_TEST_PROCESS") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			rootCmd.SetArgs(os.Args[i+1:])
			Execute()
			os.Exit(0)
		}
	}
	os.Exit(2)
}

func TestNestedScanCLI(t *testing.T) {
	for _, format := range []string{"yml", "properties"} {
		t.Run(format, func(t *testing.T) {
			dir := t.TempDir()
			config, local := filepath.Join(dir, "app."+format), filepath.Join(dir, ".env")
			content := "url: ${PRIMARY:${SECONDARY:${THIRD}}}\n"
			if format == "properties" {
				content = "url=${PRIMARY:${SECONDARY:${THIRD}}}\n"
			}
			if err := os.WriteFile(config, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			for _, tt := range []struct {
				env  string
				fail bool
			}{{"", true}, {"PRIMARY=\n", false}, {"SECONDARY=ok\n", false}, {"THIRD=ok\n", false}} {
				if err := os.WriteFile(local, []byte(tt.env), 0600); err != nil {
					t.Fatal(err)
				}
				for _, flag := range []string{"--local", "--example"} {
					cmd := exec.Command(os.Args[0], "-test.run=^TestScanConfigSubprocess$", "--", "scan-config", "--files", config, flag, local, "--strict", "--json")
					cmd.Env = append(os.Environ(), "ENVGUARD_SCAN_TEST_PROCESS=1")
					out, err := cmd.CombinedOutput()
					if (err != nil) != tt.fail {
						t.Fatalf("%s %q: %v\n%s", flag, tt.env, err, out)
					}
					var result scanJSONOutput
					if err := json.Unmarshal(out, &result); err != nil {
						t.Fatalf("invalid JSON: %v\n%s", err, out)
					}
					if len(result.Optional) != 1 || len(result.Optional[0].Fallback) != 1 || len(result.Optional[0].Fallback[0].Fallback) != 1 {
						t.Fatalf("missing JSON fallback tree: %s", out)
					}
					if tt.fail && !strings.Contains(string(out), "${PRIMARY:${SECONDARY:${THIRD}}}") {
						t.Fatalf("missing expression diagnostic: %s", out)
					}
				}
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestScanConfigSubprocess$", "--", "scan-config", "--files", config, "--quiet")
			cmd.Env = append(os.Environ(), "ENVGUARD_SCAN_TEST_PROCESS=1")
			out, err := cmd.CombinedOutput()
			if err != nil || string(out) != "PRIMARY\nSECONDARY\nTHIRD\n" {
				t.Fatalf("quiet output: %v %s", err, out)
			}
		})
	}
}
