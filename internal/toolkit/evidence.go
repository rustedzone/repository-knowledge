package toolkit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var scopedHighRiskClassifications = map[string]struct{}{
	"api": {}, "authorization": {}, "persistence": {}, "integration": {}, "configuration": {}, "deployment": {},
}

func EvidenceRun(options EvidenceRunOptions) (EvidenceRunResult, error) {
	root, session, statePath, err := activePreflightSession(options.Target, options.Token)
	if err != nil {
		return EvidenceRunResult{}, err
	}
	label := strings.TrimSpace(options.Label)
	if label == "" {
		return EvidenceRunResult{}, fmt.Errorf("verification label is required")
	}
	if len(options.Command) == 0 || strings.TrimSpace(options.Command[0]) == "" {
		return EvidenceRunResult{}, fmt.Errorf("verification command is required")
	}

	started := time.Now().UTC()
	command := exec.Command(options.Command[0], options.Command[1:]...)
	command.Dir = root
	output, commandErr := command.CombinedOutput()
	finished := time.Now().UTC()
	exitCode := 0
	if commandErr != nil {
		if exitError, ok := commandErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return EvidenceRunResult{}, fmt.Errorf("run verification %q: %w", label, commandErr)
		}
	}
	fingerprint, _, err := worktreeFingerprint(root)
	if err != nil {
		return EvidenceRunResult{}, err
	}
	outputDigest := sha256.Sum256(output)
	record := VerificationRecord{
		Label: label, Command: append([]string(nil), options.Command...),
		StartedAt: started.Format(time.RFC3339Nano), FinishedAt: finished.Format(time.RFC3339Nano),
		ExitCode: exitCode, OutputSHA256: hex.EncodeToString(outputDigest[:]), DiffFingerprint: fingerprint,
	}
	session.Verifications = append(session.Verifications, record)
	if err := writePreflightSession(statePath, session); err != nil {
		return EvidenceRunResult{}, err
	}
	status := "pass"
	if exitCode != 0 {
		status = "fail"
	}
	return EvidenceRunResult{VerificationRecord: record, Status: status, Output: string(output)}, nil
}

func EvidenceReport(options EvidenceReportOptions) (EvidenceReceipt, error) {
	root, session, _, err := activePreflightSession(options.Target, options.Token)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	impact, err := Impact(root, "", "")
	if err != nil {
		return EvidenceReceipt{}, err
	}
	if len(impact.MaterialChanges) == 0 {
		return EvidenceReceipt{}, fmt.Errorf("evidence report requires at least one material change")
	}
	fingerprint, _, err := worktreeFingerprint(root)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	matching := make([]VerificationRecord, 0)
	for _, verification := range session.Verifications {
		if verification.ExitCode == 0 && verification.DiffFingerprint == fingerprint {
			matching = append(matching, verification)
		}
	}
	if len(matching) == 0 {
		if len(session.Verifications) == 0 {
			return EvidenceReceipt{}, fmt.Errorf("evidence report requires a successful verification for the current diff")
		}
		return EvidenceReceipt{}, fmt.Errorf("no successful verification matches the current diff; rerun verification after the latest change")
	}
	if session.Workflow == WorkflowScoped {
		risky := scopedRiskClassifications(impact.Classifications)
		if len(risky) > 0 {
			return EvidenceReceipt{}, fmt.Errorf("scoped workflow cannot complete high-risk classifications %s; reactivate with --workflow standard", strings.Join(risky, ", "))
		}
	}
	evidenceRefs, err := validateEvidenceRefs(root, options.EvidenceRefs)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	impactDecision := strings.ReplaceAll(strings.TrimSpace(options.DocumentationImpact), "-", "_")
	if impactDecision != DocumentationImpactRequired && impactDecision != DocumentationImpactNotRequired {
		return EvidenceReceipt{}, fmt.Errorf("documentation impact must be required or not-required")
	}
	reason := strings.TrimSpace(options.Reason)
	documentationFiles, err := validateDocumentationDecision(root, impactDecision, reason, options.DocumentationFiles, impact.DocumentationChanges)
	if err != nil {
		return EvidenceReceipt{}, err
	}
	workflow := session.Workflow
	if workflow == "" {
		workflow = WorkflowStandard
	}
	return EvidenceReceipt{
		SchemaVersion: "1.0", Status: "complete", Root: root, Workflow: workflow,
		DiffFingerprint: fingerprint, MaterialChanges: impact.MaterialChanges, DocumentationChanges: impact.DocumentationChanges,
		Classifications: impact.Classifications, Verifications: matching, EvidenceRefs: evidenceRefs,
		DocumentationImpact: impactDecision, DocumentationFiles: documentationFiles, Reason: reason,
	}, nil
}

func activePreflightSession(target, token string) (string, preflightSession, string, error) {
	root, err := findInstalledRepositoryRoot(target)
	if err != nil {
		return "", preflightSession{}, "", err
	}
	session, statePath, err := preflightSessionByToken(token)
	if err != nil {
		return "", preflightSession{}, "", err
	}
	if session.Root != root {
		return "", preflightSession{}, "", fmt.Errorf("preflight token belongs to a different repository")
	}
	digests, _, err := preflightDigests(root)
	if err != nil {
		return "", preflightSession{}, "", err
	}
	if !session.Active || !validPreflightSession(session, root, session.Agent, digests) {
		return "", preflightSession{}, "", fmt.Errorf("active preflight session is required; activate again after expiry or knowledge changes")
	}
	return root, session, statePath, nil
}

func worktreeFingerprint(root string) (string, []Change, error) {
	changes, err := DiffChanges(root, "", "")
	if err != nil {
		return "", nil, err
	}
	digest := sha256.New()
	for _, change := range changes {
		path, err := repositoryPath(root, change.Path, "changed file")
		if err != nil {
			return "", nil, err
		}
		contentDigest := "deleted"
		info, statErr := os.Lstat(path)
		switch {
		case os.IsNotExist(statErr):
		case statErr != nil:
			return "", nil, fmt.Errorf("inspect changed file %s: %w", change.Path, statErr)
		case info.Mode()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return "", nil, fmt.Errorf("read changed symlink %s: %w", change.Path, err)
			}
			value := sha256.Sum256([]byte(target))
			contentDigest = "symlink:" + hex.EncodeToString(value[:])
		case info.Mode().IsRegular():
			contentDigest, err = sha256File(path)
			if err != nil {
				return "", nil, err
			}
		default:
			contentDigest = "mode:" + info.Mode().String()
		}
		_, _ = fmt.Fprintf(digest, "%s\t%s\t%s\n", change.Status, change.Path, contentDigest)
	}
	return hex.EncodeToString(digest.Sum(nil)), changes, nil
}

func scopedRiskClassifications(classifications map[string][]string) []string {
	result := make([]string, 0)
	for classification := range classifications {
		if _, risky := scopedHighRiskClassifications[classification]; risky {
			result = append(result, classification)
		}
	}
	sort.Strings(result)
	return result
}

func validateEvidenceRefs(root string, refs []string) ([]string, error) {
	if len(refs) == 0 {
		return nil, fmt.Errorf("at least one evidence reference is required")
	}
	seen := make(map[string]struct{}, len(refs))
	validated := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		pathPart := strings.SplitN(ref, "#", 2)[0]
		path, err := repositoryPath(root, pathPart, "evidence reference")
		if err != nil || !regularFile(path) {
			return nil, fmt.Errorf("evidence reference must name an existing repository file: %q", ref)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, fmt.Errorf("resolve evidence reference %q: %w", ref, err)
		}
		normalized := filepath.ToSlash(relative)
		if strings.Contains(ref, "#") {
			normalized += "#" + strings.SplitN(ref, "#", 2)[1]
		}
		if _, ok := seen[normalized]; !ok {
			seen[normalized] = struct{}{}
			validated = append(validated, normalized)
		}
	}
	sort.Strings(validated)
	return validated, nil
}

func validateDocumentationDecision(root, impact, reason string, requested []string, changes []Change) ([]string, error) {
	if impact == DocumentationImpactNotRequired {
		if len(reason) < 8 {
			return nil, fmt.Errorf("not-required documentation impact requires a reason of at least 8 characters")
		}
		return nil, nil
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("required documentation impact needs at least one documentation change in the current diff")
	}
	changed := make(map[string]struct{}, len(changes))
	for _, change := range changes {
		changed[change.Path] = struct{}{}
	}
	if len(requested) == 0 {
		for path := range changed {
			requested = append(requested, path)
		}
	}
	validated := make([]string, 0, len(requested))
	for _, value := range requested {
		path, err := repositoryPath(root, value, "documentation file")
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		relative = filepath.ToSlash(relative)
		if _, ok := changed[relative]; !ok {
			return nil, fmt.Errorf("documentation file is not changed in the current diff: %q", value)
		}
		validated = append(validated, relative)
	}
	sort.Strings(validated)
	return validated, nil
}
