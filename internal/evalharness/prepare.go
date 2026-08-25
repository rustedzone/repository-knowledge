package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rustedzone/repository-knowledge/internal/toolkit"
)

func Prepare(options PrepareOptions) (PrepareResult, error) {
	var result PrepareResult
	spec, caseRoot, err := LoadSpec(options.CasesRoot, options.CaseID)
	if err != nil {
		return result, err
	}
	if strings.TrimSpace(options.Agent) == "" {
		options.Agent = "codex"
	}
	output, err := filepath.Abs(options.Output)
	if err != nil {
		return result, fmt.Errorf("resolve evaluation output: %w", err)
	}
	if err := ensureEmptyOutput(output); err != nil {
		return result, err
	}
	if err := copyTree(filepath.Join(caseRoot, spec.Fixture), output); err != nil {
		return result, fmt.Errorf("copy fixture: %w", err)
	}
	if _, err := toolkit.Install(toolkit.InstallOptions{
		Target: output, Source: "evaluation-harness", Ref: "case-" + spec.ID + "-r" + spec.Revision,
		AgentAdapters: []string{options.Agent},
	}); err != nil {
		return result, fmt.Errorf("install repository-knowledge into fixture: %w", err)
	}
	files, err := hashRepository(output)
	if err != nil {
		return result, fmt.Errorf("hash evaluation baseline: %w", err)
	}
	baseline := Baseline{SchemaVersion: "1.0", CaseID: spec.ID, CaseRevision: spec.Revision, Files: files}
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return result, err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(output, ".repo-knowledge", "eval-baseline.json"), data, 0o644); err != nil {
		return result, fmt.Errorf("write evaluation baseline: %w", err)
	}
	prompt, err := os.ReadFile(filepath.Join(caseRoot, spec.Prompt))
	if err != nil {
		return result, fmt.Errorf("read evaluation prompt: %w", err)
	}
	return PrepareResult{
		CaseID: spec.ID, Revision: spec.Revision, Target: output, Agent: options.Agent,
		Prompt: strings.TrimSpace(string(prompt)), Rubric: filepath.Join(caseRoot, spec.Rubric),
	}, nil
}

func ensureEmptyOutput(path string) error {
	entries, err := os.ReadDir(path)
	if err == nil {
		if len(entries) != 0 {
			return fmt.Errorf("evaluation output must be empty: %s", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect evaluation output: %w", err)
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create evaluation output: %w", err)
	}
	return nil
}
