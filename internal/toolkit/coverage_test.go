package toolkit

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCoverageIsOptionalAndInstallPreservesReport(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(root, ".repo-knowledge", "schemas", "semantic-coverage.schema.json"))
	result, err := CheckCoverage(root)
	if err != nil || result.Status != "not_assessed" {
		t.Fatalf("missing report result = %+v, err = %v", result, err)
	}
	path := filepath.Join(root, semanticCoveragePath)
	content := []byte("{\"schema_version\":\"1.0\",\"scope\":\"partial\",\"units\":[]}\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	before, _, err := preflightDigests(root)
	if err != nil {
		t.Fatal(err)
	}
	content = []byte("{ \"schema_version\": \"1.0\", \"scope\": \"partial\", \"units\": [] }\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	afterEdit, _, err := preflightDigests(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, afterEdit) {
		t.Fatalf("editing coverage report changed preflight digests: before=%v after=%v", before, afterEdit)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true}); err != nil {
		t.Fatal(err)
	}
	after, _, err := preflightDigests(root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("coverage report changed preflight digests: before=%v after=%v", before, after)
	}
	actual, err := os.ReadFile(path)
	if err != nil || string(actual) != string(content) {
		t.Fatalf("update replaced consumer coverage: %q, err=%v", actual, err)
	}
}

func TestCoverageStampAndCheckDetectStaleEvidence(t *testing.T) {
	root := coverageFixture(t)
	result, err := CheckCoverage(root)
	if err != nil || !coverageFinding(result, "unstamped-evidence") {
		t.Fatalf("unstamped result = %+v, err=%v", result, err)
	}
	stamp, err := StampCoverage(root)
	if err != nil || stamp.Stamped != 1 {
		t.Fatalf("stamp = %+v, err=%v", stamp, err)
	}
	result, err = CheckCoverage(root)
	if err != nil || result.Status != "pass" {
		t.Fatalf("checked result = %+v, err=%v", result, err)
	}
	mustWrite(t, filepath.Join(root, "internal", "approval.go"), "package internal\nfunc Approve() { panic(\"changed\") }\n")
	result, err = CheckCoverage(root)
	if err != nil || !coverageFinding(result, "stale-evidence") {
		t.Fatalf("stale result = %+v, err=%v", result, err)
	}
	if _, err := StampCoverage(root); err == nil || !strings.Contains(err.Error(), "clear its old sha256") {
		t.Fatalf("stale stamp error = %v, want explicit review", err)
	}
}

func TestPartialCoverageIsAdvisory(t *testing.T) {
	root := coverageFixture(t)
	path := filepath.Join(root, semanticCoveragePath)
	report, err := readJSON[SemanticCoverage](path, true)
	if err != nil {
		t.Fatal(err)
	}
	report.Scope = "partial"
	report.Units[0].Status = "partial"
	report.Units[0].Gaps = []CoverageGap{{Question: "Retry behavior is not established", Inspected: []CoverageInspected{{Path: "internal/approval.go"}}}}
	if err := writeJSON(path, report); err != nil {
		t.Fatal(err)
	}
	if _, err := StampCoverage(root); err != nil {
		t.Fatal(err)
	}
	result, err := CheckCoverage(root)
	if err != nil || result.Status != "advisory" || !coverageFinding(result, "partial-coverage") {
		t.Fatalf("partial report = %+v, err=%v", result, err)
	}
}

func TestCoverageCheckRejectsFalseFullAndUnsafeReferences(t *testing.T) {
	root := coverageFixture(t)
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config, err := readJSON[RepositoryConfig](configPath, true)
	if err != nil {
		t.Fatal(err)
	}
	config.Capabilities = append(config.Capabilities, Capability{ID: "missing", Documentation: []string{"docs/index.md"}})
	if err := writeJSON(configPath, config); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(root, semanticCoveragePath)
	report, err := readJSON[SemanticCoverage](reportPath, true)
	if err != nil {
		t.Fatal(err)
	}
	report.Units = append(report.Units, report.Units[0])
	report.Units[1].Guide = "../outside.md"
	report.Units[1].Evidence[0].Path = "../outside.go"
	if err := writeJSON(reportPath, report); err != nil {
		t.Fatal(err)
	}
	result, err := CheckCoverage(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{"duplicate-unit", "invalid-guide", "invalid-evidence", "missing-capability"} {
		if !coverageFinding(result, wanted) {
			t.Fatalf("missing %s in %+v", wanted, result)
		}
	}
	if _, err := StampCoverage(root); err == nil || !strings.Contains(err.Error(), "repair coverage report") {
		t.Fatalf("StampCoverage() error = %v, want refusal", err)
	}
}

func TestCoverageReportCannotSatisfyDocumentationImpact(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(root, true, "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := runGit(root, true, "commit", "-m", "baseline"); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "internal", "new.go"), "package internal\n")
	mustWrite(t, filepath.Join(root, semanticCoveragePath), "{\"schema_version\":\"1.0\",\"scope\":\"partial\",\"units\":[]}\n")
	impact, err := Impact(root, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(impact.DocumentationChanges) != 0 || len(impact.IgnoredChanges) == 0 {
		t.Fatalf("coverage report classification = %+v", impact)
	}
	validated, err := ValidateImpact(root, "", "", "acknowledgment")
	if err != nil || validated.Status != "fail" {
		t.Fatalf("code + coverage-only impact = %+v, err=%v", validated, err)
	}
}

func coverageFixture(t *testing.T) string {
	t.Helper()
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, Source: "test", Ref: "v0.1.0"}); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "docs", "approval.md"), "# Approval\n\n## Request trace\n")
	mustWrite(t, filepath.Join(root, "internal", "approval.go"), "package internal\nfunc Approve() {}\n")
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config, err := readJSON[RepositoryConfig](configPath, true)
	if err != nil {
		t.Fatal(err)
	}
	config.Capabilities = []Capability{{ID: "approval", Documentation: []string{"docs/approval.md"}}}
	if err := writeJSON(configPath, config); err != nil {
		t.Fatal(err)
	}
	report := SemanticCoverage{SchemaVersion: "1.0", Scope: "full", Units: []CoverageUnit{{
		ID: "approval", Status: "traced", Guide: "docs/approval.md#request-trace", Trace: "Approve request",
		Evidence: []CoverageEvidence{{Path: "internal/approval.go", Symbol: "Approve"}},
	}}}
	if err := writeJSON(filepath.Join(root, semanticCoveragePath), report); err != nil {
		t.Fatal(err)
	}
	return root
}

func coverageFinding(result CoverageCheckResult, id string) bool {
	for _, finding := range result.Findings {
		if finding.ID == id {
			return true
		}
	}
	return false
}
