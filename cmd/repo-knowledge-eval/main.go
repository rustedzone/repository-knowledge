package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rustedzone/repository-knowledge/internal/evalharness"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	if len(arguments) == 0 {
		printUsage(stderr)
		return 2
	}
	switch arguments[0] {
	case "list":
		flags := flag.NewFlagSet("list", flag.ContinueOnError)
		flags.SetOutput(stderr)
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		specs, err := evalharness.ListCases(*casesRoot)
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if *jsonOutput {
			return encodeJSON(stdout, specs)
		}
		for _, spec := range specs {
			fmt.Fprintf(stdout, "%s (revision %s): %s\n", spec.ID, spec.Revision, spec.Description)
		}
		return 0
	case "prepare":
		flags := flag.NewFlagSet("prepare", flag.ContinueOnError)
		flags.SetOutput(stderr)
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		caseID := flags.String("case", "", "evaluation case id")
		output := flags.String("output", "", "empty destination directory")
		agent := flags.String("agent", "codex", "codex, claude-code, antigravity-ide, or cursor")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		if *caseID == "" || *output == "" {
			fmt.Fprintln(stderr, "error: --case and --output are required")
			return 2
		}
		result, err := evalharness.Prepare(evalharness.PrepareOptions{CasesRoot: *casesRoot, CaseID: *caseID, Output: *output, Agent: *agent})
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if *jsonOutput {
			return encodeJSON(stdout, result)
		}
		fmt.Fprintf(stdout, "prepared %s revision %s at %s for %s\nsemantic rubric: %s\n\nprompt:\n%s\n", result.CaseID, result.Revision, result.Target, result.Agent, result.Rubric, result.Prompt)
		return 0
	case "grade":
		flags := flag.NewFlagSet("grade", flag.ContinueOnError)
		flags.SetOutput(stderr)
		casesRoot := flags.String("cases", "evals/cases", "evaluation cases directory")
		caseID := flags.String("case", "", "evaluation case id")
		target := flags.String("target", "", "prepared evaluation repository")
		jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(arguments[1:]); err != nil {
			return 2
		}
		if *caseID == "" || *target == "" {
			fmt.Fprintln(stderr, "error: --case and --target are required")
			return 2
		}
		result, err := evalharness.Grade(evalharness.GradeOptions{CasesRoot: *casesRoot, CaseID: *caseID, Target: *target})
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			return 2
		}
		if *jsonOutput {
			if code := encodeJSON(stdout, result); code != 0 {
				return code
			}
		} else {
			fmt.Fprintf(stdout, "deterministic grade: %s (%d passed, %d failed)\n", result.DeterministicStatus, result.Passed, result.Failed)
			for _, check := range result.Checks {
				fmt.Fprintf(stdout, "[%s] %s: %s\n", strings.ToUpper(check.Status), check.ID, check.Detail)
			}
			fmt.Fprintf(stdout, "semantic grade: pending human/model review\nrubric: %s\noverall: %s\n", result.Rubric, result.OverallStatus)
		}
		if result.DeterministicStatus == "fail" {
			return 1
		}
		return 0
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n", arguments[0])
		return 2
	}
}

func encodeJSON(output io.Writer, value any) int {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return 2
	}
	return 0
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage: repo-knowledge-eval <list|prepare|grade> [options]")
}
