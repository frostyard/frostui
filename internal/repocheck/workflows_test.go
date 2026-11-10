package repocheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCrossTargetsAreListedEverywhere keeps the three lists of targets
// equal: this test's, the Makefile's CROSS_TARGETS and the matrix of the
// cross-build job in .github/workflows/ci.yml.
func TestCrossTargetsAreListedEverywhere(t *testing.T) {
	var makeTargets string
	for _, line := range strings.Split(readRepoFile(t, "Makefile"), "\n") {
		if rest, ok := strings.CutPrefix(line, "CROSS_TARGETS ?="); ok {
			makeTargets = strings.Join(strings.Fields(rest), " ")
		}
	}
	if want := strings.Join(crossTargets, " "); makeTargets != want {
		t.Errorf("Makefile CROSS_TARGETS = %q, want %q", makeTargets, want)
	}
	workflow := readRepoFile(t, ".github/workflows/ci.yml")
	for _, target := range crossTargets {
		if !strings.Contains(workflow, "- "+target+"\n") {
			t.Errorf(".github/workflows/ci.yml has no matrix entry %q", "- "+target)
		}
	}
}

// pinnedUses matches an external action pinned to a full commit SHA with
// its version in a trailing comment.
var pinnedUses = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+@[0-9a-f]{40} # \S.*$`)

// TestWorkflowActionsArePinned keeps every external uses: of every
// workflow on a full 40-character commit SHA with the version in a
// trailing comment (core ADR-0021). Local actions (./) are exempt.
func TestWorkflowActionsArePinned(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repoRoot(t), ".github", "workflows", "*.y*ml"))
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
			value, ok := strings.CutPrefix(trimmed, "uses:")
			if !ok {
				continue
			}
			value = strings.TrimSpace(value)
			if strings.HasPrefix(value, "./") {
				continue
			}
			seen++
			if !pinnedUses.MatchString(value) {
				t.Errorf("%s:%d: uses: %s is not pinned to a 40-character SHA with a # version comment", filepath.Base(path), i+1, value)
			}
		}
	}
	if seen < 5 {
		t.Fatalf("saw %d external uses: lines in .github/workflows, too few", seen)
	}
}

func TestPinnedUsesFixtures(t *testing.T) {
	for value, want := range map[string]bool{
		"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1":                                  true,
		"frostyard/core/.github/workflows/pr-triage.yml@c996ea3fb54bcf1a3c621113522213702a7673e2 # core main": true,
		"actions/checkout@v7": false,
		"actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1": false,
		"actions/checkout@3d3c42e5 # v7.0.1":                        false,
	} {
		if got := pinnedUses.MatchString(value); got != want {
			t.Errorf("pinnedUses(%q) = %v, want %v", value, got, want)
		}
	}
}
