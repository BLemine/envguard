package auditor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactSecretLineMasksValue(t *testing.T) {
	t.Parallel()

	got := redactSecretLine(`API_KEY=supersecretapikeyvalue1234567890`)

	if got == `API_KEY=supersecretapikeyvalue1234567890` {
		t.Fatalf("secret value was not redacted: %q", got)
	}
	if got != `API_KEY=[REDACTED]` {
		t.Fatalf("unexpected redacted line: %q", got)
	}
}

func TestRedactSecretLineHandlesQuotedSecret(t *testing.T) {
	t.Parallel()

	got := redactSecretLine(`token = "abc123456789"`)

	if got != `token = [REDACTED]` {
		t.Fatalf("unexpected redacted line: %q", got)
	}
}

func TestAuditFindsSecretsIntroducedByMerge(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+filepath.Join(repo, "no-global-config"))
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
		return strings.TrimSpace(string(output))
	}
	write := func(name, value string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(repo, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-b", "main")
	git("config", "user.name", "Test")
	git("config", "user.email", "test@example.invalid")
	write("base.txt", "base")
	git("add", ".")
	git("commit", "-m", "base")
	git("checkout", "-b", "feature")
	write("feature.txt", "feature")
	git("add", ".")
	git("commit", "-m", "feature")
	git("checkout", "main")
	write("main.txt", "main")
	git("add", ".")
	git("commit", "-m", "main")
	git("merge", "--no-ff", "--no-commit", "feature")
	write("credentials.txt", "API_KEY=fixturevalue12345678901234567890\n")
	write(".env", "VALUE=fixture\n")
	git("add", ".")
	git("commit", "-m", "merge")
	sha := git("rev-parse", "HEAD")
	result, err := Run(repo)
	if err != nil {
		t.Fatal(err)
	}
	foundSecret, foundEnv := false, false
	for _, hit := range result.Secrets {
		if strings.Contains(hit.Line, "fixturevalue") {
			t.Fatal("secret leaked in output")
		}
		if hit.CommitSHA == sha && hit.File == "credentials.txt" {
			foundSecret = true
		}
	}
	for _, hit := range result.EnvFiles {
		if hit.CommitSHA == sha && hit.File == ".env" {
			foundEnv = true
		}
	}
	if !foundSecret || !foundEnv {
		t.Fatalf("merge findings missing: %#v", result)
	}
}
