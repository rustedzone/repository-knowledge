package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	repositoryknowledge "github.com/rustedzone/repository-knowledge"
	"github.com/rustedzone/repository-knowledge/internal/toolkit"
)

type stringList []string

func (values *stringList) String() string { return strings.Join(*values, ",") }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}

type commandOutcome struct {
	value    any
	exitCode int
	json     bool
	output   string
}

func main() {
	os.Exit(runWithInput(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	return runWithInput(arguments, strings.NewReader("{}"), stdout, stderr)
}

func runWithInput(arguments []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		printUsage(stderr)
		return 2
	}
	if arguments[0] == "--version" || arguments[0] == "version" {
		fmt.Fprintf(stdout, "repo-knowledge %s\n", repositoryknowledge.Version())
		return 0
	}
	if arguments[0] == "help" || arguments[0] == "--help" || arguments[0] == "-h" {
		printUsage(stdout)
		return 0
	}
	outcome, err := dispatch(arguments, stdin, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		if outcome.json {
			_ = json.NewEncoder(stdout).Encode(map[string]string{"status": "error", "message": err.Error()})
		} else {
			fmt.Fprintf(stderr, "error: %v\n", err)
		}
		return 2
	}
	if outcome.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(outcome.value); err != nil {
			fmt.Fprintf(stderr, "error: encode command output: %v\n", err)
			return 2
		}
	} else {
		if outcome.output != "" {
			fmt.Fprint(stdout, outcome.output)
			if !strings.HasSuffix(outcome.output, "\n") {
				fmt.Fprintln(stdout)
			}
		}
		fmt.Fprintln(stdout, summarize(arguments[0], outcome.value))
	}
	return outcome.exitCode
}

func dispatch(arguments []string, stdin io.Reader, stderr io.Writer) (commandOutcome, error) {
	command := arguments[0]
	args := arguments[1:]
	switch command {
	case "install", "update":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository root")
		sourceDefault := "https://github.com/rustedzone/repository-knowledge"
		refDefault := "v" + repositoryknowledge.Version()
		if command == "update" {
			sourceDefault = ""
			refDefault = ""
		}
		source := flags.String("source", sourceDefault, "canonical binary/package source")
		ref := flags.String("ref", refDefault, "pinned toolkit release ref")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		allowDowngrade := flags.Bool("allow-downgrade", false, "allow update to an older toolkit version")
		antigravityPreflight := flags.String("antigravity-preflight", "", "Antigravity preflight mode: observe or strict (default: observe; update retains existing mode)")
		preflightContext := flags.String("preflight-context", "", "hook context profile: full or compact (default: full; update retains existing profile)")
		var agentPreflight stringList
		flags.Var(&agentPreflight, "agent-preflight", "per-agent preflight mode; repeatable: AGENT=observe or AGENT=strict")
		allAgents := flags.Bool("all-agents", false, "select every supported agent adapter; cannot be combined with --agent")
		var adapters stringList
		agentHelp := "agent adapter to install; repeatable: codex, claude-code, antigravity-ide, cursor (default: codex)"
		if command == "update" {
			agentHelp = "agent adapter selection; repeatable: codex, claude-code, antigravity-ide, cursor (default: retain installed selection)"
		}
		flags.Var(&adapters, "agent", agentHelp)
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		result, err := toolkit.Install(toolkit.InstallOptions{
			Target: *target, Source: *source, Ref: *ref, AgentAdapters: adapters,
			AllAgentAdapters: *allAgents, Update: command == "update", AllowDowngrade: *allowDowngrade, AntigravityPreflight: *antigravityPreflight, AgentPreflight: agentPreflight,
			PreflightContext: *preflightContext,
		})
		return commandOutcome{value: result, json: *jsonOutput}, err
	case "scan", "doctor":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository root")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		liveHooks := flags.Bool("live-hooks", false, "exercise supported strict preflight gate behavior in-process (doctor only)")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		if command == "scan" {
			value, err := toolkit.Scan(*target, true)
			return commandOutcome{value: value, json: *jsonOutput}, err
		}
		value, err := toolkit.Doctor(*target)
		if *liveHooks {
			value, err = toolkit.DoctorLive(*target)
		}
		exitCode := 0
		if value.Status == "fail" {
			exitCode = 1
		}
		return commandOutcome{value: value, exitCode: exitCode, json: *jsonOutput}, err
	case "hook-context":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository path or a path nested below it")
		agent := flags.String("agent", "", "agent hook protocol: codex, claude-code, antigravity-ide, or cursor")
		metrics := flags.Bool("metrics", false, "emit content-free context generation metrics instead of hook output")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{}, err
		}
		if strings.TrimSpace(*agent) == "" {
			return commandOutcome{}, fmt.Errorf("--agent is required")
		}
		if *metrics {
			value, err := toolkit.MeasureHookContext(*target, *agent)
			return commandOutcome{value: value}, err
		}
		value, err := toolkit.HookContext(*target, *agent, stdin)
		return commandOutcome{value: value}, err
	case "preflight-activate":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository path or a path nested below it")
		token := flags.String("token", "", "opaque token injected by a strict preflight hook")
		workflow := flags.String("workflow", toolkit.WorkflowStandard, "workflow profile: standard or scoped")
		var routes stringList
		flags.Var(&routes, "route", "selected documentation route; repeatable")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		if strings.TrimSpace(*token) == "" {
			return commandOutcome{json: *jsonOutput}, fmt.Errorf("--token is required")
		}
		value, err := toolkit.PreflightActivateWithOptions(toolkit.PreflightActivationOptions{Target: *target, Token: *token, Routes: routes, Workflow: *workflow})
		return commandOutcome{value: value, json: *jsonOutput}, err
	case "evidence-run":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository path or a path nested below it")
		token := flags.String("token", "", "opaque token for an active strict preflight session")
		label := flags.String("label", "", "short verification label")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON (command output is never included)")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		if strings.TrimSpace(*token) == "" || strings.TrimSpace(*label) == "" {
			return commandOutcome{json: *jsonOutput}, fmt.Errorf("--token and --label are required")
		}
		value, err := toolkit.EvidenceRun(toolkit.EvidenceRunOptions{Target: *target, Token: *token, Label: *label, Command: flags.Args()})
		exitCode := 0
		if err == nil && value.ExitCode != 0 {
			exitCode = 1
		}
		return commandOutcome{value: value, exitCode: exitCode, json: *jsonOutput, output: value.Output}, err
	case "evidence-report":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository path or a path nested below it")
		token := flags.String("token", "", "opaque token for an active strict preflight session")
		impact := flags.String("documentation-impact", "", "required or not-required")
		reason := flags.String("reason", "", "specific reason when documentation is not required")
		var evidenceRefs stringList
		flags.Var(&evidenceRefs, "evidence", "source/config/test evidence reference; repeatable, optionally path#anchor")
		var documentationFiles stringList
		flags.Var(&documentationFiles, "documentation-file", "changed documentation file; repeatable")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		if strings.TrimSpace(*token) == "" {
			return commandOutcome{json: *jsonOutput}, fmt.Errorf("--token is required")
		}
		value, err := toolkit.EvidenceReport(toolkit.EvidenceReportOptions{
			Target: *target, Token: *token, DocumentationImpact: *impact, DocumentationFiles: documentationFiles,
			EvidenceRefs: evidenceRefs, Reason: *reason,
		})
		return commandOutcome{value: value, json: *jsonOutput}, err
	case "preflight-gate":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository path or a path nested below it")
		agent := flags.String("agent", "", "agent hook protocol: codex, claude-code, antigravity-ide, or cursor")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{}, err
		}
		if strings.TrimSpace(*agent) == "" {
			return commandOutcome{}, fmt.Errorf("--agent is required")
		}
		value, err := toolkit.PreflightGate(*target, *agent, stdin)
		return commandOutcome{value: value}, err
	case "audit", "impact", "validate-doc-impact":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository root")
		base := flags.String("base", "", "base Git revision")
		head := flags.String("head", "", "head Git revision")
		mode := flags.String("mode", "", "enforcement mode")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		if command == "audit" {
			value, err := toolkit.Audit(*target, *base, *head)
			exitCode := 0
			if value.Status == "fail" {
				exitCode = 1
			}
			return commandOutcome{value: value, exitCode: exitCode, json: *jsonOutput}, err
		}
		if command == "impact" {
			value, err := toolkit.Impact(*target, *base, *head)
			return commandOutcome{value: value, json: *jsonOutput}, err
		}
		value, err := toolkit.ValidateImpact(*target, *base, *head, *mode)
		exitCode := 0
		if value.Status == "fail" {
			exitCode = 1
		}
		return commandOutcome{value: value, exitCode: exitCode, json: *jsonOutput}, err
	case "rebuild":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository root")
		apply := flags.Bool("apply", false, "write the configured structural inventory (not semantic documentation)")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		value, err := toolkit.Rebuild(*target, *apply)
		return commandOutcome{value: value, json: *jsonOutput}, err
	case "acknowledge":
		flags := newFlagSet(command, stderr)
		target := flags.String("target", ".", "repository root")
		base := flags.String("base", "", "base Git revision")
		head := flags.String("head", "", "head Git revision")
		impact := flags.String("impact", "", "required or not-required")
		reason := flags.String("reason", "", "specific reason for the decision")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args); err != nil {
			return commandOutcome{json: *jsonOutput}, err
		}
		value, err := toolkit.Acknowledge(*target, *impact, *reason, *base, *head)
		return commandOutcome{value: value, json: *jsonOutput}, err
	default:
		return commandOutcome{}, fmt.Errorf("unknown command %q", command)
	}
}

func newFlagSet(name string, output io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(output)
	return flags
}

func summarize(command string, value any) string {
	switch result := value.(type) {
	case toolkit.InstallResult:
		lines := []string{
			fmt.Sprintf("%s repository-knowledge %s in %s", result.Action, result.Version, result.Target),
			fmt.Sprintf("managed files: %d", result.ManagedFileCount),
			fmt.Sprintf("repository-owned files created: %d", len(result.RepositoryOwnedFilesCreated)),
		}
		if len(result.AgentHookRegistrations) > 0 {
			lines = append(lines, "preflight hooks: "+strings.Join(result.AgentHookRegistrations, ", "))
		}
		return strings.Join(lines, "\n")
	case toolkit.ScanState:
		return fmt.Sprintf("scan complete: %d files, %d top-level modules, %d capability signals\nwrote .repo-knowledge/scan-state.json", result.FileCount, len(result.Modules), len(result.CapabilitySignals))
	case toolkit.DoctorReport:
		lines := []string{"doctor: " + result.Status}
		for _, check := range result.Checks {
			status := "ok"
			if !check.OK {
				status = "FAIL"
			}
			lines = append(lines, fmt.Sprintf("[%s] %s: %s", status, check.ID, check.Detail))
		}
		return strings.Join(lines, "\n")
	case toolkit.AuditReport:
		lines := []string{"audit: " + result.Status}
		for _, finding := range result.Findings {
			lines = append(lines, fmt.Sprintf("[%s] %s: %s", finding.Severity, finding.ID, finding.Message))
		}
		if len(result.Findings) == 0 {
			lines = append(lines, "no drift detected by v1 checks")
		}
		return strings.Join(lines, "\n")
	case toolkit.ImpactReport:
		if command == "validate-doc-impact" {
			lines := []string{fmt.Sprintf("documentation impact validation: %s (%s)", result.Status, result.Enforcement), "change fingerprint: " + result.ChangeFingerprint}
			for _, finding := range result.Findings {
				lines = append(lines, "- "+finding.ID+": "+finding.Message)
			}
			return strings.Join(lines, "\n")
		}
		return fmt.Sprintf("documentation impact: %s\nmaterial changes: %d\ndocumentation changes: %d\nchange fingerprint: %s", result.DocumentationImpact, len(result.MaterialChanges), len(result.DocumentationChanges), result.ChangeFingerprint)
	case toolkit.ImpactAcknowledgment:
		return fmt.Sprintf("wrote .repo-knowledge/doc-impact.json\ndocumentation impact: %s\nchange fingerprint: %s", result.Impact, result.ChangeFingerprint)
	case toolkit.RebuildResult:
		if result.Applied {
			return fmt.Sprintf("structural inventory applied: %s\nthis does not generate semantic documentation\nnext: %s", result.GeneratedInventory, result.NextStep)
		}
		return fmt.Sprintf("structural inventory proposal: %s\nthis does not generate semantic documentation\nnext: review the proposal, optionally rerun with --apply, then %s", result.Proposal, result.NextStep)
	case toolkit.HookContextResult:
		return result.Output
	case toolkit.PreflightGateResult:
		return result.Output
	case toolkit.PreflightActivationResult:
		return fmt.Sprintf("preflight: %s\nworkflow: %s\nroutes: %s", result.Status, result.Workflow, strings.Join(result.Routes, ", "))
	case toolkit.EvidenceRunResult:
		return fmt.Sprintf("verification: %s\nlabel: %s\ndiff fingerprint: %s\noutput sha256: %s", result.Status, result.Label, result.DiffFingerprint, result.OutputSHA256)
	case toolkit.EvidenceReceipt:
		return fmt.Sprintf("evidence report: %s\nworkflow: %s\nmaterial changes: %d\nverifications: %d\ndiff fingerprint: %s", result.Status, result.Workflow, len(result.MaterialChanges), len(result.Verifications), result.DiffFingerprint)
	default:
		data, _ := json.MarshalIndent(value, "", "  ")
		return string(data)
	}
}

func printUsage(output io.Writer) {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(output, "Usage: %s <command> [options]\n\n", name)
	fmt.Fprintln(output, "Commands: install, update, scan, doctor, audit, rebuild, impact, acknowledge, validate-doc-impact, hook-context, preflight-activate, preflight-gate, evidence-run, evidence-report, version")
}
