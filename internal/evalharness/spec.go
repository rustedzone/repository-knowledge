package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var fullSourceCommit = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

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
	if spec.Family == "" {
		spec.Family = FamilyConformance
	}
	if spec.ID != caseID {
		return spec, "", fmt.Errorf("evaluation case directory %q contains id %q", caseID, spec.ID)
	}
	if err := validateSpec(spec, caseRoot, casesRoot); err != nil {
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

func validateSpec(spec Spec, caseRoot, casesRoot string) error {
	if spec.SchemaVersion != "1.0" {
		return fmt.Errorf("schema_version must be 1.0")
	}
	if spec.Family != FamilyConformance && spec.Family != FamilyBenchmark {
		return fmt.Errorf("family must be %s or %s", FamilyConformance, FamilyBenchmark)
	}
	if strings.TrimSpace(spec.ID) == "" || strings.TrimSpace(spec.Revision) == "" {
		return fmt.Errorf("id and revision are required")
	}
	for label, value := range map[string]string{"prompt": spec.Prompt, "rubric": spec.Rubric} {
		if value == "" || filepath.IsAbs(value) || strings.HasPrefix(filepath.Clean(value), "..") {
			return fmt.Errorf("%s must be a relative path inside the case", label)
		}
		if _, err := os.Stat(filepath.Join(caseRoot, value)); err != nil {
			return fmt.Errorf("%s %q is unavailable: %w", label, value, err)
		}
	}
	fixture, err := resolveFixturePath(casesRoot, caseRoot, spec.Fixture, spec.Family == FamilyBenchmark)
	if err != nil {
		return err
	}
	if info, err := os.Stat(fixture); err != nil || !info.IsDir() {
		return fmt.Errorf("fixture %q is unavailable or is not a directory", spec.Fixture)
	}
	for _, pattern := range spec.AllowedChanges {
		if pattern == "" || filepath.IsAbs(pattern) || strings.HasPrefix(filepath.Clean(pattern), "..") {
			return fmt.Errorf("allowed_changes entries must stay inside the prepared repository: %q", pattern)
		}
	}
	if spec.Family == FamilyBenchmark {
		if !fullSourceCommit.MatchString(spec.SourceCommit) {
			return fmt.Errorf("outcome benchmark source_commit must be a full Git commit SHA")
		}
		if len(spec.ExpectedBehavioralTrace) == 0 {
			return fmt.Errorf("outcome benchmark expected_behavioral_trace is required")
		}
		for _, step := range spec.ExpectedBehavioralTrace {
			if strings.TrimSpace(step) == "" {
				return fmt.Errorf("expected_behavioral_trace entries cannot be empty")
			}
		}
		prompt, err := os.ReadFile(filepath.Join(caseRoot, spec.Prompt))
		if err != nil {
			return err
		}
		if mentionsRepositoryKnowledge(string(prompt)) {
			return fmt.Errorf("outcome benchmark prompt must be neutral and must not mention Repository Knowledge")
		}
		rubric, err := os.ReadFile(filepath.Join(caseRoot, spec.Rubric))
		if err != nil {
			return err
		}
		lowerRubric := strings.ToLower(string(rubric))
		if mentionsRepositoryKnowledge(string(rubric)) || strings.Contains(lowerRubric, "control condition") || strings.Contains(lowerRubric, "treatment condition") {
			return fmt.Errorf("outcome benchmark rubric must be blind to the evaluated condition")
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

func resolveFixturePath(casesRoot, caseRoot, value string, allowShared bool) (string, error) {
	if value == "" || filepath.IsAbs(value) {
		return "", fmt.Errorf("fixture must be a relative path")
	}
	fixture, err := filepath.Abs(filepath.Join(caseRoot, value))
	if err != nil {
		return "", fmt.Errorf("resolve fixture: %w", err)
	}
	allowedRoot := caseRoot
	if allowShared {
		absoluteCasesRoot, err := filepath.Abs(casesRoot)
		if err != nil {
			return "", fmt.Errorf("resolve evaluation family root: %w", err)
		}
		allowedRoot = filepath.Dir(absoluteCasesRoot)
	}
	relative, err := filepath.Rel(allowedRoot, fixture)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("fixture must stay inside %s", allowedRoot)
	}
	return fixture, nil
}

func mentionsRepositoryKnowledge(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "repository-knowledge") ||
		strings.Contains(lower, "repo-knowledge") ||
		strings.Contains(lower, "repository knowledge")
}

func supportedCheckType(value string) bool {
	switch value {
	case "paths_exist", "minimum_files", "changed_files_minimum", "contains_all", "not_contains", "source_examples_minimum", "existing_evidence_minimum", "capabilities_minimum", "markdown_links_valid", "protected_files_unchanged":
		return true
	default:
		return false
	}
}
