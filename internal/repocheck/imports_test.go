package repocheck

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// allowedModules are the only modules the kit may import besides the
// standard library and its own packages: MyGo draws, purego loads GLib.
var allowedModules = []string{
	"github.com/egoist/mygo",
	"github.com/ebitengine/purego",
}

// ownPackage reports whether p is one of the kit's own packages: the
// module's packages outside internal/ and examples/.
func ownPackage(p string) bool {
	if p == modulePath {
		return true
	}
	rest, ok := strings.CutPrefix(p, modulePath+"/")
	return ok && rest != "internal" && rest != "examples" &&
		!strings.HasPrefix(rest, "internal/") && !strings.HasPrefix(rest, "examples/")
}

// importProblems checks one file of the kit, named by its slash-separated
// path from the repository root, against the kit's import rule: the
// standard library, MyGo, purego and the kit's own packages, and the
// standard library alone in the test helpers (frostuitest/). It returns
// the number of imports outside the standard library and one line per
// import that breaks the rule.
func importProblems(rel string, src []byte) (external int, problems []string, err error) {
	file, err := parser.ParseFile(token.NewFileSet(), rel, src, parser.ImportsOnly)
	if err != nil {
		return 0, nil, err
	}
	stdlibOnly := strings.HasPrefix(rel, "frostuitest/")
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return 0, nil, err
		}
		if !strings.Contains(strings.SplitN(p, "/", 2)[0], ".") {
			continue // standard library
		}
		external++
		if stdlibOnly {
			problems = append(problems, fmt.Sprintf("%s imports %s; frostuitest may import only the standard library", rel, p))
			continue
		}
		ok := ownPackage(p)
		for _, allowed := range allowedModules {
			ok = ok || p == allowed || strings.HasPrefix(p, allowed+"/")
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("%s imports %s; frostui may import only the standard library, MyGo, purego and its own packages", rel, p))
		}
	}
	return external, problems, nil
}

// TestKitImportsOnlyMyGoAndPurego keeps the kit small and portable: none
// of its files, in any of its packages and tests included, may import a
// module other than MyGo and purego, nor the repository's internal/ or
// examples/ packages. Its test helpers (frostuitest) import the standard
// library only, so an app's tests can use them whatever UI they draw.
func TestKitImportsOnlyMyGoAndPurego(t *testing.T) {
	root := repoRoot(t)
	checked, helpers, external := 0, 0, 0
	for _, f := range kitFiles(t, root) {
		src, err := os.ReadFile(f.path)
		if err != nil {
			t.Fatal(err)
		}
		n, problems, err := importProblems(f.rel, src)
		if err != nil {
			t.Fatalf("parse %s: %v", f.rel, err)
		}
		checked++
		external += n
		if strings.HasPrefix(f.rel, "frostuitest/") {
			helpers++
		}
		for _, problem := range problems {
			t.Error(problem)
		}
	}
	t.Logf("%d files of the kit (%d in frostuitest) import %d packages outside the standard library", checked, helpers, external)
	if checked < 20 || helpers == 0 || external == 0 {
		t.Fatalf("saw %d files of the kit (%d in frostuitest, %d imports outside the standard library), too few", checked, helpers, external)
	}
}

// TestKitImportFixtures proves importProblems can fail: each fixture that
// breaks the kit's import rule must be reported, and the one that keeps it
// must pass.
func TestKitImportFixtures(t *testing.T) {
	fixtures := []struct {
		name, rel string
		imports   []string
		fail      bool
	}{
		{"an app's package", "shell.go", []string{"fmt", "github.com/frostyard/chairlift/internal/config"}, true},
		{"a third-party module", "theme.go", []string{"gopkg.in/yaml.v3"}, true},
		{"golang.org/x, which is not the standard library", "appid_windows.go", []string{"golang.org/x/sys/windows"}, true},
		{"a module that only shares MyGo's prefix", "viz.go", []string{"github.com/egoist/mygoextra"}, true},
		{"a module that only shares the kit's prefix", "viz.go", []string{"github.com/frostyard/frostuikit"}, true},
		{"the repository's guards", "jobs.go", []string{"github.com/frostyard/frostui/internal/repocheck"}, true},
		{"an example", "jobs.go", []string{"github.com/frostyard/frostui/examples/commandcenter"}, true},
		{"frostuitest importing MyGo", "frostuitest/loop.go", []string{"sync", "github.com/egoist/mygo/ui"}, true},
		{"frostuitest importing the kit", "frostuitest/loop.go", []string{"github.com/frostyard/frostui"}, true},
		{"the standard library, MyGo, purego and the kit", "jobs.go", []string{
			"context", "os/exec", "github.com/egoist/mygo", "github.com/egoist/mygo/ui",
			"github.com/ebitengine/purego", "github.com/frostyard/frostui", "github.com/frostyard/frostui/frostuitest",
		}, false},
	}
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			var src strings.Builder
			src.WriteString("package frostui\n\nimport (\n")
			for _, imp := range fx.imports {
				fmt.Fprintf(&src, "\t%q\n", imp)
			}
			src.WriteString(")\n")
			_, problems, err := importProblems(fx.rel, []byte(src.String()))
			if err != nil {
				t.Fatalf("parse: %v\n%s", err, src.String())
			}
			switch {
			case fx.fail && len(problems) == 0:
				t.Errorf("%s imports %q, which breaks the rule, but was not reported", fx.rel, fx.imports)
			case !fx.fail && len(problems) > 0:
				t.Errorf("the allowed imports were reported: %q", problems)
			}
		})
	}
}

// TestModuleRequiresOnlyMyGoAndPurego keeps go.mod to the kit's two
// direct dependencies, so that a test or an example cannot add a module
// every app that imports the kit would then download.
func TestModuleRequiresOnlyMyGoAndPurego(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot(t), "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	got := directRequires(string(data))
	slices.Sort(got)
	want := slices.Clone(allowedModules)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("go.mod requires %q directly, want only %q", got, want)
	}
}

func TestDirectRequiresSkipsIndirect(t *testing.T) {
	gomod := "module m\n\ngo 1.27.3\n\nrequire example.org/single v1.0.0\n\nrequire (\n\t// a comment\n\texample.org/a v1.0.0\n\texample.org/b v1.0.0 // indirect\n)\n"
	if got := strings.Join(directRequires(gomod), ","); got != "example.org/single,example.org/a" {
		t.Errorf("directRequires = %q", got)
	}
}
