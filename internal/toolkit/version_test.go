package toolkit

import "testing"

func TestCompareVersionsUsesSemanticVersionPrecedence(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		left       string
		right      string
		comparison int
	}{
		{name: "major", left: "2.0.0", right: "1.9.9", comparison: 1},
		{name: "minor", left: "1.2.0", right: "1.1.9", comparison: 1},
		{name: "patch", left: "1.2.3", right: "1.2.4", comparison: -1},
		{name: "release after prerelease", left: "0.6.0", right: "0.6.0-rc.1", comparison: 1},
		{name: "prerelease before release", left: "0.6.0-rc.1", right: "0.6.0", comparison: -1},
		{name: "numeric prerelease", left: "1.0.0-rc.10", right: "1.0.0-rc.2", comparison: 1},
		{name: "numeric before alphanumeric", left: "1.0.0-1", right: "1.0.0-alpha", comparison: -1},
		{name: "longer prerelease", left: "1.0.0-alpha.1", right: "1.0.0-alpha", comparison: 1},
		{name: "build metadata ignored", left: "v1.0.0+build.2", right: "1.0.0+build.1", comparison: 0},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if comparison := compareVersions(test.left, test.right); comparison != test.comparison {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, comparison, test.comparison)
			}
		})
	}
}
