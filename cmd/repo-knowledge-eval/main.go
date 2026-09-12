package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
	"github.com/rustedzone/repository-knowledge/internal/evalharness"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		printUsage(stderr)
		return 2
	}
	switch arguments[0] {
	case "list":
		flags := flag.NewFlagSet("list", flag.ContinueOnError)
		flags.SetOutput(stderr)
		family := flags.String("family", "conformance", "conformance or benchmark")
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		benchmarksRoot := flags.String("benchmarks", "evals/benchmarks", "outcome benchmark cases directory")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		root, err := familyRoot(*family, *casesRoot, *benchmarksRoot)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		specs, err := evalharness.ListCases(root)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if *jsonOutput {
			return encodeJSON(stdout, specs)
		}
		for _, spec := range specs {
			fmt.Fprintf(stdout, "%s [%s] (revision %s): %s\n", spec.ID, spec.Family, spec.Revision, spec.Description)
		}
		return 0
	case "prepare":
		flags := flag.NewFlagSet("prepare", flag.ContinueOnError)
		flags.SetOutput(stderr)
		family := flags.String("family", "conformance", "conformance or benchmark")
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		benchmarksRoot := flags.String("benchmarks", "evals/benchmarks", "outcome benchmark cases directory")
		caseID := flags.String("case", "", "evaluation case id")
		output := flags.String("output", "", "empty destination directory")
		condition := flags.String("condition", "", "benchmark condition: control or treatment")
		agent := flags.String("agent", "codex", "codex, claude-code, antigravity-ide, or cursor")
		agentVersion := flags.String("agent-version", "", "exact agent host/version used for the trial")
		modelVersion := flags.String("model-version", "", "exact agent model/version used for the trial")
		reasoning := flags.String("reasoning", "", "reasoning configuration used for the trial")
		toolkitRevision := flags.String("repository-knowledge-revision", "v"+repositoryknowledge.Version(), "Repository Knowledge version or commit")
		trial := flags.Int("trial", 1, "positive trial number")
		preflightContext := flags.String("preflight-context", "", "treatment preflight context: full or compact (default: full)")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		if *caseID == "" || *output == "" {
			fmt.Fprintln(stderr, "error: --case and --output are required")
			return 2
		}
		root, err := familyRoot(*family, *casesRoot, *benchmarksRoot)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		result, err := evalharness.Prepare(evalharness.PrepareOptions{
			CasesRoot: root, CaseID: *caseID, Output: *output, Condition: *condition, Agent: *agent,
			AgentVersion: *agentVersion, ModelVersion: *modelVersion, ReasoningConfiguration: *reasoning,
			RepositoryKnowledgeRevision: *toolkitRevision, TrialNumber: *trial, PreflightContext: *preflightContext,
		})
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if *jsonOutput {
			return encodeJSON(stdout, result)
		}
		fmt.Fprintf(stdout, "prepared %s revision %s at %s for %s (%s, trial %d)\n", result.CaseID, result.Revision, result.Target, result.Agent, result.Condition, result.TrialNumber)
		if result.Rubric != "" {
			fmt.Fprintf(stdout, "semantic rubric: %s\n", result.Rubric)
		}
		fmt.Fprintf(stdout, "\nprompt:\n%s\n", result.Prompt)
		return 0
	case "grade":
		flags := flag.NewFlagSet("grade", flag.ContinueOnError)
		flags.SetOutput(stderr)
		family := flags.String("family", "conformance", "conformance or benchmark")
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		benchmarksRoot := flags.String("benchmarks", "evals/benchmarks", "outcome benchmark cases directory")
		caseID := flags.String("case", "", "evaluation case id")
		target := flags.String("target", "", "prepared evaluation repository")
		duration := flags.String("duration", "", "trial duration, such as 12m30s")
		tokens := flags.Int64("tokens", -1, "total token usage when available")
		tokenSource := flags.String("token-source", "", "usage source: codex, claude-code, antigravity-ide, cursor, manual, or unavailable")
		inputTokens := flags.Int64("input-tokens", -1, "provider-reported input tokens")
		outputTokens := flags.Int64("output-tokens", -1, "provider-reported output tokens")
		cachedTokens := flags.Int64("cached-tokens", -1, "provider-reported cached input tokens")
		totalTokens := flags.Int64("total-tokens", -1, "provider-reported total tokens")
		semanticStatus := flags.String("semantic-status", "", "explicit semantic status: pass or fail")
		semanticEarned := flags.Int("semantic-score", -1, "semantic rubric points earned")
		semanticAvailable := flags.Int("semantic-available", -1, "semantic rubric points available")
		reviewer := flags.String("reviewer", "", "semantic reviewer identity")
		resultsRoot := flags.String("results", "", "write an immutable result below this directory")
		runDate := flags.String("run-date", "", "result date in YYYY-MM-DD format")
		artifact := flags.String("artifact", "", "patch or output artifact to preserve with a result")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		if *caseID == "" || *target == "" {
			fmt.Fprintln(stderr, "error: --case and --target are required")
			return 2
		}
		root, err := familyRoot(*family, *casesRoot, *benchmarksRoot)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		var durationMillis int64
		if *duration != "" {
			parsed, err := time.ParseDuration(*duration)
			if err != nil || parsed < 0 {
				fmt.Fprintln(stderr, "error: --duration must be a non-negative Go duration")
				return 2
			}
			durationMillis = parsed.Milliseconds()
		}
		var tokenUsage *int64
		if *tokens >= 0 {
			tokenUsage = tokens
		}
		var usage *evalharness.TokenUsageDetails
		if strings.TrimSpace(*tokenSource) != "" || *inputTokens >= 0 || *outputTokens >= 0 || *cachedTokens >= 0 || *totalTokens >= 0 {
			usage = &evalharness.TokenUsageDetails{Source: *tokenSource}
			usage.InputTokens = nonNegativeInt64(inputTokens)
			usage.OutputTokens = nonNegativeInt64(outputTokens)
			usage.CachedTokens = nonNegativeInt64(cachedTokens)
			usage.TotalTokens = nonNegativeInt64(totalTokens)
			if usage.TotalTokens == nil && tokenUsage != nil {
				usage.TotalTokens = tokenUsage
			}
		}
		var score *evalharness.SemanticScore
		if *semanticEarned >= 0 || *semanticAvailable >= 0 {
			score = &evalharness.SemanticScore{Earned: *semanticEarned, Available: *semanticAvailable}
		}
		result, err := evalharness.Grade(evalharness.GradeOptions{
			CasesRoot: root, CaseID: *caseID, Target: *target, DurationMillis: durationMillis,
			TokenUsage: tokenUsage, Usage: usage, SemanticStatus: *semanticStatus, SemanticScore: score, SemanticReviewer: *reviewer,
		})
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		var output any = result
		var recorded evalharness.RecordResult
		if *resultsRoot != "" {
			if *runDate == "" || *artifact == "" || *duration == "" {
				fmt.Fprintln(stderr, "error: --results requires --run-date, --duration, and --artifact")
				return 2
			}
			recorded, err = evalharness.Record(evalharness.RecordOptions{ResultsRoot: *resultsRoot, RunDate: *runDate, Artifact: *artifact}, result)
			if err != nil {
				fmt.Fprintln(stderr, "error:", err)
				return 2
			}
			result = recorded.Result
			output = recorded
		}
		if *jsonOutput {
			if code := encodeJSON(stdout, output); code != 0 {
				return code
			}
		} else {
			fmt.Fprintf(stdout, "deterministic grade: %s (%d passed, %d failed)\n", result.DeterministicStatus, result.Passed, result.Failed)
			for _, check := range result.Checks {
				fmt.Fprintf(stdout, "[%s] %s: %s\n", strings.ToUpper(check.Status), check.ID, check.Detail)
			}
			fmt.Fprintf(stdout, "semantic grade: %s\nrubric: %s\noverall: %s\n", result.SemanticStatus, result.Rubric, result.OverallStatus)
			if recorded.ResultPath != "" {
				fmt.Fprintf(stdout, "recorded result: %s\npreserved artifact: %s\n", recorded.ResultPath, recorded.ArtifactPath)
			}
		}
		if result.DeterministicStatus == "fail" {
			return 1
		}
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n", arguments[0])
		return 2
	}
}

func nonNegativeInt64(value *int64) *int64 {
	if value == nil || *value < 0 {
		return nil
	}
	return value
}

func encodeJSON(output io.Writer, value any) int {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return 2
	}
	return 0
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage: repo-knowledge-eval <list|prepare|grade> [options]\nFamilies: conformance (default) or benchmark")
}

func familyRoot(family, casesRoot, benchmarksRoot string) (string, error) {
	switch strings.TrimSpace(family) {
	case "conformance":
		return casesRoot, nil
	case "benchmark", "outcome_benchmark":
		return benchmarksRoot, nil
	default:
		return "", fmt.Errorf("--family must be conformance or benchmark")
	}
}
