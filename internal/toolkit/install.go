package toolkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
)

const (
	agentCodex          = "codex"
	agentClaudeCode     = "claude-code"
	agentAntigravityIDE = "antigravity-ide"
	agentCursor         = "cursor"
)

var agentAdapterAliases = map[string]string{
	"codex":           agentCodex,
	"claude":          agentClaudeCode,
	"claude-code":     agentClaudeCode,
	"antigravity":     agentAntigravityIDE,
	"antigravity-ide": agentAntigravityIDE,
	"cursor":          agentCursor,
}

var supportedAgentAdapters = []string{
	agentCodex,
	agentClaudeCode,
	agentAntigravityIDE,
	agentCursor,
}

func Install(options InstallOptions) (InstallResult, error) {
	result := InstallResult{}
	target, err := filepath.Abs(options.Target)
	if err != nil {
		return result, fmt.Errorf("resolve target repository: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return result, fmt.Errorf("target repository does not exist: %s", target)
	}
	options.Target = target
	manifestPath, err := repositoryPath(target, ".repo-knowledge/toolkit.json", "toolkit manifest")
	if err != nil {
		return result, err
	}
	oldManifest, manifestExists, err := readOptionalManifest(manifestPath)
	if err != nil {
		return result, err
	}
	if options.Update && !manifestExists {
		return result, fmt.Errorf("cannot update: toolkit is not installed; run install first")
	}
	if manifestExists && !options.Update {
		return result, fmt.Errorf("toolkit is already installed; run update with the desired binary")
	}
	if options.AllAgentAdapters && len(options.AgentAdapters) > 0 {
		return result, fmt.Errorf("--all-agents cannot be combined with --agent")
	}
	if options.AllAgentAdapters {
		options.AgentAdapters = append([]string(nil), supportedAgentAdapters...)
	} else if len(options.AgentAdapters) == 0 {
		if options.Update && len(oldManifest.AgentAdapters) > 0 {
			options.AgentAdapters = append([]string(nil), oldManifest.AgentAdapters...)
		} else {
			options.AgentAdapters = []string{agentCodex}
		}
	}
	options.AgentAdapters, err = normalizeAgentAdapters(options.AgentAdapters)
	if err != nil {
		return result, err
	}
	if options.Update && options.Source == "" {
		options.Source = oldManifest.Source
	}
	if options.Update && options.Ref == "" {
		options.Ref = oldManifest.Ref
	}
	if options.Update && !options.AllowDowngrade && compareVersions(repositoryknowledge.Version(), oldManifest.ToolkitVersion) < 0 {
		return result, fmt.Errorf("refusing to downgrade toolkit from %s to %s without --allow-downgrade", oldManifest.ToolkitVersion, repositoryknowledge.Version())
	}
	options.AntigravityPreflight, err = resolveAntigravityPreflight(options, oldManifest)
	if err != nil {
		return result, err
	}
	hookRegistrations, hookMutations, err := prepareAgentHooks(target, oldManifest.AgentAdapters, options.AgentAdapters, options.Update, options.AntigravityPreflight)
	if err != nil {
		return result, err
	}

	managed, err := installEmbeddedManagedFiles(target, options.AgentAdapters)
	if err != nil {
		return result, err
	}
	if err := applyAgentHooks(hookMutations); err != nil {
		return result, err
	}
	removed, preserved, err := removeObsoleteManagedFiles(target, oldManifest, managed, options.Update)
	if err != nil {
		return result, err
	}

	if contains(options.AgentAdapters, agentCodex) {
		if err := installManagedAgentsBlock(target); err != nil {
			return result, err
		}
	} else if options.Update && contains(oldManifest.AgentAdapters, agentCodex) {
		if err := removeManagedAgentsBlock(target); err != nil {
			return result, err
		}
	}
	created, err := installConsumerOwnedDefaults(target)
	if err != nil {
		return result, err
	}

	managedFiles := make([]ManagedFile, 0, len(managed))
	for _, relative := range managed {
		path, err := repositoryPath(target, relative, "managed file path")
		if err != nil {
			return result, err
		}
		digest, err := sha256File(path)
		if err != nil {
			return result, err
		}
		managedFiles = append(managedFiles, ManagedFile{Path: relative, SHA256: digest})
	}
	manifest := ToolkitManifest{
		SchemaVersion:        "1.0",
		ToolkitVersion:       repositoryknowledge.Version(),
		Source:               options.Source,
		Ref:                  options.Ref,
		InstalledAt:          utcNow(),
		AgentAdapters:        append([]string(nil), options.AgentAdapters...),
		AntigravityPreflight: options.AntigravityPreflight,
		ManagedFiles:         managedFiles,
		Ownership: map[string]string{
			"managed_files":                  "replaced by repo-knowledge update",
			"managed_codex_agents_block":     "replaced in place; other AGENTS.md content is preserved",
			"managed_agent_hook_entries":     "merged in place; unrelated hook configuration is preserved",
			"all_other_repository_knowledge": "owned by the consuming repository",
		},
	}
	if err := writeJSON(manifestPath, manifest); err != nil {
		return result, err
	}
	action := "installed"
	if options.Update {
		action = "updated"
	}
	return InstallResult{
		Action:                         action,
		Target:                         target,
		Version:                        repositoryknowledge.Version(),
		ManagedFileCount:               len(managedFiles),
		RepositoryOwnedFilesCreated:    created,
		RepositoryOwnedFilesPreserved:  manifestExists,
		ObsoleteManagedFilesRemoved:    removed,
		ModifiedObsoleteFilesPreserved: preserved,
		AgentHookRegistrations:         hookRegistrations,
	}, nil
}

func resolveAntigravityPreflight(options InstallOptions, old ToolkitManifest) (string, error) {
	mode := strings.TrimSpace(options.AntigravityPreflight)
	if mode == "" && options.Update {
		mode = old.AntigravityPreflight
	}
	if mode == "" {
		mode = AntigravityPreflightObserve
	}
	if mode != AntigravityPreflightObserve && mode != AntigravityPreflightStrict {
		return "", fmt.Errorf("antigravity preflight mode must be %q or %q", AntigravityPreflightObserve, AntigravityPreflightStrict)
	}
	if mode == AntigravityPreflightStrict && !contains(options.AgentAdapters, agentAntigravityIDE) {
		return "", fmt.Errorf("antigravity strict preflight requires the antigravity-ide adapter")
	}
	return mode, nil
}

func readOptionalManifest(path string) (ToolkitManifest, bool, error) {
	manifest, err := readJSON[ToolkitManifest](path, false)
	if err != nil {
		return ToolkitManifest{}, false, err
	}
	if manifest.SchemaVersion == "" {
		return ToolkitManifest{}, false, nil
	}
	return manifest, true, nil
}

func installEmbeddedManagedFiles(target string, adapters []string) ([]string, error) {
	assets := make(map[string]string)
	for _, directory := range []string{"policy", "schemas"} {
		err := fs.WalkDir(repositoryknowledge.Content, directory, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				assets[".repo-knowledge/"+path] = path
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("enumerate embedded %s assets: %w", directory, err)
		}
	}
	usesAgentsSkill := contains(adapters, agentCodex) || contains(adapters, agentAntigravityIDE)
	usesClaudeSkill := contains(adapters, agentClaudeCode)
	usesCursorSkill := contains(adapters, agentCursor)
	if usesAgentsSkill || usesClaudeSkill || usesCursorSkill {
		err := fs.WalkDir(repositoryknowledge.Content, "skills/repository-knowledge", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() {
				if usesAgentsSkill {
					assets[".agents/"+path] = path
				}
				if usesClaudeSkill {
					assets[".claude/"+path] = path
				}
				if usesCursorSkill {
					assets[".cursor/"+path] = path
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("enumerate embedded repository-knowledge skill: %w", err)
		}
	}
	if contains(adapters, agentClaudeCode) {
		assets[".claude/rules/repository-knowledge.md"] = "templates/repository-knowledge-rule.md"
	}
	if contains(adapters, agentAntigravityIDE) {
		assets[".agents/rules/repository-knowledge.md"] = "templates/repository-knowledge-rule.md"
	}
	if contains(adapters, agentCursor) {
		assets[".cursor/rules/repository-knowledge.mdc"] = "templates/repository-knowledge-cursor-rule.mdc"
	}

	managed := sortedKeys(assets)
	for _, destination := range managed {
		source := assets[destination]
		data, err := repositoryknowledge.Content.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("read embedded asset %s: %w", source, err)
		}
		path, err := repositoryPath(target, destination, "managed asset path")
		if err != nil {
			return nil, err
		}
		mode := fs.FileMode(0o644)
		if strings.HasPrefix(source, "skills/repository-knowledge/scripts/") {
			mode = 0o755
		}
		if err := writeFileAtomic(path, data, mode); err != nil {
			return nil, err
		}
	}
	return managed, nil
}

func removeObsoleteManagedFiles(target string, old ToolkitManifest, managed []string, update bool) ([]string, []string, error) {
	if !update {
		return nil, nil, nil
	}
	current := make(map[string]struct{}, len(managed))
	for _, path := range managed {
		current[path] = struct{}{}
	}
	var removed []string
	var preserved []string
	for _, item := range old.ManagedFiles {
		if _, ok := current[item.Path]; ok {
			continue
		}
		path, err := repositoryPath(target, item.Path, "managed file path")
		if err != nil {
			return nil, nil, err
		}
		digest, err := sha256File(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, nil, err
		}
		if digest != item.SHA256 {
			preserved = append(preserved, item.Path)
			continue
		}
		if err := os.Remove(path); err != nil {
			return nil, nil, fmt.Errorf("remove obsolete managed file %s: %w", item.Path, err)
		}
		removed = append(removed, item.Path)
	}
	sort.Strings(removed)
	sort.Strings(preserved)
	return removed, preserved, nil
}

func installManagedAgentsBlock(target string) error {
	template, err := repositoryknowledge.Content.ReadFile("templates/AGENTS.md")
	if err != nil {
		return fmt.Errorf("read embedded AGENTS template: %w", err)
	}
	path, err := repositoryPath(target, "AGENTS.md", "AGENTS.md")
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}
	merged, err := mergeManagedAgents(string(existing), string(template))
	if err != nil {
		return err
	}
	return writeFileAtomic(path, []byte(merged), 0o644)
}

func removeManagedAgentsBlock(target string) error {
	path, err := repositoryPath(target, "AGENTS.md", "AGENTS.md")
	if err != nil {
		return err
	}
	existing, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	withoutManaged, found, err := removeManagedBlock(string(existing))
	if err != nil {
		return fmt.Errorf("remove managed AGENTS.md block: %w", err)
	}
	if !found {
		return nil
	}
	if strings.TrimSpace(withoutManaged) == "" {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove empty managed AGENTS.md: %w", err)
		}
		return nil
	}
	return writeFileAtomic(path, []byte(withoutManaged), 0o644)
}

func removeManagedBlock(existing string) (string, bool, error) {
	start := strings.Index(existing, ManagedBegin)
	if start < 0 {
		return existing, false, nil
	}
	end := strings.Index(existing[start:], ManagedEnd)
	if end < 0 {
		return "", false, fmt.Errorf("unterminated repository-knowledge managed block")
	}
	end += start + len(ManagedEnd)
	before := strings.TrimRight(existing[:start], "\n")
	after := strings.TrimLeft(existing[end:], "\n")
	switch {
	case before == "" && after == "":
		return "", true, nil
	case before == "":
		return after, true, nil
	case after == "":
		return before + "\n", true, nil
	default:
		return before + "\n\n" + after, true, nil
	}
}

func mergeManagedAgents(existing, managed string) (string, error) {
	start := strings.Index(managed, ManagedBegin)
	end := strings.Index(managed, ManagedEnd)
	if start < 0 || end < start {
		return "", fmt.Errorf("embedded AGENTS template has invalid managed markers")
	}
	block := managed[start : end+len(ManagedEnd)]
	if existingStart := strings.Index(existing, ManagedBegin); existingStart >= 0 {
		existingEnd := strings.Index(existing[existingStart:], ManagedEnd)
		if existingEnd < 0 {
			return "", fmt.Errorf("existing AGENTS.md has an unterminated repository-knowledge managed block")
		}
		existingEnd += existingStart + len(ManagedEnd)
		before := strings.TrimRight(existing[:existingStart], "\n")
		after := strings.TrimLeft(existing[existingEnd:], "\n")
		if before == "" {
			return block + "\n" + after, nil
		}
		return before + "\n\n" + block + "\n" + after, nil
	}
	if strings.TrimSpace(existing) == "" {
		return block + "\n", nil
	}
	return strings.TrimRight(existing, "\n") + "\n\n" + block + "\n", nil
}

func installConsumerOwnedDefaults(target string) ([]string, error) {
	created := make([]string, 0, 4)
	repositoryConfig := ".repo-knowledge/repository.json"
	configPath, err := repositoryPath(target, repositoryConfig, "repository configuration")
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		data, readErr := repositoryknowledge.Content.ReadFile("templates/repository.json")
		if readErr != nil {
			return nil, fmt.Errorf("read repository template: %w", readErr)
		}
		var config map[string]any
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("parse embedded repository template: %w", err)
		}
		repository, ok := config["repository"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("embedded repository template has no repository object")
		}
		repository["name"] = filepath.Base(target)
		if err := writeJSON(configPath, config); err != nil {
			return nil, err
		}
		created = append(created, repositoryConfig)
	} else if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", configPath, err)
	}

	for _, name := range []string{"local-invariants.json", "local-impact-rules.json"} {
		relative := ".repo-knowledge/" + name
		path, err := repositoryPath(target, relative, "consumer-owned default")
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			data, readErr := repositoryknowledge.Content.ReadFile("templates/" + name)
			if readErr != nil {
				return nil, fmt.Errorf("read embedded template %s: %w", name, readErr)
			}
			if err := writeFileAtomic(path, data, 0o644); err != nil {
				return nil, err
			}
			created = append(created, relative)
		} else if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", path, err)
		}
	}

	index := "docs/index.md"
	indexPath, err := repositoryPath(target, index, "documentation index")
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		data, readErr := repositoryknowledge.Content.ReadFile("templates/docs-index.md")
		if readErr != nil {
			return nil, fmt.Errorf("read embedded docs index: %w", readErr)
		}
		if err := writeFileAtomic(indexPath, data, 0o644); err != nil {
			return nil, err
		}
		created = append(created, index)
	} else if err != nil {
		return nil, fmt.Errorf("inspect %s: %w", indexPath, err)
	}
	sort.Strings(created)
	return created, nil
}

func compareVersions(left, right string) int {
	type semanticVersion struct {
		core       [3]int
		prerelease []string
	}
	parse := func(value string) semanticVersion {
		value = strings.TrimPrefix(strings.TrimSpace(value), "v")
		value = strings.SplitN(value, "+", 2)[0]
		parts := strings.SplitN(value, "-", 2)
		segments := strings.Split(parts[0], ".")
		result := semanticVersion{}
		for index := range result.core {
			if index < len(segments) {
				result.core[index], _ = strconv.Atoi(segments[index])
			}
		}
		if len(parts) == 2 {
			result.prerelease = strings.Split(parts[1], ".")
		}
		return result
	}
	a := parse(left)
	b := parse(right)
	for index := range a.core {
		if a.core[index] < b.core[index] {
			return -1
		}
		if a.core[index] > b.core[index] {
			return 1
		}
	}
	if len(a.prerelease) == 0 && len(b.prerelease) > 0 {
		return 1
	}
	if len(a.prerelease) > 0 && len(b.prerelease) == 0 {
		return -1
	}
	for index := 0; index < len(a.prerelease) && index < len(b.prerelease); index++ {
		leftIdentifier := a.prerelease[index]
		rightIdentifier := b.prerelease[index]
		leftNumber := isNumericIdentifier(leftIdentifier)
		rightNumber := isNumericIdentifier(rightIdentifier)
		switch {
		case leftNumber && rightNumber:
			if result := compareNumericIdentifiers(leftIdentifier, rightIdentifier); result != 0 {
				return result
			}
		case leftNumber:
			return -1
		case rightNumber:
			return 1
		case leftIdentifier < rightIdentifier:
			return -1
		case leftIdentifier > rightIdentifier:
			return 1
		}
	}
	if len(a.prerelease) < len(b.prerelease) {
		return -1
	}
	if len(a.prerelease) > len(b.prerelease) {
		return 1
	}
	return 0
}

func isNumericIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func compareNumericIdentifiers(left, right string) int {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if left == "" {
		left = "0"
	}
	if right == "" {
		right = "0"
	}
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func normalizeAgentAdapters(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		adapter, ok := agentAdapterAliases[strings.ToLower(strings.TrimSpace(value))]
		if !ok {
			return nil, fmt.Errorf("unsupported agent adapter %q; supported adapters: %s", value, strings.Join(supportedAgentAdapters, ", "))
		}
		if _, exists := seen[adapter]; exists {
			continue
		}
		seen[adapter] = struct{}{}
		normalized = append(normalized, adapter)
	}
	return normalized, nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
