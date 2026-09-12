package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rustedzone/repository-knowledge/internal/evalharness"
)

func TestGradeCLIRecordsSourceAttributedUsage(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	if _, err := evalharness.Prepare(evalharness.PrepareOptions{
		CasesRoot: filepath.Join("..", "..", "evals", "benchmarks"), CaseID: "frontend-onboarding", Output: target,
		Condition: evalharness.ConditionControl, Agent: "codex", AgentVersion: "test", ModelVersion: "test",
		ReasoningConfiguration: "test", RepositoryKnowledgeRevision: "test", TrialNumber: 1,
	}); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"grade", "--family", "benchmark", "--benchmarks", filepath.Join("..", "..", "evals", "benchmarks"),
		"--case", "frontend-onboarding", "--target", target, "--token-source", "codex",
		"--input-tokens", "700", "--output-tokens", "300", "--cached-tokens", "200", "--total-tokens", "1000", "--json",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("grade code=%d, want deterministic failure; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"source": "codex"`) || !strings.Contains(stdout.String(), `"total_tokens": 1000`) {
		t.Fatalf("grade output omitted usage: %q", stdout.String())
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareCLIAcceptsTreatmentPreflightContext(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"prepare", "--family", "benchmark", "--benchmarks", filepath.Join("..", "..", "evals", "benchmarks"),
		"--case", "frontend-onboarding", "--condition", "treatment", "--output", target,
		"--agent", "codex", "--agent-version", "test", "--model-version", "test", "--reasoning", "test",
		"--repository-knowledge-revision", "v0.11.1", "--trial", "1", "--preflight-context", "compact", "--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("prepare code=%d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"preflight_context": "compact"`) || !strings.Contains(stdout.String(), `"hook_payload"`) {
		t.Fatalf("prepare output omitted profile metadata: %q", stdout.String())
	}
}

func TestPrepareCLIRejectsControlPreflightContextBeforeWriting(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"prepare", "--family", "benchmark", "--benchmarks", filepath.Join("..", "..", "evals", "benchmarks"),
		"--case", "frontend-onboarding", "--condition", "control", "--output", target,
		"--agent", "codex", "--agent-version", "test", "--model-version", "test", "--reasoning", "test",
		"--repository-knowledge-revision", "v0.11.1", "--trial", "1", "--preflight-context", "full",
	}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "only valid for treatment") {
		t.Fatalf("prepare code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("invalid control profile wrote target: %v", err)
	}
}
