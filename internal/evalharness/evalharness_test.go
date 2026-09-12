package evalharness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEvaluationSpecsLoad(t *testing.T) {
	t.Parallel()
	specs, err := ListCases(testCasesRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("len(specs) = %d, want 2", len(specs))
	}
	if specs[0].ID != "backend-clean-architecture" || specs[1].ID != "frontend-nextjs" {
		t.Fatalf("unexpected cases: %q, %q", specs[0].ID, specs[1].ID)
	}
}

func TestOutcomeBenchmarkSpecIsNeutralAndComplete(t *testing.T) {
	t.Parallel()
	specs, err := ListCases(testBenchmarksRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 || specs[0].ID != "frontend-onboarding" || specs[1].ID != "scoped-bugfix-plan" {
		t.Fatalf("unexpected benchmark specs: %+v", specs)
	}
	for _, spec := range specs {
		if spec.Family != FamilyBenchmark || spec.SourceCommit == "" || len(spec.ExpectedBehavioralTrace) == 0 {
			t.Fatalf("incomplete benchmark spec: %+v", spec)
		}
		prompt, err := os.ReadFile(filepath.Join(testBenchmarksRoot(t), spec.ID, spec.Prompt))
		if err != nil {
			t.Fatal(err)
		}
		if mentionsRepositoryKnowledge(string(prompt)) {
			t.Fatalf("benchmark prompt is not neutral: %s", prompt)
		}
	}
}

func TestOutcomeBenchmarkTreatmentPreflightContextSurvivesPrepareGradeAndRecord(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	options := benchmarkPrepareOptions(target, ConditionTreatment, "codex", 4)
	options.PreflightContext = PreflightContextCompact
	prepared, err := Prepare(options)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.PreflightContext != PreflightContextCompact || prepared.HookPayload == nil {
		t.Fatalf("prepared profile metadata = %+v", prepared)
	}
	if prepared.HookPayload.Profile != PreflightContextCompact || prepared.HookPayload.Bytes < 1 || prepared.HookPayload.Characters < 1 || prepared.HookPayload.GenerationMillis < 0 {
		t.Fatalf("prepared hook payload metrics = %+v", prepared.HookPayload)
	}

	data, err := os.ReadFile(prepared.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if baseline.PreflightContext != PreflightContextCompact || baseline.HookPayload == nil || baseline.HookPayload.Profile != PreflightContextCompact {
		t.Fatalf("baseline profile metadata = %+v", baseline)
	}

	grade, err := Grade(GradeOptions{
		CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target, DurationMillis: 2500,
		Usage: &TokenUsageDetails{
			Source: "codex", InputTokens: int64Pointer(100), OutputTokens: int64Pointer(10), TotalTokens: int64Pointer(110),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if grade.PreflightContext != PreflightContextCompact || grade.HookPayload == nil || grade.HookPayload.Bytes != prepared.HookPayload.Bytes {
		t.Fatalf("grade profile metadata = %+v", grade)
	}
	if grade.Usage == nil || grade.Usage.InputTokens == nil || *grade.Usage.InputTokens != 100 {
		t.Fatalf("provider usage was not preserved separately: %+v", grade.Usage)
	}

	artifact := filepath.Join(t.TempDir(), "output.patch")
	if err := os.WriteFile(artifact, []byte("benchmark output\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resultsRoot := filepath.Join(t.TempDir(), "results")
	recorded, err := Record(RecordOptions{ResultsRoot: resultsRoot, RunDate: "2026-09-12", Artifact: artifact}, grade)
	if err != nil {
		t.Fatal(err)
	}
	if recorded.Result.PreflightContext != PreflightContextCompact || recorded.Result.HookPayload == nil || recorded.Result.HookPayload.Profile != PreflightContextCompact {
		t.Fatalf("recorded profile metadata = %+v", recorded.Result)
	}
	if !strings.Contains(filepath.Base(recorded.ResultPath), "-treatment-compact-4.json") {
		t.Fatalf("compact result path does not identify profile: %s", recorded.ResultPath)
	}
	recordedData, err := os.ReadFile(recorded.ResultPath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted GradeResult
	if err := json.Unmarshal(recordedData, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.PreflightContext != PreflightContextCompact || persisted.HookPayload == nil || persisted.Usage == nil || persisted.Usage.InputTokens == nil || *persisted.Usage.InputTokens != 100 {
		t.Fatalf("persisted profile and measurement metadata = %+v", persisted)
	}

	fullGrade := grade
	fullGrade.PreflightContext = PreflightContextFull
	fullGrade.HookPayload = &HookPayloadMetrics{Profile: PreflightContextFull, Bytes: 1, Characters: 1}
	fullRecorded, err := Record(RecordOptions{ResultsRoot: resultsRoot, RunDate: "2026-09-12", Artifact: artifact}, fullGrade)
	if err != nil {
		t.Fatalf("full and compact profiles collided: %v", err)
	}
	if fullRecorded.ResultPath == recorded.ResultPath || !strings.Contains(filepath.Base(fullRecorded.ResultPath), "-treatment-4.json") {
		t.Fatalf("full result path collided or lost compatibility: %s", fullRecorded.ResultPath)
	}
}

func TestOutcomeBenchmarkPreflightContextValidationHappensBeforeOutput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		profile   string
		want      string
	}{
		{name: "control rejects profile", condition: ConditionControl, profile: PreflightContextFull, want: "only valid for treatment"},
		{name: "treatment rejects unknown profile", condition: ConditionTreatment, profile: "minimal", want: "full or compact"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			options := benchmarkPrepareOptions(target, test.condition, "codex", 1)
			options.PreflightContext = test.profile
			if _, err := Prepare(options); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Prepare() error = %v, want %q", err, test.want)
			}
			if _, err := os.Stat(target); !os.IsNotExist(err) {
				t.Fatalf("invalid profile wrote target: %v", err)
			}
			if _, err := os.Stat(evaluationBaselinePath(target, FamilyBenchmark)); !os.IsNotExist(err) {
				t.Fatalf("invalid profile wrote sidecar: %v", err)
			}
		})
	}

	t.Run("conformance rejects profile", func(t *testing.T) {
		t.Parallel()
		target := filepath.Join(t.TempDir(), "target")
		_, err := Prepare(PrepareOptions{
			CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target,
			Condition: ConditionConformance, Agent: "codex", PreflightContext: PreflightContextCompact,
		})
		if err == nil || !strings.Contains(err.Error(), "only valid for treatment") {
			t.Fatalf("Prepare() error = %v, want treatment-only rejection", err)
		}
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Fatalf("invalid conformance profile wrote target: %v", err)
		}
	})
}

func TestOutcomeBenchmarkTreatmentDefaultsToFullPreflightContext(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		requested string
	}{
		{name: "default"},
		{name: "explicit", requested: PreflightContextFull},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			options := benchmarkPrepareOptions(target, ConditionTreatment, "cursor", 1)
			options.PreflightContext = test.requested
			prepared, err := Prepare(options)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.PreflightContext != PreflightContextFull || prepared.HookPayload == nil || prepared.HookPayload.Profile != PreflightContextFull {
				t.Fatalf("full treatment profile = %+v", prepared)
			}
		})
	}
}

func TestOutcomeBenchmarkControlAndTreatmentIsolation(t *testing.T) {
	t.Parallel()
	controlTarget := filepath.Join(t.TempDir(), "control")
	treatmentTarget := filepath.Join(t.TempDir(), "treatment")
	control, err := Prepare(benchmarkPrepareOptions(controlTarget, ConditionControl, "claude-code", 2))
	if err != nil {
		t.Fatal(err)
	}
	treatment, err := Prepare(benchmarkPrepareOptions(treatmentTarget, ConditionTreatment, "claude-code", 2))
	if err != nil {
		t.Fatal(err)
	}
	if control.Prompt != treatment.Prompt {
		t.Fatal("control and treatment received different task prompts")
	}
	if control.Rubric != "" || treatment.Rubric != "" {
		t.Fatalf("blind benchmark rubric leaked during preparation: control=%q treatment=%q", control.Rubric, treatment.Rubric)
	}
	if control.PreflightContext != "" || control.HookPayload != nil {
		t.Fatalf("control contains treatment-only preflight metadata: %+v", control)
	}
	if treatment.PreflightContext != PreflightContextFull || treatment.HookPayload == nil || treatment.HookPayload.Profile != PreflightContextFull {
		t.Fatalf("treatment omitted default preflight metadata: %+v", treatment)
	}
	for _, prepared := range []PrepareResult{control, treatment} {
		if prepared.AgentVersion != "test-agent-1" || prepared.ModelVersion != "test-model-1" || prepared.ReasoningConfiguration != "high" || prepared.TrialNumber != 2 || prepared.SourceCommit == "" || prepared.Baseline == "" {
			t.Fatalf("prepared metadata was not preserved: %+v", prepared)
		}
	}
	for _, forbidden := range []string{
		".repo-knowledge",
		".agents/skills/repository-knowledge/SKILL.md",
		".claude/skills/repository-knowledge/SKILL.md",
		".cursor/skills/repository-knowledge/SKILL.md",
		"AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(controlTarget, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Errorf("control unexpectedly installed toolkit asset %s", forbidden)
		}
	}
	if _, err := os.Stat(control.Baseline); err != nil {
		t.Fatalf("control sidecar baseline is missing: %v", err)
	}
	manifestData, err := os.ReadFile(filepath.Join(treatmentTarget, ".repo-knowledge", "toolkit.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		AgentAdapters []string `json:"agent_adapters"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.AgentAdapters) != 1 || manifest.AgentAdapters[0] != "claude-code" {
		t.Fatalf("treatment adapters = %v, want only claude-code", manifest.AgentAdapters)
	}
	if _, err := os.Stat(filepath.Join(treatmentTarget, ".claude", "skills", "repository-knowledge", "SKILL.md")); err != nil {
		t.Fatalf("requested treatment adapter is missing: %v", err)
	}
	for _, forbidden := range []string{
		".agents/skills/repository-knowledge/SKILL.md",
		".cursor/skills/repository-knowledge/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(treatmentTarget, filepath.FromSlash(forbidden))); !os.IsNotExist(err) {
			t.Errorf("treatment installed unrequested adapter asset %s", forbidden)
		}
	}

	spec, caseRoot, err := LoadSpec(testBenchmarksRoot(t), "frontend-onboarding")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := resolveFixturePath(testBenchmarksRoot(t), caseRoot, spec.Fixture, true)
	if err != nil {
		t.Fatal(err)
	}
	fixtureFiles, err := hashRepository(fixture)
	if err != nil {
		t.Fatal(err)
	}
	controlFiles, err := hashRepository(controlTarget)
	if err != nil {
		t.Fatal(err)
	}
	if len(controlFiles) != len(fixtureFiles) {
		t.Fatalf("control contains %d files, fixture contains %d", len(controlFiles), len(fixtureFiles))
	}
	for relative, digest := range fixtureFiles {
		for condition, target := range map[string]string{"control": controlTarget, "treatment": treatmentTarget} {
			actual, err := hashFile(filepath.Join(target, filepath.FromSlash(relative)))
			if err != nil || actual != digest {
				t.Errorf("%s application file %s differs from fixture: digest=%s err=%v", condition, relative, actual, err)
			}
		}
	}

	for condition, target := range map[string]string{"control": controlTarget, "treatment": treatmentTarget} {
		grade, err := Grade(GradeOptions{CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target})
		if err != nil {
			t.Fatal(err)
		}
		if grade.Condition != condition || grade.TrialNumber != 2 || grade.AgentVersion != "test-agent-1" || grade.ModelVersion != "test-model-1" || grade.ReasoningConfiguration != "high" {
			t.Errorf("%s grade metadata was not preserved: %+v", condition, grade)
		}
		if findCheck(t, grade, "fixture-source-preserved").Status != "pass" {
			t.Errorf("%s untouched source did not pass protection: %+v", condition, grade.Checks)
		}
	}
}

func TestOutcomeBenchmarkProtectedSourceChecksBothConditions(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{ConditionControl, ConditionTreatment} {
		condition := condition
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			if _, err := Prepare(benchmarkPrepareOptions(target, condition, "codex", 1)); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(target, "src", "data", "intelligence.json")
			if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			grade, err := Grade(GradeOptions{CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target})
			if err != nil {
				t.Fatal(err)
			}
			check := findCheck(t, grade, "fixture-source-preserved")
			if check.Status != "fail" || !strings.Contains(check.Detail, "src/data/intelligence.json (modified)") {
				t.Fatalf("%s source protection check = %+v", condition, check)
			}
		})
	}
}

func TestOutcomeBenchmarkConditionErrorsDoNotCreateOutput(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{"", "experiment"} {
		name := condition
		if name == "" {
			name = "missing"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			options := benchmarkPrepareOptions(target, condition, "codex", 1)
			_, err := Prepare(options)
			if err == nil || !strings.Contains(err.Error(), "--condition control or --condition treatment") {
				t.Fatalf("Prepare() error = %v, want condition error", err)
			}
			if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
				t.Fatalf("invalid condition wrote output at %s", target)
			}
		})
	}
}

func TestOutcomeBenchmarkRunMetadataIsRequiredBeforeOutput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		remove func(*PrepareOptions)
		want   string
	}{
		{name: "agent version", remove: func(options *PrepareOptions) { options.AgentVersion = "" }, want: "agent version is required"},
		{name: "model version", remove: func(options *PrepareOptions) { options.ModelVersion = "" }, want: "model version is required"},
		{name: "reasoning", remove: func(options *PrepareOptions) { options.ReasoningConfiguration = "" }, want: "reasoning configuration is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			options := benchmarkPrepareOptions(target, ConditionControl, "codex", 1)
			test.remove(&options)
			_, err := Prepare(options)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Prepare() error = %v, want %q", err, test.want)
			}
			if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
				t.Fatalf("missing metadata wrote output at %s", target)
			}
		})
	}
}

func TestOutcomeBenchmarkResultRecordingIsImmutable(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(benchmarkPrepareOptions(target, ConditionControl, "codex", 3)); err != nil {
		t.Fatal(err)
	}
	tokens := int64(1234)
	grade, err := Grade(GradeOptions{
		CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target,
		DurationMillis: 9500, TokenUsage: &tokens,
	})
	if err != nil {
		t.Fatal(err)
	}
	if grade.DeterministicStatus != "fail" || grade.SemanticStatus != "pending_human_or_model_review" {
		t.Fatalf("failed trial metadata was overstated: %+v", grade)
	}
	artifact := filepath.Join(t.TempDir(), "output.tar.gz")
	if err := os.WriteFile(artifact, []byte("preserved output\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	resultsRoot := filepath.Join(t.TempDir(), "results")
	recorded, err := Record(RecordOptions{ResultsRoot: resultsRoot, RunDate: "2026-09-06", Artifact: artifact}, grade)
	if err != nil {
		t.Fatal(err)
	}
	if recorded.Result.RunDate != "2026-09-06" || recorded.Result.PreservedArtifact == "" || recorded.Result.TokenUsage == nil || *recorded.Result.TokenUsage != tokens {
		t.Fatalf("recorded result metadata is incomplete: %+v", recorded.Result)
	}
	data, err := os.ReadFile(recorded.ResultPath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted GradeResult
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.Condition != ConditionControl || persisted.TrialNumber != 3 || persisted.DeterministicStatus != "fail" || persisted.PreservedArtifact == "" || filepath.IsAbs(persisted.Rubric) {
		t.Fatalf("persisted result is incomplete: %+v", persisted)
	}
	if _, err := Record(RecordOptions{ResultsRoot: resultsRoot, RunDate: "2026-09-06", Artifact: artifact}, grade); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second Record() error = %v, want immutable output refusal", err)
	}
}

func TestPreparedShallowFixturesFailDeterministicGrade(t *testing.T) {
	t.Parallel()
	for _, caseID := range []string{"frontend-nextjs", "backend-clean-architecture"} {
		caseID := caseID
		t.Run(caseID, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			prepared, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: caseID, Output: target, Agent: "codex"})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(prepared.Prompt, "generation request") {
				t.Fatalf("prompt does not preserve generation intent: %q", prepared.Prompt)
			}
			if _, err := os.Stat(filepath.Join(target, ".agents", "skills", "repository-knowledge", "SKILL.md")); err != nil {
				t.Fatalf("installed skill: %v", err)
			}
			grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: caseID, Target: target})
			if err != nil {
				t.Fatal(err)
			}
			if grade.DeterministicStatus != "fail" || grade.Failed == 0 {
				t.Fatalf("initial grade = %+v, want deterministic failures", grade)
			}
			if findCheck(t, grade, "fixture-source-preserved").Status != "pass" {
				t.Fatalf("untouched fixture source should pass protection: %+v", grade.Checks)
			}
		})
	}
}

func TestFrontendObjectiveCandidatePassesDeterministicChecks(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "claude-code"}); err != nil {
		t.Fatal(err)
	}
	writeFrontendObjectiveCandidate(t, target)
	grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if grade.DeterministicStatus != "pass" || grade.Failed != 0 {
		t.Fatalf("deterministic grade = %+v, want pass", grade)
	}
	if grade.OverallStatus != "pending_semantic_review" || grade.SemanticStatus != "pending_human_or_model_review" {
		t.Fatalf("semantic result was overstated: %+v", grade)
	}
}

func TestSemanticSuccessRequiresExplicitReviewMetadata(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "codex"}); err != nil {
		t.Fatal(err)
	}
	writeFrontendObjectiveCandidate(t, target)
	score := &SemanticScore{Earned: 18, Available: 18}
	if _, err := Grade(GradeOptions{
		CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target, SemanticScore: score,
	}); err == nil || !strings.Contains(err.Error(), "explicit semantic status") {
		t.Fatalf("Grade() error = %v, want explicit semantic status requirement", err)
	}
	grade, err := Grade(GradeOptions{
		CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target,
		SemanticStatus: "pass", SemanticScore: score, SemanticReviewer: "reviewer@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if grade.DeterministicStatus != "pass" || grade.SemanticStatus != "pass" || grade.OverallStatus != "pass" || grade.SemanticReviewer == "" {
		t.Fatalf("explicit reviewed grade = %+v", grade)
	}
}

func TestProtectedSourceMutationFails(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "antigravity-ide"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, "src", "data", "intelligence.json")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target})
	if err != nil {
		t.Fatal(err)
	}
	check := findCheck(t, grade, "fixture-source-preserved")
	if check.Status != "fail" || !strings.Contains(check.Detail, "src/data/intelligence.json (modified)") {
		t.Fatalf("source protection check = %+v", check)
	}
}

func TestPrepareRefusesNonEmptyOutput(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "keep.txt"), []byte("preserve\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "codex"})
	if err == nil || !strings.Contains(err.Error(), "must be empty") {
		t.Fatalf("Prepare() error = %v, want non-empty refusal", err)
	}
}

func TestPrepareSupportsCursorAdapter(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	prepared, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "cursor"})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Agent != "cursor" {
		t.Fatalf("prepared agent = %q, want cursor", prepared.Agent)
	}
	for _, relative := range []string{
		".cursor/rules/repository-knowledge.mdc",
		".cursor/skills/repository-knowledge/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("installed Cursor adapter file %s: %v", relative, err)
		}
	}
}

func TestGlobMatchSupportsZeroOrMoreDirectories(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"docs/index.md", "docs/frontend/state.md", "docs/a/b/c.md"} {
		if !globMatch("docs/**/*.md", value) {
			t.Errorf("glob did not match %q", value)
		}
	}
	if globMatch("docs/**/*.md", "src/index.md") {
		t.Fatal("docs glob matched source file")
	}
}

func testCasesRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "evals", "cases"))
}

func testBenchmarksRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "evals", "benchmarks"))
}

func benchmarkPrepareOptions(target, condition, agent string, trial int) PrepareOptions {
	return PrepareOptions{
		CasesRoot: testBenchmarkRootFromSource(), CaseID: "frontend-onboarding", Output: target,
		Condition: condition, Agent: agent, AgentVersion: "test-agent-1",
		ModelVersion: "test-model-1", ReasoningConfiguration: "high",
		RepositoryKnowledgeRevision: "test-toolkit-commit", TrialNumber: trial,
	}
}

func testBenchmarkRootFromSource() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "evals/benchmarks"
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "evals", "benchmarks"))
}

func findCheck(t *testing.T, grade GradeResult, id string) CheckResult {
	t.Helper()
	for _, check := range grade.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("check %q not found", id)
	return CheckResult{}
}

func writeFrontendObjectiveCandidate(t *testing.T, target string) {
	t.Helper()
	content := `# Verified frontend mechanisms

The application declares React 19 in ` + "`package.json`" + ` and resolves it in ` + "`package-lock.json`" + `.

Intelligence is static: ` + "`IntelligencePage`" + ` imports ` + "`src/data/intelligence.json`" + ` directly from ` + "`src/app/(protected)/intelligence/page.tsx`" + `.

Protected navigation starts at ` + "`src/app/(protected)/layout.tsx`" + `. ` + "`ProtectedLayout`" + ` calls ` + "`requireSession`" + ` from ` + "`src/lib/session.ts`" + ` and supplies permissions to ` + "`useCheckPermission`" + ` in ` + "`src/auth/PermissionContext.tsx`" + `.

Forms use ` + "`src/features/users/UserForm.tsx`" + ` with ` + "`userSchema`" + ` from ` + "`src/features/users/schema.ts`" + `. Server state uses ` + "`src/features/users/useUsers.ts`" + `. UI composition uses ` + "`src/components/PageShell.tsx`" + `.

The BFF boundary is ` + "`src/pages/api/[...paths]/index.ts`" + `.

Source: ` + "`src/lib/session.ts`" + ` — ` + "`requireSession`" + `.

` + "```ts" + `
const session = await requireSession();
` + "```" + `

Source: ` + "`src/features/users/UserForm.tsx`" + ` — ` + "`UserForm`" + `.

` + "```tsx" + `
const result = userSchema.safeParse(values);
` + "```" + `

Source: ` + "`src/pages/api/[...paths]/index.ts`" + ` — ` + "`handler`" + `.

` + "```ts" + `
response.status(upstream.status).json(payload);
` + "```" + `
`
	write := func(relative, value string) {
		t.Helper()
		path := filepath.Join(target, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("docs/index.md", "# Repository knowledge index\n\n- [Overview](repository-overview.md)\n- [Architecture](architecture.md)\n- [Features](features.md)\n- [Integrations](integrations.md)\n- [Development](development.md)\n- [State](state.md)\n")
	write("docs/repository-overview.md", content)
	write("docs/architecture.md", content)
	write("docs/features.md", content)
	write("docs/integrations.md", content)
	write("docs/development.md", content)
	write("docs/state.md", content)
	write(".repo-knowledge/repository.json", `{
  "$schema": "schemas/repository.schema.json",
  "schema_version": "1.0",
  "repository": {"name": "eval-frontend", "kind": "frontend", "owners": []},
  "documentation": {"index": "docs/index.md", "generated_inventory": "docs/repository-inventory.md"},
  "capabilities": [
    {"id": "navigation", "documentation": ["docs/architecture.md"], "evidence": ["src/app/"]},
    {"id": "permissions", "documentation": ["docs/architecture.md"], "evidence": ["src/auth/"]},
    {"id": "users", "documentation": ["docs/features.md"], "evidence": ["src/features/users/"]},
    {"id": "bff", "documentation": ["docs/integrations.md"], "evidence": ["src/pages/api/"]}
  ],
  "ci": {"enforcement": "advisory"}
}
`)
}
