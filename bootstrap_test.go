package repositoryknowledge

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUnixBootstrapScriptSyntax(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell test")
	}
	command := exec.Command("sh", "-n", "skills/repository-knowledge/scripts/install-binary.sh")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sh -n: %v\n%s", err, output)
	}
}

func TestUnixBootstrapInstallsVerifiedPinnedFixture(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("bootstrap release mapping supports macOS and Linux")
	}
	architecture := map[string]string{"amd64": "amd64", "arm64": "arm64"}[runtime.GOARCH]
	if architecture == "" {
		t.Skipf("unsupported test architecture %s", runtime.GOARCH)
	}
	platform := map[string]string{"darwin": "darwin", "linux": "linux"}[runtime.GOOS]
	artifact := fmt.Sprintf("repo-knowledge-%s-%s", platform, architecture)
	payload := []byte("#!/bin/sh\nprintf 'repo-knowledge 0.8.0\\n'\n")
	digest := sha256.Sum256(payload)

	root := t.TempDir()
	releaseDir := filepath.Join(root, "release")
	toolsDir := filepath.Join(root, "tools")
	binDir := filepath.Join(root, "bin")
	for _, directory := range []string{releaseDir, toolsDir, binDir} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(releaseDir, artifact), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	checksum := fmt.Sprintf("%x  %s\n", digest, artifact)
	if err := os.WriteFile(filepath.Join(releaseDir, "SHA256SUMS"), []byte(checksum), 0o644); err != nil {
		t.Fatal(err)
	}
	fakeCurl := `#!/bin/sh
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --output) output=$2; shift 2 ;;
    *) url=$1; shift ;;
  esac
done
cp "$FAKE_RELEASE_DIR/${url##*/}" "$output"
`
	if err := os.WriteFile(filepath.Join(toolsDir, "curl"), []byte(fakeCurl), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "gh"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	command := exec.Command("sh", "skills/repository-knowledge/scripts/install-binary.sh",
		"--version", "v0.8.0", "--bin-dir", binDir, "--yes")
	command.Env = append(os.Environ(),
		"FAKE_RELEASE_DIR="+releaseDir,
		"HOME="+filepath.Join(root, "home"),
		"PATH="+strings.Join([]string{toolsDir, binDir, os.Getenv("PATH")}, string(os.PathListSeparator)),
	)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("bootstrap: %v\n%s", err, output)
	}
	destination := filepath.Join(binDir, "repo-knowledge")
	installed, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(installed) != string(payload) {
		t.Fatalf("installed payload = %q, want %q", installed, payload)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("installed mode = %v, want 0755", info.Mode().Perm())
	}
}

func TestUnixBootstrapRequiresConfirmationBeforeDownload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell test")
	}
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("sh", "skills/repository-knowledge/scripts/install-binary.sh",
		"--version", "v0.8.0", "--bin-dir", binDir)
	command.Env = append(os.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "confirmation required") {
		t.Fatalf("bootstrap error = %v, output = %q", err, output)
	}
	if _, err := os.Stat(filepath.Join(binDir, "repo-knowledge")); !os.IsNotExist(err) {
		t.Fatalf("unconfirmed bootstrap wrote destination; stat error = %v", err)
	}
}

func TestUnixBootstrapRejectsUnpinnedVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell test")
	}
	command := exec.Command("sh", "skills/repository-knowledge/scripts/install-binary.sh",
		"--version", "latest", "--dry-run")
	output, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "pinned semantic release tag") {
		t.Fatalf("bootstrap error = %v, output = %q", err, output)
	}
}
