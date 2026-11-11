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

	// The widths the top bar had in the last frame: the sidebar toggle,
	// the end of a bar of segments (barEnd) and the trailing controls in
	// it (trailRow).
	toggleW, endW, trailW measured
	// room is the width of the last segment shown less its start, worked
	// out in this frame: the room the bar's end fits the palette field to.
	room float32
}

// PaneFrame is what the app gives the shell to draw each frame. With Panes
// the top bar is made of each pane's Bar, as wide as the panes under them;
// without, the bar has Start, the palette field in the middle, and End.
type PaneFrame struct {
	// SidebarPane is the ID of the pane the toggle at the far left of the
	// top bar shows and hides, of SidebarPanes, or of the shell's Panes
	// when that is nil; the pane must be Collapsible. "" draws no toggle.
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
	// element around the focus hears the key first. With a Table it is
	// not read: CommandTable.Editing says.
	InEditor bool
	// Commands are the palette's commands. The console's is added when
	// there is a console and none has its ID.
	Commands []Command
	// Table, when set, is the app's command table, which owns the window's
	// shortcuts: the shell adds the palette's command (Palette.Command)
	// and the console's to it, binds its keys (CommandTable.Bind) and
	// binds none of its own, and the palette lists Commands and then the
	// table's commands.
	Table *CommandTable
	// Connection is the status bar's first item; its Connecting state
	// draws the connecting screen over the whole window. Nil for none.
	Connection *Connection
	// Status and StatusEnd are the status bar's items at its start, after
	// the connection, and at its end.
	Status, StatusEnd []StatusItem
	// ConsoleData is the console the console pane shows, nil for no
	// console.
	ConsoleData *Console
	// ToastCloseLabel names the close button of the window's toasts;
	// empty is "Close". The shell draws the toasts as TopToasts does,
	// under the top bar, so that a toast does not land on the button that
	// raised it and stop its own time under the pointer.
	ToastCloseLabel string
}

// View draws the shell. Ctrl K opens the palette, except inside a text
// editor, which keeps the key for itself (PaneFrame.InEditor, or the
// Table's Editing); Ctrl Shift P opens it from everywhere. Ctrl J shows or
// hides the console.
func (s *PaneShell) View(c *ui.Context, f PaneFrame) {
	t := Adapt(c)
	s.Palette.RunPending()
	connecting := f.Connection != nil && f.Connection.State == Connecting
	if f.Table != nil {
		s.joinTable(f, connecting)
		f.Table.Bind(c)
	} else {
		if !f.InEditor {
			c.OnWindowShortcut(ui.Cmd, ui.KeyK, func() { s.Palette.Show() })
		}
		c.OnWindowShortcut(ui.Cmd|ui.Shift, ui.KeyP, func() { s.Palette.Show() })
	}
	if connecting {
		connectingWindow(c, *f.Connection, func() { s.mainMenu(c, f) })
	} else {
		if f.ConsoleData != nil && f.Table == nil {
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
	closeLabel := f.ToastCloseLabel
	if closeLabel == "" {
		closeLabel = "Close"
	}
	toastsAt(c, closeLabel, TopBarHeight+SpaceS)
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
	return consoleCommand(verb+s.Console.title(), s.toggleConsole)
}

// consoleCommand is the command that shows or hides a frame's console.
func consoleCommand(title string, toggle func()) Command {
	return Command{ID: ConsoleCommandID, Title: title, Group: "View", Icon: "utilities-terminal",
		Keys: []string{"Ctrl", "J"}, Bindings: []KeyBinding{{Mods: ui.Cmd, Key: ui.KeyJ, Scope: Everywhere}}, Run: toggle}
}

// joinConsole brings a frame's console command to the table, whose keys
// are then the window's only ones, or takes it back with show false. An
// app's own console command in cmds goes in the table, so its key works.
func joinConsole(t *CommandTable, console Command, cmds []Command, show bool) {
	if !show {
		t.withdraw(ConsoleCommandID)
		return
	}
	for _, cmd := range cmds {
		if cmd.ID == ConsoleCommandID {
			bindings := console.Bindings
			console = cmd
			if len(console.Bindings) == 0 {
				console.Bindings = bindings
			}
		}
	}
	t.provide(console)
}

// joinTable brings the palette's and the console's commands to the table.
// There is no console to toggle while the connecting screen covers the
// window.
func (s *PaneShell) joinTable(f PaneFrame, connecting bool) {
	f.Table.provide(s.Palette.Command())
	joinConsole(f.Table, s.ConsoleCommand(), f.Commands, f.ConsoleData != nil && !connecting)
}

func (s *PaneShell) commands(f PaneFrame) []Command {
	if f.Table != nil {
		return f.Table.merged(f.Commands)
	}
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
				s.Panes.view(c, f.Panes, func(pn Pane, first, last bool, width, after float32) {
					s.segment(c, f, pn, first, last, width, after)
				}).Grow(1)
				s.barEnd(c, f)
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
// left, so the palette field is in the middle of the window. The toggle
// and the trailing controls keep their size. The field gives way first,
// staying in the middle, down to its least width; then Start, then the
// field again, to its compact button.
func (s *PaneShell) topBar(c *ui.Context, f PaneFrame) {
	t := c.Theme()
	tb := c.TitleBar()
	w, _ := c.Size()
	var lead float32
	if f.SidebarPane != "" {
		lead = s.toggleW.w
		if f.Start != nil {
			lead += SpaceXS
		}
	}
	room := w - 2*SpaceL - tb.Left - tb.Right - 2*SpaceM
	var width float32 // compact
	if room-lead-s.trailW.w >= paletteFieldLeast {
		// As wide as sides of the same width allow, so that it stays in
		// the middle, and no narrower than its least.
		width = min(max(room-2*max(lead, s.trailW.w), paletteFieldLeast), paletteFieldWidth)
	}
	ui.Row(c).Height(TopBarHeight).Padding(0, SpaceL+tb.Right, 0, SpaceL+tb.Left).Gap(SpaceM).AlignItems(ui.Center).
		Background(Chrome(t)).DragWindow().Children(func() {
		ui.Row(c).Grow(1).Shrink(1).Basis(0).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Children(func() {
			s.sidebarToggle(c, f)
			if f.Start != nil {
				ui.Row(c).Grow(1).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Clip().Children(f.Start)
			}
		})
		paletteControl(c, &s.Palette, width)
		// No MinWidth: the side is never narrower than the trailing
		// controls.
		ui.Row(c).Grow(1).Shrink(1).Basis(0).AlignItems(ui.Center).Justify(ui.End).Children(func() {
			s.trailRow(c, f)
		})
	})
}

// segment is the part of the top bar over one pane, width DIPs wide, with
// after DIPs of panes after it: the first shown starts with the sidebar
// toggle. The bar's end (barEnd) lies over the end of the last one shown,
// and over the ones before it when the last is too narrow for it: the
// part of a segment under it is padding, so the pane's Bar gives way.
func (s *PaneShell) segment(c *ui.Context, f PaneFrame, pn Pane, first, last bool, width, after float32) {
	t := c.Theme()
	start, end := barEdges(c)
	lead := float32(SpaceM)
	if first {
		lead += start
	}
	under := s.endW.w + SpaceXS + SpaceM + end - after
	trail := max(min(under, width-lead), SpaceM)
	row := ui.Row(c).Height(TopBarHeight).Shrink(0).Gap(SpaceXS).AlignItems(ui.Center).
		Background(Chrome(t)).DragWindow().Clip()
	if c.Direction() == ui.RightToLeft {
		row.Padding(0, lead, 0, trail)
	} else {
		row.Padding(0, trail, 0, lead)
	}
	row.Children(func() {
		if first {
			s.sidebarToggle(c, f)
		}
		items := ui.Row(c).Grow(1).Shrink(1).Basis(0).MinWidth(0).Gap(SpaceXS).AlignItems(ui.Center).Clip()
		if pn.Bar != nil {
			items.Children(pn.Bar)
		}
	})
	if last {
		s.room = width - lead
		if first && f.SidebarPane != "" {
			s.room -= s.toggleW.w + SpaceXS
		}
	}
	ui.Box(c).Height(1).Shrink(0).Background(t.Border)
}

// barEnd draws the end of a bar of segments: the palette field, End, the
// console's button and the main menu, last before the window controls. It
// is the window's, not the last pane's: it lies over the segments at its
// own width, above the separators, and they give way under it. The field
// fits the room the last segment has; with less than its least width there
// it is drawn compact, and when the end is still wider than the last
// segment it reaches over the segments before it.
func (s *PaneShell) barEnd(c *ui.Context, f PaneFrame) {
	t := c.Theme()
	_, end := barEdges(c)
	row := ui.Row(c.Key("bar-end")).Absolute().Top(0).Height(TopBarHeight).Gap(SpaceXS).AlignItems(ui.Center).
		Background(Chrome(t)).DragWindow().ZIndex(1)
	if c.Direction() == ui.RightToLeft {
		row.Left(SpaceM + end)
	} else {
		row.Right(SpaceM + end)
	}
	s.endW.of(c, row)
	// The room left for the field: the last segment less the gap before
	// the end, the trailing controls with the gap before them, and the
	// room by the window controls.
	room := s.room - SpaceXS - s.trailW.w - SpaceXS - SpaceM - end
	row.Children(func() {
		paletteControl(c, &s.Palette, fieldWidth(room, segmentFieldWidth, segmentFieldLeast))
		s.trailRow(c, f)
	})
}

// barEdges returns the room the window controls take at the bar's start,
// where the sidebar toggle is, and at its end, where the main menu is:
// right to left, the title bar's right and left.
func barEdges(c *ui.Context) (start, end float32) {
	tb := c.TitleBar()
	if c.Direction() == ui.RightToLeft {
		return tb.Right, tb.Left
	}
	return tb.Left, tb.Right
}

// trailRow draws the trailing controls, which never shrink.
func (s *PaneShell) trailRow(c *ui.Context, f PaneFrame) {
	row := ui.Row(c).Shrink(0).Gap(SpaceXS).AlignItems(ui.Center)
	s.trailW.of(c, row)
	row.Children(func() { s.trailing(c, f) })
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
	b := toggleButton(c, "sidebar-show", "Sidebar", open).Shrink(0)
	s.toggleW.of(c, b)
	if b.Clicked() {
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
	field := ui.ButtonBase(c).Row().Width(paletteFieldWidth).Height(32).Padding(0, 10).Gap(SpaceS).Radius(16).AlignItems(ui.Center).
		Background(t.Background).Border(1, t.Border).Cursor(ui.CursorPointer).
		Label(paletteLabel).Tooltip(paletteTooltip).KeyShortcuts(ui.Cmd, ui.KeyK)
	if field.Hovered() {
		field.Border(1, t.Accent.Alpha(0.6))
	}
	if field.Clicked() {
		p.Show()
	}
	return field.Children(func() {
		Icon(c, "system-search").TextColor(t.TextMuted)
		Muted(c, paletteLabel).Grow(1).Shrink(1).SingleLine()
		Shortcut(c, "Ctrl", "K")
	})
}

// The palette field's name and tooltip, which its compact button shares.
const (
	paletteLabel   = "Search or run a command"
	paletteTooltip = "Command palette (Ctrl K)"
)

// The palette field's widths in DIPs: what it takes when the bar has room,
// and the least it gives way to before it is drawn compact. The field over
// a bar of segments is narrower than the one in the middle of a bar.
const (
	paletteFieldWidth = 380
	paletteFieldLeast = 180
	segmentFieldWidth = 300
	segmentFieldLeast = 160
)

// fieldWidth is the palette field's width in a bar with room DIPs for it:
// want, or less down to least; 0, for less room than least, draws the
// field compact.
func fieldWidth(room, want, least float32) float32 {
	if room < least {
		return 0
	}
	return min(room, want)
}

// paletteControl draws the palette field width DIPs wide, or its compact
// button for a width of 0.
func paletteControl(c *ui.Context, p *Palette, width float32) ui.Element {
	if width <= 0 {
		return paletteButton(c, p)
	}
	return PaletteField(c, p).Width(width)
}

// paletteButton is the palette field drawn compact, for a bar without
// room for it: its icon on a button with its name, its tooltip and its
// key, which opens the palette.
func paletteButton(c *ui.Context, p *Palette) ui.Element {
	b := ui.Button(c, "").Label(paletteLabel).Tooltip(paletteTooltip).KeyShortcuts(ui.Cmd, ui.KeyK).Shrink(0)
	if b.Clicked() {
		p.Show()
	}
	return b.Children(func() { Icon(c, "system-search") })
}

// measured is the width an element had in the last frame, which a bar
// reads before it builds the element. of asks for one more frame when the
// width changed, and after the element's first build, so the frame after
// settles.
type measured struct {
	w     float32
	built bool
}

func (m *measured) of(c *ui.Context, e ui.Element) {
	if w := e.LayoutBounds().W; w != m.w || !m.built {
		m.w, m.built = w, true
		c.AnimationFrame()
	}
}
