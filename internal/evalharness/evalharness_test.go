package evalharness

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEvaluationSpecsLoad(t *testing.T) {
	t.Parallel()
	specs, err := ListCases(testCasesRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("len(specs) = %d, want 2", len(specs))
	}
	if specs[0].ID != "backend-clean-architecture" || specs[1].ID != "frontend-nextjs" {
		t.Fatalf("unexpected cases: %q, %q", specs[0].ID, specs[1].ID)
	}
}

func TestPreparedShallowFixturesFailDeterministicGrade(t *testing.T) {
	t.Parallel()
	for _, caseID := range []string{"frontend-nextjs", "backend-clean-architecture"} {
		caseID := caseID
		t.Run(caseID, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "target")
			prepared, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: caseID, Output: target, Agent: "codex"})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(prepared.Prompt, "generation request") {
				t.Fatalf("prompt does not preserve generation intent: %q", prepared.Prompt)
			}
			if _, err := os.Stat(filepath.Join(target, ".agents", "skills", "repository-knowledge", "SKILL.md")); err != nil {
				t.Fatalf("installed skill: %v", err)
			}
			grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: caseID, Target: target})
			if err != nil {
				t.Fatal(err)
			}
			if grade.DeterministicStatus != "fail" || grade.Failed == 0 {
				t.Fatalf("initial grade = %+v, want deterministic failures", grade)
			}
			if findCheck(t, grade, "fixture-source-preserved").Status != "pass" {
				t.Fatalf("untouched fixture source should pass protection: %+v", grade.Checks)
			}
		})
	}
}

func TestFrontendObjectiveCandidatePassesDeterministicChecks(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "claude-code"}); err != nil {
		t.Fatal(err)
	}
	writeFrontendObjectiveCandidate(t, target)
	grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if grade.DeterministicStatus != "pass" || grade.Failed != 0 {
		t.Fatalf("deterministic grade = %+v, want pass", grade)
	}
	if grade.OverallStatus != "pending_semantic_review" || grade.SemanticStatus != "pending_human_or_model_review" {
		t.Fatalf("semantic result was overstated: %+v", grade)
	}
}

func TestProtectedSourceMutationFails(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	if _, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "antigravity-ide"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, "src", "data", "intelligence.json")
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	grade, err := Grade(GradeOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Target: target})
	if err != nil {
		t.Fatal(err)
	}
	check := findCheck(t, grade, "fixture-source-preserved")
	if check.Status != "fail" || !strings.Contains(check.Detail, "src/data/intelligence.json (modified)") {
		t.Fatalf("source protection check = %+v", check)
	}
}

func TestPrepareRefusesNonEmptyOutput(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "keep.txt"), []byte("preserve\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "codex"})
	if err == nil || !strings.Contains(err.Error(), "must be empty") {
		t.Fatalf("Prepare() error = %v, want non-empty refusal", err)
	}
}

func TestPrepareSupportsCursorAdapter(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "target")
	prepared, err := Prepare(PrepareOptions{CasesRoot: testCasesRoot(t), CaseID: "frontend-nextjs", Output: target, Agent: "cursor"})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Agent != "cursor" {
		t.Fatalf("prepared agent = %q, want cursor", prepared.Agent)
	}
	for _, relative := range []string{
		".cursor/rules/repository-knowledge.mdc",
		".cursor/skills/repository-knowledge/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("installed Cursor adapter file %s: %v", relative, err)
		}
	}
}

func TestGlobMatchSupportsZeroOrMoreDirectories(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"docs/index.md", "docs/frontend/state.md", "docs/a/b/c.md"} {
		if !globMatch("docs/**/*.md", value) {
			t.Errorf("glob did not match %q", value)
		}
	}
	if globMatch("docs/**/*.md", "src/index.md") {
		t.Fatal("docs glob matched source file")
	}
}

func testCasesRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "evals", "cases"))
}

func findCheck(t *testing.T, grade GradeResult, id string) CheckResult {
	t.Helper()
	for _, check := range grade.Checks {
		if check.ID == id {
			return check
		}
	}
	t.Fatalf("check %q not found", id)
	return CheckResult{}
}

func writeFrontendObjectiveCandidate(t *testing.T, target string) {
	t.Helper()
	content := `# Verified frontend mechanisms

The application declares React 19 in ` + "`package.json`" + ` and resolves it in ` + "`package-lock.json`" + `.

Intelligence is static: ` + "`IntelligencePage`" + ` imports ` + "`src/data/intelligence.json`" + ` directly from ` + "`src/app/(protected)/intelligence/page.tsx`" + `.

Protected navigation starts at ` + "`src/app/(protected)/layout.tsx`" + `. ` + "`ProtectedLayout`" + ` calls ` + "`requireSession`" + ` from ` + "`src/lib/session.ts`" + ` and supplies permissions to ` + "`useCheckPermission`" + ` in ` + "`src/auth/PermissionContext.tsx`" + `.

Forms use ` + "`src/features/users/UserForm.tsx`" + ` with ` + "`userSchema`" + ` from ` + "`src/features/users/schema.ts`" + `. Server state uses ` + "`src/features/users/useUsers.ts`" + `. UI composition uses ` + "`src/components/PageShell.tsx`" + `.

The BFF boundary is ` + "`src/pages/api/[...paths]/index.ts`" + `.

Source: ` + "`src/lib/session.ts`" + ` — ` + "`requireSession`" + `.

` + "```ts" + `
const session = await requireSession();
` + "```" + `

Source: ` + "`src/features/users/UserForm.tsx`" + ` — ` + "`UserForm`" + `.

` + "```tsx" + `
const result = userSchema.safeParse(values);
` + "```" + `

Source: ` + "`src/pages/api/[...paths]/index.ts`" + ` — ` + "`handler`" + `.

` + "```ts" + `
response.status(upstream.status).json(payload);
` + "```" + `
`
	write := func(relative, value string) {
		t.Helper()
		path := filepath.Join(target, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("docs/index.md", "# Repository knowledge index\n\n- [Overview](repository-overview.md)\n- [Architecture](architecture.md)\n- [Features](features.md)\n- [Integrations](integrations.md)\n- [Development](development.md)\n- [State](state.md)\n")
	write("docs/repository-overview.md", content)
	write("docs/architecture.md", content)
	write("docs/features.md", content)
	write("docs/integrations.md", content)
	write("docs/development.md", content)
	write("docs/state.md", content)
	write(".repo-knowledge/repository.json", `{
  "$schema": "schemas/repository.schema.json",
  "schema_version": "1.0",
  "repository": {"name": "eval-frontend", "kind": "frontend", "owners": []},
  "documentation": {"index": "docs/index.md", "generated_inventory": "docs/repository-inventory.md"},
  "capabilities": [
    {"id": "navigation", "documentation": ["docs/architecture.md"], "evidence": ["src/app/"]},
    {"id": "permissions", "documentation": ["docs/architecture.md"], "evidence": ["src/auth/"]},
    {"id": "users", "documentation": ["docs/features.md"], "evidence": ["src/features/users/"]},
    {"id": "bff", "documentation": ["docs/integrations.md"], "evidence": ["src/pages/api/"]}
  ],
  "ci": {"enforcement": "advisory"}
}
`)
}
