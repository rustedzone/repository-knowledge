package toolkit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStrictPreflightBlocksRepositoryToolsForEverySupportedAgent(t *testing.T) {
	tests := []struct {
		agent       string
		contextInput string
		gateInput    string
		denyMarker   string
	}{
		{
			agent:       agentCodex,
			contextInput: `{"session_id":"codex-sso","source":"startup"}`,
			gateInput:    `{"session_id":"codex-sso","tool_name":"Bash","tool_input":{"command":"git status"}}`,
			denyMarker:   `"permissionDecision":"deny"`,
		},
		{
			agent:       agentClaudeCode,
			contextInput: `{"session_id":"claude-sso","source":"startup"}`,
			gateInput:    `{"session_id":"claude-sso","tool_name":"Bash","tool_input":{"command":"git status"}}`,
			denyMarker:   `"permissionDecision":"deny"`,
		},
		{
			agent:       agentCursor,
			contextInput: `{"conversation_id":"cursor-sso"}`,
			gateInput:    `{"conversation_id":"cursor-sso","tool_name":"Shell","tool_input":{"command":"git status"}}`,
			denyMarker:   `"permission":"deny"`,
		},
	}

	for _, test := range tests {
		t.Run(test.agent, func(t *testing.T) {
			root := newGitRepository(t)
			if _, err := Install(InstallOptions{
				Target: root, AgentAdapters: []string{test.agent}, PreflightModes: map[string]string{test.agent: AntigravityPreflightStrict},
			}); err != nil {
				t.Fatalf("Install() error = %v", err)
			}
			context, err := HookContext(root, test.agent, strings.NewReader(test.contextInput))
			if err != nil {
				t.Fatalf("HookContext() error = %v", err)
			}
			token := activationTokenFromContext(t, test.agent, context.Output)

			pending, err := PreflightGate(root, test.agent, strings.NewReader(test.gateInput))
			if err != nil || pending.Decision != "deny" || !strings.Contains(pending.Output, test.denyMarker) {
				t.Fatalf("pending gate = %+v, %v", pending, err)
			}
			if _, err := PreflightActivate(root, token, []string{"docs/index.md"}); err != nil {
				t.Fatalf("PreflightActivate() error = %v", err)
			}
			active, err := PreflightGate(root, test.agent, strings.NewReader(test.gateInput))
			if err != nil || active.Decision != "allow" {
				t.Fatalf("active gate = %+v, %v", active, err)
			}
			config, _, err := readHookConfig(root + "/" + hookConfigPaths[test.agent])
			if err != nil || !containsHookCommand(config, managedPreflightGateCommand(test.agent)) {
				t.Fatalf("strict hook registration = %+v, %v", config, err)
			}
		})
	}
}

func TestPreflightModesRetainStrictSettingsForEverySelectedAdapter(t *testing.T) {
	root := newGitRepository(t)
	modes := map[string]string{
		agentCodex:      AntigravityPreflightStrict,
		agentClaudeCode: AntigravityPreflightStrict,
		agentCursor:     AntigravityPreflightStrict,
	}
	if _, err := Install(InstallOptions{Target: root, AgentAdapters: []string{agentCodex, agentClaudeCode, agentCursor}, PreflightModes: modes}); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(InstallOptions{Target: root, Update: true}); err != nil {
		t.Fatal(err)
	}
	manifest, err := readJSON[ToolkitManifest](root+"/.repo-knowledge/toolkit.json", true)
	if err != nil {
		t.Fatal(err)
	}
	for agent, mode := range modes {
		if manifest.PreflightModes[agent] != mode {
			t.Fatalf("mode for %s = %q, want %q; manifest = %+v", agent, manifest.PreflightModes[agent], mode, manifest)
		}
	}
}

func activationTokenFromContext(t *testing.T, agent, output string) string {
	t.Helper()
	if agent == agentAntigravityIDE {
		return activationToken(t, output)
	}
	if agent == agentCursor {
		var body struct {
			AdditionalContext string `json:"additional_context"`
		}
		if err := json.Unmarshal([]byte(output), &body); err != nil {
			t.Fatalf("parse Cursor context %q: %v", output, err)
		}
		return preflightTokenFromText(t, body.AdditionalContext)
	}
	return preflightTokenFromText(t, output)
}
