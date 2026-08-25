package toolkit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
)

func loadRules(root string) (ImpactRules, error) {
	var base ImpactRules
	installed := filepath.Join(root, ".repo-knowledge", "policy", "default-impact-rules.json")
	if _, err := os.Stat(installed); err == nil {
		var readErr error
		base, readErr = readJSON[ImpactRules](installed, true)
		if readErr != nil {
			return ImpactRules{}, readErr
		}
	} else if os.IsNotExist(err) {
		data, readErr := repositoryknowledge.Content.ReadFile("policy/default-impact-rules.json")
		if readErr != nil {
			return ImpactRules{}, fmt.Errorf("read embedded impact rules: %w", readErr)
		}
		if err := unmarshalJSON(data, &base, "embedded impact rules"); err != nil {
			return ImpactRules{}, err
		}
	} else {
		return ImpactRules{}, fmt.Errorf("inspect installed impact rules: %w", err)
	}
	local, err := readJSON[ImpactRules](filepath.Join(root, ".repo-knowledge", "local-impact-rules.json"), false)
	if err != nil {
		return ImpactRules{}, err
	}
	base.DocumentationGlobs = append(base.DocumentationGlobs, local.DocumentationGlobs...)
	base.IgnoredGlobs = append(base.IgnoredGlobs, local.IgnoredGlobs...)
	base.Classifiers = append(base.Classifiers, local.Classifiers...)
	if local.RequiredDocMappings != nil {
		base.RequiredDocMappings = local.RequiredDocMappings
	}
	if base.RequiredDocMappings == nil {
		base.RequiredDocMappings = make(map[string][]string)
	}
	return base, nil
}

func DiffChanges(root, base, head string) ([]Change, error) {
	if !isGitRepository(root) {
		return nil, fmt.Errorf("documentation impact requires a Git repository")
	}
	arguments := []string{"diff", "--name-status", "--find-renames"}
	switch {
	case base != "" && head != "":
		arguments = append(arguments, base, head)
	case base != "":
		arguments = append(arguments, base)
	case head != "":
		arguments = append(arguments, head)
	default:
		arguments = append(arguments, "HEAD")
	}
	output, err := runGit(root, true, arguments...)
	if err != nil {
		return nil, err
	}
	unique := make(map[string]Change)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		change := Change{Status: fields[0], Path: filepath.ToSlash(fields[len(fields)-1])}
		unique[change.Status+"\x00"+change.Path] = change
	}
	if head == "" {
		untracked, err := runGit(root, true, "ls-files", "--others", "--exclude-standard", "-z")
		if err != nil {
			return nil, err
		}
		for _, value := range strings.Split(untracked, "\x00") {
			if value != "" {
				change := Change{Status: "?", Path: filepath.ToSlash(value)}
				unique[change.Status+"\x00"+change.Path] = change
			}
		}
	}
	changes := make([]Change, 0, len(unique))
	for _, change := range unique {
		changes = append(changes, change)
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Status < changes[j].Status
		}
		return changes[i].Path < changes[j].Path
	})
	return changes, nil
}

func Impact(root, base, head string) (ImpactReport, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return ImpactReport{}, fmt.Errorf("resolve repository root: %w", err)
	}
	rules, err := loadRules(root)
	if err != nil {
		return ImpactReport{}, err
	}
	changes, err := DiffChanges(root, base, head)
	if err != nil {
		return ImpactReport{}, err
	}
	material := make([]Change, 0)
	documentation := make([]Change, 0)
	ignored := make([]Change, 0)
	for _, change := range changes {
		switch {
		case matchesAny(change.Path, rules.IgnoredGlobs):
			ignored = append(ignored, change)
		case matchesAny(change.Path, rules.DocumentationGlobs):
			documentation = append(documentation, change)
		default:
			material = append(material, change)
		}
	}
	classifications := make(map[string][]string)
	for _, classifier := range rules.Classifiers {
		for _, change := range material {
			lower := strings.ToLower(change.Path)
			for _, token := range classifier.Tokens {
				if strings.Contains(lower, strings.ToLower(token)) {
					classifications[classifier.ID] = append(classifications[classifier.ID], change.Path)
					break
				}
			}
		}
	}
	if len(material) > 0 && len(classifications) == 0 {
		for _, change := range material {
			classifications["general"] = append(classifications["general"], change.Path)
		}
	}
	digest := sha256.New()
	for _, change := range material {
		status := change.Status
		if status == "?" {
			status = "A"
		}
		_, _ = fmt.Fprintf(digest, "%s\t%s\n", status, change.Path)
	}
	documentationImpact := "not_required"
	if len(material) > 0 {
		documentationImpact = "required"
	}
	return ImpactReport{
		SchemaVersion:        "1.0",
		Base:                 base,
		Head:                 head,
		MaterialChanges:      material,
		DocumentationChanges: documentation,
		IgnoredChanges:       ignored,
		Classifications:      classifications,
		ChangeFingerprint:    hex.EncodeToString(digest.Sum(nil)),
		DocumentationImpact:  documentationImpact,
	}, nil
}

func ValidateImpact(root, base, head, mode string) (ImpactReport, error) {
	report, err := Impact(root, base, head)
	if err != nil {
		return ImpactReport{}, err
	}
	config, err := readJSON[RepositoryConfig](filepath.Join(root, ".repo-knowledge", "repository.json"), true)
	if err != nil {
		return ImpactReport{}, err
	}
	if mode == "" {
		mode = config.CI.Enforcement
	}
	if mode != "advisory" && mode != "acknowledgment" && mode != "enforced" {
		return ImpactReport{}, fmt.Errorf("unsupported enforcement mode: %s", mode)
	}

	acknowledgment, err := readJSON[ImpactAcknowledgment](filepath.Join(root, ".repo-knowledge", "doc-impact.json"), false)
	if err != nil {
		return ImpactReport{}, err
	}
	acknowledged := acknowledgment.ChangeFingerprint == report.ChangeFingerprint &&
		(acknowledgment.Impact == "required" || acknowledgment.Impact == "not_required") &&
		len(strings.TrimSpace(acknowledgment.Reason)) >= 8
	findings := make([]ImpactFinding, 0)
	if len(report.MaterialChanges) > 0 && len(report.DocumentationChanges) == 0 && !acknowledged {
		findings = append(findings, ImpactFinding{
			ID:      "missing-impact-decision",
			Message: "material changes have neither documentation changes nor a matching impact acknowledgment",
		})
	}
	if mode == "enforced" && len(report.DocumentationChanges) > 0 {
		rules, err := loadRules(root)
		if err != nil {
			return ImpactReport{}, err
		}
		for _, classification := range sortedKeys(report.Classifications) {
			patterns := rules.RequiredDocMappings[classification]
			if len(patterns) == 0 {
				continue
			}
			matched := false
			for _, change := range report.DocumentationChanges {
				if matchesAny(change.Path, patterns) {
					matched = true
					break
				}
			}
			if !matched {
				findings = append(findings, ImpactFinding{
					ID:      "missing-required-documentation-route",
					Message: fmt.Sprintf("%s changes require documentation matching %v", classification, patterns),
				})
			}
		}
	}
	status := "pass"
	if len(findings) > 0 {
		if mode == "advisory" {
			status = "advisory"
		} else {
			status = "fail"
		}
	}
	report.Enforcement = mode
	report.AcknowledgmentMatches = acknowledged
	report.Findings = findings
	report.Status = status
	return report, nil
}

func Acknowledge(root, impact, reason, base, head string) (ImpactAcknowledgment, error) {
	normalized := strings.ReplaceAll(impact, "-", "_")
	if normalized != "required" && normalized != "not_required" {
		return ImpactAcknowledgment{}, fmt.Errorf("impact must be required or not-required")
	}
	if len(strings.TrimSpace(reason)) < 8 {
		return ImpactAcknowledgment{}, fmt.Errorf("reason must be at least 8 characters and specific to the diff")
	}
	report, err := Impact(root, base, head)
	if err != nil {
		return ImpactAcknowledgment{}, err
	}
	value := ImpactAcknowledgment{
		Schema:            "schemas/doc-impact.schema.json",
		SchemaVersion:     "1.0",
		ChangeFingerprint: report.ChangeFingerprint,
		Impact:            normalized,
		Reason:            strings.TrimSpace(reason),
	}
	path, err := repositoryPath(root, ".repo-knowledge/doc-impact.json", "documentation impact acknowledgment")
	if err != nil {
		return ImpactAcknowledgment{}, err
	}
	if err := writeJSON(path, value); err != nil {
		return ImpactAcknowledgment{}, err
	}
	return value, nil
}
