package repocheck

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// modulePath is the kit's module path.
const modulePath = "github.com/frostyard/frostui"

// repoRoot returns the directory of the kit's go.mod, found upward from
// the test's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.HasPrefix(string(data), "module "+modulePath+"\n") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod of %s above the working directory", modulePath)
		}
		dir = parent
	}
}

// kitFile is one Go file of the kit: rel is its slash-separated path from
// the repository root.
type kitFile struct {
	rel, path string
}

// kitFiles returns every Go file of the kit, tests included: the files of
// every package in the module except internal/ (the guards) and examples/
// (programs that use the kit). It skips what the go tool's ./... pattern
// skips: directories that start with "." or "_", testdata, and nested
// modules.
func kitFiles(t *testing.T, root string) []kitFile {
	t.Helper()
	var files []kitFile
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			name := d.Name()
			if path == root {
				return nil
			}
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" ||
				rel == "internal" || rel == "examples" {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(rel, ".go") {
			files = append(files, kitFile{rel: rel, path: path})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// directRequires returns the module paths go.mod requires directly: every
// require line without an "// indirect" comment.
func directRequires(gomod string) []string {
	var out []string
	inBlock := false
	for s := bufio.NewScanner(strings.NewReader(gomod)); s.Scan(); {
		line := strings.TrimSpace(s.Text())
		switch {
		case line == "require (":
			inBlock = true
			continue
		case inBlock && line == ")":
			inBlock = false
			continue
		case strings.HasPrefix(line, "require "):
			line = strings.TrimSpace(strings.TrimPrefix(line, "require "))
		case !inBlock:
			continue
		}
		if line == "" || strings.HasPrefix(line, "//") || strings.Contains(line, "// indirect") {
			continue
		}
		out = append(out, strings.Fields(line)[0])
	}
	return out
}

func TestKitFilesSkipsWhatTheGoToolSkips(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{
		"go.mod", "a.go", "frostuitest/b.go", "sub/c.go",
		"internal/repocheck/x.go", "examples/demo/main.go", "testdata/t.go", "sub/testdata/u.go",
		".hidden/h.go", "_skip/s.go", "nested/go.mod", "nested/n.go", "notes.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package x\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var got []string
	for _, f := range kitFiles(t, root) {
		got = append(got, f.rel)
	}
	if strings.Join(got, ",") != "a.go,frostuitest/b.go,sub/c.go" {
		t.Errorf("kitFiles = %q", got)
	}
}
