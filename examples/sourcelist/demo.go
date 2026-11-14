package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// folderPrefix starts the IDs of folders in the tree; other items are
// notes.
const folderPrefix = "folder:"

// note is one note of the example.
type note struct {
	id, title, folder, body string
	pinned                  bool
}

// folder is one folder of the example.
type folder struct {
	name, problem string
	syncing       bool
}

// demo is the example's state. It belongs to the UI thread, as frostui's
// models do; the palette source's worker reads only its own copy.
type demo struct {
	post func(func())

	list     frostui.SourceList[string]
	palette  frostui.Palette
	info     frostui.Info
	settings frostui.Settings
	table    *frostui.CommandTable
	editor   ui.Handle

	folders []*folder
	notes   []*note
	current *note
	nextID  int

	spellcheck, sync bool
	theme, author    string
	fontSize         float64
}

// newDemo returns the example's state. post hands work to the UI thread
// (Window.Update in the program, a frostuitest.Loop in the tests).
func newDemo(post func(func())) *demo {
	d := &demo{post: post, theme: "System", author: "Ada", fontSize: 14, sync: true}
	d.folders = []*folder{{name: "Projects"}, {name: "Journal", syncing: true}, {name: "Archive", problem: "The hub refused this folder"}}
	for _, n := range []struct{ title, folder, body string }{
		{"Garden plan", "Projects", "Tomatoes along the south wall.\n\nPress Ctrl K here to insert a link."},
		{"Reading list", "Projects", "Books for the winter."},
		{"Monday", "Journal", "Rain all day."},
		{"Old ideas", "Archive", ""},
	} {
		d.addNote(n.title, n.folder, n.body)
	}
	d.notes[0].pinned = true
	d.setUpList()
	d.setUpCommands()
	d.settings.Search = true
	d.palette.Sources = []frostui.PaletteSource{d.noteSource()}
	d.list.SetOpen("vault", folderPrefix+"Projects", true)
	d.open(d.notes[0])
	return d
}

func (d *demo) addNote(title, folder, body string) *note {
	d.nextID++
	n := &note{id: fmt.Sprintf("note:%d", d.nextID), title: title, folder: folder, body: body}
	d.notes = append(d.notes, n)
	return n
}

func (d *demo) note(id string) *note {
	for _, n := range d.notes {
		if n.id == id {
			return n
		}
	}
	return nil
}

func (d *demo) open(n *note) {
	d.current = n
	d.list.Select("vault", n.id)
}

// setUpList says how the tree shows and what it does.
func (d *demo) setUpList() {
	l := &d.list
	l.Label = "Notes"
	l.Children = func(item string) []string {
		name, ok := strings.CutPrefix(item, folderPrefix)
		if !ok {
			return nil
		}
		ids := []string{}
		for _, n := range d.notes {
			if n.folder == name {
				ids = append(ids, n.id)
			}
		}
		return ids
	}
	l.Row = func(item string) frostui.SourceRow {
		if name, ok := strings.CutPrefix(item, folderPrefix); ok {
			f := d.folderNamed(name)
			return frostui.SourceRow{Title: name, Icon: "folder", Count: fmt.Sprint(len(l.Children(item))),
				Busy: f.syncing, Alert: f.problem}
		}
		n := d.note(item)
		if n == nil {
			return frostui.SourceRow{Title: item}
		}
		return frostui.SourceRow{Title: n.title, Icon: "text-x-generic", Dim: n.body == ""}
	}
	l.OnSelect = func(item string) {
		if n := d.note(item); n != nil {
			d.current = n
		}
	}
	l.Menu = func(m *ui.Menu, item, section string) {
		n := d.note(item)
		if n == nil {
			d.table.MenuItem(m, "new")
			return
		}
		if m.Item("Rename").Shortcut(0, ui.KeyF2).Chosen() {
			l.StartRename(section, item)
		}
		if m.Item("Pinned").Checked(n.pinned).Chosen() {
			n.pinned = !n.pinned
		}
	}
	l.CanRename = func(item string) bool { return d.note(item) != nil }
	l.Rename = func(item, name string) { d.note(item).title = name }
	l.DragItems = func(items []string) any {
		for _, id := range items {
			if d.note(id) == nil {
				return nil
			}
		}
		return items
	}
	frostui.DropOn(l, func(item string, _ []string) bool { return strings.HasPrefix(item, folderPrefix) },
		func(item string, ids []string, _ ui.Modifiers) {
			for _, id := range ids {
				d.note(id).folder = strings.TrimPrefix(item, folderPrefix)
			}
		})
}

func (d *demo) folderNamed(name string) *folder {
	for _, f := range d.folders {
		if f.name == name {
			return f
		}
	}
	return &folder{name: name}
}

// sections are the tree's sections: the pinned notes and the folders.
func (d *demo) sections() []frostui.SourceSection[string] {
	pinned := []string{}
	for _, n := range d.notes {
		if n.pinned {
			pinned = append(pinned, n.id)
		}
	}
	folders := make([]string, len(d.folders))
	for i, f := range d.folders {
		folders[i] = folderPrefix + f.name
	}
	return []frostui.SourceSection[string]{
		{ID: "pinned", Title: "Pinned", Items: pinned,
			Tooltip: "Notes pinned from a note's menu or with Pin Note"},
		{ID: "vault", Title: "Vault", Items: folders},
	}
}

// setUpCommands fills the one table behind the keys, the menu, the
// palette and the shortcuts dialog.
func (d *demo) setUpCommands() {
	hasNote := func() (bool, string) { return d.current != nil, "No note is open" }
	d.table = frostui.NewCommandTable(
		frostui.Command{ID: "new", Title: "New Note", Group: "Notes", Icon: "document-new",
			Bindings: []frostui.KeyBinding{{Mods: ui.Cmd, Key: ui.KeyN, Scope: frostui.Everywhere}},
			Run: func() {
				n := d.addNote("Untitled", "Projects", "")
				d.open(n)
			}},
		frostui.Command{ID: "rename", Title: "Rename Note", Group: "Notes", Icon: "document-edit", Enabled: hasNote,
			Run: func() { d.list.StartRename("vault", d.current.id) }},
		frostui.Command{ID: "pin", Title: "Pin Note", Group: "Notes", Enabled: hasNote,
			Checked: func() bool { return d.current != nil && d.current.pinned },
			Run:     func() { d.current.pinned = !d.current.pinned }},
		frostui.Command{ID: "switcher", Title: "Quick Switcher", Group: "Go to", Icon: "system-search",
			Bindings: []frostui.KeyBinding{{Mods: ui.Cmd, Key: ui.KeyO, Scope: frostui.Everywhere}},
			Run:      func() { d.palette.ShowSource("notes") }},
		frostui.Command{ID: "commands", Title: "Commands", Group: "Go to", Unlisted: true,
			Bindings: []frostui.KeyBinding{{Mods: ui.Cmd, Key: ui.KeyP, Scope: frostui.Everywhere}},
			Run:      func() { d.palette.ShowSource(frostui.PaletteCommands) }},
		frostui.Command{ID: "settings", Title: "Settings", Group: "App", Icon: "preferences-system",
			Bindings: []frostui.KeyBinding{{Mods: ui.Cmd, Key: ui.KeyComma, Scope: frostui.Everywhere}},
			Run:      func() { d.settings.Show("") }},
		frostui.Command{ID: "shortcuts", Title: "Keyboard Shortcuts", Group: "App",
			Bindings: []frostui.KeyBinding{{Key: ui.KeyF1, Scope: frostui.Everywhere}},
			Run:      func() { d.info.Show("Keyboard shortcuts", frostui.ShortcutsBody(d.table.Shortcuts())) }},
	)
	d.table.Editing = func(c *ui.Context) bool { return c.FocusWithin(d.editor) }
}

// noteSource finds notes by title on a worker, and posts them back.
func (d *demo) noteSource() frostui.PaletteSource {
	type found struct{ id, title, folder string }
	return frostui.PaletteSource{ID: "notes", Title: "Notes", Placeholder: "Find a note", Limit: 5, Query: func(q *frostui.PaletteQuery) {
		all := make([]found, len(d.notes))
		for i, n := range d.notes {
			all[i] = found{n.id, n.title, n.folder}
		}
		go func() {
			hits := frostui.Rank(all, q.Text, 50, func(f found) string { return f.title })
			items := make([]frostui.PaletteItem, len(hits))
			for i, h := range hits {
				items[i] = frostui.PaletteItem{ID: h.id, Title: h.title, Detail: h.folder, Icon: "text-x-generic", Run: func() {
					if n := d.note(h.id); n != nil {
						d.open(n)
					}
				}}
			}
			d.post(func() { q.Reply(items) })
		}()
	}}
}

// settingsPages are the settings dialog's sections.
func (d *demo) settingsPages() []frostui.SettingsPage {
	return []frostui.SettingsPage{
		{ID: "editor", Title: "Editor", Icon: "accessories-text-editor", Groups: []frostui.SettingsGroup{{
			Title: "Writing", Settings: []frostui.Setting{
				{Title: "Spell check", Subtitle: "Underline words it does not know", Switch: &d.spellcheck},
				{Title: "Font size", Keywords: "text", Number: &d.fontSize, Min: 8, Max: 32, Step: 1},
				{Title: "Author", Text: &d.author, Placeholder: "Your name"},
			},
		}}},
		{ID: "appearance", Title: "Appearance", Icon: "preferences-desktop-theme", Groups: []frostui.SettingsGroup{{
			Settings: []frostui.Setting{{Title: "Theme", Keywords: "dark light", Choice: &d.theme, Options: []string{"System", "Light", "Dark"}}},
		}}},
		{ID: "sync", Title: "Sync", Icon: "emblem-synchronizing", Keywords: "hub", Groups: []frostui.SettingsGroup{{
			Title: "Hub", Settings: []frostui.Setting{{Title: "Sync in the background", Switch: &d.sync}},
			Values: []frostui.KeyValue{{Key: "Host", Value: "hub.example.org"}, {Key: "Device", Value: "laptop"}},
		}}},
	}
}

// view is the window: a bar, the tree and the open note.
func (d *demo) view(c *ui.Context) {
	t := frostui.Adapt(c)
	d.palette.RunPending()
	d.table.BindKeys(c, frostui.PaletteBindings(), d.palette.Show)
	d.table.Bind(c)
	ui.Column(c).Fill().AlignItems(ui.Stretch).Background(t.Background).Children(func() {
		d.bar(c, t)
		ui.Box(c).Height(1).Background(t.Border)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(260).AlignItems(ui.Stretch).Background(frostui.Chrome(t)).Children(func() {
				d.list.View(c, d.sections()).Grow(1).Padding(frostui.SpaceS)
			})
			ui.Box(c).Width(1).Background(t.Border)
			d.page(c)
		})
	})
	d.info.View(c)
	d.settings.Dialog(c, "Settings", d.settingsPages())
	d.palette.View(c, d.table.Commands())
}

// bar is the top bar: the palette's button and the main menu, built from
// the table.
func (d *demo) bar(c *ui.Context, t *ui.Theme) {
	ui.Row(c).Height(frostui.TopBarHeight).Padding(0, frostui.SpaceL).Gap(frostui.SpaceM).AlignItems(ui.Center).
		Background(frostui.Chrome(t)).Children(func() {
		ui.Text(c, "Source list").Bold()
		ui.Spacer(c)
		if frostui.LabeledButton(c, "system-search", "Search or run a command", false).Clicked() {
			d.palette.Show()
		}
		ui.Spacer(c)
		ui.MenuButton(c, "", func(m *ui.Menu) {
			d.table.Menu("new", "", "switcher", "settings", "shortcuts")(m)
			m.Submenu("Note", d.table.Menu("rename", "pin"))
		}).Label("Main menu").Tooltip("Main menu").Children(func() { frostui.Icon(c, "open-menu") })
	})
}

// page draws the open note: an editor that keeps Ctrl K for itself.
func (d *demo) page(c *ui.Context) {
	ui.Column(c).Grow(1).Padding(frostui.SpaceXL).Gap(frostui.SpaceM).AlignItems(ui.Stretch).Children(func() {
		if d.current == nil {
			frostui.Empty(c, "text-x-generic", "No note is open", "Choose one on the left, or press Ctrl O.")
			return
		}
		n := d.current
		frostui.Title(c, n.title)
		frostui.Muted(c, n.folder+" · Ctrl K here inserts a link; Ctrl+Shift+P opens the palette")
		ui.TextArea(c.Key(n.id), &n.body).Label("Note").Grow(1).Bind(&d.editor).OnTextKey(func(k ui.TextKey) bool {
			if k.Mods == ui.Cmd && k.Key == ui.KeyK {
				k.SetText(k.Text() + "[link](https://example.org)")
				return true
			}
			return false
		})
		frostui.Caption(c, fmt.Sprintf("%d words", len(strings.Fields(n.body))))
	})
}
