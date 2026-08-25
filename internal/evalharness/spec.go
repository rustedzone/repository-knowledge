package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadSpec(casesRoot, caseID string) (Spec, string, error) {
	var spec Spec
	if strings.TrimSpace(caseID) == "" || filepath.Base(caseID) != caseID {
		return spec, "", fmt.Errorf("case must be a simple identifier: %q", caseID)
	}
	caseRoot, err := filepath.Abs(filepath.Join(casesRoot, caseID))
	if err != nil {
		return spec, "", fmt.Errorf("resolve case root: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(caseRoot, "eval.json"))
	if err != nil {
		return spec, "", fmt.Errorf("read evaluation case %q: %w", caseID, err)
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return spec, "", fmt.Errorf("parse evaluation case %q: %w", caseID, err)
	}
	if spec.ID != caseID {
		return spec, "", fmt.Errorf("evaluation case directory %q contains id %q", caseID, spec.ID)
	}
	if err := validateSpec(spec, caseRoot); err != nil {
		return spec, "", fmt.Errorf("invalid evaluation case %q: %w", caseID, err)
	}
	return spec, caseRoot, nil
}

func ListCases(casesRoot string) ([]Spec, error) {
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		return nil, fmt.Errorf("list evaluation cases: %w", err)
	}
	var specs []Spec
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		spec, _, err := LoadSpec(casesRoot, entry.Name())
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].ID < specs[j].ID })
	return specs, nil
}

func validateSpec(spec Spec, caseRoot string) error {
	if spec.SchemaVersion != "1.0" {
		return fmt.Errorf("schema_version must be 1.0")
	}
	if strings.TrimSpace(spec.ID) == "" || strings.TrimSpace(spec.Revision) == "" {
		return fmt.Errorf("id and revision are required")
	}
	for label, value := range map[string]string{"fixture": spec.Fixture, "prompt": spec.Prompt, "rubric": spec.Rubric} {
		if value == "" || filepath.IsAbs(value) || strings.HasPrefix(filepath.Clean(value), "..") {
			return fmt.Errorf("%s must be a relative path inside the case", label)
		}
		if _, err := os.Stat(filepath.Join(caseRoot, value)); err != nil {
			return fmt.Errorf("%s %q is unavailable: %w", label, value, err)
		}
	}
	if len(spec.Checks) == 0 {
		return fmt.Errorf("at least one deterministic check is required")
	}
	seen := make(map[string]struct{})
	for _, check := range spec.Checks {
		if check.ID == "" || check.Type == "" {
			return fmt.Errorf("every check requires id and type")
		}
		if _, exists := seen[check.ID]; exists {
			return fmt.Errorf("duplicate check id %q", check.ID)
		}
		seen[check.ID] = struct{}{}
		if !supportedCheckType(check.Type) {
			return fmt.Errorf("check %q uses unsupported type %q", check.ID, check.Type)
		}
	}
	return nil
}

func supportedCheckType(value string) bool {
	switch value {
	case "paths_exist", "minimum_files", "changed_files_minimum", "contains_all", "not_contains", "source_examples_minimum", "existing_evidence_minimum", "capabilities_minimum", "markdown_links_valid", "protected_files_unchanged":
		return true
	default:
		return false
	}
}
