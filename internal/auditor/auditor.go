package auditor

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BLemine/envguard/internal/parser"
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

func Run(repoPath string, format parser.Format) (*Result, error) {
	result := &Result{}
	if err := findConfigFiles(repoPath, format, result); err != nil {
		return nil, err
	}
	if err := scanSecrets(repoPath, result); err != nil {
		return nil, err
	}
	return result, nil
}

func findConfigFiles(repoPath string, format parser.Format, result *Result) error {
	var currentSHA string
	return streamGitLines(repoPath, []string{"log", "--all", "--diff-filter=A", "--name-only", "--format=COMMIT:%H"}, func(line string) {
		if strings.HasPrefix(line, "COMMIT:") {
			currentSHA = strings.TrimPrefix(line, "COMMIT:")
			return
		}
		if line == "" || currentSHA == "" {
			return
		}
		if matchesConfigFile(line, format) {
			result.EnvFiles = append(result.EnvFiles, EnvFileHit{
				CommitSHA: currentSHA,
				File:      line,
			})
		}
	})
}

// matchesConfigFile returns true for config files likely to contain secrets for the given format.
func matchesConfigFile(path string, format parser.Format) bool {
	base := path
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		base = path[i+1:]
	}
	lower := strings.ToLower(base)
	switch format {
	case parser.FormatEnv:
		if !strings.HasPrefix(lower, ".env") {
			return false
		}
		for _, suffix := range []string{".example", ".sample", ".template"} {
			if strings.HasSuffix(lower, suffix) {
				return false
			}
		}
		return true
	case parser.FormatYAML:
		ext := strings.ToLower(filepath.Ext(lower))
		if ext != ".yml" && ext != ".yaml" {
			return false
		}
		return lower == "application.yml" || lower == "application.yaml" || lower == "docker-compose.yml" || lower == "docker-compose.yaml"
	case parser.FormatProps:
		return lower == "application.properties"
	default:
		return false
	}
}

func scanSecrets(repoPath string, result *Result) error {
	var currentSHA, currentFile string
	seen := make(map[string]bool)

	return streamGitLines(repoPath, []string{"log", "--all", "--no-merges", "-p", "--no-color", "--format=COMMIT:%H"}, func(line string) {
		if strings.HasPrefix(line, "COMMIT:") {
			currentSHA = strings.TrimPrefix(line, "COMMIT:")
			currentFile = ""
			return
		}
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			return
		}
		if len(line) == 0 || line[0] != '+' || strings.HasPrefix(line, "+++") {
			return
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
						Line:      truncate(redactSecretLine(strings.TrimSpace(content)), 80),
					})
				}
				break
			}
		}
	})
}

func streamGitLines(repoPath string, args []string, consume func(string)) error {
	cmdArgs := append([]string{"-C", repoPath}, args...)
	cmd := exec.Command("git", cmdArgs...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("starting git command: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting git command: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

	for scanner.Scan() {
		consume(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return fmt.Errorf("reading git output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("git %s failed: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

var assignmentSecretRE = regexp.MustCompile(`^(\s*[\w.-]+\s*[=:]\s*)(?:"[^"]*"|'[^']*'|[^\s#]+)(\s*(?:#.*)?)$`)

func redactSecretLine(line string) string {
	if matches := assignmentSecretRE.FindStringSubmatch(line); len(matches) == 3 {
		return matches[1] + "[REDACTED]" + matches[2]
	}
	if strings.Contains(line, "-----BEGIN ") && strings.Contains(line, "PRIVATE KEY-----") {
		return "[REDACTED PRIVATE KEY HEADER]"
	}
	return "[REDACTED]"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
