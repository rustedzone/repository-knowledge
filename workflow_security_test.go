package repositoryknowledge

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	fullCommitSHA = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	dockerDigest  = regexp.MustCompile(`@sha256:[0-9a-fA-F]{64}$`)
)

func TestGitHubWorkflowActionReferencesAreImmutable(t *testing.T) {
	t.Parallel()
	workflowCount := 0
	err := filepath.WalkDir(".github/workflows", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || (filepath.Ext(path) != ".yml" && filepath.Ext(path) != ".yaml") {
			return nil
		}
		workflowCount++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, issue := range immutableUsesIssues(string(data)) {
			t.Errorf("%s: %s", path, issue)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if workflowCount == 0 {
		t.Fatal("no GitHub workflow files were examined")
	}
}

func TestImmutableUsesValidationRejectsMutableReferences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		reference string
		wantIssue bool
	}{
		{name: "mutable version tag", reference: "actions/checkout@v6", wantIssue: true},
		{name: "branch", reference: "actions/checkout@main", wantIssue: true},
		{name: "short SHA", reference: "actions/checkout@d23441a", wantIssue: true},
		{name: "full commit SHA", reference: "actions/checkout@d23441a48e516b6c34aea4fa41551a30e30af803"},
		{name: "local action", reference: "./.github/actions/check"},
		{name: "mutable Docker tag", reference: "docker://alpine:3.21", wantIssue: true},
		{name: "Docker digest", reference: "docker://alpine@sha256:" + strings.Repeat("a", 64)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workflow := "steps:\n  - uses: " + test.reference + "\n"
			issues := immutableUsesIssues(workflow)
			if test.wantIssue && len(issues) == 0 {
				t.Fatalf("reference %q was accepted; want immutable-reference error", test.reference)
			}
			if !test.wantIssue && len(issues) != 0 {
				t.Fatalf("reference %q was rejected: %v", test.reference, issues)
			}
		})
	}
}

func TestReleaseWorkflowPermissionBoundary(t *testing.T) {
	t.Parallel()
	workflow := readWorkflow(t, ".github/workflows/release.yml")
	build, publish, found := strings.Cut(workflow, "\n  publish:\n")
	if !found {
		t.Fatal("release workflow must separate build and publish jobs")
	}
	for _, forbidden := range []string{"contents: write", "id-token: write", "attestations: write"} {
		if strings.Contains(build, forbidden) {
			t.Errorf("release build boundary contains %q", forbidden)
		}
	}
	for _, required := range []string{
		"permissions: {}",
		"contents: read",
		"persist-credentials: false",
		"actions/upload-artifact@ea165f8d65b6e75b540449e92b4886f43607fa02",
	} {
		if !strings.Contains(build, required) {
			t.Errorf("release build boundary missing %q", required)
		}
	}
	for _, required := range []string{
		"needs: build",
		"contents: write",
		"id-token: write",
		"attestations: write",
		"actions/download-artifact@634f93cb2916e3fdff6788551b99b062d0335ce0",
		"actions/attest@1e69f48acb82d1966a394da916b4c1698aa569d6",
	} {
		if !strings.Contains(publish, required) {
			t.Errorf("release publish boundary missing %q", required)
		}
	}
}

func TestCIWorkflowGoCompatibilityContract(t *testing.T) {
	t.Parallel()
	workflow := readWorkflow(t, ".github/workflows/ci.yml")
	for _, required := range []string{
		"'1.22.0'",
		"'1.25.14'",
		"go-version: ${{ matrix.go-version }}",
		"make check",
		"make build",
		"./cmd/repo-knowledge-eval",
		"go test -race ./...",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("CI compatibility contract missing %q", required)
		}
	}
}

func immutableUsesIssues(workflow string) []string {
	var issues []string
	for index, line := range strings.Split(workflow, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		if !strings.HasPrefix(trimmed, "uses:") {
			continue
		}
		reference := strings.TrimSpace(strings.TrimPrefix(trimmed, "uses:"))
		if fields := strings.Fields(reference); len(fields) > 0 {
			reference = strings.Trim(fields[0], `"'`)
		} else {
			reference = ""
		}
		if immutableUsesReference(reference) {
			continue
		}
		issues = append(issues, fmt.Sprintf("line %d uses mutable or unverified reference %q", index+1, reference))
	}
	return issues
}

func immutableUsesReference(reference string) bool {
	if strings.HasPrefix(reference, "./") {
		return true
	}
	if strings.HasPrefix(reference, "docker://") {
		return dockerDigest.MatchString(reference)
	}
	separator := strings.LastIndex(reference, "@")
	return separator > 0 && fullCommitSHA.MatchString(reference[separator+1:])
}

func readWorkflow(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
