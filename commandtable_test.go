package frostui

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/egoist/mygo/ui"
)

// keyTable returns a table whose commands count their runs in ran.
func keyTable(ran map[string]int, more ...Command) *CommandTable {
	run := func(id string) func() { return func() { ran[id]++ } }
	t := NewCommandTable(
		Command{ID: "new", Title: "New Note", Group: "File", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyN, Scope: Everywhere}}, Run: run("new")},
		Command{ID: "bold", Title: "Toggle Bold", Group: "Edit", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyB}}, Run: run("bold")},
		Command{ID: "next", Title: "Next", Group: "Go", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyJ, Scope: OutsideText}}, Run: run("next")},
	)
	t.Add(more...)
	return t
}

func TestCommandTableKeysRunCommands(t *testing.T) {
	ran := map[string]int{}
	enabled := false
	table := keyTable(ran, Command{ID: "reply", Title: "Reply", Group: "Mail",
		Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyR}, {Mods: ui.Alt, Key: ui.KeyR}},
		Enabled:  func() (bool, string) { return enabled, "No message is shown" },
		Run:      func() { ran["reply"]++ }})
	tt := ui.NewTester(func(c *ui.Context) {
		table.Bind(c)
		Title(c, "Page")
	}, 400, 300)
	tt.Key(ui.Cmd, ui.KeyN)
	if ran["new"] != 1 {
		t.Errorf("Ctrl+N ran New %d times", ran["new"])
	}
	tt.Key(ui.Cmd, ui.KeyR)
	if ran["reply"] != 0 {
		t.Error("the key of a disabled command ran it")
	}
	enabled = true
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyR)
	tt.Key(ui.Alt, ui.KeyR)
	if ran["reply"] != 2 {
		t.Errorf("both keys of Reply ran it %d times, want 2", ran["reply"])
	}
}

func TestCommandTableFirstCommandTakesASharedKey(t *testing.T) {
	first, second := 0, 0
	key := []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyG}}
	table := NewCommandTable(
		Command{ID: "a", Title: "A", Bindings: key, Run: func() { first++ }},
		Command{ID: "b", Title: "B", Bindings: key, Run: func() { second++ }},
	)
	tt := ui.NewTester(func(c *ui.Context) { table.Bind(c); Title(c, "Page") }, 400, 300)
	tt.Key(ui.Cmd, ui.KeyG)
	if first != 1 || second != 0 {
		t.Errorf("runs = %d, %d, want the first command only", first, second)
	}
}

func TestCommandTableAddReplacesAndRemoveForgets(t *testing.T) {
	table := NewCommandTable(Command{ID: "a", Title: "A"}, Command{ID: "b", Title: "B"}, Command{ID: "c", Title: "C"})
	table.Add(Command{ID: "b", Title: "Bee"})
	table.Remove("a")
	var titles []string
	for _, cmd := range table.Commands() {
		titles = append(titles, cmd.Title)
	}
	if !reflect.DeepEqual(titles, []string{"Bee", "C"}) {
		t.Errorf("titles = %v", titles)
	}
	if _, ok := table.Command("a"); ok {
		t.Error("a removed command is still found")
	}
	if cmd, ok := table.Command("c"); !ok || cmd.Title != "C" {
		t.Errorf("Command(c) = %+v, %v after a removal moved it", cmd, ok)
	}
}

func TestCommandResolveReadsTheEnabledRule(t *testing.T) {
	cmd := Command{ID: "a", Title: "A", Disabled: false, Reason: "stale",
		Enabled:  func() (bool, string) { return false, "Nothing selected" },
		Bindings: []KeyBinding{{Mods: ui.Ctrl | ui.Shift, Key: ui.KeyP}},
	}.Resolve()
	if !cmd.Disabled || cmd.Reason != "Nothing selected" {
		t.Errorf("Resolve = disabled %v, reason %q", cmd.Disabled, cmd.Reason)
	}
	if runtime.GOOS != "darwin" && !reflect.DeepEqual(cmd.Keys, []string{"Ctrl", "Shift", "P"}) {
		t.Errorf("Keys = %v", cmd.Keys)
	}
	on := Command{Enabled: func() (bool, string) { return true, "ignored" }}.Resolve()
	if on.Disabled || on.Reason != "" {
		t.Errorf("an enabled command = disabled %v, reason %q", on.Disabled, on.Reason)
	}
	kept := Command{Keys: []string{"Ctrl", "U"}, Bindings: []KeyBinding{{Mods: ui.Alt, Key: ui.KeyU}}}.Resolve()
	if !reflect.DeepEqual(kept.Keys, []string{"Ctrl", "U"}) {
		t.Errorf("Resolve replaced the given Keys: %v", kept.Keys)
	}
}

func TestKeyBindingNamesItsKeys(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("macOS writes symbols")
	}
	for _, tc := range []struct {
		b    KeyBinding
		want string
	}{
		{KeyBinding{Mods: ui.Ctrl, Key: ui.KeyK}, "Ctrl+K"},
		{KeyBinding{Mods: ui.Ctrl | ui.Shift, Key: ui.KeyP}, "Ctrl+Shift+P"},
		{KeyBinding{Mods: ui.Ctrl | ui.Alt, Key: ui.KeyEnter}, "Ctrl+Alt+Enter"},
		{KeyBinding{Key: ui.KeyF2}, "F2"},
		{KeyBinding{Key: ui.KeyF12}, "F12"},
		{KeyBinding{Mods: ui.Alt, Key: ui.Key3}, "Alt+3"},
		{KeyBinding{Mods: ui.Ctrl, Key: ui.KeyComma}, "Ctrl+,"},
		{KeyBinding{Key: ui.KeyLeft}, "←"},
	} {
		if got := tc.b.String(); got != tc.want {
			t.Errorf("String = %q, want %q", got, tc.want)
		}
	}
}

// editorWindow is a window with a text editor, a one-line field and the
// shell's frame with a command table.
type editorWindow struct {
	shell   Shell
	table   *CommandTable
	editor  ui.Handle
	text    string
	field   string
	claimed []ui.Key
	elemKey bool
	linked  int
}

func (w *editorWindow) view(c *ui.Context) {
	w.shell.View(c, Frame{Table: w.table, Content: func() {
		ed := ui.TextArea(c, &w.text).Label("Editor").Height(120).Bind(&w.editor).OnTextKey(func(k ui.TextKey) bool {
			if k.Mods == ui.Cmd && (k.Key == ui.KeyK || k.Key == ui.KeyB) {
				w.claimed = append(w.claimed, k.Key)
				return true
			}
			return false
		})
		if w.elemKey {
			ed.OnShortcut(ui.Cmd, ui.KeyL, func() { w.linked++ })
		}
		ui.TextInput(c, &w.field).Label("Field")
		ui.Button(c, "Button")
	}})
}

func newEditorWindow(ran map[string]int) (*editorWindow, *ui.Tester) {
	w := &editorWindow{table: keyTable(ran)}
	w.table.Editing = func(c *ui.Context) bool { return c.FocusWithin(w.editor) }
	return w, ui.NewTester(w.view, 1000, 700)
}

func TestCtrlKYieldsToAFocusedEditor(t *testing.T) {
	ran := map[string]int{}
	w, tt := newEditorWindow(ran)
	if err := tt.Click("Editor"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Editor") {
		t.Fatal("the editor did not take the focus")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if w.shell.Palette.Open {
		t.Fatal("Ctrl K opened the palette in the editor")
	}
	if len(w.claimed) != 1 || w.claimed[0] != ui.KeyK {
		t.Fatalf("the editor's own key handler saw %v, want Ctrl K", w.claimed)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !w.shell.Palette.Open {
		t.Fatal("Ctrl+Shift+P did not open the palette in the editor")
	}
}

func TestCtrlKOpensThePaletteOutsideEditors(t *testing.T) {
	ran := map[string]int{}
	w, tt := newEditorWindow(ran)
	if err := tt.Click("Field"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Field") {
		t.Fatal("the field did not take the focus")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !w.shell.Palette.Open {
		t.Fatal("Ctrl K did not open the palette from a one-line field")
	}
	tt.Key(0, ui.KeyEscape)
	if err := tt.Click("Button"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !w.shell.Palette.Open {
		t.Fatal("Ctrl K did not open the palette from a button")
	}
}

func TestKeyScopesFollowTheFocus(t *testing.T) {
	ran := map[string]int{}
	w, tt := newEditorWindow(ran)
	press := func(place string) map[string]int {
		t.Helper()
		if err := tt.Click(place); err != nil {
			t.Fatal(err)
		}
		clear(ran)
		w.claimed = nil
		tt.Key(ui.Cmd, ui.KeyN) // Everywhere
		tt.Key(ui.Cmd, ui.KeyB) // OutsideEditors
		tt.Key(ui.Cmd, ui.KeyJ) // OutsideText
		return map[string]int{"new": ran["new"], "bold": ran["bold"], "next": ran["next"]}
	}
	if got, want := press("Editor"), map[string]int{"new": 1, "bold": 0, "next": 0}; !reflect.DeepEqual(got, want) {
		t.Errorf("in the editor: ran %v, want %v", got, want)
	}
	if len(w.claimed) != 1 || w.claimed[0] != ui.KeyB {
		t.Errorf("the editor saw %v, want Ctrl+B, which the table left to it", w.claimed)
	}
	if got, want := press("Field"), map[string]int{"new": 1, "bold": 1, "next": 0}; !reflect.DeepEqual(got, want) {
		t.Errorf("in the field: ran %v, want %v", got, want)
	}
	if got, want := press("Button"), map[string]int{"new": 1, "bold": 1, "next": 1}; !reflect.DeepEqual(got, want) {
		t.Errorf("on the button: ran %v, want %v", got, want)
	}
}

func TestAnEditorsOwnShortcutComesBeforeTheWindows(t *testing.T) {
	ran := map[string]int{}
	w, tt := newEditorWindow(ran)
	w.elemKey = true
	w.table.Add(Command{ID: "lines", Title: "Lines", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyL, Scope: Everywhere}}, Run: func() { ran["lines"]++ }})
	if err := tt.Click("Editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyL)
	if w.linked != 1 || ran["lines"] != 0 {
		t.Errorf("element's shortcut ran %d times, the window's %d: the element comes first", w.linked, ran["lines"])
	}
}

func TestWithoutEditingAnyFieldCountsAsAnEditor(t *testing.T) {
	ran := map[string]int{}
	w, tt := newEditorWindow(ran)
	w.table.Editing = nil
	if err := tt.Click("Field"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if w.shell.Palette.Open {
		t.Error("with no Editing rule, Ctrl K went past a text field")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !w.shell.Palette.Open {
		t.Error("Ctrl+Shift+P did not open the palette")
	}
}

func TestAShellWithoutATableKeepsItsCtrlK(t *testing.T) {
	var s Shell
	text := ""
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{Content: func() { ui.TextArea(c, &text).Label("Editor").Height(120) }})
	}, 1000, 700)
	if err := tt.Click("Editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("the shell without a table no longer opens the palette with Ctrl K")
	}
}

func TestPaletteListsTheTableAndSkipsUnlisted(t *testing.T) {
	var s Shell
	ran := 0
	table := NewCommandTable(
		Command{ID: "a", Title: "Archive", Group: "Mail", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyE}}, Run: func() { ran++ }},
		Command{ID: "z", Title: "Reply", Group: "Mail", Enabled: func() (bool, string) { return false, "No message" }, Run: func() { ran += 10 }},
		Command{ID: "h", Title: "Hidden thing", Group: "Mail", Unlisted: true, Run: func() {}},
	)
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{Table: table, Commands: []Command{{ID: "o", Title: "Old command", Group: "Old"}}, Content: func() { Title(c, "Page") }})
	}, 1280, 800)
	tt.Key(ui.Cmd, ui.KeyK)
	for _, want := range []string{"Old command", "Archive", "Reply", "No message"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("Hidden thing") {
		t.Error("an unlisted command shows in the palette")
	}
	if err := tt.Click("Reply"); err != nil {
		t.Fatal(err)
	}
	if ran != 0 || !s.Palette.Open {
		t.Fatalf("a command disabled by its rule ran (%d) or closed the palette", ran)
	}
	if err := tt.Click("Archive"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ran != 1 {
		t.Errorf("Archive ran %d times", ran)
	}
}

func TestCommandTableBuildsMenus(t *testing.T) {
	ran := ""
	checked := true
	table := NewCommandTable(
		Command{ID: "reply", Title: "Reply", Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyR}}, Run: func() { ran = "reply" }},
		Command{ID: "flag", Title: "Flagged", Checked: func() bool { return checked }, Run: func() { ran = "flag" }},
		Command{ID: "gone", Title: "Delete", Enabled: func() (bool, string) { return false, "Read only" }, Run: func() { ran = "gone" }},
	)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Gap(SpaceM).Children(func() {
			ui.Text(c, "Row").ContextMenu(table.Menu("reply", "", "flag", "gone", "missing"))
			ui.MenuButton(c, "More", func(m *ui.Menu) { m.Submenu("Mark", table.Menu("flag")) })
		})
	}, 400, 300)
	if err := tt.RightClick("Row"); err != nil {
		t.Fatal(err)
	}
	got := tt.Menu()
	if len(got) != 4 {
		t.Fatalf("menu = %q, want Reply, a separator, Flagged and Delete", got)
	}
	if err := tt.ChooseMenuItem("Delete"); err == nil && ran != "" {
		t.Errorf("choosing a disabled item ran %q", ran)
	}
	tt.CloseMenu()
	if err := tt.RightClick("Row"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Reply"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ran != "reply" {
		t.Errorf("ran = %q, want reply", ran)
	}
	if err := tt.Click("More"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Mark", "Flagged"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ran != "flag" {
		t.Errorf("ran = %q, want flag from the submenu", ran)
	}
}

func TestCommandTableFeedsTheShortcutsDialog(t *testing.T) {
	table := NewCommandTable(
		Command{ID: "a", Title: "Archive", Group: "Mail", Bindings: []KeyBinding{{Mods: ui.Ctrl, Key: ui.KeyE}}},
		Command{ID: "n", Title: "Next pane", Group: "View", Bindings: []KeyBinding{{Key: ui.KeyF6}}},
		Command{ID: "x", Title: "No keys", Group: "Mail"},
		Command{ID: "r", Title: "Reply", Group: "Mail", Bindings: []KeyBinding{{Mods: ui.Ctrl, Key: ui.KeyR}}},
	)
	var titles []string
	for _, e := range table.Shortcuts() {
		titles = append(titles, e.Group+"/"+e.Title)
	}
	if want := []string{"Mail/Archive", "Mail/Reply", "View/Next pane"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("Shortcuts = %v, want %v", titles, want)
	}
	if runtime.GOOS != "darwin" {
		if got := table.Tooltip("r"); got != "Reply (Ctrl+R)" {
			t.Errorf("Tooltip = %q", got)
		}
	}
	if got := table.Tooltip("x"); got != "No keys" {
		t.Errorf("Tooltip without a key = %q", got)
	}
	tt := ui.NewTester(func(c *ui.Context) { ShortcutsBody(table.Shortcuts())(c) }, 500, 400)
	for _, want := range []string{"Mail", "View", "Archive", "Reply", "Next pane", "F6"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}
