package auditor

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type EnvFileHit struct {
	CommitSHA string
	File      string
}

type SecretHit struct {
	CommitSHA string
	File      string
	Pattern   string
	Line      string
}

type Result struct {
	EnvFiles []EnvFileHit
	Secrets  []SecretHit
}

var patterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"AWS Access Key ID", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"Private Key Header", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"GitHub Token", regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`)},
	{"Slack Token", regexp.MustCompile(`xox[baprs]-[0-9A-Za-z\-]{10,48}`)},
	{"Stripe Key", regexp.MustCompile(`(?:sk|pk)_(?:live|test)_[0-9a-zA-Z]{24,}`)},
	{"Generic API Key", regexp.MustCompile(`(?i)(?:api[_-]?key|apikey)\s*[=:]\s*["']?[a-zA-Z0-9_\-]{20,}["']?`)},
	{"Generic Secret/Token/Password", regexp.MustCompile(`(?i)(?:secret|token|password|passwd)\s*[=:]\s*["']?[^\s"'<>{}\[\]]{8,}["']?`)},
}

func Run(repoPath string) (*Result, error) {
	result := &Result{}
	if err := findEnvFiles(repoPath, result); err != nil {
		return nil, err
	}
	if err := scanSecrets(repoPath, result); err != nil {
		return nil, err
	}
	return result, nil
}

func findEnvFiles(repoPath string, result *Result) error {
	cmd := exec.Command("git", "-C", repoPath, "log", "--all", "--diff-filter=A", "--name-only", "--format=COMMIT:%H")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git log failed: %w", err)
	}

	var currentSHA string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMIT:") {
			currentSHA = strings.TrimPrefix(line, "COMMIT:")
			continue
		}
		if line == "" || currentSHA == "" {
			continue
		}
		if matchesEnvFile(line) {
			result.EnvFiles = append(result.EnvFiles, EnvFileHit{
				CommitSHA: currentSHA,
				File:      line,
			})
		}
	}
	return scanner.Err()
}

// matchesEnvFile returns true for .env files that likely contain real secrets
// (.env, .env.local, .env.production) but not templates (.env.example, .env.local.example).
func matchesEnvFile(path string) bool {
	base := path
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		base = path[i+1:]
	}
	lower := strings.ToLower(base)
	if !strings.HasPrefix(lower, ".env") {
		return false
	}
	for _, suffix := range []string{".example", ".sample", ".template"} {
		if strings.HasSuffix(lower, suffix) {
			return false
		}
	}
	return true
}

func scanSecrets(repoPath string, result *Result) error {
	cmd := exec.Command("git", "-C", repoPath, "log", "--all", "--no-merges", "-p", "--no-color", "--format=COMMIT:%H")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git log -p failed: %w", err)
	}

	var currentSHA, currentFile string
	seen := make(map[string]bool)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "COMMIT:") {
			currentSHA = strings.TrimPrefix(line, "COMMIT:")
			currentFile = ""
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			continue
		}
		if len(line) == 0 || line[0] != '+' || strings.HasPrefix(line, "+++") {
			continue
		}

		content := line[1:]
		for _, p := range patterns {
			if p.re.MatchString(content) {
				key := currentSHA + "|" + currentFile + "|" + content
				if !seen[key] {
					seen[key] = true
					result.Secrets = append(result.Secrets, SecretHit{
						CommitSHA: currentSHA,
						File:      currentFile,
						Pattern:   p.name,
						Line:      truncate(strings.TrimSpace(content), 80),
					})
				}
				break
			}
		}
	}
	return scanner.Err()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
