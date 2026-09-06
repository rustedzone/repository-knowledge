package toolkit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestHookContextProfilesExposeContentFreeMetrics(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentCodex}}); err != nil {
		t.Fatal(err)
	}

	full, err := HookContext(root, agentCodex, strings.NewReader(`{"session_id":"full-context"}`))
	if err != nil {
		t.Fatal(err)
	}
	if full.Metrics.Profile != PreflightContextFull || full.Metrics.Bytes != len(full.Output) || full.Metrics.Characters != utf8.RuneCountInString(full.Output) || full.Metrics.GenerationMillis < 0 {
		t.Fatalf("full context metrics = %+v for %d bytes", full.Metrics, len(full.Output))
	}
	encodedMetrics, err := json.Marshal(full.Metrics)
	if err != nil || strings.Contains(string(encodedMetrics), "Installed toolkit policy contract") {
		t.Fatalf("metrics leaked context content: %s, %v", encodedMetrics, err)
	}

	if _, err := Install(InstallOptions{Target: root, Update: true, PreflightContext: PreflightContextCompact}); err != nil {
		t.Fatal(err)
	}
	compact, err := HookContext(root, agentCodex, strings.NewReader(`{"session_id":"compact-context"}`))
	if err != nil {
		t.Fatal(err)
	}
	if compact.Metrics.Profile != PreflightContextCompact || compact.Metrics.Bytes != len(compact.Output) {
		t.Fatalf("compact context metrics = %+v for %d bytes", compact.Metrics, len(compact.Output))
	}
	if compact.Metrics.Bytes >= full.Metrics.Bytes {
		t.Fatalf("compact context is not smaller: compact=%d full=%d", compact.Metrics.Bytes, full.Metrics.Bytes)
	}
	if strings.Contains(compact.Output, "Installed toolkit policy contract") {
		t.Fatalf("compact context included the full policy payload: %s", compact.Output)
	}

	if _, err := Install(InstallOptions{Target: root, Update: true}); err != nil {
		t.Fatal(err)
	}
	manifest, err := readJSON[ToolkitManifest](filepath.Join(root, ".repo-knowledge", "toolkit.json"), true)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.PreflightContext != PreflightContextCompact {
		t.Fatalf("updated context profile = %q, want retained compact", manifest.PreflightContext)
	}
}

func TestEvidenceReportRequiresCurrentSuccessfulVerification(t *testing.T) {
	root, token := strictEvidenceRepository(t, "evidence-current")
	if _, err := PreflightActivateWithOptions(PreflightActivationOptions{
		Target: root, Token: token, Routes: []string{"docs/index.md"}, Workflow: WorkflowStandard,
	}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "internal", "worker.go")
	mustWrite(t, path, "package internal\n\nfunc Work() {}\n")

	if _, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "The internal helper does not change documented behavior.", EvidenceRefs: []string{"internal/worker.go#Work"},
	}); err == nil || !strings.Contains(err.Error(), "successful verification") {
		t.Fatalf("report without verification error = %v", err)
	}

	run, err := EvidenceRun(EvidenceRunOptions{
		Target: root, Token: token, Label: "unit-tests",
		Command: []string{os.Args[0], "-test.run=TestEvidenceCommandHelper", "--", "success"},
	})
	if err != nil || run.ExitCode != 0 || run.OutputSHA256 == "" || run.DiffFingerprint == "" {
		t.Fatalf("EvidenceRun() = %+v, %v", run, err)
	}

	mustWrite(t, path, "package internal\n\nfunc Work() int { return 1 }\n")
	if _, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "The internal helper does not change documented behavior.", EvidenceRefs: []string{"internal/worker.go#Work"},
	}); err == nil || !strings.Contains(err.Error(), "current diff") {
		t.Fatalf("stale verification error = %v", err)
	}

	if _, err := EvidenceRun(EvidenceRunOptions{
		Target: root, Token: token, Label: "unit-tests",
		Command: []string{os.Args[0], "-test.run=TestEvidenceCommandHelper", "--", "success"},
	}); err != nil {
		t.Fatal(err)
	}
	report, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "The internal helper does not change documented behavior.", EvidenceRefs: []string{"internal/worker.go#Work"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "complete" || len(report.Verifications) != 1 || report.DiffFingerprint == "" {
		t.Fatalf("evidence report = %+v", report)
	}
	if strings.Contains(fmt.Sprintf("%+v", report.Verifications[0]), "PASS") {
		t.Fatalf("verification report persisted command output: %+v", report.Verifications[0])
	}
}

func TestScopedWorkflowEscalatesHighRiskChanges(t *testing.T) {
	root, token := strictEvidenceRepository(t, "scoped-risk")
	activated, err := PreflightActivateWithOptions(PreflightActivationOptions{
		Target: root, Token: token, Routes: []string{"docs/index.md"}, Workflow: WorkflowScoped,
	})
	if err != nil || activated.Workflow != WorkflowScoped {
		t.Fatalf("scoped activation = %+v, %v", activated, err)
	}
	mustWrite(t, filepath.Join(root, "internal", "api", "users.go"), "package api\n\nfunc Users() {}\n")
	if _, err := EvidenceRun(EvidenceRunOptions{
		Target: root, Token: token, Label: "unit-tests",
		Command: []string{os.Args[0], "-test.run=TestEvidenceCommandHelper", "--", "success"},
	}); err != nil {
		t.Fatal(err)
	}
	_, err = EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "This API file is covered by its existing repository route.", EvidenceRefs: []string{"internal/api/users.go#Users"},
	})
	if err == nil || !strings.Contains(err.Error(), "scoped workflow") || !strings.Contains(err.Error(), "api") {
		t.Fatalf("high-risk scoped report error = %v", err)
	}

	if _, err := PreflightActivateWithOptions(PreflightActivationOptions{
		Target: root, Token: token, Routes: []string{"docs/index.md"}, Workflow: WorkflowStandard,
	}); err != nil {
		t.Fatal(err)
	}
	report, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "This API file is covered by its existing repository route.", EvidenceRefs: []string{"internal/api/users.go#Users"},
	})
	if err != nil || report.Workflow != WorkflowStandard || report.Status != "complete" {
		t.Fatalf("standard evidence report = %+v, %v", report, err)
	}
}

func TestEvidenceReportValidatesAttributionAndDocumentationDecision(t *testing.T) {
	root, token := strictEvidenceRepository(t, "evidence-validation")
	if _, err := PreflightActivateWithOptions(PreflightActivationOptions{
		Target: root, Token: token, Routes: []string{"docs/index.md"}, Workflow: WorkflowStandard,
	}); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "service.go"), "package service\n")
	if _, err := EvidenceRun(EvidenceRunOptions{
		Target: root, Token: token, Label: "unit-tests",
		Command: []string{os.Args[0], "-test.run=TestEvidenceCommandHelper", "--", "success"},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactNotRequired,
		Reason: "No public behavior changed.", EvidenceRefs: []string{"missing.go#Thing"},
	}); err == nil || !strings.Contains(err.Error(), "evidence reference") {
		t.Fatalf("missing evidence error = %v", err)
	}
	if _, err := EvidenceReport(EvidenceReportOptions{
		Target: root, Token: token, DocumentationImpact: DocumentationImpactRequired,
		EvidenceRefs: []string{"service.go#service"},
	}); err == nil || !strings.Contains(err.Error(), "documentation change") {
		t.Fatalf("missing documentation error = %v", err)
	}
}

func strictEvidenceRepository(t *testing.T, conversation string) (string, string) {
	t.Helper()
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentCodex}, PreflightModes: map[string]string{agentCodex: AntigravityPreflightStrict},
	}); err != nil {
		t.Fatal(err)
	}
	commitAll(t, root, "install repository knowledge")
	context, err := HookContext(root, agentCodex, strings.NewReader(`{"session_id":"`+conversation+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	return root, preflightTokenFromText(t, context.Output)
}

func TestEvidenceCommandHelper(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-1] != "success" {
		return
	}
	fmt.Fprintln(os.Stdout, "verification output must be hashed, not persisted")
}
