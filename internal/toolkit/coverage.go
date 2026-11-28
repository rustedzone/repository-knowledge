package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const semanticCoveragePath = ".repo-knowledge/semantic-coverage.json"

type CoverageEvidence struct {
	Path   string `json:"path"`
	Symbol string `json:"symbol"`
	SHA256 string `json:"sha256,omitempty"`
}

type CoverageGap struct {
	Question  string              `json:"question"`
	Inspected []CoverageInspected `json:"inspected"`
}

type CoverageInspected struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type CoverageUnit struct {
	ID       string             `json:"id"`
	Status   string             `json:"status"`
	Guide    string             `json:"guide,omitempty"`
	Trace    string             `json:"trace,omitempty"`
	Evidence []CoverageEvidence `json:"evidence,omitempty"`
	Gaps     []CoverageGap      `json:"gaps,omitempty"`
}

type SemanticCoverage struct {
	SchemaVersion string         `json:"schema_version"`
	Scope         string         `json:"scope"`
	Units         []CoverageUnit `json:"units"`
}

type CoverageFinding struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type CoverageCheckResult struct {
	Status   string            `json:"status"`
	Scope    string            `json:"scope,omitempty"`
	Units    int               `json:"units"`
	Findings []CoverageFinding `json:"findings"`
}

type CoverageStampResult struct {
	Status  string `json:"status"`
	Path    string `json:"path"`
	Stamped int    `json:"stamped"`
}

func CheckCoverage(root string) (CoverageCheckResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CoverageCheckResult{}, fmt.Errorf("resolve repository root: %w", err)
	}
	report, exists, err := readSemanticCoverage(root)
	if err != nil {
		return CoverageCheckResult{}, err
	}
	if !exists {
		return CoverageCheckResult{Status: "not_assessed", Findings: []CoverageFinding{}}, nil
	}
	config, err := readJSON[RepositoryConfig](filepath.Join(root, ".repo-knowledge", "repository.json"), true)
	if err != nil {
		return CoverageCheckResult{}, err
	}
	result := CoverageCheckResult{Status: "pass", Scope: report.Scope, Units: len(report.Units), Findings: []CoverageFinding{}}
	add := func(id, message string) {
		result.Findings = append(result.Findings, CoverageFinding{ID: id, Message: message})
	}
	if report.SchemaVersion != "1.0" {
		add("schema-version", "semantic coverage schema_version must be 1.0")
	}
	if report.Scope != "full" && report.Scope != "partial" {
		add("scope", "semantic coverage scope must be full or partial")
	}
	if report.Units == nil {
		add("missing-units", "semantic coverage units must be an array")
	}
	if report.Scope == "partial" {
		add("partial-coverage", "report covers only a bounded portion of the repository")
	}
	if report.Scope == "full" && len(report.Units) == 0 {
		add("empty-full-report", "a full report needs at least one material unit")
	}
	index, err := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if err != nil {
		return CoverageCheckResult{}, err
	}
	docRoot := filepath.Dir(index)
	hashes := make(map[string]string)
	fileDigest := func(path string) (string, error) {
		if digest, found := hashes[path]; found {
			return digest, nil
		}
		digest, err := sha256File(path)
		if err == nil {
			hashes[path] = digest
		}
		return digest, err
	}
	seen := make(map[string]struct{}, len(report.Units))
	for _, unit := range report.Units {
		label := "unit " + unit.ID
		if strings.TrimSpace(unit.ID) == "" {
			add("unit-id", "coverage unit id is required")
		} else if _, duplicate := seen[unit.ID]; duplicate {
			add("duplicate-unit", label+" is duplicated")
		}
		seen[unit.ID] = struct{}{}
		if unit.Status == "partial" || unit.Status == "unassessed" || len(unit.Gaps) > 0 {
			add("partial-coverage", label+" has an untraced or uncertain behavior")
		}
		switch unit.Status {
		case "traced":
			if strings.TrimSpace(unit.Trace) == "" || strings.TrimSpace(unit.Guide) == "" || len(unit.Evidence) == 0 {
				add("incomplete-trace", label+" needs a named trace, guide, and evidence")
			}
		case "partial":
			if len(unit.Gaps) == 0 {
				add("missing-gap", label+" is partial but has no recorded gap")
			}
		case "unassessed":
			if report.Scope == "full" {
				add("unassessed-full-report", label+" cannot be unassessed in a full report")
			}
		default:
			add("unit-status", label+" has an unsupported status")
		}
		if unit.Guide != "" {
			guide := strings.SplitN(unit.Guide, "#", 2)[0]
			resolved, pathErr := repositoryPath(root, guide, "coverage guide")
			if pathErr != nil || !pathWithin(docRoot, resolved) || !regularFile(resolved) {
				add("invalid-guide", label+" guide must be an existing file under the documentation root: "+unit.Guide)
			}
		}
		for _, evidence := range unit.Evidence {
			resolved, pathErr := repositoryPath(root, evidence.Path, "coverage evidence")
			if pathErr != nil || !regularFile(resolved) {
				add("invalid-evidence", label+" evidence must be an existing repository file: "+evidence.Path)
				continue
			}
			if strings.TrimSpace(evidence.Symbol) == "" {
				add("missing-symbol", label+" evidence needs a symbol or exact evidence locator: "+evidence.Path)
			}
			if evidence.SHA256 == "" {
				add("unstamped-evidence", label+" evidence has no content fingerprint: "+evidence.Path)
				continue
			}
			digest, hashErr := fileDigest(resolved)
			if hashErr != nil {
				return CoverageCheckResult{}, hashErr
			}
			if digest != evidence.SHA256 {
				add("stale-evidence", label+" evidence changed since review: "+evidence.Path)
			}
		}
		for _, gap := range unit.Gaps {
			if strings.TrimSpace(gap.Question) == "" {
				add("empty-gap", label+" has a gap without a question")
			}
			if len(gap.Inspected) == 0 {
				add("uninspected-gap", label+" gap needs at least one inspected repository file")
			}
			for _, inspected := range gap.Inspected {
				resolved, pathErr := repositoryPath(root, inspected.Path, "inspected evidence")
				if pathErr != nil || !regularFile(resolved) {
					add("invalid-inspected", label+" gap references missing or unsafe evidence: "+inspected.Path)
					continue
				}
				if inspected.SHA256 == "" {
					add("unstamped-evidence", label+" gap evidence has no content fingerprint: "+inspected.Path)
					continue
				}
				digest, hashErr := fileDigest(resolved)
				if hashErr != nil {
					return CoverageCheckResult{}, hashErr
				}
				if digest != inspected.SHA256 {
					add("stale-evidence", label+" gap evidence changed since review: "+inspected.Path)
				}
			}
		}
	}
	if report.Scope == "full" {
		for _, capability := range config.Capabilities {
			if _, exists := seen[capability.ID]; !exists {
				add("missing-capability", "declared capability is absent from full report: "+capability.ID)
			}
		}
	}
	if len(result.Findings) > 0 {
		result.Status = "advisory"
	}
	return result, nil
}

func StampCoverage(root string) (CoverageStampResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CoverageStampResult{}, fmt.Errorf("resolve repository root: %w", err)
	}
	report, exists, err := readSemanticCoverage(root)
	if err != nil {
		return CoverageStampResult{}, err
	}
	if !exists {
		return CoverageStampResult{}, fmt.Errorf("semantic coverage report does not exist: %s", semanticCoveragePath)
	}
	checked, err := CheckCoverage(root)
	if err != nil {
		return CoverageStampResult{}, err
	}
	for _, finding := range checked.Findings {
		if finding.ID == "stale-evidence" {
			return CoverageStampResult{}, fmt.Errorf("review changed evidence and clear its old sha256 before stamping: %s", finding.Message)
		}
		if finding.ID != "unstamped-evidence" && finding.ID != "partial-coverage" {
			return CoverageStampResult{}, fmt.Errorf("repair coverage report before stamping: %s: %s", finding.ID, finding.Message)
		}
	}
	count := 0
	for i := range report.Units {
		for j := range report.Units[i].Evidence {
			evidence := &report.Units[i].Evidence[j]
			if evidence.SHA256 != "" {
				continue
			}
			path, pathErr := repositoryPath(root, evidence.Path, "coverage evidence")
			if pathErr != nil || !regularFile(path) {
				return CoverageStampResult{}, fmt.Errorf("coverage evidence must be an existing repository file: %q", evidence.Path)
			}
			evidence.SHA256, err = sha256File(path)
			if err != nil {
				return CoverageStampResult{}, err
			}
			count++
		}
		for j := range report.Units[i].Gaps {
			for k := range report.Units[i].Gaps[j].Inspected {
				inspected := &report.Units[i].Gaps[j].Inspected[k]
				if inspected.SHA256 != "" {
					continue
				}
				path, pathErr := repositoryPath(root, inspected.Path, "inspected evidence")
				if pathErr != nil || !regularFile(path) {
					return CoverageStampResult{}, fmt.Errorf("inspected evidence must be an existing repository file: %q", inspected.Path)
				}
				inspected.SHA256, err = sha256File(path)
				if err != nil {
					return CoverageStampResult{}, err
				}
				count++
			}
		}
	}
	path, err := repositoryPath(root, semanticCoveragePath, "semantic coverage report")
	if err != nil {
		return CoverageStampResult{}, err
	}
	if count > 0 {
		if err := writeJSON(path, report); err != nil {
			return CoverageStampResult{}, err
		}
	}
	return CoverageStampResult{Status: "stamped", Path: semanticCoveragePath, Stamped: count}, nil
}

func readSemanticCoverage(root string) (SemanticCoverage, bool, error) {
	path, err := repositoryPath(root, semanticCoveragePath, "semantic coverage report")
	if err != nil {
		return SemanticCoverage{}, false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) || os.IsNotExist(err) {
		return SemanticCoverage{}, false, nil
	}
	if err != nil {
		return SemanticCoverage{}, false, fmt.Errorf("read %s: %w", path, err)
	}
	var report SemanticCoverage
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return SemanticCoverage{}, false, fmt.Errorf("parse semantic coverage report: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return SemanticCoverage{}, false, fmt.Errorf("semantic coverage report contains trailing JSON")
	}
	return report, true, nil
}
