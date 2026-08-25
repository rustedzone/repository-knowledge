package repositoryknowledge

import (
	"encoding/json"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

func TestEmbeddedJSONAssetsParse(t *testing.T) {
	t.Parallel()
	for _, directory := range []string{"policy", "schemas", "templates"} {
		err := fs.WalkDir(Content, directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".json") {
				return nil
			}
			data, err := Content.ReadFile(path)
			if err != nil {
				t.Errorf("Content.ReadFile(%q): %v", path, err)
				return nil
			}
			var value any
			if err := json.Unmarshal(data, &value); err != nil {
				t.Errorf("embedded JSON %s: %v", path, err)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WalkDir(%q): %v", directory, err)
		}
	}
}

func TestVersionIsSemantic(t *testing.T) {
	t.Parallel()
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(Version()) {
		t.Fatalf("Version() = %q, want semantic version", Version())
	}
}

func TestAgentAdapterAssetsAreEmbedded(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"skills/repository-knowledge/SKILL.md",
		"skills/repository-knowledge/references/documentation-generation.md",
		"skills/repository-knowledge/references/documentation-quality.md",
		"skills/repository-knowledge/references/repository-type-profiles.md",
		"skills/repository-knowledge/references/source-of-truth.md",
		"skills/repository-knowledge/references/operations.md",
		"templates/AGENTS.md",
		"templates/repository-knowledge-rule.md",
		"templates/repository-knowledge-cursor-rule.mdc",
	} {
		if _, err := Content.ReadFile(path); err != nil {
			t.Errorf("Content.ReadFile(%q): %v", path, err)
		}
	}
}
