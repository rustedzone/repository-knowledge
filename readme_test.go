package repositoryknowledge

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestREADMEExplainsUsageEvidenceAndLimits(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	required := []string{
		"# Repository Knowledge",
		"repo-knowledge install --target .",
		"repo-knowledge doctor --target .",
		"evals/benchmarks/frontend-onboarding/eval.json",
		"policy/contract.json",
		"docs/installation.md",
		"documentation-impact",
	}
	for _, claim := range required {
		if !strings.Contains(readme, claim) {
			t.Errorf("README is missing %q", claim)
		}
	}
	if !regexp.MustCompile(`(?i)\b(?:cannot|does not) guarantee\b`).MatchString(readme) {
		t.Error("README must state the product's non-guarantee")
	}
	firstSection := strings.SplitN(readme, "## ", 2)[0]
	if !strings.Contains(firstSection, "Evidence status:") {
		t.Error("README must disclose outcome-evidence status before the detailed sections")
	}
}

func TestREADMEEvidenceStatusMatchesCommittedResults(t *testing.T) {
	t.Parallel()
	resultCount := 0
	err := filepath.WalkDir("evals/results", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && filepath.Ext(path) == ".json" {
			resultCount++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	noResultsClaim := regexp.MustCompile(`(?i)\bno outcome results\b.{0,40}\b(?:published|available)\b`).MatchString(readme)
	if resultCount == 0 && !noResultsClaim {
		t.Error("README must disclose that no outcome results are published")
	}
	if resultCount > 0 && noResultsClaim {
		t.Error("README claims no outcome results although committed result JSON exists")
	}
}
