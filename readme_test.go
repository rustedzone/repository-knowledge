package repositoryknowledge

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestREADMELeadsWithOutcomesAndEvidence(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	requiredOrder := []string{
		"Repository Knowledge helps coding agents avoid stale architectural assumptions and makes documentation-impact decisions visible.",
		"## 60-second hostile demonstration",
		"## Measured results",
		"## Three-command quickstart",
		"## Honest boundary",
		"## How it works",
		"## Detailed documentation",
	}
	last := -1
	for _, section := range requiredOrder {
		index := strings.Index(readme, section)
		if index < 0 {
			t.Errorf("README is missing %q", section)
			continue
		}
		if index <= last {
			t.Errorf("README section %q is out of order", section)
		}
		last = index
	}
	if !strings.Contains(readme, "does not guarantee agent compliance or understanding") {
		t.Error("README must state the product's non-guarantee")
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
	noResultsClaim := strings.Contains(readme, "no causal outcome-benchmark trial has been published yet") &&
		strings.Contains(readme, "No outcome results are currently available")
	if resultCount == 0 && !noResultsClaim {
		t.Error("README must disclose that no outcome results are published")
	}
	if resultCount > 0 && noResultsClaim {
		t.Error("README claims no outcome results although committed result JSON exists")
	}
}
