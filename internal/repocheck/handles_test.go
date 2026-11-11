package repocheck

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
func handleProblems(rel string, src []byte) (views int, problems []string, err error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, rel, src, parser.SkipObjectResolution)
	if err != nil {
		return 0, nil, err
	}
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
	// own: a Handle, or an array, slice or map of them (not of pointers).
	var holdsHandles func(e ast.Expr) bool
	holdsHandles = func(e ast.Expr) bool {
		switch x := e.(type) {
		case *ast.ParenExpr:
			return holdsHandles(x.X)
		case *ast.ArrayType:
			return holdsHandles(x.Elt)
		case *ast.MapType:
			return holdsHandles(x.Value)
		}
		return uiNamed(e, "Handle", "Ref")
	}
	// keepsHandle reports whether ui.Local keeping a value of type e keeps
	// a Handle: the value, or what it points to, holds one.
	keepsHandle := func(e ast.Expr) bool {
		if star, ok := e.(*ast.StarExpr); ok {
			e = star.X
		}
		return holdsHandles(e)
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
				if id, ok := x.Fun.(*ast.Ident); ok && view && (id.Name == "new" || id.Name == "make") && len(x.Args) > 0 && holdsHandles(x.Args[0]) {
					report(x, id.Name+" makes a ui.Handle in a view")
				}
			case *ast.CompositeLit:
				if view && x.Type != nil && holdsHandles(x.Type) {
					report(x, "a ui.Handle literal in a view")
				}
			case *ast.ValueSpec:
				if view && x.Type != nil && holdsHandles(x.Type) {
					report(x, "a ui.Handle variable in a view")
				}
			}
			return true
		})
	}
	visit(file, false)
	return views, problems, nil
}

// TestKitMakesNoHandleInAView keeps the kit's ui.Handles in structs: no
// file of the kit, tests included, makes one in a view or with ui.Local
// (handleProblems), where each build or each element built anew would
// register another with the window for good.
func TestKitMakesNoHandleInAView(t *testing.T) {
	root := repoRoot(t)
	files, views := 0, 0
	for _, f := range kitFiles(t, root) {
		src, err := os.ReadFile(f.path)
		if err != nil {
			t.Fatal(err)
		}
		n, problems, err := handleProblems(f.rel, src)
		if err != nil {
			t.Fatalf("parse %s: %v", f.rel, err)
		}
		files++
		views += n
		for _, problem := range problems {
			t.Error(problem)
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
