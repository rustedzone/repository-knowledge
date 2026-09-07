package evalharness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func Grade(options GradeOptions) (GradeResult, error) {
	var result GradeResult
	if options.DurationMillis < 0 {
		return result, fmt.Errorf("duration cannot be negative")
	}
	if options.TokenUsage != nil && *options.TokenUsage < 0 {
		return result, fmt.Errorf("token usage cannot be negative")
	}
	usage, legacyTotal, err := validateTokenUsage(options.Usage, options.TokenUsage)
	if err != nil {
		return result, err
	}
	semanticStatus, err := validateSemanticResult(options.SemanticStatus, options.SemanticScore, options.SemanticReviewer)
	if err != nil {
		return result, err
	}
	spec, caseRoot, err := LoadSpec(options.CasesRoot, options.CaseID)
	if err != nil {
		return result, err
	}
	target, err := filepath.Abs(options.Target)
	if err != nil {
		return result, fmt.Errorf("resolve evaluation target: %w", err)
	}
	baseline, err := readBaseline(target, spec.Family)
	if err != nil {
		return result, err
	}
	if baseline.CaseID != spec.ID || baseline.CaseRevision != spec.Revision {
		return result, fmt.Errorf("baseline is for %s revision %s, not %s revision %s", baseline.CaseID, baseline.CaseRevision, spec.ID, spec.Revision)
	}
	if baseline.Family != spec.Family {
		return result, fmt.Errorf("baseline family is %s, not %s", baseline.Family, spec.Family)
	}
	result = GradeResult{
		SchemaVersion: "1.0", Family: baseline.Family, Condition: baseline.Condition,
		CaseID: spec.ID, Revision: spec.Revision, SourceCommit: baseline.SourceCommit, Target: target,
		Agent: baseline.Agent, AgentVersion: baseline.AgentVersion, ModelVersion: baseline.ModelVersion,
		ReasoningConfiguration:      baseline.ReasoningConfiguration,
		RepositoryKnowledgeRevision: baseline.RepositoryKnowledgeRevision,
		TrialNumber:                 baseline.TrialNumber, DurationMillis: options.DurationMillis, TokenUsage: legacyTotal, Usage: usage,
		SemanticStatus: semanticStatus, SemanticScore: options.SemanticScore,
		SemanticReviewer: strings.TrimSpace(options.SemanticReviewer), OverallStatus: "pending_semantic_review",
		Rubric: filepath.Join(caseRoot, spec.Rubric),
	}
	for _, check := range spec.Checks {
		checkResult := runCheck(target, spec, baseline, check)
		result.Checks = append(result.Checks, checkResult)
		if checkResult.Status == "pass" {
			result.Passed++
		} else {
			result.Failed++
		}
	}
	if result.Failed == 0 {
		result.DeterministicStatus = "pass"
	} else {
		result.DeterministicStatus = "fail"
		result.OverallStatus = "fail"
	}
	if result.DeterministicStatus == "pass" {
		switch result.SemanticStatus {
		case "pass":
			result.OverallStatus = "pass"
		case "fail":
			result.OverallStatus = "fail"
		}
	}
	return result, nil
}

func validateTokenUsage(usage *TokenUsageDetails, legacy *int64) (*TokenUsageDetails, *int64, error) {
	if usage == nil {
		return nil, legacy, nil
	}
	copy := *usage
	copy.Source = strings.TrimSpace(copy.Source)
	if copy.Source == "" {
		return nil, nil, fmt.Errorf("detailed token usage requires a source")
	}
	supportedSources := map[string]struct{}{
		"codex": {}, "claude-code": {}, "antigravity-ide": {}, "cursor": {}, "manual": {}, TokenSourceUnavailable: {},
	}
	if _, ok := supportedSources[copy.Source]; !ok {
		return nil, nil, fmt.Errorf("unsupported token usage source %q", copy.Source)
	}
	values := []*int64{copy.InputTokens, copy.OutputTokens, copy.CachedTokens, copy.TotalTokens}
	for _, value := range values {
		if value != nil && *value < 0 {
			return nil, nil, fmt.Errorf("token usage cannot be negative")
		}
	}
	if copy.Source == TokenSourceUnavailable {
		for _, value := range values {
			if value != nil {
				return nil, nil, fmt.Errorf("unavailable token usage cannot include numeric values")
			}
		}
		if legacy != nil {
			return nil, nil, fmt.Errorf("unavailable token usage cannot include a legacy total")
		}
		return &copy, nil, nil
	}
	if copy.TotalTokens == nil && copy.InputTokens != nil && copy.OutputTokens != nil {
		total := *copy.InputTokens + *copy.OutputTokens
		copy.TotalTokens = &total
	}
	if copy.TotalTokens == nil && legacy != nil {
		copy.TotalTokens = legacy
	}
	if copy.TotalTokens != nil && copy.InputTokens != nil && copy.OutputTokens != nil && *copy.TotalTokens != *copy.InputTokens+*copy.OutputTokens {
		return nil, nil, fmt.Errorf("total token usage must equal input plus output tokens")
	}
	if legacy != nil && copy.TotalTokens != nil && *legacy != *copy.TotalTokens {
		return nil, nil, fmt.Errorf("legacy and detailed total token usage disagree")
	}
	if legacy == nil {
		legacy = copy.TotalTokens
	}
	return &copy, legacy, nil
}

func readBaseline(target, family string) (Baseline, error) {
	var baseline Baseline
	path := evaluationBaselinePath(target, family)
	data, err := os.ReadFile(path)
	if err != nil {
		return baseline, fmt.Errorf("read evaluation baseline; prepare the case first: %w", err)
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		return baseline, fmt.Errorf("parse evaluation baseline: %w", err)
	}
	if baseline.SchemaVersion != "2.0" {
		return baseline, fmt.Errorf("evaluation baseline schema_version must be 2.0; prepare the case again")
	}
	return baseline, nil
}

func validateSemanticResult(status string, score *SemanticScore, reviewer string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" || status == "pending_human_or_model_review" {
		if score != nil || strings.TrimSpace(reviewer) != "" {
			return "", fmt.Errorf("semantic score and reviewer require an explicit semantic status of pass or fail")
		}
		return "pending_human_or_model_review", nil
	}
	if status != "pass" && status != "fail" {
		return "", fmt.Errorf("semantic status must be pass, fail, or pending_human_or_model_review")
	}
	if score == nil || score.Available < 1 || score.Earned < 0 || score.Earned > score.Available {
		return "", fmt.Errorf("semantic pass or fail requires a valid earned and available score")
	}
	if strings.TrimSpace(reviewer) == "" {
		return "", fmt.Errorf("semantic pass or fail requires a reviewer")
	}
	return status, nil
}

func runCheck(target string, spec Spec, baseline Baseline, check CheckSpec) CheckResult {
	result := CheckResult{ID: check.ID, Type: check.Type, Status: "fail"}
	fail := func(format string, values ...any) CheckResult {
		result.Detail = fmt.Sprintf(format, values...)
		return result
	}
	pass := func(format string, values ...any) CheckResult {
		result.Status = "pass"
		result.Detail = fmt.Sprintf(format, values...)
		return result
	}

	switch check.Type {
	case "paths_exist":
		var missing []string
		for _, relative := range check.Paths {
			if info, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil || !info.Mode().IsRegular() {
				missing = append(missing, relative)
			}
		}
		if len(missing) > 0 {
			return fail("missing required paths: %s", strings.Join(missing, ", "))
		}
		return pass("all %d required paths exist", len(check.Paths))
	case "minimum_files":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		if len(files) < check.Minimum {
			return fail("%d files match %s; need at least %d", len(files), check.Glob, check.Minimum)
		}
		return pass("%d files match %s", len(files), check.Glob)
	case "changed_files_minimum":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		changed := 0
		for _, relative := range files {
			digest, err := hashFile(filepath.Join(target, filepath.FromSlash(relative)))
			if err == nil && baseline.Files[relative] != digest {
				changed++
			}
		}
		if changed < check.Minimum {
			return fail("%d matching files changed; need at least %d", changed, check.Minimum)
		}
		return pass("%d matching files changed", changed)
	case "contains_all", "not_contains":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		content, err := concatenateFiles(target, files)
		if err != nil {
			return fail("read files: %v", err)
		}
		lower := strings.ToLower(content)
		var violations []string
		for _, pattern := range check.Patterns {
			contains := strings.Contains(lower, strings.ToLower(pattern))
			if (check.Type == "contains_all" && !contains) || (check.Type == "not_contains" && contains) {
				violations = append(violations, pattern)
			}
		}
		if len(violations) > 0 {
			verb := "missing"
			if check.Type == "not_contains" {
				verb = "forbidden content present"
			}
			return fail("%s: %s", verb, strings.Join(violations, ", "))
		}
		return pass("content requirement satisfied across %d files", len(files))
	case "source_examples_minimum":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		content, err := concatenateFiles(target, files)
		if err != nil {
			return fail("read files: %v", err)
		}
		fences := len(regexp.MustCompile("(?m)^```[A-Za-z0-9_-]*[ \\t]*$").FindAllString(content, -1)) / 2
		sources := uniqueExistingEvidence(target, content)
		count := fences
		if len(sources) < count {
			count = len(sources)
		}
		if count < check.Minimum {
			return fail("found %d fenced examples and %d existing source anchors; need %d attributed examples", fences, len(sources), check.Minimum)
		}
		return pass("found at least %d attributed source examples", check.Minimum)
	case "existing_evidence_minimum":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		content, err := concatenateFiles(target, files)
		if err != nil {
			return fail("read files: %v", err)
		}
		evidence := uniqueExistingEvidence(target, content)
		if len(evidence) < check.Minimum {
			return fail("found %d distinct existing source anchors; need %d", len(evidence), check.Minimum)
		}
		return pass("found %d distinct existing source anchors", len(evidence))
	case "capabilities_minimum":
		data, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(check.Path)))
		if err != nil {
			return fail("read capability metadata: %v", err)
		}
		var value struct {
			Capabilities []json.RawMessage `json:"capabilities"`
		}
		if err := json.Unmarshal(data, &value); err != nil {
			return fail("parse capability metadata: %v", err)
		}
		if len(value.Capabilities) < check.Minimum {
			return fail("found %d capabilities; need %d", len(value.Capabilities), check.Minimum)
		}
		return pass("found %d verified capability routes", len(value.Capabilities))
	case "markdown_links_valid":
		files, err := matchingFiles(target, check.Glob)
		if err != nil {
			return fail("match files: %v", err)
		}
		broken := brokenMarkdownLinks(target, files)
		if len(broken) > 0 {
			return fail("broken local links: %s", strings.Join(broken, ", "))
		}
		return pass("local Markdown links are valid across %d files", len(files))
	case "protected_files_unchanged":
		violations, err := protectedViolations(target, baseline, spec.AllowedChanges)
		if err != nil {
			return fail("inspect protected files: %v", err)
		}
		if len(violations) > 0 {
			return fail("protected fixture changes: %s", strings.Join(violations, ", "))
		}
		return pass("fixture source and toolkit-managed files are unchanged")
	default:
		return fail("unsupported check type %q", check.Type)
	}
}

func uniqueExistingEvidence(target, content string) []string {
	matcher := regexp.MustCompile("`([^`\\n]+)`")
	values := make(map[string]struct{})
	for _, match := range matcher.FindAllStringSubmatch(content, -1) {
		candidate := strings.TrimSpace(match[1])
		if strings.ContainsAny(candidate, " *{}|<>") || filepath.IsAbs(candidate) || strings.HasPrefix(candidate, "docs/") || strings.HasPrefix(candidate, ".repo-knowledge/") {
			continue
		}
		if index := strings.LastIndex(candidate, ":"); index > 1 {
			if _, err := strconv.Atoi(candidate[index+1:]); err == nil {
				candidate = candidate[:index]
			}
		}
		if candidate == "" || (!strings.Contains(candidate, "/") && !strings.Contains(candidate, ".")) {
			continue
		}
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(candidate))); err == nil {
			values[candidate] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func brokenMarkdownLinks(target string, files []string) []string {
	matcher := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	var broken []string
	for _, relative := range files {
		data, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(relative)))
		if err != nil {
			broken = append(broken, relative+": unreadable")
			continue
		}
		for _, match := range matcher.FindAllStringSubmatch(string(data), -1) {
			link := strings.TrimSpace(match[1])
			if link == "" || strings.HasPrefix(link, "#") || strings.Contains(link, "://") || strings.HasPrefix(link, "mailto:") {
				continue
			}
			link = strings.SplitN(link, "#", 2)[0]
			path := filepath.Join(target, filepath.Dir(filepath.FromSlash(relative)), filepath.FromSlash(link))
			if _, err := os.Stat(path); err != nil {
				broken = append(broken, relative+" -> "+link)
			}
		}
	}
	sort.Strings(broken)
	return broken
}

func protectedViolations(target string, baseline Baseline, allowed []string) ([]string, error) {
	current, err := hashRepository(target)
	if err != nil {
		return nil, err
	}
	var violations []string
	for path, original := range baseline.Files {
		if allowedPath(path, allowed) {
			continue
		}
		value, exists := current[path]
		if !exists {
			violations = append(violations, path+" (deleted)")
		} else if value != original {
			violations = append(violations, path+" (modified)")
		}
	}
	for path := range current {
		if path == ".repo-knowledge/eval-baseline.json" || allowedPath(path, allowed) {
			continue
		}
		if _, exists := baseline.Files[path]; !exists {
			violations = append(violations, path+" (added)")
		}
	}
	sort.Strings(violations)
	return violations, nil
}

func allowedPath(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if globMatch(pattern, path) {
			return true
		}
	}
	return false
}
