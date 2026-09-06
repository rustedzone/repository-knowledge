package repositoryknowledge

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGitHubAdapterValidatesCommittedRange(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("GitHub adapter executes on an Ubuntu runner")
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "repo-knowledge")
	runGitHubAdapterCommand(t, root, nil, "go", "build", "-trimpath", "-o", binary, "./cmd/repo-knowledge")

	target := t.TempDir()
	runGitHubAdapterCommand(t, target, nil, "git", "init", "-q")
	runGitHubAdapterCommand(t, target, nil, "git", "config", "user.email", "test@example.com")
	runGitHubAdapterCommand(t, target, nil, "git", "config", "user.name", "Repository Knowledge Test")
	runGitHubAdapterCommand(t, root, nil, binary, "install", "--target", target, "--source", "test", "--ref", "v0.9.0")
	runGitHubAdapterCommand(t, target, nil, "git", "add", ".")
	runGitHubAdapterCommand(t, target, nil, "git", "commit", "-q", "-m", "baseline")
	base := strings.TrimSpace(runGitHubAdapterCommand(t, target, nil, "git", "rev-parse", "HEAD"))

	if err := os.MkdirAll(filepath.Join(target, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "src", "logic.go"), []byte("package fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitHubAdapterCommand(t, target, nil, "git", "add", ".")
	runGitHubAdapterCommand(t, target, nil, "git", "commit", "-q", "-m", "change behavior")
	head := strings.TrimSpace(runGitHubAdapterCommand(t, target, nil, "git", "rev-parse", "HEAD"))

	adapter := filepath.Join(root, "adapters", "github", "documentation-check.sh")
	advisoryOutput := filepath.Join(t.TempDir(), "advisory.json")
	environment := []string{
		"REPO_KNOWLEDGE_BINARY=" + binary,
		"REPO_KNOWLEDGE_TARGET=" + target,
		"REPO_KNOWLEDGE_OUTPUT=" + advisoryOutput,
		"REPO_KNOWLEDGE_ENFORCEMENT=advisory",
		"REPO_KNOWLEDGE_BASE_SHA=" + base,
		"REPO_KNOWLEDGE_HEAD_SHA=" + head,
	}
	runGitHubAdapterCommand(t, root, environment, "sh", adapter)
	report := readGitHubAdapterReport(t, advisoryOutput)
	if report.Status != "advisory" || report.Enforcement != "advisory" || report.Base != base || report.Head != head {
		t.Fatalf("advisory report = %+v", report)
	}
	initialOutput := filepath.Join(t.TempDir(), "initial-push.json")
	initialEnvironment := append([]string(nil), environment...)
	initialEnvironment[2] = "REPO_KNOWLEDGE_OUTPUT=" + initialOutput
	initialEnvironment[4] = "REPO_KNOWLEDGE_BASE_SHA=" + strings.Repeat("0", 40)
	runGitHubAdapterCommand(t, root, initialEnvironment, "sh", adapter)
	initialReport := readGitHubAdapterReport(t, initialOutput)
	if initialReport.Base == strings.Repeat("0", 40) || initialReport.Base == strings.Repeat("0", 64) ||
		(len(initialReport.Base) != 40 && len(initialReport.Base) != 64) || initialReport.Head != head {
		t.Fatalf("initial-push report = %+v", initialReport)
	}

	acknowledgmentOutput := filepath.Join(t.TempDir(), "acknowledgment.json")
	environment[2] = "REPO_KNOWLEDGE_OUTPUT=" + acknowledgmentOutput
	environment[3] = "REPO_KNOWLEDGE_ENFORCEMENT=acknowledgment"
	command := exec.Command("sh", adapter)
	command.Dir = root
	command.Env = append(os.Environ(), environment...)
	output, err := command.CombinedOutput()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
		t.Fatalf("acknowledgment adapter error = %v, output = %s", err, output)
	}
	report = readGitHubAdapterReport(t, acknowledgmentOutput)
	if report.Status != "fail" || report.Enforcement != "acknowledgment" {
		t.Fatalf("acknowledgment report = %+v", report)
	}
}

func TestGitHubAdapterRejectsInvalidInputs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("GitHub adapter executes on an Ubuntu runner")
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	adapter := filepath.Join(root, "adapters", "github", "documentation-check.sh")
	baseEnvironment := []string{
		"REPO_KNOWLEDGE_BINARY=/usr/bin/true",
		"REPO_KNOWLEDGE_TARGET=" + t.TempDir(),
		"REPO_KNOWLEDGE_OUTPUT=" + filepath.Join(t.TempDir(), "report.json"),
		"REPO_KNOWLEDGE_HEAD_SHA=" + strings.Repeat("a", 40),
	}
	tests := []struct {
		name        string
		environment string
		want        string
	}{
		{name: "mode", environment: "REPO_KNOWLEDGE_ENFORCEMENT=disabled", want: "unsupported enforcement mode"},
		{name: "base object", environment: "REPO_KNOWLEDGE_BASE_SHA=main", want: "base must be a Git object ID"},
		{name: "missing base object", environment: "REPO_KNOWLEDGE_BASE_SHA=" + strings.Repeat("b", 40), want: "Git object is unavailable in the full checkout"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command("sh", adapter)
			command.Dir = root
			command.Env = append(os.Environ(), append(baseEnvironment, test.environment)...)
			output, err := command.CombinedOutput()
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) || exitError.ExitCode() != 2 || !strings.Contains(string(output), test.want) {
				t.Fatalf("adapter error = %v, output = %s", err, output)
			}
		})
	}
}

func TestGitHubReusableWorkflowSecurityContract(t *testing.T) {
	if runtime.GOOS != "windows" {
		adapter := "adapters/github/documentation-check.sh"
		if output, err := exec.Command("sh", "-n", adapter).CombinedOutput(); err != nil {
			t.Fatalf("sh -n %s: %v\n%s", adapter, err, output)
		}
		info, err := os.Stat(adapter)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Fatalf("GitHub adapter mode = %v, want 0755", info.Mode().Perm())
		}
	}
	data, err := os.ReadFile(".github/workflows/documentation-check.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	for _, required := range []string{
		"workflow_call:",
		"contents: read",
		"attestations: read",
		"fetch-depth: 0",
		"persist-credentials: false",
		"gh attestation verify",
		"actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803",
		"actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02",
		"if: always()",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("reusable workflow missing %q", required)
		}
	}
	for _, forbidden := range []string{"pull_request_target", "contents: write", "secrets: inherit"} {
		if strings.Contains(workflow, forbidden) {
			t.Errorf("reusable workflow contains unsafe contract %q", forbidden)
		}
	}

	makefile, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(makefile), "repo-knowledge-github-adapter.sh") {
		t.Fatal("release Makefile does not package the GitHub adapter")
	}
}

type githubAdapterReport struct {
	Base        string `json:"base"`
	Head        string `json:"head"`
	Enforcement string `json:"enforcement"`
	Status      string `json:"status"`
}

func readGitHubAdapterReport(t *testing.T, path string) githubAdapterReport {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report githubAdapterReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("parse adapter report: %v\n%s", err, data)
	}
	return report
}

func runGitHubAdapterCommand(t *testing.T, directory string, environment []string, name string, arguments ...string) string {
	t.Helper()
	command := exec.Command(name, arguments...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(arguments, " "), err, output)
	}
	return string(output)
}
