package frostui

import (
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// KeyScope says where a key of a command works, by what has the keyboard
// focus. Whatever the scope, a text input still takes the keys it types
// and edits with (letters, Backspace, the arrows, Ctrl+C), and an element
// around the focus that binds the key itself (Element.OnShortcut) takes
// it first.
type KeyScope uint8

const (
	// OutsideEditors keys work wherever the focus is, except in a text
	// editor (CommandTable.Editing), which takes them for itself: Ctrl K
	// opens the palette, but is Insert Link in an editor. The zero value.
	OutsideEditors KeyScope = iota
	// OutsideText keys step aside while any element that takes text has
	// the focus: a one-line field as well as an editor.
	OutsideText
	// Everywhere keys work wherever the focus is, editors included:
	// Ctrl+Shift+P opens the palette everywhere.
	Everywhere
)

// KeyBinding is a key that runs a command: the key, its modifiers (ui.Cmd
// is Ctrl, or Command on macOS) and where it works.
type KeyBinding struct {
	Mods  ui.Modifiers
	Key   ui.Key
	Scope KeyScope
}

// Caps returns the key caps of the binding, as {"Ctrl", "Shift", "P"},
// for Shortcut and the palette: Ctrl, Alt, Shift and Super (Win on
// Windows) in that order, and the symbols of macOS there (⌃ ⌥ ⇧ ⌘).
func (b KeyBinding) Caps() []string {
	var caps []string
	mods := [...]struct {
		m          ui.Modifiers
		name, symb string
	}{{ui.Ctrl, "Ctrl", "⌃"}, {ui.Alt, "Alt", "⌥"}, {ui.Shift, "Shift", "⇧"}, {ui.Super, "Super", "⌘"}}
	for _, m := range mods {
		if b.Mods&m.m == 0 {
			continue
		}
		switch {
		case runtime.GOOS == "darwin":
			caps = append(caps, m.symb)
		case m.m == ui.Super && runtime.GOOS == "windows":
			caps = append(caps, "Win")
		default:
			caps = append(caps, m.name)
		}
	}
	return append(caps, keyName(b.Key))
}

// String returns the binding as a menu or a tooltip writes it: "Ctrl+K",
// or "⌘K" on macOS.
func (b KeyBinding) String() string {
	sep := "+"
	if runtime.GOOS == "darwin" {
		sep = ""
	}
	return strings.Join(b.Caps(), sep)
}

// keyName is the cap of a key, as a keyboard prints it.
func keyName(k ui.Key) string {
	switch {
	case k >= ui.KeyA && k <= ui.KeyZ:
		return string(rune('A' + k - ui.KeyA))
	case k >= ui.Key0 && k <= ui.Key9:
		return string(rune('0' + k - ui.Key0))
	case k >= ui.KeyF1 && k <= ui.KeyF12:
		return "F" + strconv.Itoa(int(k-ui.KeyF1)+1)
	}
	if n, ok := keyNames[k]; ok {
		return n
	}
	return "?"
}

var keyNames = map[ui.Key]string{
	ui.KeyEnter: "Enter", ui.KeyEscape: "Esc", ui.KeyBackspace: "Backspace", ui.KeyTab: "Tab",
	ui.KeySpace: "Space", ui.KeyDelete: "Delete", ui.KeyInsert: "Insert", ui.KeyHome: "Home",
	ui.KeyEnd: "End", ui.KeyPageUp: "Page Up", ui.KeyPageDown: "Page Down", ui.KeyLeft: "←",
	ui.KeyRight: "→", ui.KeyUp: "↑", ui.KeyDown: "↓", ui.KeyMinus: "-", ui.KeyEqual: "=",
	ui.KeyComma: ",", ui.KeyPeriod: ".", ui.KeySlash: "/", ui.KeySemicolon: ";", ui.KeyQuote: "'",
	ui.KeyBracketLeft: "[", ui.KeyBracketRight: "]", ui.KeyBackslash: "\\", ui.KeyBackquote: "`",
	ui.KeyContextMenu: "Menu", ui.KeyBack: "Back", ui.KeyForward: "Forward",
}

// PaletteBindings are the keys that open the command palette in every
// frostyard window: Ctrl K, except while a text editor has the focus,
// where it is the editor's (Insert Link), and Ctrl+Shift+P everywhere,
// editors included.
func PaletteBindings() []KeyBinding {
	return []KeyBinding{
		{Mods: ui.Cmd, Key: ui.KeyK, Scope: OutsideEditors},
		{Mods: ui.Cmd | ui.Shift, Key: ui.KeyP, Scope: Everywhere},
	}
}

// CommandTable is an app's commands, one entry per command: its ID,
// title, group, keys (Command.Bindings), enabled rule (Command.Enabled)
// and run function. From the one table the app binds the keys at the
// window (Bind), fills the palette (Commands, or Frame.Table), builds
// menus (Menu, MenuItem) and lists the shortcuts (Shortcuts). It belongs
// to the UI thread, and binds the keys of one window: give each window
// its own table.
type CommandTable struct {
	// Editing reports whether a text editor has the keyboard focus in
	// c's window: keys of the scope OutsideEditors step aside while it
	// does. Nil counts any element that takes text as an editor
	// (c.FocusTakesText), so that a key never goes past a field; an app
	// with editors names them, as with handles bound to them:
	//
	//	table.Editing = func(c *ui.Context) bool { return c.FocusWithin(app.editor) }
	Editing func(c *ui.Context) bool

	cmds  []Command
	index map[string]int
	// provided are the IDs of the commands a frame brought (provide).
	provided map[string]bool
}

// NewCommandTable returns a table of cmds, in their order.
func NewCommandTable(cmds ...Command) *CommandTable {
	t := &CommandTable{}
	t.Add(cmds...)
	return t
}

// Add adds commands at the end of the table; one whose ID the table
// holds replaces that command in its place.
func (t *CommandTable) Add(cmds ...Command) {
	if t.index == nil {
		t.index = map[string]int{}
	}
	for _, cmd := range cmds {
		delete(t.provided, cmd.ID)
		if i, ok := t.index[cmd.ID]; ok {
			t.cmds[i] = cmd
			continue
		}
		t.index[cmd.ID] = len(t.cmds)
		t.cmds = append(t.cmds, cmd)
	}
}

// Remove takes the commands with these IDs out of the table.
func (t *CommandTable) Remove(ids ...string) {
	gone := map[string]bool{}
	for _, id := range ids {
		gone[id] = true
		delete(t.provided, id)
	}
	kept := t.cmds[:0]
	for _, cmd := range t.cmds {
		if !gone[cmd.ID] {
			kept = append(kept, cmd)
		}
	}
	clear(t.cmds[len(kept):])
	t.cmds = kept
	t.index = map[string]int{}
	for i, cmd := range t.cmds {
		t.index[cmd.ID] = i
	}
}

// provide adds a command a frame brings to the table, as the palette's,
// or refreshes the one it brought in an earlier frame. A command of the
// app's with the ID stays as it is, and the frame's gives up the keys
// another command binds, so that no key is bound twice.
func (t *CommandTable) provide(cmd Command) {
	if _, ok := t.index[cmd.ID]; ok && !t.provided[cmd.ID] {
		return
	}
	cmd.Bindings = slices.DeleteFunc(slices.Clone(cmd.Bindings), func(b KeyBinding) bool {
		return slices.ContainsFunc(t.cmds, func(other Command) bool {
			return other.ID != cmd.ID && slices.ContainsFunc(other.Bindings, func(o KeyBinding) bool {
				return o.Mods == b.Mods && o.Key == b.Key
			})
		})
	})
	t.Add(cmd)
	if t.provided == nil {
		t.provided = map[string]bool{}
	}
	t.provided[cmd.ID] = true
}

// withdraw takes back a command a frame brought; one of the app's stays.
func (t *CommandTable) withdraw(id string) {
	if t.provided[id] {
		t.Remove(id)
	}
}

// merged returns cmds, less those whose ID the table holds, then the
// table's commands: what a frame gives the palette.
func (t *CommandTable) merged(cmds []Command) []Command {
	out := make([]Command, 0, len(cmds)+len(t.cmds))
	for _, cmd := range cmds {
		if _, ok := t.index[cmd.ID]; !ok {
			out = append(out, cmd)
		}
	}
	return append(out, t.Commands()...)
}

// Command returns the command with the ID, as Commands gives it.
func (t *CommandTable) Command(id string) (Command, bool) {
	i, ok := t.index[id]
	if !ok {
		return Command{}, false
	}
	return t.cmds[i].Resolve(), true
}

// Commands returns the table's commands in order, each as Resolve gives
// it (Disabled and Reason from Enabled, Keys from Bindings): the
// palette's commands, as Frame.Commands.
func (t *CommandTable) Commands() []Command {
	out := make([]Command, len(t.cmds))
	for i, cmd := range t.cmds {
		out[i] = cmd.Resolve()
	}
	return out
}

// Enabled reports whether the command with the ID runs now, and why
// not: false with "" for an ID the table does not hold.
func (t *CommandTable) Enabled(id string) (ok bool, reason string) {
	cmd, found := t.Command(id)
	if !found {
		return false, ""
	}
	return !cmd.Disabled && cmd.Run != nil, cmd.Reason
}

// Run runs the command with the ID when it is enabled, and reports
// whether it ran.
func (t *CommandTable) Run(id string) bool {
	if ok, _ := t.Enabled(id); !ok {
		return false
	}
	t.cmds[t.index[id]].Run()
	return true
}

// Tooltip returns the command's title with its first key, as "Reply
// (Ctrl+R)", for an icon button's label and tooltip; the title alone
// without a key, "" for an ID the table does not hold.
func (t *CommandTable) Tooltip(id string) string {
	i, ok := t.index[id]
	if !ok {
		return ""
	}
	cmd := t.cmds[i]
	if len(cmd.Bindings) == 0 {
		return cmd.Title
	}
	return cmd.Title + " (" + cmd.Bindings[0].String() + ")"
}

// Bind binds the keys of the table's enabled commands at c's window,
// with ui.Context.OnWindowShortcut, where their scopes let them work:
// call it in every frame, as the shell does with Frame.Table. When two
// commands bind one key, the first enabled one in the table takes it, and
// the key is bound once. Keys bound behind a dialog wait, as every window
// shortcut does.
func (t *CommandTable) Bind(c *ui.Context) {
	editing, text := t.focus(c)
	type chord struct {
		mods ui.Modifiers
		key  ui.Key
	}
	bound := map[chord]bool{}
	for _, cmd := range t.cmds {
		if len(cmd.Bindings) == 0 {
			continue
		}
		if ok, _ := t.Enabled(cmd.ID); !ok {
			continue
		}
		id := cmd.ID
		for _, b := range cmd.Bindings {
			if k := (chord{b.Mods, b.Key}); !bound[k] && b.works(editing, text) {
				bound[k] = true
				c.OnWindowShortcut(b.Mods, b.Key, func() { t.Run(id) })
			}
		}
	}
}

// BindKeys binds keys that run run at c's window, with the table's rule
// of where each scope works, for keys outside the table, as a window of
// the app's own binds PaletteBindings.
func (t *CommandTable) BindKeys(c *ui.Context, keys []KeyBinding, run func()) {
	editing, text := t.focus(c)
	for _, b := range keys {
		if b.works(editing, text) {
			c.OnWindowShortcut(b.Mods, b.Key, run)
		}
	}
}

// focus tells whether an editor, and any element taking text, has the
// keyboard focus in c's window.
func (t *CommandTable) focus(c *ui.Context) (editing, text bool) {
	text = c.FocusTakesText()
	if t != nil && t.Editing != nil {
		return t.Editing(c), text
	}
	return text, text
}

// works reports whether the binding works with the focus in an editor
// (editing) or another element taking text (text).
func (b KeyBinding) works(editing, text bool) bool {
	switch b.Scope {
	case Everywhere:
		return true
	case OutsideText:
		return !text && !editing
	default:
		return !editing
	}
}

// MenuItem adds the command with the ID to m: its title, its first key,
// a check mark when Checked says so, grayed out with the reason as its
// tooltip when it does not run. Choosing it runs the command. It returns
// the item, nil for an ID the table does not hold.
func (t *CommandTable) MenuItem(m *ui.Menu, id string) *ui.MenuItem {
	cmd, ok := t.Command(id)
	if !ok {
		return nil
	}
	it := m.Item(cmd.Title).ID(id)
	if len(cmd.Bindings) > 0 {
		it.Shortcut(cmd.Bindings[0].Mods, cmd.Bindings[0].Key)
	}
	if cmd.Checked != nil {
		it.Checked(cmd.Checked())
	}
	if cmd.Disabled || cmd.Run == nil {
		it.Disabled(true)
		if cmd.Reason != "" {
			it.Tooltip(cmd.Reason)
		}
	}
	if it.Chosen() {
		t.Run(id)
	}
	return it
}

// Menu returns a function that builds a menu of the commands with these
// IDs, in order, "" adding a separator: for Frame.Menu, ui.MenuButton,
// Element.ContextMenu and Element.PopoverMenu. Nest one in another with
// ui.Menu.Submenu:
//
//	row.ContextMenu(table.Menu("reply", "forward", "", "archive"))
//	m.Submenu("Mark", table.Menu("mark-read", "mark-unread"))
func (t *CommandTable) Menu(ids ...string) func(m *ui.Menu) {
	return func(m *ui.Menu) {
		for _, id := range ids {
			if id == "" {
				m.Separator()
				continue
			}
			t.MenuItem(m, id)
		}
	}
}

// Shortcuts lists the commands that have keys, in the table's order, for
// ShortcutsBody: grouped by Group, in the order the groups first come,
// with a line for each key.
func (t *CommandTable) Shortcuts() []ShortcutEntry {
	order := map[string]int{}
	var groups [][]ShortcutEntry
	for _, cmd := range t.cmds {
		if len(cmd.Bindings) == 0 {
			continue
		}
		g, ok := order[cmd.Group]
		if !ok {
			g = len(groups)
			order[cmd.Group] = g
			groups = append(groups, nil)
		}
		for _, b := range cmd.Bindings {
			groups[g] = append(groups[g], ShortcutEntry{Group: cmd.Group, Title: cmd.Title, Keys: b.Caps()})
		}
	}
	var out []ShortcutEntry
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}
