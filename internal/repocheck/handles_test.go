package repocheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

// myGoUI is MyGo's widget package. Its Handle (and Ref, an alias of it)
// gives a control an identity across builds.
const myGoUI = "github.com/egoist/mygo/ui"

// handleProblems checks one Go file of the kit against the rule that it
// never makes a MyGo ui.Handle where it would be made again and again: in
// a view, or with ui.Local. The first Bind of a Handle registers it with
// the window, and nothing takes it out until the window closes, so a
// Handle made at each build, or each time an element is built again (a
// row scrolling back into view, a section opening), piles up for the life
// of the window, and every build pass walks them all.
//
// A view is a function or function literal that takes the build's
// *ui.Context or a ui.Element, and every function literal inside one: the
// code that runs as the view builds. Making a Handle is a composite literal
// of one (or of a collection of them), new or make of one, a variable of
// one, or a ui.Local that keeps one (by its type argument or its init's
// result); what a ui.Local's init makes counts as made in a view, since the
// element keeps it. It returns the number of views it saw and one line per
// Handle made.
//
// It resolves the named types the package declares: a struct with a field
// holding a Handle, an array, slice or map of them, a named type of one, or
// an alias of one. It reads the types of the file it is given; for the
// package's other files, use handleProblemsWith (and packageTypes). An
// instance of a generic type resolves to the generic type. It does not
// resolve a type from another package, a type declared inside a function,
// or a type parameter, so a Handle in an imported struct, in a function's
// own type, or in a field whose type is a type parameter, is not seen.
func handleProblems(rel string, src []byte) (views int, problems []string, err error) {
	return handleProblemsWith(rel, src, nil)
}

// handleProblemsWith is handleProblems with the type declarations of the
// file's package, by name: a named type declared in another file of the
// package then resolves too.
func handleProblemsWith(rel string, src []byte, types map[string]ast.Expr) (views int, problems []string, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, rel, src, parser.SkipObjectResolution)
	if err != nil {
		return 0, nil, err
	}
	// The named types of the package: those given, and the file's own.
	named := make(map[string]ast.Expr, len(types))
	for k, v := range types {
		named[k] = v
	}
	addTypeSpecs(named, file)
	name := "" // what the file calls MyGo's ui package
	for _, imp := range file.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return 0, nil, err
		}
		if p == myGoUI {
			name = "ui"
			if imp.Name != nil {
				name = imp.Name.Name
			}
		}
	}
	if name == "" || name == "_" {
		return 0, nil, nil
	}
	// uiNamed reports whether e names one of the ui package's idents.
	uiNamed := func(e ast.Expr, idents ...string) bool {
		switch x := e.(type) {
		case *ast.SelectorExpr:
			pkg, ok := x.X.(*ast.Ident)
			return ok && pkg.Name == name && slices.Contains(idents, x.Sel.Name)
		case *ast.Ident:
			return name == "." && slices.Contains(idents, x.Name)
		}
		return false
	}
	// holdsHandles reports whether a value of type e holds Handles of its
	// own: a Handle, an array, slice or map of them (not of pointers), an
	// inline struct with a field of one, or a named type of the package that
	// is one. seen holds the named types resolved so far in one query, so a
	// recursive type (a node that holds more nodes) stops.
	var holdsHandles func(e ast.Expr, seen map[string]bool) bool
	holdsHandles = func(e ast.Expr, seen map[string]bool) bool {
		switch x := e.(type) {
		case *ast.ParenExpr:
			return holdsHandles(x.X, seen)
		case *ast.ArrayType:
			return holdsHandles(x.Elt, seen)
		case *ast.MapType:
			return holdsHandles(x.Value, seen)
		case *ast.IndexExpr: // an instance of a generic type: pair[int]
			return holdsHandles(x.X, seen)
		case *ast.IndexListExpr:
			return holdsHandles(x.X, seen)
		case *ast.StructType:
			for _, f := range x.Fields.List {
				if holdsHandles(f.Type, seen) {
					return true
				}
			}
			return false
		case *ast.Ident:
			if decl, ok := named[x.Name]; ok && !seen[x.Name] {
				seen[x.Name] = true
				return holdsHandles(decl, seen)
			}
		}
		return uiNamed(e, "Handle", "Ref")
	}
	holds := func(e ast.Expr) bool { return holdsHandles(e, map[string]bool{}) }
	// keepsHandle reports whether ui.Local keeping a value of type e keeps
	// a Handle: the value, or what it points to, holds one.
	keepsHandle := func(e ast.Expr) bool {
		if star, ok := e.(*ast.StarExpr); ok {
			e = star.X
		}
		return holds(e)
	}
	returnsHandle := func(ft *ast.FuncType) bool {
		return ft != nil && ft.Results != nil && len(ft.Results.List) == 1 && keepsHandle(ft.Results.List[0].Type)
	}
	takesView := func(ft *ast.FuncType) bool {
		for _, p := range ft.Params.List {
			if star, ok := p.Type.(*ast.StarExpr); ok && uiNamed(star.X, "Context") || uiNamed(p.Type, "Element") {
				return true
			}
		}
		return false
	}
	// The file's functions, by name, for an init passed to ui.Local by name.
	funcs := map[string]*ast.FuncType{}
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && fd.Recv == nil {
			funcs[fd.Name.Name] = fd.Type
		}
	}
	report := func(n ast.Node, what string) {
		problems = append(problems, fmt.Sprintf("%s: %s; a ui.Handle stays registered with the window until it closes, so keep it in a struct", fset.Position(n.Pos()), what))
	}
	var visit func(n ast.Node, view bool)
	visit = func(n ast.Node, view bool) {
		ast.Inspect(n, func(m ast.Node) bool {
			switch x := m.(type) {
			case *ast.FuncDecl:
				if x.Body != nil {
					v := takesView(x.Type)
					if v {
						views++
					}
					visit(x.Body, v)
				}
				return false
			case *ast.FuncLit:
				v := takesView(x.Type)
				if v {
					views++
				}
				visit(x.Body, view || v)
				return false
			case *ast.CallExpr:
				fun, typeArgs := x.Fun, []ast.Expr(nil)
				switch f := fun.(type) {
				case *ast.IndexExpr:
					fun, typeArgs = f.X, []ast.Expr{f.Index}
				case *ast.IndexListExpr:
					fun, typeArgs = f.X, f.Indices
				}
				if uiNamed(fun, "Local") {
					keeps := len(typeArgs) == 1 && keepsHandle(typeArgs[0])
					if len(x.Args) == 3 {
						switch init := x.Args[2].(type) {
						case *ast.FuncLit:
							keeps = keeps || returnsHandle(init.Type)
						case *ast.Ident:
							keeps = keeps || returnsHandle(funcs[init.Name])
						}
					}
					if keeps {
						report(x, "ui.Local keeps a ui.Handle, made again each time its element is built anew")
					}
					for _, arg := range x.Args {
						visit(arg, true)
					}
					return false
				}
				if id, ok := x.Fun.(*ast.Ident); ok && view && (id.Name == "new" || id.Name == "make") && len(x.Args) > 0 && holds(x.Args[0]) {
					report(x, id.Name+" makes a ui.Handle in a view")
				}
			case *ast.CompositeLit:
				if view && x.Type != nil && holds(x.Type) {
					report(x, "a ui.Handle literal in a view")
				}
			case *ast.ValueSpec:
				if view && x.Type != nil && holds(x.Type) {
					report(x, "a ui.Handle variable in a view")
				}
			}
			return true
		})
	}
	visit(file, false)
	return views, problems, nil
}

// packageName returns the package clause of one Go file.
func packageName(rel string, src []byte) (string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), rel, src, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}
	return file.Name.Name, nil
}

// packageTypes returns the type declarations of one Go package, by name,
// from the sources of its files (rel → source): the type specs of every file
// whose package clause is pkg. handleProblemsWith resolves a named type
// against them, so a type declared in one file resolves in another file of
// the same package.
func packageTypes(pkg string, files map[string][]byte) (map[string]ast.Expr, error) {
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	slices.Sort(rels)
	types := map[string]ast.Expr{}
	for _, rel := range rels {
		file, err := parser.ParseFile(token.NewFileSet(), rel, files[rel], parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		if file.Name.Name != pkg {
			continue
		}
		addTypeSpecs(types, file)
	}
	return types, nil
}

// addTypeSpecs adds a file's type declarations to types, by name.
func addTypeSpecs(types map[string]ast.Expr, file *ast.File) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				types[ts.Name.Name] = ts.Type
			}
		}
	}
}

// TestKitMakesNoHandleInAView keeps the kit's ui.Handles in structs: no
// file of the kit, tests included, makes one in a view or with ui.Local
// (handleProblems), where each build or each element built anew would
// register another with the window for good.
func TestKitMakesNoHandleInAView(t *testing.T) {
	root := repoRoot(t)
	// The kit's files, grouped by package: a directory and its package
	// clause (a frostui_test file is in another package than the frostui
	// files beside it). The guard resolves a named type declared in any file
	// of the package, so each group is checked with its own types.
	type group struct {
		pkg   string
		files map[string][]byte
	}
	groups := map[string]*group{}
	for _, f := range kitFiles(t, root) {
		src, err := os.ReadFile(f.path)
		if err != nil {
			t.Fatal(err)
		}
		pkg, err := packageName(f.rel, src)
		if err != nil {
			t.Fatalf("parse %s: %v", f.rel, err)
		}
		key := filepath.Dir(f.rel) + " " + pkg
		if groups[key] == nil {
			groups[key] = &group{pkg: pkg, files: map[string][]byte{}}
		}
		groups[key].files[f.rel] = src
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	files, views := 0, 0
	for _, key := range keys {
		g := groups[key]
		types, err := packageTypes(g.pkg, g.files)
		if err != nil {
			t.Fatalf("parse package %s: %v", key, err)
		}
		rels := make([]string, 0, len(g.files))
		for rel := range g.files {
			rels = append(rels, rel)
		}
		slices.Sort(rels)
		for _, rel := range rels {
			n, problems, err := handleProblemsWith(rel, g.files[rel], types)
			if err != nil {
				t.Fatalf("parse %s: %v", rel, err)
			}
			files++
			views += n
			for _, problem := range problems {
				t.Error(problem)
			}
		}
	}
	t.Logf("%d files of the kit hold %d views", files, views)
	if files < 20 || views < 100 {
		t.Fatalf("saw %d views in %d files of the kit, too few: the check no longer finds the views", views, files)
	}
}

// TestHandleFixtures proves handleProblems can fail: each fixture that
// makes a ui.Handle in a view or with ui.Local must be reported, and each
// that keeps its Handles in structs, or only passes them around, must
// pass.
func TestHandleFixtures(t *testing.T) {
	fixtures := []struct {
		name, src string
		fail      bool
	}{
		{"ui.Local keeps a Handle per mount (finding B2)", `package frostui
import "github.com/egoist/mygo/ui"
type list struct{ on func() }
func (s *list) item(c *ui.Context) {
	content := ui.Row(c).Height(28)
	handle := ui.Local(content, "middle-click", func() ui.Handle { return ui.Handle{} })
	content.Bind(handle)
}
`, true},
		{"ui.Local with a Handle type argument, outside a view", `package frostui
import "github.com/egoist/mygo/ui"
type list struct{ row ui.Element }
func (s *list) handle() *ui.Handle { return ui.Local[ui.Handle](s.row, "h", nil) }
`, true},
		{"ui.Local of a *Ref made by a named function", `package frostui
import "github.com/egoist/mygo/ui"
func newRef() *ui.Ref { return new(ui.Ref) }
type list struct{ row ui.Element }
func (s *list) ref() **ui.Ref { return ui.Local(s.row, "r", newRef) }
`, true},
		{"ui.Local whose init makes a struct holding a Handle", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
type list struct{ row ui.Element }
func (s *list) state() *rowState {
	return *ui.Local(s.row, "s", func() *rowState { return &rowState{h: ui.Handle{}} })
}
`, true},
		{"a Handle variable in a view", `package frostui
import "github.com/egoist/mygo/ui"
func Field(c *ui.Context, text *string) ui.Element {
	var h ui.Handle
	return ui.TextInput(c, text).Bind(&h)
}
`, true},
		{"a Handle literal in a test's view", `package frostui
import (
	"testing"
	"github.com/egoist/mygo/ui"
)
func TestField(t *testing.T) {
	text := ""
	ui.NewTester(func(c *ui.Context) {
		h := ui.Handle{}
		ui.TextInput(c, &text).Bind(&h)
	}, 100, 100)
}
`, true},
		{"new(ui.Ref) in a closure inside a view", `package frostui
import "github.com/egoist/mygo/ui"
type pane struct{}
func (p *pane) View(c *ui.Context) ui.Element {
	return ui.Column(c).Children(func() { ui.Text(c, "a").Bind(new(ui.Ref)) })
}
`, true},
		{"an array of Handles in a function taking an Element", `package frostui
import "github.com/egoist/mygo/ui"
func decorate(e ui.Element) {
	var hs [2]ui.Handle
	e.Bind(&hs[0])
}
`, true},
		{"a map of Handles made in a view", `package frostui
import "github.com/egoist/mygo/ui"
func View(c *ui.Context) {
	hs := make(map[string]ui.Handle)
	_ = hs
}
`, true},
		{"a renamed import", `package frostui
import mui "github.com/egoist/mygo/ui"
func View(c *mui.Context) {
	h := mui.Handle{}
	mui.Row(c).Bind(&h)
}
`, true},
		{"Handles kept in the struct", `package frostui
import "github.com/egoist/mygo/ui"
type pane struct {
	editor ui.Handle
	rows   map[int]*ui.Handle
}
func (p *pane) row(i int) *ui.Handle {
	h := p.rows[i]
	if h == nil {
		h = new(ui.Handle)
		p.rows[i] = h
	}
	return h
}
func (p *pane) View(c *ui.Context, text *string) {
	ui.TextArea(c, text).Bind(&p.editor)
	ui.Row(c).Bind(p.row(1))
	if c.Focused(p.editor) {
		ui.Text(c, "editing")
	}
}
`, false},
		{"a test's Handle made outside the Tester's view", `package frostui
import (
	"testing"
	"github.com/egoist/mygo/ui"
)
func TestEditor(t *testing.T) {
	var editor ui.Handle
	text := ""
	ui.NewTester(func(c *ui.Context) { ui.TextArea(c, &text).Bind(&editor) }, 100, 100)
	editor.Focus()
}
`, false},
		{"Handles passed in, and a ui.Local of another type", `package frostui
import "github.com/egoist/mygo/ui"
func bind(c *ui.Context, h *ui.Handle, hs []*ui.Handle) *ui.Handle {
	n := ui.Local(ui.Row(c).Bind(h), "n", func() int { return 0 })
	_ = n
	_ = hs
	return h
}
`, false},
		{"the kit's own type named Handle", `package frostui
import "github.com/egoist/mygo/ui"
type Handle struct{}
func View(c *ui.Context) {
	h := Handle{}
	_ = h
	_ = new(Handle)
}
`, false},
		{"a file without MyGo's ui", `package frostui
type ui struct{ Handle int }
func View() { _ = ui{Handle: 1} }
`, false},
		{"ui.Local whose init makes a named struct holding a Handle (finding: a named type of the package)", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
type list struct{}
func (s *list) item(c *ui.Context) {
	row := ui.Row(c)
	st := ui.Local(row, "s", func() rowState { return rowState{} })
	row.Bind(&st.h)
}
`, true},
		{"a variable of a named struct holding a Handle, in a view", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
func View(c *ui.Context) {
	var st rowState
	ui.Row(c).Bind(&st.h)
}
`, true},
		{"new of a named struct holding a Handle, in a view", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
func View(c *ui.Context) {
	st := new(rowState)
	ui.Row(c).Bind(&st.h)
}
`, true},
		{"make of a named slice of Handles, in a view", `package frostui
import "github.com/egoist/mygo/ui"
type hs []ui.Handle
func View(c *ui.Context) {
	handles := make(hs, 2)
	ui.Row(c).Bind(&handles[0])
}
`, true},
		{"a struct embedding a struct that holds a Handle, made in a view", `package frostui
import "github.com/egoist/mygo/ui"
type inner struct{ h ui.Handle }
type outer struct{ inner }
func View(c *ui.Context) {
	o := outer{}
	ui.Row(c).Bind(&o.h)
}
`, true},
		{"an alias of a Handle, made in a view", `package frostui
import "github.com/egoist/mygo/ui"
type H = ui.Handle
func View(c *ui.Context) {
	h := H{}
	ui.Row(c).Bind(&h)
}
`, true},
		{"an instance of a generic struct holding a Handle, made in a view", `package frostui
import "github.com/egoist/mygo/ui"
type pair[K comparable, V any] struct {
	h ui.Handle
	v V
}
type one[V any] struct{ p pair[int, V] }
func View(c *ui.Context) {
	p := one[string]{}
	ui.Row(c).Bind(&p.p.h)
}
`, true},
		{"a named struct whose only Handle field is a pointer, made in a view", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h *ui.Handle }
func View(c *ui.Context) {
	ui.Row(c).Bind(new(rowState).h)
}
`, false},
		{"a recursive named type, made in a view: the cycle guard stops", `package frostui
import "github.com/egoist/mygo/ui"
type node struct {
	next *node
	kids []node
}
func View(c *ui.Context) {
	n := node{}
	ui.Row(c).Bind(&n.kids)
}
`, false},
		{"a named struct holding a Handle, made in a constructor outside a view", `package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
func newRowState() *rowState { return &rowState{} }
func View(c *ui.Context) {
	ui.Row(c).Bind(&newRowState().h)
}
`, false},
	}
	for _, fx := range fixtures {
		t.Run(fx.name, func(t *testing.T) {
			_, problems, err := handleProblems("fixture.go", []byte(fx.src))
			if err != nil {
				t.Fatalf("parse: %v\n%s", err, fx.src)
			}
			switch {
			case fx.fail && len(problems) == 0:
				t.Errorf("not reported:\n%s", fx.src)
			case !fx.fail && len(problems) > 0:
				t.Errorf("reported %q", problems)
			}
		})
	}
}

// TestHandlePackageWide proves the guard resolves a named type declared in
// another file of the same package: made in a view of file B, the struct of
// file A is reported, and file B alone does not see it.
func TestHandlePackageWide(t *testing.T) {
	files := map[string][]byte{
		"paneshell.go": []byte(`package frostui
import "github.com/egoist/mygo/ui"
type rowState struct{ h ui.Handle }
`),
		"sourcelist.go": []byte(`package frostui
import "github.com/egoist/mygo/ui"
func View(c *ui.Context) {
	row := ui.Row(c)
	st := ui.Local(row, "s", func() rowState { return rowState{} })
	row.Bind(&st.h)
}
`),
	}
	types, err := packageTypes("frostui", files)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, problems, err := handleProblemsWith("sourcelist.go", files["sourcelist.go"], types)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(problems) == 0 {
		t.Error("the type declared in another file of the package was not resolved")
	}
	// The same file without the package's types: nothing to resolve.
	if _, alone, err := handleProblems("sourcelist.go", files["sourcelist.go"]); err != nil {
		t.Fatalf("parse: %v", err)
	} else if len(alone) > 0 {
		t.Errorf("a named type of another file resolved without the package's types: %q", alone)
	}
}
