package toolkit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	ManagedBegin   = "<!-- repository-knowledge:begin managed; toolkit replaces only this block -->"
	ManagedEnd     = "<!-- repository-knowledge:end managed -->"
	GeneratedBegin = "<!-- repository-knowledge:generated inventory; do not hand edit -->"
	GeneratedEnd   = "<!-- repository-knowledge:end generated inventory -->"
)

func utcNow() string {
	return time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
}

func readJSON[T any](path string, required bool) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		if !required && errors.Is(err, fs.ErrNotExist) {
			return value, nil
		}
		return value, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("parse JSON %s: %w", path, err)
	}
	return value, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON for %s: %w", path, err)
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data, 0o644)
}

func unmarshalJSON(data []byte, value any, label string) error {
	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("parse JSON %s: %w", label, err)
	}
	return nil
}

func writeFileAtomic(path string, data []byte, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", path, err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".repo-knowledge-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %s: %w", path, err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary file for %s: %w", path, err)
	}
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("set mode for %s: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary file for %s: %w", path, err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}

func repositoryPath(root, value, field string) (string, error) {
	if strings.TrimSpace(value) == "" || filepath.IsAbs(value) {
		return "", fmt.Errorf("%s must be a non-empty path inside the repository: %q", field, value)
	}
	clean := filepath.Clean(value)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s resolves outside the repository: %q", field, value)
	}
	rootAbsolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	rootResolved, err := filepath.EvalSymlinks(rootAbsolute)
	if err != nil {
		return "", fmt.Errorf("resolve repository root symlinks: %w", err)
	}
	candidate := filepath.Join(rootAbsolute, clean)
	ancestor := candidate
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("inspect %s: %w", ancestor, err)
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", fmt.Errorf("cannot find existing parent for %s", candidate)
		}
		ancestor = parent
	}
	ancestorResolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", ancestor, err)
	}
	relative, err := filepath.Rel(rootResolved, ancestorResolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s resolves outside the repository: %q", field, value)
	}
	return candidate, nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", fmt.Errorf("hash %s: %w", path, err)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func runGit(root string, check bool, arguments ...string) (string, error) {
	args := append([]string{"-C", root}, arguments...)
	command := exec.Command("git", args...)
	output, err := command.CombinedOutput()
	if err != nil && check {
		return "", fmt.Errorf("git %s failed: %s", strings.Join(arguments, " "), strings.TrimSpace(string(output)))
	}
	return string(output), err
}

func isGitRepository(root string) bool {
	output, err := runGit(root, false, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(output) == "true"
}

func currentCommit(root string) string {
	if !isGitRepository(root) {
		return ""
	}
	output, err := runGit(root, false, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(output)
}

func globMatch(pattern, value string) bool {
	var expression strings.Builder
	expression.WriteString("^")
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '*':
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				expression.WriteString(".*")
				index += 2
			} else {
				expression.WriteString("[^/]*")
				index++
			}
		case '?':
			expression.WriteString("[^/]")
			index++
		default:
			expression.WriteString(regexp.QuoteMeta(string(pattern[index])))
			index++
		}
	}
	expression.WriteString("$")
	matched, err := regexp.MatchString(expression.String(), filepath.ToSlash(value))
	return err == nil && matched
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if globMatch(pattern, filepath.ToSlash(path)) {
			return true
		}
	}
	return false
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
