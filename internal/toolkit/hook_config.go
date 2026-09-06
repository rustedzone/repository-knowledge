package toolkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const antigravityHookKey = "repository-knowledge-preflight"

var hookConfigPaths = map[string]string{
	agentCodex:          ".codex/hooks.json",
	agentClaudeCode:     ".claude/settings.json",
	agentAntigravityIDE: ".agents/hooks.json",
	agentCursor:         ".cursor/hooks.json",
}

type hookConfigMutation struct {
	path    string
	config  map[string]any
	existed bool
}

func prepareAgentHooks(target string, oldAdapters, adapters []string, update bool, antigravityPreflight string) ([]string, []hookConfigMutation, error) {
	registrations := make([]string, 0, len(adapters))
	mutations := make([]hookConfigMutation, 0, len(adapters))
	for _, agent := range supportedAgentAdapters {
		enabled := contains(adapters, agent)
		if !enabled && (!update || !contains(oldAdapters, agent)) {
			continue
		}
		mutation, err := prepareAgentHookMutation(target, agent, enabled, antigravityPreflight)
		if err != nil {
			return nil, nil, err
		}
		mutations = append(mutations, mutation)
		if enabled {
			registrations = append(registrations, hookConfigPaths[agent])
		}
	}
	sort.Strings(registrations)
	return registrations, mutations, nil
}

func applyAgentHooks(mutations []hookConfigMutation) error {
	for _, mutation := range mutations {
		if err := writeOrRemoveHookConfig(mutation.path, mutation.config, mutation.existed); err != nil {
			return err
		}
	}
	return nil
}

func prepareAgentHookMutation(target, agent string, enabled bool, antigravityPreflight string) (hookConfigMutation, error) {
	relative := hookConfigPaths[agent]
	path, err := repositoryPath(target, relative, "agent hook configuration")
	if err != nil {
		return hookConfigMutation{}, err
	}
	config, exists, err := readHookConfig(path)
	if err != nil {
		return hookConfigMutation{}, err
	}

	if !enabled {
		if err := removeAgentHookFromConfig(config, agent); err != nil {
			return hookConfigMutation{}, fmt.Errorf("reconcile %s hook in %s: %w", agent, relative, err)
		}
		return hookConfigMutation{path: path, config: config, existed: exists}, nil
	}

	switch agent {
	case agentCodex:
		err = reconcileMatcherHook(config, "SessionStart", managedHookCommand(agent), true, map[string]any{
			"matcher": "startup|resume|clear|compact",
			"hooks": []any{map[string]any{
				"type":                   "command",
				"command":                managedHookCommand(agent),
				"timeout":                float64(10),
				"statusMessage":          "Loading repository knowledge",
				"additionalContextLimit": float64(65536),
			}},
		})
	case agentClaudeCode:
		err = reconcileMatcherHook(config, "SessionStart", managedHookCommand(agent), true, map[string]any{
			"matcher": "startup|resume|clear|compact|fork",
			"hooks": []any{map[string]any{
				"type":    "command",
				"command": managedHookCommand(agent),
				"timeout": float64(10),
			}},
		})
	case agentAntigravityIDE:
		err = reconcileAntigravityHook(config, true, antigravityPreflight == AntigravityPreflightStrict)
	case agentCursor:
		err = reconcileCursorHook(config, true)
	default:
		err = fmt.Errorf("unsupported agent hook adapter %q", agent)
	}
	if err != nil {
		return hookConfigMutation{}, fmt.Errorf("reconcile %s hook in %s: %w", agent, relative, err)
	}
	return hookConfigMutation{path: path, config: config, existed: exists}, nil
}

func removeAgentHookFromConfig(config map[string]any, agent string) error {
	switch agent {
	case agentCodex, agentClaudeCode:
		return reconcileMatcherHook(config, "SessionStart", managedHookCommand(agent), false, nil)
	case agentAntigravityIDE:
		return reconcileAntigravityHook(config, false, false)
	case agentCursor:
		return reconcileCursorHook(config, false)
	default:
		return fmt.Errorf("unsupported agent hook adapter %q", agent)
	}
}

func managedHookCommand(agent string) string {
	return "repo-knowledge hook-context --agent " + agent
}

func managedPreflightGateCommand(agent string) string {
	return "repo-knowledge preflight-gate --agent " + agent
}

func readHookConfig(path string) (map[string]any, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return make(map[string]any), false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("read %s: %w", path, err)
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, false, fmt.Errorf("parse JSON %s before merging repository-knowledge hook: %w", path, err)
	}
	if config == nil {
		config = make(map[string]any)
	}
	return config, true, nil
}

func reconcileMatcherHook(config map[string]any, event, command string, enabled bool, canonical map[string]any) error {
	hooks, err := objectField(config, "hooks", enabled)
	if err != nil {
		return err
	}
	if hooks == nil {
		return nil
	}
	entries, err := arrayField(hooks, event)
	if err != nil {
		return err
	}
	filtered := make([]any, 0, len(entries)+1)
	for _, entry := range entries {
		group, ok := entry.(map[string]any)
		if !ok {
			filtered = append(filtered, entry)
			continue
		}
		handlers, ok := group["hooks"].([]any)
		if !ok {
			filtered = append(filtered, entry)
			continue
		}
		kept := make([]any, 0, len(handlers))
		for _, handler := range handlers {
			if hookCommand(handler) != command {
				kept = append(kept, handler)
			}
		}
		if len(kept) > 0 {
			copy := cloneObject(group)
			copy["hooks"] = kept
			filtered = append(filtered, copy)
		}
	}
	if enabled {
		filtered = append(filtered, canonical)
	}
	if len(filtered) == 0 {
		delete(hooks, event)
	} else {
		hooks[event] = filtered
	}
	if len(hooks) == 0 {
		delete(config, "hooks")
	}
	return nil
}

func reconcileAntigravityHook(config map[string]any, enabled, strict bool) error {
	if existing, ok := config[antigravityHookKey]; ok {
		if !containsHookCommand(existing, managedHookCommand(agentAntigravityIDE)) {
			return fmt.Errorf("top-level key %q already belongs to the consuming repository", antigravityHookKey)
		}
		delete(config, antigravityHookKey)
	}
	if enabled {
		preToolUse := []any{}
		if strict {
			preToolUse = append(preToolUse, map[string]any{
				"type":    "command",
				"command": managedPreflightGateCommand(agentAntigravityIDE),
				"timeout": float64(10),
			})
		}
		config[antigravityHookKey] = map[string]any{
			"enabled": true,
			"PreInvocation": []any{map[string]any{
				"type":    "command",
				"command": managedHookCommand(agentAntigravityIDE),
				"timeout": float64(10),
			}},
		}
		if len(preToolUse) > 0 {
			config[antigravityHookKey].(map[string]any)["PreToolUse"] = preToolUse
		}
	}
	return nil
}

func reconcileCursorHook(config map[string]any, enabled bool) error {
	if version, ok := config["version"]; ok && version != float64(1) {
		return fmt.Errorf("unsupported Cursor hooks version %v", version)
	}
	hooks, err := objectField(config, "hooks", enabled)
	if err != nil {
		return err
	}
	if hooks == nil {
		return nil
	}
	entries, err := arrayField(hooks, "sessionStart")
	if err != nil {
		return err
	}
	filtered := make([]any, 0, len(entries)+1)
	for _, entry := range entries {
		if hookCommand(entry) != managedHookCommand(agentCursor) {
			filtered = append(filtered, entry)
		}
	}
	if enabled {
		filtered = append(filtered, map[string]any{"command": managedHookCommand(agentCursor)})
		config["version"] = float64(1)
	}
	if len(filtered) == 0 {
		delete(hooks, "sessionStart")
	} else {
		hooks["sessionStart"] = filtered
	}
	if len(hooks) == 0 {
		delete(config, "hooks")
	}
	if !enabled && len(config) == 1 && config["version"] == float64(1) {
		delete(config, "version")
	}
	return nil
}

func objectField(parent map[string]any, name string, create bool) (map[string]any, error) {
	value, ok := parent[name]
	if !ok {
		if !create {
			return nil, nil
		}
		object := make(map[string]any)
		parent[name] = object
		return object, nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a JSON object", name)
	}
	return object, nil
}

func arrayField(parent map[string]any, name string) ([]any, error) {
	value, ok := parent[name]
	if !ok {
		return nil, nil
	}
	array, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a JSON array", name)
	}
	return array, nil
}

func hookCommand(value any) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	command, _ := object["command"].(string)
	return command
}

func containsHookCommand(value any, command string) bool {
	switch typed := value.(type) {
	case map[string]any:
		if hookCommand(typed) == command {
			return true
		}
		for _, nested := range typed {
			if containsHookCommand(nested, command) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsHookCommand(nested, command) {
				return true
			}
		}
	}
	return false
}

func cloneObject(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = item
	}
	return result
}

func writeOrRemoveHookConfig(path string, config map[string]any, existed bool) error {
	if len(config) == 0 {
		if !existed {
			return nil
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove empty hook configuration %s: %w", path, err)
		}
		return nil
	}
	return writeJSON(path, config)
}

func hasAgentHookRegistration(root, agent string) (bool, string) {
	relative := hookConfigPaths[agent]
	path := filepath.Join(root, filepath.FromSlash(relative))
	config, _, err := readHookConfig(path)
	if err != nil {
		return false, err.Error()
	}
	command := managedHookCommand(agent)
	switch agent {
	case agentCodex, agentClaudeCode:
		hooks, ok := config["hooks"].(map[string]any)
		if !ok {
			return false, relative
		}
		entries, ok := hooks["SessionStart"].([]any)
		if !ok {
			return false, relative
		}
		return containsHookCommand(entries, command), relative
	case agentAntigravityIDE:
		value, ok := config[antigravityHookKey]
		if !ok || !containsHookCommand(value, command) {
			return false, relative
		}
		return true, relative
	case agentCursor:
		hooks, ok := config["hooks"].(map[string]any)
		if !ok {
			return false, relative
		}
		entries, ok := hooks["sessionStart"].([]any)
		return ok && containsHookCommand(entries, command), relative
	default:
		return false, relative
	}
}
