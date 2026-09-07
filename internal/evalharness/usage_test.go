package evalharness

import (
	"strings"
	"testing"
)

func TestGradeRecordsSourceAttributedTokenUsage(t *testing.T) {
	target := t.TempDir() + "/target"
	if _, err := Prepare(benchmarkPrepareOptions(target, ConditionControl, "codex", 1)); err != nil {
		t.Fatal(err)
	}
	usage := &TokenUsageDetails{
		Source: "codex", InputTokens: int64Pointer(700), OutputTokens: int64Pointer(300),
		CachedTokens: int64Pointer(200), TotalTokens: int64Pointer(1000),
	}
	grade, err := Grade(GradeOptions{
		CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target, Usage: usage,
	})
	if err != nil {
		t.Fatal(err)
	}
	if grade.Usage == nil || grade.Usage.Source != "codex" || grade.Usage.TotalTokens == nil || *grade.Usage.TotalTokens != 1000 {
		t.Fatalf("grade usage = %+v", grade.Usage)
	}
	if grade.TokenUsage == nil || *grade.TokenUsage != 1000 {
		t.Fatalf("legacy token total was not retained: %+v", grade)
	}
}

func TestGradeRejectsUnattributedOrInconsistentTokenUsage(t *testing.T) {
	target := t.TempDir() + "/target"
	if _, err := Prepare(benchmarkPrepareOptions(target, ConditionControl, "codex", 1)); err != nil {
		t.Fatal(err)
	}
	total := int64(100)
	for _, test := range []struct {
		name  string
		usage *TokenUsageDetails
		want  string
	}{
		{name: "missing source", usage: &TokenUsageDetails{TotalTokens: &total}, want: "source"},
		{name: "unavailable with value", usage: &TokenUsageDetails{Source: TokenSourceUnavailable, TotalTokens: &total}, want: "unavailable"},
		{name: "inconsistent total", usage: &TokenUsageDetails{Source: "codex", InputTokens: int64Pointer(80), OutputTokens: int64Pointer(30), TotalTokens: &total}, want: "total"},
		{name: "unsupported source", usage: &TokenUsageDetails{Source: "estimated", TotalTokens: &total}, want: "unsupported"},
		{name: "negative detail", usage: &TokenUsageDetails{Source: "codex", InputTokens: int64Pointer(-1)}, want: "negative"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := Grade(GradeOptions{CasesRoot: testBenchmarksRoot(t), CaseID: "frontend-onboarding", Target: target, Usage: test.usage})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Grade() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestTokenUsageCompatibilityAndUnavailableState(t *testing.T) {
	legacy := int64(42)
	usage, total, err := validateTokenUsage(&TokenUsageDetails{Source: "manual"}, &legacy)
	if err != nil || usage.TotalTokens == nil || *usage.TotalTokens != legacy || total == nil || *total != legacy {
		t.Fatalf("legacy compatibility = %+v, %v, %v", usage, total, err)
	}
	usage, total, err = validateTokenUsage(&TokenUsageDetails{
		Source: "cursor", InputTokens: int64Pointer(30), OutputTokens: int64Pointer(12),
	}, nil)
	if err != nil || usage.TotalTokens == nil || *usage.TotalTokens != 42 || total == nil || *total != 42 {
		t.Fatalf("derived total = %+v, %v, %v", usage, total, err)
	}
	usage, total, err = validateTokenUsage(&TokenUsageDetails{Source: TokenSourceUnavailable}, nil)
	if err != nil || usage.Source != TokenSourceUnavailable || total != nil {
		t.Fatalf("unavailable usage = %+v, %v, %v", usage, total, err)
	}
	other := int64(43)
	if _, _, err := validateTokenUsage(&TokenUsageDetails{Source: "codex", TotalTokens: &other}, &legacy); err == nil || !strings.Contains(err.Error(), "disagree") {
		t.Fatalf("mismatched legacy total error = %v", err)
	}
}

func int64Pointer(value int64) *int64 { return &value }
