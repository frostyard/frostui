package frostui

import "github.com/egoist/mygo/ui"

// ConsoleCommandID is the ID of PaneShell.ConsoleCommand.
const ConsoleCommandID = "frostui.console"

// The console split's sizes, in DIPs.
const (
	consoleDefaultHeight = 240
	consoleMinHeight     = 100
	consoleMainMinHeight = 120
	consoleStep          = 16
)

// PaneShell is the window of an app whose content is its own panes: a top
// bar, full-height unscrolled content with no rail, an optional console
// under it and a status bar. A top bar's far left is the sidebar toggle and
// its end the palette field's side, the console's button and the main menu,
// in every app. Keep one PaneShell in the app's state and call View from
// the root view every frame.
type PaneShell struct {
	Palette Palette
	Console ConsoleView
	// Dialogs are the confirmations the shell draws; share one with the
	// app's views. Nil draws none.
	Dialogs *Dialogs
	Info    Info
	// Panes is the split of PaneFrame.Panes; save and restore its Layout.
	Panes Panes
	// ConsoleHeight is the console's height in DIPs, which the user
	// changes and the app may save and restore; 0 is 240.
	ConsoleHeight float32

	window float32 // the window's height in the last build
	total  float32 // the height the console split last laid out
}

// PaneFrame is what the app gives the shell to draw each frame. With Panes
// the top bar is made of each pane's Bar, as wide as the panes under them;
// without, the bar has Start, the palette field in the middle, and End.
type PaneFrame struct {
	// SidebarPane is the ID of the pane the toggle at the far left of the
	// top bar shows and hides, of SidebarPanes, or of the shell's Panes
	// when that is nil. "" draws no toggle.
	SidebarPane  string
	SidebarPanes *Panes
	// Panes are the window's panes, which take the room under the top bar.
	Panes []Pane
	// Start draws a bar without Panes' items, after the toggle.
	Start func()
	// End draws the items after the palette field, before the console's
	// button and the main menu.
	End func()
	// Content draws the window's content without Panes: full height, not
	// scrolled and not padded.
	Content func()
	// Menu builds the main menu, the last item before the window controls;
	// nil for none.
	Menu func(m *ui.Menu)
	// InEditor says a text editor has the focus, so Ctrl K stays with it.
	// An editor that binds Ctrl K on its own element needs no flag: the
	// element around the focus hears the key first.
	InEditor bool
	// Commands are the palette's commands. The console's is added when
	// there is a console and none has its ID.
	Commands []Command
	// Connection is the status bar's first item; its Connecting state
	// draws the connecting screen over the whole window. Nil for none.
	Connection *Connection
	// Status and StatusEnd are the status bar's items at its start, after
	// the connection, and at its end.
	Status, StatusEnd []StatusItem
	// ConsoleData is the console the console pane shows, nil for no
	// console.
	ConsoleData *Console
}

// View draws the shell. Ctrl K opens the palette, except inside a text
// editor, which keeps the key for itself (PaneFrame.InEditor); Ctrl Shift P
// opens it from everywhere. Ctrl J shows or hides the console.
func (s *PaneShell) View(c *ui.Context, f PaneFrame) {
	t := Adapt(c)
	s.Palette.RunPending()
	if !f.InEditor {
		c.OnWindowShortcut(ui.Cmd, ui.KeyK, func() { s.Palette.Show() })
	}
	c.OnWindowShortcut(ui.Cmd|ui.Shift, ui.KeyP, func() { s.Palette.Show() })
	if f.Connection != nil && f.Connection.State == Connecting {
		connectingWindow(c, *f.Connection, func() { s.mainMenu(c, f) })
	} else {
		if f.ConsoleData != nil {
			c.OnWindowShortcut(ui.Cmd, ui.KeyJ, s.toggleConsole)
		}
		ui.Column(c).Fill().AlignItems(ui.Stretch).Background(t.Background).Children(func() {
			if len(f.Panes) == 0 {
				s.topBar(c, f)
				ui.Box(c).Height(1).Background(t.Border)
			}
			s.area(c, f)
			if f.Connection != nil || len(f.Status) > 0 || len(f.StatusEnd) > 0 {
				ui.Box(c).Height(1).Background(t.Border)
				s.statusBar(c, f)
			}
		})
	}
	if s.Dialogs != nil {
		s.Dialogs.View(c)
	}
	s.Info.View(c)
	s.Palette.View(c, s.commands(f))
}

func (s *PaneShell) toggleConsole() { s.Console.Open = !s.Console.Open }

// ConsoleCommand is the palette's command that shows or hides the console,
// in the group "View". PaneFrame adds it; an app that keeps its commands in
// a table of its own may add it there, with another Group.
func (s *PaneShell) ConsoleCommand() Command {
	verb := "Show "
	if s.Console.Open {
		verb = "Hide "
	}
	return Command{ID: ConsoleCommandID, Title: verb + s.Console.title(), Group: "View", Icon: "utilities-terminal",
		Keys: []string{"Ctrl", "J"}, Run: s.toggleConsole}
}

func (s *PaneShell) commands(f PaneFrame) []Command {
	if f.ConsoleData == nil {
		return f.Commands
	}
	for _, cmd := range f.Commands {
		if cmd.ID == ConsoleCommandID {
			return f.Commands
		}
	}
	return append(f.Commands[:len(f.Commands):len(f.Commands)], s.ConsoleCommand())
}

// area draws what lies between the bars: the panes or the content, with
// the console under them when it is open. The content stays the first
// child either way, so opening the console does not rebuild it.
func (s *PaneShell) area(c *ui.Context, f PaneFrame) {
	area := ui.Column(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Clip()
	open := s.Console.Open && f.ConsoleData != nil
	total := area.LayoutBounds().H
	_, window := c.Size()
	if total != s.total || window != s.window {
		s.total, s.window = total, window
		c.AnimationFrame()
	}
	area.Children(func() {
		main := ui.Column(c).Grow(1).Shrink(1).Basis(0).MinHeight(0).AlignItems(ui.Stretch).Clip()
		main.Children(func() {
			switch {
			case len(f.Panes) > 0:
				s.Panes.view(c, f.Panes, func(pn Pane, first, last bool) { s.segment(c, f, pn, first, last) }).Grow(1)
			case f.Content != nil:
				f.Content()
			}
		})
		if !open {
			return
		}
		height := s.consoleHeight(total)
		hi := height
		if total > 0 {
			hi = max(total-1-consoleMainMinHeight, consoleMinHeight)
		}
		drawGrip(c.Key("console-separator"), grip{
			label: s.Console.title(), lo: consoleMinHeight, hi: max(hi, height), now: height, step: consoleStep, sign: -1,
			set:   func(v float32) { s.ConsoleHeight = v },
			reset: func() { s.ConsoleHeight = 0 },
		})
		ui.Column(c).Height(height).Shrink(0).AlignItems(ui.Stretch).Clip().Children(func() {
			s.Console.View(c, f.ConsoleData)
		})
	})
}

// consoleHeight is the console's height in an area total high, 0 for one
// not yet laid out.
func (s *PaneShell) consoleHeight(total float32) float32 {
	h := s.ConsoleHeight
	if h <= 0 {
		h = consoleDefaultHeight
	}
	if total > 0 {
		h = min(h, total-1-consoleMainMinHeight)
	}
	return max(h, consoleMinHeight)
}

// topBar is the bar of a frame without panes: its sides share the room
// left, so the palette field is in the middle of the window.
func (s *PaneShell) topBar(c *ui.Context, f PaneFrame) {
	t := c.Theme()
	tb := c.TitleBar()
	ui.Row(c).Height(TopBarHeight).Padding(0, SpaceL+tb.Right, 0, SpaceL+tb.Left).Gap(SpaceM).AlignItems(ui.Center).
		Background(Chrome(t)).DragWindow().Children(func() {
		ui.Row(c).Grow(1).Shrink(1).Basis(0).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Children(func() {
			s.sidebarToggle(c, f)
			if f.Start != nil {
				f.Start()
			}
		})
		PaletteField(c, &s.Palette).Shrink(1).MinWidth(180)
		ui.Row(c).Grow(1).Shrink(1).Basis(0).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Justify(ui.End).Children(func() {
			s.trailing(c, f)
		})
	})
}

// segment is the part of the top bar over one pane: the first shown starts
// with the sidebar toggle and the last shown ends with the palette field,
// End, the console's button and the main menu.
func (s *PaneShell) segment(c *ui.Context, f PaneFrame, pn Pane, first, last bool) {
	t := c.Theme()
	tb := c.TitleBar()
	var left, right float32 = SpaceM, SpaceM
	if first {
		left += tb.Left
	}
	if last {
		right += tb.Right
	}
	ui.Row(c).Height(TopBarHeight).Shrink(0).Padding(0, right, 0, left).Gap(SpaceXS).AlignItems(ui.Center).
		Background(Chrome(t)).DragWindow().Clip().Children(func() {
		if first {
			s.sidebarToggle(c, f)
		}
		items := ui.Row(c).Grow(1).Shrink(1).Basis(0).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Clip()
		if pn.Bar != nil {
			items.Children(pn.Bar)
		}
		if last {
			PaletteField(c, &s.Palette).Width(300).Shrink(1).MinWidth(160)
			s.trailing(c, f)
		}
	})
	ui.Box(c).Height(1).Shrink(0).Background(t.Border)
}

func (s *PaneShell) sidebarToggle(c *ui.Context, f PaneFrame) {
	if f.SidebarPane == "" {
		return
	}
	panes := f.SidebarPanes
	if panes == nil {
		panes = &s.Panes
	}
	open := panes.Shown(f.SidebarPane)
	if toggleButton(c, "sidebar-show", "Sidebar", open).Clicked() {
		panes.Toggle(f.SidebarPane)
	}
}

// trailing is the end of the top bar, in the order every app shares: the
// app's items, the console's button, and the main menu last.
func (s *PaneShell) trailing(c *ui.Context, f PaneFrame) {
	if f.End != nil {
		f.End()
	}
	if f.ConsoleData != nil {
		if toggleButton(c, "utilities-terminal", s.Console.title()+" (Ctrl J)", s.Console.Open).Clicked() {
			s.toggleConsole()
		}
	}
	s.mainMenu(c, f)
}

func (s *PaneShell) mainMenu(c *ui.Context, f PaneFrame) {
	if f.Menu != nil {
		ui.MenuButton(c, "", f.Menu).Label("Main menu").Tooltip("Main menu").Children(func() {
			Icon(c, "open-menu")
		})
	}
}

// toggleButton is an IconButton that stays pressed while on, and says so
// to assistive technology.
func toggleButton(c *ui.Context, icon, label string, on bool) ui.Element {
	return IconButton(c, icon, label).Selected(on).Role(ui.RoleToggleButton).Checked(on)
}

func (s *PaneShell) statusBar(c *ui.Context, f PaneFrame) {
	t := c.Theme()
	ui.Row(c).Height(28).Shrink(0).Padding(0, SpaceM).Gap(SpaceL).AlignItems(ui.Center).Background(Chrome(t)).Role(ui.RoleStatus).Children(func() {
		if f.Connection != nil {
			ConnectionStatus(c.Key("connection"), *f.Connection)
		}
		for i, it := range f.Status {
			statusItem(c.Key([2]any{"s", i}), it)
		}
		ui.Spacer(c)
		for i, it := range f.StatusEnd {
			statusItem(c.Key([2]any{"e", i}), it)
		}
	})
}

// PaletteField is the field that opens the command palette, the same in
// every window: it reads "Search or run a command" and shows its key. It
// is a button, 380 DIPs wide unless the caller sets another width.
func PaletteField(c *ui.Context, p *Palette) ui.Element {
	t := c.Theme()
	field := ui.ButtonBase(c).Row().Width(380).Height(32).Padding(0, 10).Gap(SpaceS).Radius(16).AlignItems(ui.Center).
		Background(t.Background).Border(1, t.Border).Cursor(ui.CursorPointer).
		Label("Search or run a command").Tooltip("Command palette (Ctrl K)").KeyShortcuts(ui.Cmd, ui.KeyK)
	if field.Hovered() {
		field.Border(1, t.Accent.Alpha(0.6))
	}
	if field.Clicked() {
		p.Show()
	}
	return field.Children(func() {
		Icon(c, "system-search").TextColor(t.TextMuted)
		Muted(c, "Search or run a command").Grow(1).Shrink(1).SingleLine()
		Shortcut(c, "Ctrl", "K")
	})
}
