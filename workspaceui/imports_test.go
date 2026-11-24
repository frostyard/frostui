package workspaceui

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// nsgoModule is the one module, beside the standard library, that the kit
// imports and that its go.mod requires directly.
const nsgoModule = "github.com/nsgo-dev/nsgo"

// allowedImport reports whether a file of the kit may import path: the
// standard library (no dot in the first path element) and NSGo's packages.
func allowedImport(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	if !strings.Contains(first, ".") {
		return true
	}
	return path == nsgoModule || strings.HasPrefix(path, nsgoModule+"/")
}

// TestKitImportsOnlyNSGo keeps every file of the module, tests included, to
// the standard library and NSGo: an app that imports the kit takes no other
// requirement from it, and never MyGo, which the repository's root module
// draws with.
func TestKitImportsOnlyNSGo(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 6 {
		t.Fatalf("found %d Go files in the module's directory: %q", len(files), files)
	}
	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !allowedImport(path) {
				t.Errorf("%s imports %q, want only the standard library and %s", file, path, nsgoModule)
			}
		}
	}
}

func TestAllowedImport(t *testing.T) {
	for path, want := range map[string]bool{
		"strings":                          true,
		"go/parser":                        true,
		nsgoModule:                         true,
		nsgoModule + "/ui":                 true,
		nsgoModule + "-extras/ui":          false,
		"github.com/egoist/mygo/ui":        false,
		"github.com/frostyard/frostui":     false,
		"golang.org/x/text/unicode/norm":   false,
		"github.com/nsgo-dev/other/widget": false,
	} {
		if got := allowedImport(path); got != want {
			t.Errorf("allowedImport(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestModuleRequiresOnlyNSGo keeps go.mod to one direct requirement.
func TestModuleRequiresOnlyNSGo(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	var direct []string
	inBlock := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "require (":
			inBlock = true
			continue
		case line == ")":
			inBlock = false
			continue
		case strings.HasPrefix(line, "require "):
			line = strings.TrimPrefix(line, "require ")
		case !inBlock:
			continue
		}
		if line == "" || strings.HasSuffix(line, "// indirect") {
			continue
		}
		module, _, _ := strings.Cut(line, " ")
		direct = append(direct, module)
	}
	if len(direct) != 1 || direct[0] != nsgoModule {
		t.Errorf("go.mod requires %q directly, want only %q", direct, nsgoModule)
	}
}
