package toolkit

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const preflightSessionTTL = 30 * time.Minute

var preflightClock = time.Now

type preflightSession struct {
	SchemaVersion string            `json:"schema_version"`
	Token         string            `json:"token"`
	Root          string            `json:"root"`
	Agent         string            `json:"agent"`
	Conversation  string            `json:"conversation"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	Active        bool              `json:"active"`
	Routes        []string          `json:"routes,omitempty"`
	Digests       map[string]string `json:"digests"`
}

type antigravityHookEvent struct {
	ConversationID string `json:"conversationId"`
	InvocationNum  int    `json:"invocationNum"`
	ToolCall       struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"toolCall"`
}

type preflightHookEvent struct {
	ConversationID string
	InvocationNum  int
	ToolName       string
	ToolInput      json.RawMessage
}

func preflightStateDirectory() (string, error) {
	path := filepath.Join(os.TempDir(), "repository-knowledge", "preflight")
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("create preflight state directory: %w", err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return "", fmt.Errorf("secure preflight state directory: %w", err)
	}
	return path, nil
}

func preflightSessionPath(root, agent, conversation string) (string, error) {
	directory, err := preflightStateDirectory()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(root + "\x00" + agent + "\x00" + conversation))
	return filepath.Join(directory, hex.EncodeToString(digest[:])+".json"), nil
}

func preflightDigests(root string) (map[string]string, RepositoryConfig, error) {
	configPath := filepath.Join(root, ".repo-knowledge", "repository.json")
	config, err := readJSON[RepositoryConfig](configPath, true)
	if err != nil {
		return nil, RepositoryConfig{}, err
	}
	indexPath, err := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if err != nil {
		return nil, RepositoryConfig{}, err
	}
	paths := map[string]string{
		"contract":   filepath.Join(root, ".repo-knowledge", "policy", "contract.json"),
		"repository": configPath,
		"index":      indexPath,
	}
	digests := make(map[string]string, len(paths))
	for name, path := range paths {
		digest, err := sha256File(path)
		if err != nil {
			return nil, RepositoryConfig{}, fmt.Errorf("hash preflight %s: %w", name, err)
		}
		digests[name] = digest
	}
	return digests, config, nil
}

func startOrResumePreflight(root, agent, conversation string) (preflightSession, error) {
	digests, _, err := preflightDigests(root)
	if err != nil {
		return preflightSession{}, err
	}
	path, err := preflightSessionPath(root, agent, conversation)
	if err != nil {
		return preflightSession{}, err
	}
	if existing, err := readPreflightSession(path); err == nil && validPreflightSession(existing, root, agent, digests) {
		return existing, nil
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return preflightSession{}, err
	}
	token, err := newPreflightToken()
	if err != nil {
		return preflightSession{}, err
	}
	now := preflightClock().UTC()
	session := preflightSession{
		SchemaVersion: "1.0",
		Token:         token,
		Root:          root,
		Agent:         agent,
		Conversation:  conversation,
		CreatedAt:     now,
		ExpiresAt:     now.Add(preflightSessionTTL),
		Digests:       digests,
	}
	if err := writePreflightSession(path, session); err != nil {
		return preflightSession{}, err
	}
	return session, nil
}

func newPreflightToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate preflight token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func readPreflightSession(path string) (preflightSession, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return preflightSession{}, err
	}
	var session preflightSession
	if err := json.Unmarshal(data, &session); err != nil {
		return preflightSession{}, fmt.Errorf("parse preflight state: %w", err)
	}
	return session, nil
}

func writePreflightSession(path string, session preflightSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("encode preflight state: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create preflight state parent: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".repo-knowledge-preflight-*")
	if err != nil {
		return fmt.Errorf("create temporary preflight state: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary preflight state: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure temporary preflight state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary preflight state: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace preflight state: %w", err)
	}
	return nil
}

func validPreflightSession(session preflightSession, root, agent string, digests map[string]string) bool {
	if session.SchemaVersion != "1.0" || session.Root != root || session.Agent != agent || session.Token == "" || !session.ExpiresAt.After(preflightClock().UTC()) {
		return false
	}
	if len(session.Digests) != len(digests) {
		return false
	}
	for name, value := range digests {
		if session.Digests[name] != value {
			return false
		}
	}
	return true
}

func preflightSessionByToken(token string) (preflightSession, string, error) {
	directory, err := preflightStateDirectory()
	if err != nil {
		return preflightSession{}, "", err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return preflightSession{}, "", fmt.Errorf("list preflight state: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		session, err := readPreflightSession(path)
		if err != nil {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(session.Token), []byte(token)) == 1 {
			return session, path, nil
		}
	}
	return preflightSession{}, "", fmt.Errorf("unknown preflight token")
}

func PreflightActivate(target, token string, routes []string) (PreflightActivationResult, error) {
	root, err := findInstalledRepositoryRoot(target)
	if err != nil {
		return PreflightActivationResult{}, err
	}
	session, statePath, err := preflightSessionByToken(token)
	if err != nil {
		return PreflightActivationResult{}, err
	}
	if session.Root != root {
		return PreflightActivationResult{}, fmt.Errorf("preflight token belongs to a different repository")
	}
	digests, config, err := preflightDigests(root)
	if err != nil {
		return PreflightActivationResult{}, err
	}
	if !validPreflightSession(session, root, session.Agent, digests) {
		return PreflightActivationResult{}, fmt.Errorf("preflight token expired or repository knowledge changed; start a new task invocation")
	}
	validated, err := validatePreflightRoutes(root, config, routes)
	if err != nil {
		return PreflightActivationResult{}, err
	}
	session.Active = true
	session.Routes = validated
	if err := writePreflightSession(statePath, session); err != nil {
		return PreflightActivationResult{}, err
	}
	return PreflightActivationResult{Status: "active", Root: root, Routes: validated}, nil
}

func validatePreflightRoutes(root string, config RepositoryConfig, routes []string) ([]string, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("at least one --route is required")
	}
	indexPath, err := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if err != nil {
		return nil, err
	}
	documentationRoot := filepath.Dir(indexPath)
	seen := make(map[string]struct{}, len(routes))
	validated := make([]string, 0, len(routes))
	for _, route := range routes {
		path, err := repositoryPath(root, route, "preflight route")
		if err != nil {
			return nil, err
		}
		if !pathWithin(documentationRoot, path) || !regularFile(path) {
			return nil, fmt.Errorf("preflight route must be an existing documentation file below %s: %q", filepath.ToSlash(filepath.Dir(config.Documentation.Index)), route)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, fmt.Errorf("resolve preflight route %q: %w", route, err)
		}
		relative = filepath.ToSlash(relative)
		if _, exists := seen[relative]; exists {
			continue
		}
		seen[relative] = struct{}{}
		validated = append(validated, relative)
	}
	sort.Strings(validated)
	return validated, nil
}

func pathWithin(parent, child string) bool {
	relative, err := filepath.Rel(parent, child)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func parseAntigravityHookEvent(input io.Reader) (antigravityHookEvent, error) {
	var event antigravityHookEvent
	if input == nil {
		return event, nil
	}
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&event); err != nil && err != io.EOF {
		return antigravityHookEvent{}, fmt.Errorf("parse Antigravity hook input: %w", err)
	}
	return event, nil
}

func parsePreflightHookEvent(agent string, input io.Reader) (preflightHookEvent, error) {
	if input == nil {
		return preflightHookEvent{}, nil
	}
	var raw struct {
		SessionID      string          `json:"session_id"`
		ConversationID string          `json:"conversation_id"`
		InvocationNum  int             `json:"invocationNum"`
		ToolName       string          `json:"tool_name"`
		ToolInput      json.RawMessage `json:"tool_input"`
		AGConversation string          `json:"conversationId"`
		ToolCall       struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		} `json:"toolCall"`
	}
	decoder := json.NewDecoder(input)
	if err := decoder.Decode(&raw); err != nil && err != io.EOF {
		return preflightHookEvent{}, fmt.Errorf("parse %s hook input: %w", displayAgentName(agent), err)
	}
	event := preflightHookEvent{InvocationNum: raw.InvocationNum, ToolName: raw.ToolName, ToolInput: raw.ToolInput}
	switch agent {
	case agentAntigravityIDE:
		event.ConversationID, event.ToolName, event.ToolInput = raw.AGConversation, raw.ToolCall.Name, raw.ToolCall.Arguments
	case agentCodex, agentClaudeCode:
		event.ConversationID = raw.SessionID
	case agentCursor:
		event.ConversationID = raw.ConversationID
	}
	return event, nil
}

func displayAgentName(agent string) string {
	if agent == agentAntigravityIDE {
		return "Antigravity"
	}
	return agent
}

func PreflightGate(target, agent string, input io.Reader) (PreflightGateResult, error) {
	adapters, err := normalizeAgentAdapters([]string{agent})
	if err != nil {
		return PreflightGateResult{}, err
	}
	agent = adapters[0]
	root, err := findInstalledRepositoryRoot(target)
	if err != nil {
		return PreflightGateResult{}, err
	}
	manifest, _, err := readOptionalManifest(filepath.Join(root, ".repo-knowledge", "toolkit.json"))
	if err != nil {
		return PreflightGateResult{}, err
	}
	if preflightMode(manifest, agent) != AntigravityPreflightStrict {
		return preflightGateResult(agent, root, "allow", ""), nil
	}
	event, err := parsePreflightHookEvent(agent, input)
	if err != nil {
		return PreflightGateResult{}, err
	}
	digests, _, digestErr := preflightDigests(root)
	statePath, pathErr := preflightSessionPath(root, agent, event.ConversationID)
	session, stateErr := readPreflightSession(statePath)
	if digestErr == nil && pathErr == nil && stateErr == nil && session.Active && validPreflightSession(session, root, agent, digests) {
		return preflightGateResult(agent, root, "allow", ""), nil
	}
	if preflightPendingToolAllowed(root, agent, event) {
		return preflightGateResult(agent, root, "allow", ""), nil
	}
	return preflightGateResult(agent, root, "deny", "Repository Knowledge preflight is pending. Read the installed skill, contract, repository configuration, and documentation routes, then run the injected repo-knowledge preflight-activate command before repository discovery or changes."), nil
}

func preflightGateResult(agent, root, decision, reason string) PreflightGateResult {
	payload := any(map[string]string{"decision": decision})
	switch agent {
	case agentCodex, agentClaudeCode:
		output := map[string]any{"hookSpecificOutput": map[string]string{"hookEventName": "PreToolUse", "permissionDecision": decision}}
		if reason != "" {
			output["hookSpecificOutput"].(map[string]string)["permissionDecisionReason"] = reason
		}
		payload = output
	case agentCursor:
		output := map[string]string{"permission": decision}
		if reason != "" {
			output["user_message"], output["agent_message"] = reason, reason
		}
		payload = output
	default:
		output := map[string]string{"decision": decision}
		if reason != "" {
			output["reason"] = reason
		}
		payload = output
	}
	encoded, _ := json.Marshal(payload)
	return PreflightGateResult{Agent: agent, Root: root, Decision: decision, Reason: reason, Output: string(encoded)}
}

func preflightPendingToolAllowed(root, agent string, event preflightHookEvent) bool {
	name := strings.ToLower(strings.TrimSpace(event.ToolName))
	switch name {
	case "search_web", "read_url_content", "websearch", "webfetch":
		return true
	case "run_command", "bash", "shell", "exec_command":
		return activationCommandAllowed(event.ToolInput)
	case "view_file", "read_file", "list_dir", "read":
		return knowledgePathReadAllowed(root, event.ToolInput)
	default:
		return false
	}
}

func activationCommandAllowed(arguments json.RawMessage) bool {
	var values map[string]any
	if json.Unmarshal(arguments, &values) != nil {
		return false
	}
	command, _ := values["command"].(string)
	if strings.ContainsAny(command, ";|&$`()<>\n\r") {
		return false
	}
	parts := strings.Fields(command)
	if len(parts) < 6 || parts[0] != "repo-knowledge" || parts[1] != "preflight-activate" {
		return false
	}
	hasToken := false
	hasRoute := false
	for index := 2; index < len(parts); index += 2 {
		if index+1 >= len(parts) || strings.HasPrefix(parts[index+1], "--") {
			return false
		}
		switch parts[index] {
		case "--token":
			hasToken = true
		case "--route":
			hasRoute = true
		case "--target":
		default:
			return false
		}
	}
	return hasToken && hasRoute
}

func knowledgePathReadAllowed(root string, arguments json.RawMessage) bool {
	var values map[string]any
	if json.Unmarshal(arguments, &values) != nil {
		return false
	}
	path, _ := values["path"].(string)
	if strings.TrimSpace(path) == "" {
		path, _ = values["filePath"].(string)
	}
	if strings.TrimSpace(path) == "" {
		path, _ = values["file_path"].(string)
	}
	if strings.TrimSpace(path) == "" {
		return false
	}
	candidate, err := preflightPath(root, path)
	if err != nil {
		return false
	}
	config, err := readJSON[RepositoryConfig](filepath.Join(root, ".repo-knowledge", "repository.json"), true)
	if err != nil {
		return false
	}
	indexPath, err := repositoryPath(root, config.Documentation.Index, "documentation.index")
	if err != nil {
		return false
	}
	allowed := []string{
		filepath.Join(root, ".repo-knowledge"),
		filepath.Join(root, ".agents", "skills", "repository-knowledge"),
		filepath.Dir(indexPath),
	}
	for _, directory := range allowed {
		if pathWithin(directory, candidate) {
			return true
		}
	}
	return false
}

func preflightPath(root, value string) (string, error) {
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return repositoryPath(root, value, "preflight tool path")
}
