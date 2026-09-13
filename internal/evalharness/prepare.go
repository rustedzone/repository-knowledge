package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
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
	condition, err := prepareCondition(spec.Family, options.Condition)
	if err != nil {
		return result, err
	}
	if options.TrialNumber == 0 {
		options.TrialNumber = 1
	}
	if options.TrialNumber < 1 {
		return result, fmt.Errorf("trial number must be at least 1")
	}
	if strings.TrimSpace(options.RepositoryKnowledgeRevision) == "" {
		options.RepositoryKnowledgeRevision = "v" + repositoryknowledge.Version()
	}
	if spec.Family == FamilyBenchmark {
		if strings.TrimSpace(options.AgentVersion) == "" {
			return result, fmt.Errorf("agent version is required for outcome benchmarks")
		}
		if strings.TrimSpace(options.ModelVersion) == "" {
			return result, fmt.Errorf("model version is required for outcome benchmarks")
		}
		if strings.TrimSpace(options.ReasoningConfiguration) == "" {
			return result, fmt.Errorf("reasoning configuration is required for outcome benchmarks")
		}
	} else {
		if strings.TrimSpace(options.AgentVersion) == "" {
			options.AgentVersion = "unrecorded"
		}
		if strings.TrimSpace(options.ModelVersion) == "" {
			options.ModelVersion = "unrecorded"
		}
		if strings.TrimSpace(options.ReasoningConfiguration) == "" {
			options.ReasoningConfiguration = "unrecorded"
		}
	}
	output, err := filepath.Abs(options.Output)
	if err != nil {
		return result, fmt.Errorf("resolve evaluation output: %w", err)
	}
	baselinePath := evaluationBaselinePath(output, spec.Family)
	if spec.Family == FamilyBenchmark {
		if _, err := os.Lstat(baselinePath); err == nil {
			return result, fmt.Errorf("evaluation baseline already exists: %s", baselinePath)
		} else if !os.IsNotExist(err) {
			return result, fmt.Errorf("inspect evaluation baseline: %w", err)
		}
	}
	if err := ensureEmptyOutput(output); err != nil {
		return result, err
	}
	fixture, err := resolveFixturePath(options.CasesRoot, caseRoot, spec.Fixture, spec.Family == FamilyBenchmark)
	if err != nil {
		return result, err
	}
	if err := copyTree(fixture, output); err != nil {
		return result, fmt.Errorf("copy fixture: %w", err)
	}
	if condition != ConditionControl {
		if _, err := toolkit.Install(toolkit.InstallOptions{
			Target: output, Source: "evaluation-harness", Ref: options.RepositoryKnowledgeRevision,
			AgentAdapters: []string{options.Agent},
		}); err != nil {
			return result, fmt.Errorf("install repository-knowledge into fixture: %w", err)
		}
	}
	files, err := hashRepository(output)
	if err != nil {
		return result, fmt.Errorf("hash evaluation baseline: %w", err)
	}
	baseline := Baseline{
		SchemaVersion: "2.0", Family: spec.Family, Condition: condition,
		CaseID: spec.ID, CaseRevision: spec.Revision, SourceCommit: spec.SourceCommit,
		Agent: options.Agent, AgentVersion: options.AgentVersion, ModelVersion: options.ModelVersion,
		ReasoningConfiguration:      options.ReasoningConfiguration,
		RepositoryKnowledgeRevision: options.RepositoryKnowledgeRevision,
		TrialNumber:                 options.TrialNumber, Files: files,
	}
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return result, err
	}
	data = append(data, '\n')
	if spec.Family == FamilyConformance {
		if err := os.MkdirAll(filepath.Dir(baselinePath), 0o755); err != nil {
			return result, fmt.Errorf("create evaluation metadata directory: %w", err)
		}
	}
	if err := writeExclusive(baselinePath, data); err != nil {
		return result, fmt.Errorf("write evaluation baseline: %w", err)
	}
	prompt, err := os.ReadFile(filepath.Join(caseRoot, spec.Prompt))
	if err != nil {
		return result, fmt.Errorf("read evaluation prompt: %w", err)
	}
	rubric := filepath.Join(caseRoot, spec.Rubric)
	if spec.Family == FamilyBenchmark {
		rubric = ""
	}
	return PrepareResult{
		Family: spec.Family, Condition: condition, CaseID: spec.ID, Revision: spec.Revision,
		SourceCommit: spec.SourceCommit, Target: output, Baseline: baselinePath,
		Agent: options.Agent, AgentVersion: options.AgentVersion,
		ModelVersion: options.ModelVersion, ReasoningConfiguration: options.ReasoningConfiguration,
		RepositoryKnowledgeRevision: options.RepositoryKnowledgeRevision, TrialNumber: options.TrialNumber,
		Prompt: strings.TrimSpace(string(prompt)), Rubric: rubric,
	}, nil
}

func evaluationBaselinePath(target, family string) string {
	if family == FamilyBenchmark {
		return filepath.Clean(target) + ".repo-knowledge-eval-baseline.json"
	}
	return filepath.Join(target, ".repo-knowledge", "eval-baseline.json")
}

func prepareCondition(family, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if family == FamilyConformance {
		if requested == "" || requested == ConditionConformance {
			return ConditionConformance, nil
		}
		return "", fmt.Errorf("conformance evaluations use condition %q", ConditionConformance)
	}
	if requested == ConditionControl || requested == ConditionTreatment {
		return requested, nil
	}
	return "", fmt.Errorf("outcome benchmarks require --condition control or --condition treatment")
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
