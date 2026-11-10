package frostui

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

// noKeyBoundTwice fails t when two commands of the table bind one key.
func noKeyBoundTwice(t *testing.T, table *CommandTable) {
	t.Helper()
	seen := map[string]string{}
	for _, cmd := range table.Commands() {
		for _, b := range cmd.Bindings {
			k := fmt.Sprint(b.Mods, "+", b.Key)
			if other, ok := seen[k]; ok {
				t.Errorf("%s is bound by %s and %s", b, other, cmd.ID)
			}
			seen[k] = cmd.ID
		}
	}
}

func TestPaneShellTableOwnsTheKeys(t *testing.T) {
	var s PaneShell
	s.Console.Title = "Activity"
	ran := 0
	table := NewCommandTable(Command{ID: "mine", Title: "My command", Group: "App",
		Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyU}}, Run: func() { ran++ }})
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{Table: table, ConsoleData: con, Content: func() { Title(c, "Page") }})
	}, 1000, 600)
	tt.Frame()
	noKeyBoundTwice(t, table)
	tt.Key(ui.Cmd, ui.KeyU)
	if ran != 1 {
		t.Errorf("the table's key ran its command %d times", ran)
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Fatal("one Ctrl J did not open the console: it fired twice, or not at all")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Fatal("Ctrl K did not open the palette")
	}
	for _, want := range []string{"My command", "Hide Activity"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("Command Palette") {
		t.Error("the palette lists the command that opens it")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !s.Palette.Open {
		t.Error("Ctrl Shift P did not open the palette")
	}
	noKeyBoundTwice(t, table)
}

func TestPaneShellTableCommandTakesTheKeyFromTheShell(t *testing.T) {
	var s PaneShell
	ran := 0
	table := NewCommandTable(Command{ID: "mine", Title: "Mine",
		Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyK}}, Run: func() { ran++ }})
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{Table: table}) }, 800, 500)
	tt.Frame()
	noKeyBoundTwice(t, table)
	tt.Key(ui.Cmd, ui.KeyK)
	if ran != 1 || s.Palette.Open {
		t.Errorf("Ctrl K: the app's command ran %d times, palette %v; the shell must not bind it too", ran, s.Palette.Open)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !s.Palette.Open {
		t.Error("the palette kept no key of its own")
	}
}

func TestPaneShellCtrlKYieldsToAnEditorUnderTheTable(t *testing.T) {
	var s PaneShell
	var editor ui.Handle
	var text, field string
	var claimed []ui.Key
	table := NewCommandTable()
	table.Editing = func(c *ui.Context) bool { return c.FocusWithin(editor) }
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{Table: table, Content: func() {
			ui.TextArea(c, &text).Label("Editor").Height(200).Bind(&editor).OnTextKey(func(k ui.TextKey) bool {
				if k.Mods == ui.Cmd && k.Key == ui.KeyK {
					claimed = append(claimed, k.Key)
					return true
				}
				return false
			})
			ui.TextInput(c, &field).Label("Filter")
		}})
	}, 1000, 600)
	tt.Frame()
	if err := tt.Click("Editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if s.Palette.Open || len(claimed) != 1 {
		t.Fatalf("Ctrl K in the editor: palette %v, the editor saw it %d times", s.Palette.Open, len(claimed))
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !s.Palette.Open {
		t.Fatal("Ctrl Shift P in the editor did not open the palette")
	}
	tt.Key(0, ui.KeyEscape)
	if err := tt.Click("Filter"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("Ctrl K in a one-line field did not open the palette")
	}
}

func TestPaneShellTableDoesNotReadInEditor(t *testing.T) {
	var s PaneShell
	table := NewCommandTable()
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{Table: table, InEditor: true}) }, 800, 500)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("with a table, InEditor kept Ctrl K with no editor focused")
	}
}

func TestPaneShellWithoutATableBindsItsOwnKeys(t *testing.T) {
	var s PaneShell
	s.Console.Title = "Activity"
	con := NewConsole(16)
	inEditor := true
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{InEditor: inEditor, ConsoleData: con, Commands: []Command{{ID: "a", Title: "A command", Group: "App"}}})
	}, 1000, 600)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if s.Palette.Open {
		t.Fatal("InEditor did not keep Ctrl K")
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Fatal("one Ctrl J did not open the console")
	}
	inEditor = false
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Fatal("Ctrl K did not open the palette")
	}
	for _, want := range []string{"A command", "Hide Activity"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
}

func TestPaneShellTableTakesAnAppsConsoleCommandOnce(t *testing.T) {
	var s PaneShell
	con := NewConsole(16)
	table := NewCommandTable()
	own := s.ConsoleCommand()
	own.Group, own.Bindings = "App", nil
	f := PaneFrame{Table: table, ConsoleData: con, Commands: []Command{own}}
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, f) }, 800, 500)
	tt.Frame()
	n := 0
	for _, cmd := range s.commands(f) {
		if cmd.ID == ConsoleCommandID {
			n++
			if cmd.Group != "App" {
				t.Errorf("the app's Group was lost: %+v", cmd)
			}
		}
	}
	if n != 1 {
		t.Errorf("the console's command is listed %d times", n)
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Error("Ctrl J does not work with the app's command")
	}
}

func TestPaneShellLeavesAConsoleCommandTheAppAddedToTheTable(t *testing.T) {
	var s PaneShell
	table := NewCommandTable()
	own := s.ConsoleCommand()
	own.Group = "Mine"
	table.Add(own)
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{Table: table, ConsoleData: NewConsole(16)}) }, 800, 500)
	tt.Frame()
	if cmd, _ := table.Command(ConsoleCommandID); cmd.Group != "Mine" {
		t.Errorf("the shell replaced the app's command: %+v", cmd)
	}
}

func TestPaneShellTableWithdrawsTheConsoleWhileConnecting(t *testing.T) {
	var s PaneShell
	table := NewCommandTable()
	con := NewConsole(16)
	conn := Connection{Daemon: "exampled", State: Connecting}
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{Table: table, ConsoleData: con, Connection: &conn}) }, 800, 500)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyJ)
	if s.Console.Open {
		t.Error("Ctrl J toggled a console the connecting screen covers")
	}
	conn.State = Connected
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Error("Ctrl J does not work once connected")
	}
}

func TestShellTableAddsThePaletteCommandOnce(t *testing.T) {
	var s Shell
	ran := 0
	table := NewCommandTable(Command{ID: "mine", Title: "Mine",
		Bindings: []KeyBinding{{Mods: ui.Cmd | ui.Shift, Key: ui.KeyP}}, Run: func() { ran++ }})
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, Frame{Table: table, Content: func() { Title(c, "Page") }}) }, 1000, 600)
	tt.Frame()
	noKeyBoundTwice(t, table)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if ran != 1 || s.Palette.Open {
		t.Errorf("Ctrl Shift P: ran %d, palette %v", ran, s.Palette.Open)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("Ctrl K did not open the palette")
	}
}

func TestBindPassesAKeyToTheNextEnabledCommand(t *testing.T) {
	first, second := 0, 0
	key := []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyG}}
	table := NewCommandTable(
		Command{ID: "a", Title: "A", Bindings: key, Enabled: func() (bool, string) { return false, "" }, Run: func() { first++ }},
		Command{ID: "b", Title: "B", Bindings: key, Run: func() { second++ }},
	)
	tt := ui.NewTester(func(c *ui.Context) { table.Bind(c); Title(c, "Page") }, 400, 300)
	tt.Key(ui.Cmd, ui.KeyG)
	if first != 0 || second != 1 {
		t.Errorf("runs = %d, %d, want the enabled command alone", first, second)
	}
}

func TestShellTableOwnsCtrlJ(t *testing.T) {
	var s Shell
	con := NewConsole(16)
	table := NewCommandTable()
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{Table: table, ConsoleData: con, Content: func() { Title(c, "Page") }})
	}, 1000, 600)
	tt.Frame()
	noKeyBoundTwice(t, table)
	if _, ok := table.Command(ConsoleCommandID); !ok {
		t.Fatal("the console's command did not join the table")
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Fatal("one Ctrl J did not open the console")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !tt.HasText("Hide Console") {
		t.Errorf("the palette does not offer to hide the console: %q", tt.Texts())
	}
}

func TestShellLeavesCtrlJToATableCommand(t *testing.T) {
	var s Shell
	ran, enabled := 0, false
	table := NewCommandTable(Command{ID: "mine", Title: "Mine",
		Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyJ}},
		Enabled:  func() (bool, string) { return enabled, "" }, Run: func() { ran++ }})
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{Table: table, ConsoleData: con, Content: func() { Title(c, "Page") }})
	}, 1000, 600)
	tt.Frame()
	noKeyBoundTwice(t, table)
	// While the command is disabled the table binds nothing, and a binding
	// of the shell's would show here.
	tt.Key(ui.Cmd, ui.KeyJ)
	if ran != 0 || s.Console.Open {
		t.Fatalf("Ctrl J with the command disabled: ran %d, console %v; the shell must not bind it", ran, s.Console.Open)
	}
	enabled = true
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyJ)
	if ran != 1 || s.Console.Open {
		t.Errorf("Ctrl J: the app's command ran %d times, console %v", ran, s.Console.Open)
	}
}

func TestShellWithoutATableKeepsCtrlJ(t *testing.T) {
	var s Shell
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{ConsoleData: con, Content: func() { Title(c, "Page") }})
	}, 1000, 600)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyJ)
	if !s.Console.Open {
		t.Error("Ctrl J did not open the console")
	}
}
