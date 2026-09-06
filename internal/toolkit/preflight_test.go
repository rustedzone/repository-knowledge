package toolkit

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
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
}

func TestStrictAntigravityPreflightRejectsExpiredOrCrossRepositoryActivation(t *testing.T) {
	root := newGitRepository(t)
	other := newGitRepository(t)
	if _, err := Install(InstallOptions{
		Target: root, AgentAdapters: []string{agentAntigravityIDE}, AntigravityPreflight: AntigravityPreflightStrict,
	}); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	context, err := HookContext(root, agentAntigravityIDE, strings.NewReader(`{"conversationId":"isolated","invocationNum":0}`))
	if err != nil {
		t.Fatalf("HookContext() error = %v", err)
	}
	if _, err := PreflightActivate(other, activationToken(t, context.Output), []string{"docs/index.md"}); err == nil || !strings.Contains(err.Error(), "different repository") {
		t.Fatalf("cross-repository activation error = %v, want rejection", err)
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
