package repocheck

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/api.txt from the kit's source")

// apiFile is the kit's exported API, one declaration per line, sorted.
const apiFile = "testdata/api.txt"

// kitPackages are the kit's packages, by directory from the repository
// root, with the name their API lines carry.
var kitPackages = map[string]string{".": "frostui", "frostuitest": "frostuitest"}

// exportedAPI lists what the packages under root export: every exported
// function, method, type, struct field, interface method, constant and
// variable, with its type and without parameter names, one per line,
// sorted. It reads the non-test files only, whatever their build tags.
func exportedAPI(root string) ([]string, error) {
	var lines []string
	for dir, pkg := range kitPackages {
		paths, err := filepath.Glob(filepath.Join(root, dir, "*.go"))
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				return nil, err
			}
			lines = append(lines, fileAPI(fset, pkg, file)...)
		}
	}
	slices.Sort(lines)
	return lines, nil
}

// fileAPI lists the exported declarations of one file.
func fileAPI(fset *token.FileSet, pkg string, file *ast.File) []string {
	var lines []string
	add := func(format string, args ...any) { lines = append(lines, fmt.Sprintf(format, args...)) }
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Recv == nil {
				add("func %s.%s%s", pkg, d.Name.Name, signature(fset, d.Type))
				continue
			}
			recv := d.Recv.List[0].Type
			if !ast.IsExported(baseType(recv)) {
				continue
			}
			add("method (%s) %s.%s%s", node(fset, recv), pkg, d.Name.Name, signature(fset, d.Type))
		case *ast.GenDecl:
			var lastType ast.Expr
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if !s.Name.IsExported() {
						continue
					}
					name := pkg + "." + s.Name.Name
					tparams := ""
					if s.TypeParams != nil {
						var params []string
						for _, f := range s.TypeParams.List {
							var names []string
							for _, n := range f.Names {
								names = append(names, n.Name)
							}
							params = append(params, strings.Join(names, ", ")+" "+node(fset, f.Type))
						}
						tparams = "[" + strings.Join(params, ", ") + "]"
					}
					switch tt := s.Type.(type) {
					case *ast.StructType:
						add("type %s%s struct", name, tparams)
						for _, f := range tt.Fields.List {
							if len(f.Names) == 0 {
								if ast.IsExported(baseType(f.Type)) {
									add("embed %s %s", name, node(fset, f.Type))
								}
								continue
							}
							for _, n := range f.Names {
								if n.IsExported() {
									add("field %s.%s %s", name, n.Name, node(fset, f.Type))
								}
							}
						}
					case *ast.InterfaceType:
						add("type %s%s interface", name, tparams)
						for _, m := range tt.Methods.List {
							for _, n := range m.Names {
								add("imethod %s.%s%s", name, n.Name, signature(fset, m.Type.(*ast.FuncType)))
							}
							if len(m.Names) == 0 {
								add("imethod %s %s", name, node(fset, m.Type))
							}
						}
					default:
						assign := " "
						if s.Assign.IsValid() {
							assign = " = "
						}
						add("type %s%s%s%s", name, tparams, assign, node(fset, s.Type))
					}
				case *ast.ValueSpec:
					if d.Tok == token.CONST && s.Type == nil && len(s.Values) == 0 {
						// An implicit repetition: the spec takes the type of
						// the last one that named it.
					} else {
						lastType = s.Type
					}
					for i, n := range s.Names {
						if !n.IsExported() {
							continue
						}
						kind := strings.ToLower(d.Tok.String())
						switch {
						case lastType != nil:
							add("%s %s.%s %s", kind, pkg, n.Name, node(fset, lastType))
						case i < len(s.Values):
							v := s.Values[i]
							if lit, ok := v.(*ast.CompositeLit); ok && lit.Type != nil {
								add("%s %s.%s %s", kind, pkg, n.Name, node(fset, lit.Type))
							} else {
								add("%s %s.%s = %s", kind, pkg, n.Name, node(fset, v))
							}
						default:
							add("%s %s.%s", kind, pkg, n.Name)
						}
					}
				}
			}
		}
	}
	return lines
}

// signature prints a function type without its parameter and result
// names: renaming a parameter changes no caller.
func signature(fset *token.FileSet, ft *ast.FuncType) string {
	strip := func(fl *ast.FieldList) *ast.FieldList {
		if fl == nil {
			return nil
		}
		out := &ast.FieldList{}
		for _, f := range fl.List {
			for range max(1, len(f.Names)) {
				out.List = append(out.List, &ast.Field{Type: f.Type})
			}
		}
		return out
	}
	s := node(fset, &ast.FuncType{TypeParams: ft.TypeParams, Params: strip(ft.Params), Results: strip(ft.Results)})
	return strings.TrimPrefix(s, "func")
}

// baseType is the name of a type expression's named type: T for T, *T,
// T[P] and pkg.T.
func baseType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return baseType(t.X)
	case *ast.IndexExpr:
		return baseType(t.X)
	case *ast.IndexListExpr:
		return baseType(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

// node prints n on one line.
func node(fset *token.FileSet, n any) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, fset, n); err != nil {
		return fmt.Sprintf("<%v>", err)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// TestExportedAPIIsUnchanged pins the kit's exported API to
// testdata/api.txt. An added line is a minor change; a removed or changed
// line breaks callers and needs a major version (a minor one while the
// module is v0) and a note in docs/design/frostui.md (AGENTS.md, "API
// stability"). After a deliberate change, rewrite the file with
//
//	go test ./internal/repocheck -run TestExportedAPIIsUnchanged -update
func TestExportedAPIIsUnchanged(t *testing.T) {
	root := repoRoot(t)
	got, err := exportedAPI(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 100 {
		t.Fatalf("the kit exports %d declarations, too few: the walk missed files", len(got))
	}
	path := filepath.Join(root, "internal", "repocheck", apiFile)
	if *update {
		if err := os.WriteFile(path, []byte(strings.Join(got, "\n")+"\n"), 0o644); err != nil { //nolint:gosec // a tracked text file
			t.Fatal(err)
		}
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var removed, added []string
	for _, line := range want {
		if !slices.Contains(got, line) {
			removed = append(removed, line)
		}
	}
	for _, line := range got {
		if !slices.Contains(want, line) {
			added = append(added, line)
		}
	}
	for _, line := range removed {
		t.Errorf("removed or changed (breaks callers: major version, or minor in v0, plus a note): %s", line)
	}
	for _, line := range added {
		t.Errorf("added (a minor version): %s", line)
	}
	if len(removed)+len(added) > 0 {
		t.Log("after a deliberate API change: go test ./internal/repocheck -run TestExportedAPIIsUnchanged -update")
	}
}

// knownOS and knownArch are the GOOS and GOARCH values a file name's
// suffix can carry (go tool dist list).
var (
	knownOS   = []string{"aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "zos"}
	knownArch = []string{"386", "amd64", "arm", "arm64", "loong64", "mips", "mipsle", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "wasm"}
)

// platformFile reports whether a Go file builds only on some platforms:
// it has a //go:build line, or its name ends in _GOOS, _GOARCH or
// _GOOS_GOARCH.
func platformFile(name string, src []byte) bool {
	base := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(name), ".go"), "_test")
	parts := strings.Split(base, "_")
	if n := len(parts); n > 1 {
		last := parts[n-1]
		if slices.Contains(knownOS, last) || slices.Contains(knownArch, last) {
			return true
		}
	}
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			break
		}
		if constraint.IsGoBuild(line) || constraint.IsPlusBuild(line) {
			return true
		}
	}
	return false
}

// TestPlatformFilesExportNothing keeps the API the same on every
// platform: a file that builds only on some platforms (appid_linux.go,
// icons_other.go) declares no exported identifier. The exported function
// lives in an untagged file with its doc comment and calls a lower-case
// one that each platform file defines.
func TestPlatformFilesExportNothing(t *testing.T) {
	root := repoRoot(t)
	seen := 0
	for _, f := range kitFiles(t, root) {
		if strings.HasSuffix(f.rel, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f.path)
		if err != nil {
			t.Fatal(err)
		}
		if !platformFile(f.rel, src) {
			continue
		}
		seen++
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, f.rel, src, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range fileAPI(fset, file.Name.Name, file) {
			t.Errorf("%s builds only on some platforms but exports %s", f.rel, line)
		}
	}
	if seen == 0 {
		t.Fatal("found no platform file: the kit has appid_linux.go and appid_other.go")
	}
}

func TestPlatformFileFixtures(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      bool
	}{
		{"appid_linux.go", "package frostui\n", true},
		{"icons_windows_arm64.go", "package frostui\n", true},
		{"icons_amd64.go", "package frostui\n", true},
		{"icons_other.go", "//go:build !linux\n\npackage frostui\n", true},
		{"icons_other.go", "// +build !linux\n\npackage frostui\n", true},
		{"icons.go", "package frostui\n\n//go:build linux\n", false},
		{"shell.go", "// Package frostui.\npackage frostui\n", false},
		{"linux.go", "package frostui\n", false},
	} {
		if got := platformFile(tc.name, []byte(tc.src)); got != tc.want {
			t.Errorf("platformFile(%s, %q) = %v, want %v", tc.name, tc.src, got, tc.want)
		}
	}
}

func TestExportedAPIFixture(t *testing.T) {
	src := `package p

type Shown struct {
	A, b int
	Embedded
	*hidden
}

type hidden struct{ X int }

type Iface interface {
	M(x, y int) (n int, err error)
	io.Reader
}

type Alias = string

type Gen[T any] []T

const (
	One Kind = iota
	Two
	three
	Four = 4
)

var Names = [2]string{"a", "b"}

var Err error

func F(a, b string, rest ...int) error { return nil }

func (s *Shown) Method(x int) {}

func (h hidden) Visible() {}

func unexported() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	got := fileAPI(fset, "p", file)
	slices.Sort(got)
	want := []string{
		"const p.Four = 4",
		"const p.One Kind",
		"const p.Two Kind",
		"embed p.Shown Embedded",
		"field p.Shown.A int",
		"func p.F(string, string, ...int) error",
		"imethod p.Iface io.Reader",
		"imethod p.Iface.M(int, int) (int, error)",
		"method (*Shown) p.Method(int)",
		"type p.Alias = string",
		"type p.Gen[T any] []T",
		"type p.Iface interface",
		"type p.Shown struct",
		"var p.Err error",
		"var p.Names [2]string",
	}
	if !slices.Equal(got, want) {
		t.Errorf("fileAPI =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
