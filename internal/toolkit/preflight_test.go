package toolkit

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestStrictAntigravityPreflightBlocksDiscoveryUntilActivation(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"sso-plan","invocationNum":0}`))
	if err != nil {
		t.Fatalf("HookContext() error = %v", err)
	}
	token := activationToken(t, context.Output)

	pending, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{
  "conversationId":"sso-plan",
  "toolCall":{"name":"grep_search","arguments":{"query":"Google SSO","path":"internal"}}
}`))
	if err != nil {
		t.Fatalf("PreflightGate() pending error = %v", err)
	}
	if pending.Decision != "deny" || !strings.Contains(pending.Reason, "preflight is pending") {
		t.Fatalf("pending gate = %+v, want preflight denial", pending)
	}

	knowledgeRead, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{
  "conversationId":"sso-plan",
  "toolCall":{"name":"view_file","arguments":{"path":"docs/index.md"}}
}`))
	if err != nil {
		t.Fatalf("PreflightGate() knowledge read error = %v", err)
	}
	if knowledgeRead.Decision != "allow" {
		t.Fatalf("knowledge read gate = %+v, want allow", knowledgeRead)
	}

	activated, err := PreflightActivate(root, token, []string{"docs/index.md"})
	if err != nil {
		t.Fatalf("PreflightActivate() error = %v", err)
	}
	if activated.Status != "active" || strings.Join(activated.Routes, ",") != "docs/index.md" {
		t.Fatalf("activation = %+v", activated)
	}

	active, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{
  "conversationId":"sso-plan",
  "toolCall":{"name":"grep_search","arguments":{"query":"Google SSO","path":"internal"}}
}`))
	if err != nil {
		t.Fatalf("PreflightGate() active error = %v", err)
	}
	if active.Decision != "allow" {
		t.Fatalf("active gate = %+v, want allow", active)
	}
	resumed, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"sso-plan","invocationNum":1}`))
	if err != nil || !strings.Contains(resumed.Output, "status: active") {
		t.Fatalf("resumed strict context = %+v, %v", resumed, err)
	}
}

func TestStrictAntigravityPreflightRejectsCrossRepositoryActivation(t *testing.T) {
	root := newGitRepository(t)
	other := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := Install(InstallOptions{Target: other, AgentAdapters: []string{agentAntigravityIDE}}); err != nil {
		t.Fatalf("Install(other) error = %v", err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"isolated","invocationNum":0}`))
	if err != nil {
		t.Fatalf("HookContext() error = %v", err)
	}
	if _, err := PreflightActivate(other, activationToken(t, context.Output), []string{"docs/index.md"}); err == nil || !strings.Contains(err.Error(), "different repository") {
		t.Fatalf("cross-repository activation error = %v, want rejection", err)
	}
}

func TestStrictAntigravityPreflightExpiresBeforeActivation(t *testing.T) {
	now := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	previousClock := preflightClock
	preflightClock = func() time.Time { return now }
	t.Cleanup(func() { preflightClock = previousClock })

	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"expires","invocationNum":0}`))
	if err != nil {
		t.Fatalf("HookContext() error = %v", err)
	}
	now = now.Add(preflightSessionTTL + time.Second)
	if _, err := PreflightActivate(root, activationToken(t, context.Output), []string{"docs/index.md"}); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expired activation error = %v, want expiry rejection", err)
	}
}

func TestStrictAntigravityInstallRegistersGateAndDoctorRequiresIt(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	config, _, err := readHookConfig(root + "/.agents/hooks.json")
	if err != nil {
		t.Fatal(err)
	}
	if !containsHookCommand(config[antigravityHookKey], managedPreflightGateCommand(agentAntigravityIDE)) {
		t.Fatalf("strict hook config does not contain gate: %+v", config)
	}
	report, err := Doctor(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != "pass" {
		t.Fatalf("Doctor() = %+v, want pass", report)
	}
}

func TestStrictAntigravityDoctorLiveHooksExercisesPendingAndActiveGate(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	report, err := DoctorLive(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range report.Checks {
		if check.ID == "antigravity-ide-preflight-gate-self-test" {
			if !check.OK {
				t.Fatalf("live self-test failed: %+v", check)
			}
			return
		}
	}
	t.Fatalf("DoctorLive() did not report the strict gate self-test: %+v", report)
}

func TestStrictAntigravityGateOnlyAllowsSafeActivationCommand(t *testing.T) {
	allowed := json.RawMessage(`{"command":"repo-knowledge preflight-activate --token abc --route docs/index.md"}`)
	if !activationCommandAllowed(allowed) {
		t.Fatal("valid activation command was denied")
	}
	unsafe := json.RawMessage(`{"command":"repo-knowledge preflight-activate --token abc --route docs/index.md; git status"}`)
	if activationCommandAllowed(unsafe) {
		t.Fatal("shell-chained activation command was allowed")
	}
}

func TestStrictAntigravityPreflightRejectsSourceRoutesAndDeduplicatesDocumentationRoutes(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"routes","invocationNum":0}`))
	if err != nil {
		t.Fatal(err)
	}
	token := activationToken(t, context.Output)
	if _, err := PreflightActivate(root, token, []string{"internal/toolkit/install.go"}); err == nil || !strings.Contains(err.Error(), "documentation file") {
		t.Fatalf("source route error = %v, want documentation rejection", err)
	}
	activated, err := PreflightActivate(root, token, []string{"docs/index.md", "docs/index.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(activated.Routes, ",") != "docs/index.md" {
		t.Fatalf("deduplicated routes = %v", activated.Routes)
	}
}

func TestStrictAntigravityGateAllowsOnlyKnowledgeReadsAndExternalResearchWhilePending(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"allowlist","invocationNum":0}`)); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{
		`{"conversationId":"allowlist","toolCall":{"name":"search_web","arguments":{"query":"Google OIDC"}}}`,
		`{"conversationId":"allowlist","toolCall":{"name":"read_url_content","arguments":{"url":"https://example.test"}}}`,
		`{"conversationId":"allowlist","toolCall":{"name":"view_file","arguments":{"path":".agents/skills/repository-knowledge/SKILL.md"}}}`,
	} {
		result, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(input))
		if err != nil || result.Decision != "allow" {
			t.Fatalf("allowlisted gate input %s = %+v, %v", input, result, err)
		}
	}
	denied, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"allowlist","toolCall":{"name":"view_file","arguments":{"path":"internal/toolkit/install.go"}}}`))
	if err != nil || denied.Decision != "deny" {
		t.Fatalf("source read = %+v, %v, want deny", denied, err)
	}
}

func TestObserveAntigravityPreflightDoesNotGateToolsAndUpdatesRetainStrictMode(t *testing.T) {
	observeRoot := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: observeRoot, AgentAdapters: []string{agentAntigravityIDE}}); err != nil {
		t.Fatal(err)
	}
	result, err := PreflightGate(observeRoot, agentAntigravityIDE, strings.NewReader(`{"toolCall":{"name":"grep_search","arguments":{"query":"anything"}}}`))
	if err != nil || result.Decision != "allow" {
		t.Fatalf("observe gate = %+v, %v", result, err)
	}

	strictRoot := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: strictRoot, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict}); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{Target: strictRoot, Update: true}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	manifest, err := readJSON[ToolkitManifest](strictRoot+"/.repo-knowledge/toolkit.json", true)
	if err != nil || manifest.AntigravityPreflight != AntigravityPreflightStrict {
		t.Fatalf("updated manifest = %+v, %v", manifest, err)
	}
}

func TestStrictAntigravityGateAllowsExactActivationCommandAndRejectsMalformedEvents(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict}); err != nil {
		t.Fatal(err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"command","invocationNum":0}`))
	if err != nil {
		t.Fatal(err)
	}
	token := activationToken(t, context.Output)
	allowed, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"command","toolCall":{"name":"run_command","arguments":{"command":"repo-knowledge preflight-activate --token `+token+` --route docs/index.md"}}}`))
	if err != nil || allowed.Decision != "allow" {
		t.Fatalf("activation command gate = %+v, %v", allowed, err)
	}
	if _, err := PreflightGate(root, agentAntigravityIDE, strings.NewReader(`{not-json}`)); err == nil || !strings.Contains(err.Error(), "parse Antigravity hook input") {
		t.Fatalf("malformed event error = %v", err)
	}
}

func TestPreflightActivateRequiresRoutes(t *testing.T) {
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict}); err != nil {
		t.Fatal(err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"no-routes","invocationNum":0}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PreflightActivate(root, activationToken(t, context.Output), nil); err == nil || !strings.Contains(err.Error(), "at least one --route") {
		t.Fatalf("missing route error = %v", err)
	}
	if _, err := PreflightActivate(root, strings.Repeat("a", 64), []string{"docs/index.md"}); err == nil || !strings.Contains(err.Error(), "unknown preflight token") {
		t.Fatalf("unknown token error = %v", err)
	}
}

func TestPreflightGateRejectsOtherAgentsAndAllowsAbsoluteKnowledgePaths(t *testing.T) {
	if event, err := parseAntigravityHookEvent(nil); err != nil || event.ConversationID != "" {
		t.Fatalf("nil event parse = %+v, %v", event, err)
	}
	if _, err := PreflightGate(".", agentCodex, strings.NewReader(`{}`)); err == nil || !strings.Contains(err.Error(), "only supports") {
		t.Fatalf("non-Antigravity gate error = %v", err)
	}
	root := newGitRepository(t)
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict}); err != nil {
		t.Fatal(err)
	}
	if !knowledgePathReadAllowed(root, json.RawMessage(`{"path":"`+filepath.Join(root, "docs", "index.md")+`"}`)) {
		t.Fatal("absolute documentation path was not allowlisted")
	}
}

func activationToken(t *testing.T, output string) string {
	t.Helper()
	var body struct {
		InjectSteps []struct {
			EphemeralMessage string `json:"ephemeralMessage"`
		} `json:"injectSteps"`
	}
	if err := json.Unmarshal([]byte(output), &body); err != nil || len(body.InjectSteps) != 1 {
		t.Fatalf("parse hook output %q: %v", output, err)
	}
	match := regexp.MustCompile(`--token ([a-f0-9]+)`).FindStringSubmatch(body.InjectSteps[0].EphemeralMessage)
	if len(match) != 2 {
		t.Fatalf("no activation token in hook output: %s", body.InjectSteps[0].EphemeralMessage)
	}
	return match[1]
}
