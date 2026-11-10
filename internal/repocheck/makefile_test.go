package repocheck

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// makeTargets returns the targets a Makefile defines and the names its
// .PHONY line declares.
func makeTargets(makefile string) (targets, phony []string) {
	rule := regexp.MustCompile(`^([A-Za-z0-9_.-]+):([^=]|$)`)
	for _, line := range strings.Split(makefile, "\n") {
		if rest, ok := strings.CutPrefix(line, ".PHONY:"); ok {
			phony = append(phony, strings.Fields(rest)...)
			continue
		}
		if m := rule.FindStringSubmatch(line); m != nil && !strings.HasPrefix(m[1], ".") {
			targets = append(targets, m[1])
		}
	}
	return targets, phony
}

// TestMakefileTargetsArePhony keeps make from comparing a target with a
// file of its name: every target is a command, so .PHONY lists exactly the
// targets the Makefile defines. A directory named like a target (test/ is
// one) otherwise makes the target "up to date" and run nothing.
func TestMakefileTargetsArePhony(t *testing.T) {
	targets, phony := makeTargets(readRepoFile(t, "Makefile"))
	if len(targets) < 10 {
		t.Fatalf("found %d targets: %q", len(targets), targets)
	}
	for _, target := range targets {
		if !slices.Contains(phony, target) {
			t.Errorf("target %s is not in .PHONY", target)
		}
	}
	for _, name := range phony {
		if !slices.Contains(targets, name) {
			t.Errorf(".PHONY names %s, which the Makefile does not define", name)
		}
	}
	for _, gate := range []string{"verify", "check", "ci"} {
		if !slices.Contains(targets, gate) {
			t.Errorf("the Makefile has no %s target (core ADR-0044)", gate)
		}
	}
}

func TestMakeTargetsFixture(t *testing.T) {
	targets, phony := makeTargets(".PHONY: a b\n\nVAR := 1\nOTHER ?= x:y\na: b\n\techo\nb:\n.SUFFIXES:\n")
	if strings.Join(targets, ",") != "a,b" || strings.Join(phony, ",") != "a,b" {
		t.Errorf("targets %q phony %q", targets, phony)
	}
}
