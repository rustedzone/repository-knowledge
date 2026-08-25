package toolkit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TrackedEntry struct {
	Path     string
	Identity string
}

type ModuleInventory struct {
	Path        string `json:"path"`
	FileCount   int    `json:"file_count"`
	Fingerprint string `json:"fingerprint"`
}

type CapabilitySignal struct {
	ID         string   `json:"id"`
	Confidence string   `json:"confidence"`
	DetectedBy []string `json:"detected_by"`
}

type ScanState struct {
	SchemaVersion     string             `json:"schema_version"`
	GeneratedAt       string             `json:"generated_at"`
	ScannedCommit     string             `json:"scanned_commit,omitempty"`
	FileCount         int                `json:"file_count"`
	Manifests         []string           `json:"manifests"`
	RootFiles         []string           `json:"root_files"`
	Documentation     []string           `json:"documentation"`
	Modules           []ModuleInventory  `json:"modules"`
	CapabilitySignals []CapabilitySignal `json:"capability_signals"`
	Limitations       []string           `json:"limitations"`
}

var manifestPatterns = []string{
	"package.json", "pyproject.toml", "requirements.txt", "go.mod", "Cargo.toml",
	"pom.xml", "build.gradle", "build.gradle.kts", "composer.json", "Gemfile",
	"mix.exs", "Package.swift", "*.csproj", "*.sln", "Dockerfile",
	"docker-compose.yml", "docker-compose.yaml", "Chart.yaml", "terraform.tf",
}

var capabilitySignals = map[string][]string{
	"api":                          {"api", "openapi", "controller", "endpoint", "graphql", "proto"},
	"ui":                           {"ui", "page", "pages", "screen", "screens", "route", "routes", "view", "views"},
	"persistence":                  {"migration", "migrations", "schema", "database", "persistence"},
	"authentication":               {"auth", "authentication", "login", "session"},
	"authorization":                {"authorization", "permission", "permissions", "role", "roles"},
	"integrations":                 {"integration", "integrations", "client", "clients", "webhook", "webhooks"},
	"events_or_queues":             {"event", "events", "queue", "queues", "worker", "workers"},
	"scheduled_jobs":               {"cron", "schedule", "scheduled", "jobs"},
	"configuration":                {"config", "configuration", "environment", "env"},
	"deployment_or_infrastructure": {"deploy", "deployment", "infra", "terraform", "kubernetes", "helm"},
	"observability":                {"observability", "metric", "metrics", "trace", "tracing", "logging", "alerts"},
	"tests":                        {"test", "tests", "spec", "specs"},
	"packages_or_modules":          {"package", "packages", "module", "modules", "libs", "libraries"},
}

func Scan(root string, write bool) (ScanState, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ScanState{}, fmt.Errorf("resolve repository root: %w", err)
	}
	entries, err := trackedEntries(root)
	if err != nil {
		return ScanState{}, err
	}
	filtered := entries[:0]
	for _, entry := range entries {
		value := filepath.ToSlash(entry.Path)
		if isToolkitManagedKnowledgePath(value) {
			continue
		}
		filtered = append(filtered, entry)
	}
	entries = filtered

	modules := make(map[string][]TrackedEntry)
	rootFiles := make([]string, 0)
	tokens := make(map[string]map[string]struct{})
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		value := filepath.ToSlash(entry.Path)
		paths = append(paths, value)
		module := structuralModule(value)
		if module == "" {
			rootFiles = append(rootFiles, value)
		} else {
			modules[module] = append(modules[module], entry)
		}
		for _, part := range strings.Split(value, "/") {
			token := strings.ToLower(strings.SplitN(part, ".", 2)[0])
			if tokens[token] == nil {
				tokens[token] = make(map[string]struct{})
			}
			tokens[token][value] = struct{}{}
		}
	}

	moduleInventory := make([]ModuleInventory, 0, len(modules))
	for _, name := range sortedKeys(modules) {
		members := modules[name]
		sort.Slice(members, func(i, j int) bool { return members[i].Path < members[j].Path })
		digest := sha256.New()
		for _, member := range members {
			_, _ = fmt.Fprintf(digest, "%s\x00%s\n", filepath.ToSlash(member.Path), member.Identity)
		}
		moduleInventory = append(moduleInventory, ModuleInventory{
			Path:        name,
			FileCount:   len(members),
			Fingerprint: hex.EncodeToString(digest.Sum(nil)),
		})
	}

	detected := make([]CapabilitySignal, 0)
	for _, capability := range sortedKeys(capabilitySignals) {
		evidence := make(map[string]struct{})
		for _, signal := range capabilitySignals[capability] {
			for value := range tokens[signal] {
				evidence[value] = struct{}{}
			}
		}
		if len(evidence) == 0 {
			continue
		}
		values := sortedKeys(evidence)
		if len(values) > 12 {
			values = values[:12]
		}
		detected = append(detected, CapabilitySignal{
			ID: capability, Confidence: "discovery_signal_only", DetectedBy: values,
		})
	}

	manifests := make([]string, 0)
	documentation := make([]string, 0)
	for _, value := range paths {
		if isManifest(value) {
			manifests = append(manifests, value)
		}
		if value == "README.md" || strings.HasPrefix(value, "docs/") {
			documentation = append(documentation, value)
		}
	}
	sort.Strings(manifests)
	sort.Strings(rootFiles)
	sort.Strings(documentation)
	state := ScanState{
		SchemaVersion:     "1.0",
		GeneratedAt:       utcNow(),
		ScannedCommit:     currentCommit(root),
		FileCount:         len(entries),
		Manifests:         manifests,
		RootFiles:         rootFiles,
		Documentation:     documentation,
		Modules:           moduleInventory,
		CapabilitySignals: detected,
		Limitations: []string{
			"capability signals are routing hints, not verified architectural claims",
			"semantic behavior and runtime truth require targeted evidence inspection",
		},
	}
	if write {
		output, err := repositoryPath(root, ".repo-knowledge/scan-state.json", "scan state")
		if err != nil {
			return ScanState{}, err
		}
		if err := writeJSON(output, state); err != nil {
			return ScanState{}, err
		}
	}
	return state, nil
}

func structuralModule(value string) string {
	parts := strings.Split(filepath.ToSlash(value), "/")
	if len(parts) < 2 {
		return ""
	}
	if len(parts) >= 3 && isSourceContainer(parts[0]) {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

func isSourceContainer(value string) bool {
	switch value {
	case "app", "apps", "cmd", "internal", "lib", "libs", "modules", "packages", "pkg", "services", "src":
		return true
	default:
		return false
	}
}

func isToolkitManagedKnowledgePath(value string) bool {
	return strings.HasPrefix(value, ".repo-knowledge/") ||
		strings.HasPrefix(value, ".agents/skills/repository-knowledge/") ||
		strings.HasPrefix(value, ".claude/skills/repository-knowledge/") ||
		strings.HasPrefix(value, ".cursor/skills/repository-knowledge/") ||
		value == ".agents/rules/repository-knowledge.md" ||
		value == ".claude/rules/repository-knowledge.md" ||
		value == ".cursor/rules/repository-knowledge.mdc"
}

func trackedEntries(root string) ([]TrackedEntry, error) {
	if !isGitRepository(root) {
		return filesystemEntries(root)
	}
	output, err := runGit(root, true, "ls-files", "-s", "-z")
	if err != nil {
		return nil, err
	}
	matcher := regexp.MustCompile(`^\d+ ([a-f0-9]+) \d+\t(.+)$`)
	entries := make([]TrackedEntry, 0)
	tracked := make(map[string]struct{})
	for _, record := range strings.Split(output, "\x00") {
		if record == "" {
			continue
		}
		matches := matcher.FindStringSubmatch(record)
		if len(matches) != 3 {
			continue
		}
		entries = append(entries, TrackedEntry{Path: matches[2], Identity: matches[1]})
		tracked[matches[2]] = struct{}{}
	}
	untracked, err := runGit(root, true, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	for _, value := range strings.Split(untracked, "\x00") {
		if value == "" {
			continue
		}
		if _, ok := tracked[value]; ok {
			continue
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(value)))
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		identity := fmt.Sprintf("untracked:%d:%d", info.Size(), info.ModTime().UnixNano())
		entries = append(entries, TrackedEntry{Path: value, Identity: identity})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func filesystemEntries(root string) ([]TrackedEntry, error) {
	entries := make([]TrackedEntry, 0)
	err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, filePath)
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		identity := fmt.Sprintf("filesystem:%d:%d", info.Size(), info.ModTime().UnixNano())
		entries = append(entries, TrackedEntry{Path: filepath.ToSlash(relative), Identity: identity})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("inventory filesystem: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func isManifest(value string) bool {
	name := path.Base(filepath.ToSlash(value))
	for _, pattern := range manifestPatterns {
		if matched, _ := path.Match(pattern, name); matched {
			return true
		}
	}
	return false
}
