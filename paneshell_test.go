package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// shellPanes are three panes, each with a segment of the top bar.
func shellPanes(c *ui.Context) []Pane {
	pane := func(id, label string, width, minimum float32) Pane {
		return Pane{ID: id, Label: label, Width: width, Min: minimum, Collapsible: id == "side",
			Bar:     func() { ui.Text(c, id+" bar") },
			Content: func() { ui.Text(c, id+" body") }}
	}
	return []Pane{pane("side", "Sidebar", 200, 160), pane("list", "List", 300, 200), pane("read", "Reader", 0, 300)}
}

func segmentedFrame(c *ui.Context, con *Console) PaneFrame {
	return PaneFrame{
		SidebarPane: "side",
		Panes:       shellPanes(c),
		End:         func() { IconButton(c, "view-refresh", "App item") },
		Menu:        func(m *ui.Menu) { m.Item("About") },
		Connection:  &Connection{Daemon: "exampled"},
		Status:      []StatusItem{{Text: "Synced", Progress: -1}},
		StatusEnd:   []StatusItem{{Text: "12 messages", Progress: -1}},
		ConsoleData: con,
	}
}

func rect(t *testing.T, tt *ui.Tester, s string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(s)
	if !ok {
		t.Fatalf("no %q in %q", s, tt.Texts())
	}
	return r
}

func TestPaneShellSegmentsFollowThePanes(t *testing.T) {
	var s PaneShell
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, con)) }, 1200, 700)
	tt.Frame()
	aligned := func(id string) {
		t.Helper()
		bar, body := rect(t, tt, id+" bar"), rect(t, tt, id+" body")
		if bar.X != body.X+SpaceM {
			t.Errorf("%s's segment starts at %v over a pane at %v", id, bar.X, body.X)
		}
	}
	if bar, body := rect(t, tt, "list bar"), rect(t, tt, "list body"); bar.X != body.X+SpaceM {
		t.Errorf("the list's segment starts at %v over a pane at %v", bar.X, body.X)
	}
	aligned("read")
	drag(tt, 200, 400, 260, 400)
	if got := s.Panes.Layout.Widths["side"]; got != 260 {
		t.Fatalf("side is %v after the drag", got)
	}
	aligned("list")
	aligned("read")
	if body := rect(t, tt, "list body"); body.X != 261 {
		t.Errorf("list body at %v, want 261", body.X)
	}
}

func TestPaneShellBarOrder(t *testing.T) {
	var s PaneShell
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, con)) }, 1200, 700)
	tt.Frame()
	order := []string{"Sidebar", "side bar", "list bar", "read bar", "Search or run a command", "App item", "Activity (Ctrl J)", "Main menu"}
	s.Console.Title = "Activity"
	tt.Frame()
	last := float32(-1)
	for _, name := range order {
		r := rect(t, tt, name)
		if r.X <= last {
			t.Errorf("%q at %v is not after the item before it at %v", name, r.X, last)
		}
		last = r.X
	}
}

func TestPaneShellKeepsClearOfTheWindowControls(t *testing.T) {
	for _, segmented := range []bool{true, false} {
		var s PaneShell
		view := func(c *ui.Context) {
			f := PaneFrame{SidebarPane: "side", Menu: func(m *ui.Menu) { m.Item("About") }}
			if segmented {
				f.Panes = shellPanes(c)
			} else {
				f.Content = func() {}
			}
			s.View(c, f)
		}
		tt := ui.NewTester(view, 1200, 700)
		tt.SetTitleBar(ui.TitleBar{Height: TopBarHeight, Left: 70, Right: 110})
		tt.Frame()
		if r := rect(t, tt, "Sidebar"); r.X < 70 {
			t.Errorf("segmented %v: the toggle at %v is under the left controls", segmented, r.X)
		}
		if r := rect(t, tt, "Main menu"); r.X+r.W > 1200-110 {
			t.Errorf("segmented %v: the menu ends at %v, under the right controls", segmented, r.X+r.W)
		}
	}
}

func TestPaneShellSidebarToggleStaysAtTheFarLeft(t *testing.T) {
	var s PaneShell
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, NewConsole(16))) }, 1200, 700)
	tt.Frame()
	a, ok := tt.AccessByRole(ui.RoleToggleButton, "Sidebar")
	if !ok || !a.Checked {
		t.Fatalf("the toggle: %+v, %v", a, ok)
	}
	if err := tt.Click("Sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.Panes.Shown("side") || tt.HasText("side bar") || tt.HasText("side body") {
		t.Errorf("collapsed: %q", tt.Texts())
	}
	r := rect(t, tt, "Sidebar")
	if r.X > SpaceM+1 {
		t.Errorf("the toggle moved to %v", r.X)
	}
	if a, _ := tt.AccessByRole(ui.RoleToggleButton, "Sidebar"); a.Checked {
		t.Error("the toggle reads pressed with the sidebar hidden")
	}
	if err := tt.Click("Sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("side body") {
		t.Errorf("shown again: %q", tt.Texts())
	}
}

func TestPaneShellSimpleBarCentersThePalette(t *testing.T) {
	var s PaneShell
	var side Panes
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{
			SidebarPane: "left", SidebarPanes: &side,
			Start:   func() { ui.Text(c, "Brand") },
			End:     func() { IconButton(c, "list-add", "New") },
			Menu:    func(m *ui.Menu) { m.Item("About") },
			Content: func() { side.View(c, shellPanesNoBar(c)).Grow(1) },
		})
	}, 1200, 700)
	tt.Frame()
	field := rect(t, tt, "Search or run a command")
	if mid := field.X + field.W/2; mid < 599 || mid > 601 {
		t.Errorf("the field's middle is at %v in a window 1200 wide", mid)
	}
	if rect(t, tt, "Sidebar").X >= rect(t, tt, "Brand").X {
		t.Error("the toggle is not before the bar's items")
	}
	if rect(t, tt, "New").X >= rect(t, tt, "Main menu").X {
		t.Error("the main menu is not last")
	}
	if err := tt.Click("Sidebar"); err != nil {
		t.Fatal(err)
	}
	if side.Shown("left") {
		t.Error("the toggle did not hide the pane of SidebarPanes")
	}
}

func shellPanesNoBar(c *ui.Context) []Pane {
	return []Pane{
		{ID: "left", Width: 200, Collapsible: true, Content: func() { ui.Text(c, "left body") }},
		{ID: "work", Content: func() { ui.Text(c, "work body") }},
	}
}

func TestPaneShellContentIsFullHeightAndUnscrolled(t *testing.T) {
	var s PaneShell
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{
			Content:    func() { ui.Box(c).Grow(1).Label("the content") },
			Connection: &Connection{},
		})
	}, 1000, 600)
	tt.Frame()
	r := rect(t, tt, "the content")
	if want := float32(600 - TopBarHeight - 1 - 1 - 28); r.H != want || r.W != 1000 || r.X != 0 {
		t.Errorf("content %+v, want the whole width and %v high", r, want)
	}
	if _, ok := tt.AccessByRole(ui.RoleScroll, ""); ok {
		t.Error("the content scrolls")
	}
}

func TestPaneShellStatusBarStartsWithTheConnection(t *testing.T) {
	var s PaneShell
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, NewConsole(16))) }, 1200, 700)
	tt.Frame()
	if _, ok := tt.AccessByRole(ui.RoleStatus, ""); !ok {
		t.Error("no status bar")
	}
	if a, b := rect(t, tt, "Connection: Local"), rect(t, tt, "Synced"); a.X >= b.X {
		t.Errorf("the connection at %v is not before the first item at %v", a.X, b.X)
	}
	if r := rect(t, tt, "12 messages"); r.X < 900 {
		t.Errorf("the end item at %v is not at the end", r.X)
	}
}

func TestPaneShellConsoleIsClosedAndTogglesThreeWays(t *testing.T) {
	var s PaneShell
	s.Console.Title = "Activity"
	con := NewConsole(16)
	con.Add(LevelInfo, "Work", "Inbox synced")
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, con)) }, 1200, 700)
	tt.Frame()
	if s.Console.Open || tt.HasText("Inbox synced") {
		t.Fatalf("the console is open at first: %q", tt.Texts())
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	tt.Frame()
	if !s.Console.Open || !tt.HasText("Inbox synced") || !tt.HasText("Activity") {
		t.Fatalf("Ctrl J: %q", tt.Texts())
	}
	if a, ok := tt.AccessByRole(ui.RoleToggleButton, "Activity (Ctrl J)"); !ok || !a.Checked {
		t.Errorf("the button reads %+v", a)
	}
	if err := tt.Click("Activity (Ctrl J)"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.Console.Open {
		t.Fatal("the button did not close the console")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	tt.Frame()
	if err := tt.Click("Show Activity"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !s.Console.Open {
		t.Error("the palette's command did not open the console")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	tt.Frame()
	if !tt.HasText("Hide Activity") {
		t.Errorf("open, the command should hide: %q", tt.Texts())
	}
}

func TestPaneShellConsoleCommandIsAddedOnce(t *testing.T) {
	var s PaneShell
	own := s.ConsoleCommand()
	own.Group = "App"
	f := PaneFrame{ConsoleData: NewConsole(16), Commands: []Command{own}}
	if got := s.commands(f); len(got) != 1 || got[0].Group != "App" {
		t.Errorf("the app's own command was doubled: %+v", got)
	}
	f.Commands = []Command{{ID: "other"}}
	if got := s.commands(f); len(got) != 2 || got[1].ID != ConsoleCommandID {
		t.Errorf("commands = %+v", got)
	}
	f.ConsoleData = nil
	if got := s.commands(f); len(got) != 1 {
		t.Errorf("no console, commands = %+v", got)
	}
}

func TestPaneShellConsoleSplitResizes(t *testing.T) {
	var s PaneShell
	s.Console.Open = true
	s.Console.Title = "Activity"
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, segmentedFrame(c, NewConsole(16))) }, 1200, 700)
	tt.Frame()
	a, ok := tt.AccessByRole(ui.RoleSeparator, "Activity")
	if !ok || a.Now != consoleDefaultHeight || a.Min != consoleMinHeight || !a.Horizontal || !a.Focusable {
		t.Fatalf("the console's separator: %+v, %v", a, ok)
	}
	sep := a.Bounds
	y := sep.Y + sep.H/2
	drag(tt, 600, y, 600, y-60)
	if s.ConsoleHeight != consoleDefaultHeight+60 {
		t.Errorf("dragged 60 up: the console is %v", s.ConsoleHeight)
	}
	drag(tt, 600, y-60, 600, 2)
	if a, _ := tt.AccessByRole(ui.RoleSeparator, "Activity"); s.ConsoleHeight != float32(a.Max) {
		t.Errorf("dragged to the top: %v, max %v", s.ConsoleHeight, a.Max)
	}
}

func TestPaneShellPaletteKeys(t *testing.T) {
	var s PaneShell
	insertLink := 0
	text := ""
	tt := ui.NewTester(func(c *ui.Context) {
		s.View(c, PaneFrame{Content: func() {
			ui.TextArea(c, &text).Label("Editor").Height(200).OnShortcut(ui.Cmd, ui.KeyK, func() { insertLink++ })
			ui.TextInput(c, &text).Label("Filter")
		}})
	}, 1000, 600)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Fatal("Ctrl K with no focus did not open the palette")
	}
	s.Palette.Open = false
	tt.Frame()
	if err := tt.Click("Editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if s.Palette.Open || insertLink != 1 {
		t.Fatalf("Ctrl K in the editor: palette %v, Insert Link ran %d times", s.Palette.Open, insertLink)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !s.Palette.Open {
		t.Error("Ctrl Shift P in the editor did not open the palette")
	}
}

func TestPaneShellInEditorKeepsCtrlK(t *testing.T) {
	var s PaneShell
	inEditor := true
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{InEditor: inEditor}) }, 800, 500)
	tt.Key(ui.Cmd, ui.KeyK)
	if s.Palette.Open {
		t.Fatal("Ctrl K opened the palette inside an editor")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !s.Palette.Open {
		t.Fatal("Ctrl Shift P did not open the palette inside an editor")
	}
	s.Palette.Open, inEditor = false, false
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("Ctrl K did not open the palette outside an editor")
	}
}

func TestPaletteFieldIsAButton(t *testing.T) {
	var p Palette
	tt := ui.NewTester(func(c *ui.Context) { PaletteField(c, &p); p.View(c, nil) }, 800, 400)
	a, ok := tt.AccessByRole(ui.RoleButton, "Search or run a command")
	if !ok || !a.Focusable || a.KeyShortcuts != "Ctrl+K" {
		t.Fatalf("the field: %+v, %v", a, ok)
	}
	if !tt.HasText("Ctrl") || !tt.HasText("K") {
		t.Errorf("no key shown: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnter)
	if !p.Open {
		t.Error("Enter on the field did not open the palette")
	}
}

func TestPaneShellConnectingCoversTheWindow(t *testing.T) {
	var s PaneShell
	retried := false
	conn := &Connection{State: Connecting, Daemon: "exampled", OnRetry: func() { retried = true }}
	tt := ui.NewTester(func(c *ui.Context) {
		f := segmentedFrame(c, NewConsole(16))
		f.Connection = conn
		s.View(c, f)
	}, 1200, 700)
	tt.Frame()
	if !tt.HasText("Connecting to exampled…") || tt.HasText("side body") || tt.HasText("Synced") {
		t.Fatalf("%q", tt.Texts())
	}
	if _, ok := tt.Find("Main menu"); !ok {
		t.Error("no main menu while connecting: Settings would be out of reach")
	}
	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	if !retried {
		t.Error("Retry did not run")
	}
	tt.Key(ui.Cmd, ui.KeyK)
	if !s.Palette.Open {
		t.Error("no palette while connecting")
	}
	s.Palette.Open = false
	conn.State = Reconnecting
	tt.Frame()
	if !tt.HasText("side body") {
		t.Errorf("reconnecting hides the content: %q", tt.Texts())
	}
}

func TestPaneShellDrawsDialogsAndInfo(t *testing.T) {
	var s PaneShell
	s.Dialogs = &Dialogs{}
	s.Dialogs.Ask(Confirm{Title: "Sure?", Action: "Yes"})
	tt := ui.NewTester(func(c *ui.Context) { s.View(c, PaneFrame{}) }, 800, 500)
	tt.Frame()
	if !tt.HasText("Sure?") {
		t.Errorf("%q", tt.Texts())
	}
}

func TestConsoleViewTakesATitleAndItsLevels(t *testing.T) {
	v := ConsoleView{Title: "Activity", Subtitle: "what the service did", EmptyNote: "No activity yet."}
	v.Offered = 1<<LevelInfo | 1<<LevelWarn | 1<<LevelError | 1<<LevelDebug
	con := NewConsole(16)
	tt := ui.NewTester(func(c *ui.Context) { v.View(c, con) }, 1000, 300)
	tt.Frame()
	if !tt.HasText("Activity") || !tt.HasText("No activity yet.") || tt.HasText("Console") {
		t.Errorf("%q", tt.Texts())
	}
	if _, ok := tt.AccessByRole(ui.RoleRegion, "Activity"); !ok {
		t.Error("the region is not named by the title")
	}
	for _, name := range []string{"Info", "Warn", "Error", "Debug"} {
		if !tt.HasText(name) {
			t.Errorf("no %s toggle in %q", name, tt.Texts())
		}
	}
	for _, name := range []string{"Output", "Command", "Dry run"} {
		if tt.HasText(name) {
			t.Errorf("%s is offered: %q", name, tt.Texts())
		}
	}
	if _, ok := tt.Find("Close Activity"); !ok {
		t.Error("no Close Activity button")
	}
	con.Add(LevelCommand, "x", "a line of a level not offered")
	con.Add(LevelInfo, "x", "a line of a level offered")
	tt.Frame()
	if tt.HasText("a line of a level not offered") || !tt.HasText("a line of a level offered") {
		t.Errorf("%q", tt.Texts())
	}
}

func TestConsoleViewDefaultsAreUnchanged(t *testing.T) {
	var v ConsoleView
	tt := ui.NewTester(func(c *ui.Context) { v.View(c, NewConsole(16)) }, 1000, 300)
	tt.Frame()
	for _, want := range []string{"Console", "every command the app runs, as it runs", "Commands and their output show here as they run.", "Output", "Command", "Dry run"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	if _, ok := tt.Find("Close the console"); !ok {
		t.Error("no Close the console button")
	}
}
