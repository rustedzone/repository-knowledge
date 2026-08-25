package evalharness

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("fixture symlinks are not allowed: %s", path)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err != nil {
			_ = input.Close()
			return err
		}
		if _, err := io.Copy(output, input); err != nil {
			_ = input.Close()
			_ = output.Close()
			return err
		}
		if err := input.Close(); err != nil {
			_ = output.Close()
			return err
		}
		return output.Close()
	})
}

func regularFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relative))
		return nil
	})
	sort.Strings(files)
	return files, err
}

func matchingFiles(root, pattern string) ([]string, error) {
	files, err := regularFiles(root)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, file := range files {
		if globMatch(pattern, file) {
			matches = append(matches, file)
		}
	}
	return matches, nil
}

func globMatch(pattern, value string) bool {
	pattern = filepath.ToSlash(pattern)
	value = filepath.ToSlash(value)
	var expression strings.Builder
	expression.WriteString("^")
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '*':
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				if index+2 < len(pattern) && pattern[index+2] == '/' {
					expression.WriteString("(?:.*/)?")
					index += 3
				} else {
					expression.WriteString(".*")
					index += 2
				}
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
	return regexp.MustCompile(expression.String()).MatchString(value)
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func hashRepository(root string) (map[string]string, error) {
	files, err := regularFiles(root)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(files))
	for _, relative := range files {
		if relative == ".repo-knowledge/eval-baseline.json" {
			continue
		}
		digest, err := hashFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		result[relative] = digest
	}
	return result, nil
}

func concatenateFiles(root string, files []string) (string, error) {
	var content strings.Builder
	for _, relative := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return "", err
		}
		content.Write(data)
		content.WriteByte('\n')
	}
	return content.String(), nil
}
