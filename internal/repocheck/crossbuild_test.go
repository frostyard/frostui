package repocheck

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// crossTargets are the GOOS/GOARCH pairs the kit builds for with
// CGO_ENABLED=0: the desktops frostyard's apps ship on, and js/wasm as a
// compile check for reuse on the web.
var crossTargets = []string{
	"linux/amd64", "linux/arm64",
	"windows/amd64", "windows/arm64",
	"darwin/amd64", "darwin/arm64",
	"js/wasm",
}

// TestKitCrossCompiles builds and vets every package of the module, tests
// included (go vet type-checks them), for every target with CGO disabled,
// so a Linux-only call outside a _linux.go file fails here, not on a
// Windows user's machine.
func TestKitCrossCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: the cross-build check runs in make verify and in the cross-build jobs of CI")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("the go command is not on PATH, so the cross-build check cannot run go build and go vet for %s: %v", strings.Join(crossTargets, ", "), err)
	}
	root := repoRoot(t)
	for _, target := range crossTargets {
		t.Run(strings.ReplaceAll(target, "/", "_"), func(t *testing.T) {
			t.Parallel()
			goos, goarch, _ := strings.Cut(target, "/")
			for _, args := range [][]string{{"build", "./..."}, {"vet", "./..."}} {
				cmd := exec.CommandContext(t.Context(), goBin, args...)
				cmd.Dir = root
				cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+goarch)
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Errorf("CGO_ENABLED=0 GOOS=%s GOARCH=%s go %s: %v\n%s", goos, goarch, strings.Join(args, " "), err, out)
				}
			}
		})
	}
}
