package frostui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// SourceSection is a titled group of a SourceList's items, as Favorites
// or an account's mailboxes.
type SourceSection[K comparable] struct {
	// ID identifies the section among the list's.
	ID string
	// Title heads the section; a click on it hides and shows its items.
	// Empty, the section has no header and its items always show.
	Title string
	// Items are the section's items at the top of its tree. One item may
	// show in several sections, as a favorite mailbox does.
	Items []K
}

// SourceRow is how a SourceList shows an item.
type SourceRow struct {
	// Title is the item's name, which assistive technology reads and
	// typing its first letters chooses; Icon is a freedesktop icon name.
	Title, Icon string
	// Count is the badge at the row's end, as an unread count, "" for
	// none; CountTone colors it.
	Count     string
	CountTone Tone
	// Tag is a small label after the title, as a file's extension.
	Tag string
	// Busy shows a spinner at the row's end, as while a mailbox syncs.
	Busy bool
	// Alert, when set, shows a warning sign at the row's end with the
	// text as its tooltip and name, as why a mailbox fails to sync.
	Alert string
	// Dim draws the row muted, as an item not loaded or not supported.
	Dim bool
	// Spoken, when set, is what assistive technology reads after the
	// title, as "12 unread, syncing"; empty, it reads the tag, the count,
	// "busy" and the alert.
	Spoken string
}

// accessName is the name assistive technology reads for the row: a tree
// item is one name, which holds no elements of its own.
func (r SourceRow) accessName() string {
	parts := []string{r.Title}
	switch {
	case r.Spoken != "":
		parts = append(parts, r.Spoken)
	default:
		if r.Tag != "" {
			parts = append(parts, r.Tag)
		}
		if r.Count != "" {
			parts = append(parts, r.Count)
		}
		if r.Busy {
			parts = append(parts, "busy")
		}
		if r.Alert != "" {
			parts = append(parts, r.Alert)
		}
	}
	return strings.Join(parts, ", ")
}

// SourceList is a sidebar of items in a tree, under section headers, as
// a mail app's mailboxes or a notes app's files. It draws on MyGo's
// Outline, which builds only the rows in view, so a tree of thousands of
// items is as fast as a small one. A click or the arrows choose an item;
// Left and Right close and open it; a click on a section's header hides
// and shows its items. Rows show an icon, the title, a tag, a spinner or
// an alert, and a count. Items take context menus (Menu), values dropped
// on them (DropOn) and new names typed in place (Rename). Assistive
// technology sees a tree of items with their levels, open or closed.
//
// Keep one SourceList per window in the app's state, and call View from
// the view every frame. Its fields are read as View runs.
type SourceList[K comparable] struct {
	// Label names the tree for assistive technology, as "Mailboxes".
	Label string
	// Children returns an item's children, nil for an item without any
	// (a leaf); an empty, non-nil slice for a branch that holds nothing
	// now. View asks it for every item open, every frame: keep it cheap.
	Children func(item K) []K
	// Row returns how an item shows.
	Row func(item K) SourceRow
	// Menu, when set, builds an item's context menu.
	Menu func(m *ui.Menu, item K)
	// OnSelect runs as the user chooses an item, by a click or the keys.
	OnSelect func(item K)
	// OnActivate runs on Enter or a double click on an item, with the
	// modifier keys held: Ctrl+Enter opens a note in a new tab.
	OnActivate func(item K, mods ui.Modifiers)
	// Rename, when set, lets the user rename items in place: F2 on the
	// item chosen, or StartRename, shows a field of its title; Enter, or
	// the focus leaving it, keeps the name, and Escape gives up. Rename
	// gets the new name, unless it is empty or the title.
	Rename func(item K, name string)
	// CanRename, with Rename, reports whether an item can be renamed; nil
	// lets every item be.
	CanRename func(item K) bool
	// DragItems, when set, lets the user drag items out of the list,
	// onto an item of it (DropOn) or elsewhere: it returns the value they
	// drag, as ui.Drop takes it, nil for items that do not drag.
	DragItems func(items []K) any

	state    ui.OutlineState[sourceKey[K]]
	row      int
	sel      sourceKey[K]
	hasSel   bool
	want     *sourceKey[K]
	wantAny  bool
	seen     map[string]bool
	sections []SourceSection[K]
	cur      ui.OutlineRow
	drops    []func(c *ui.Context, row ui.Element, item K) bool

	renaming   bool
	renameKey  sourceKey[K]
	renameFrom string
	draft      string
}

// sourceKey is a row of a SourceList: a section's header, or an item in a
// section, which tells apart an item shown in two sections.
type sourceKey[K comparable] struct {
	section string
	item    K
	header  bool
}

// DropOn makes the items of l take values of type T dropped on them, as
// messages on a mailbox or notes on a folder: accept reports whether item
// takes v, and the row shows that it does while v is over it; drop gets
// what is dropped there, with the modifier keys held (Ctrl or Alt to copy
// rather than move). Call it once, as the list is set up; call it again
// for other types.
func DropOn[K comparable, T any](l *SourceList[K], accept func(item K, v T) bool, drop func(item K, v T, mods ui.Modifiers)) {
	l.drops = append(l.drops, func(c *ui.Context, row ui.Element, item K) bool {
		v, over := ui.DragOver[T](row)
		if over && accept != nil && !accept(item, v) {
			over = false
		}
		if v, ok := ui.Drop[T](row); ok && (accept == nil || accept(item, v)) && drop != nil {
			drop(item, v, c.Modifiers())
		}
		return over
	})
}

// Selected returns the item chosen and the section it shows in; ok is
// false when none is.
func (s *SourceList[K]) Selected() (item K, section string, ok bool) {
	return s.sel.item, s.sel.section, s.hasSel
}

// Select chooses item in section, or in the first section that holds it
// for "": View opens the items above it and scrolls to it. OnSelect does
// not run.
func (s *SourceList[K]) Select(section string, item K) {
	s.want = &sourceKey[K]{section: section, item: item}
	s.wantAny = section == ""
}

// ClearSelection chooses no item.
func (s *SourceList[K]) ClearSelection() {
	s.row, s.hasSel, s.want = -1, false, nil
	s.sel = sourceKey[K]{}
}

// SetOpen opens or closes item in section, showing or hiding its
// children.
func (s *SourceList[K]) SetOpen(section string, item K, open bool) {
	k := sourceKey[K]{section: section, item: item}
	if open {
		s.state.Open.Add(k)
	} else {
		s.state.Open.Remove(k)
	}
}

// IsOpen reports whether item in section shows its children.
func (s *SourceList[K]) IsOpen(section string, item K) bool {
	return s.state.Open.Has(sourceKey[K]{section: section, item: item})
}

// OpenItems returns the items of section that show their children, to
// keep them open the next time the app starts (SetOpen).
func (s *SourceList[K]) OpenItems(section string) []K {
	var out []K
	for k := range s.state.Open.All() {
		if !k.header && k.section == section {
			out = append(out, k.item)
		}
	}
	return out
}

// SetSectionOpen shows or hides the items of the section with the ID, as
// a click on its header does.
func (s *SourceList[K]) SetSectionOpen(id string, open bool) {
	if s.seen == nil {
		s.seen = map[string]bool{}
	}
	s.seen[id] = true
	h := sourceKey[K]{section: id, header: true}
	if open {
		s.state.Open.Add(h)
	} else {
		s.state.Open.Remove(h)
	}
}

// SectionOpen reports whether the section with the ID shows its items. A
// section shows them until the user or SetSectionOpen hides them.
func (s *SourceList[K]) SectionOpen(id string) bool {
	return !s.seen[id] || s.state.Open.Has(sourceKey[K]{section: id, header: true})
}

// CollapseAll closes every item, leaving the sections as they are.
func (s *SourceList[K]) CollapseAll() {
	var headers []sourceKey[K]
	for k := range s.state.Open.All() {
		if k.header {
			headers = append(headers, k)
		}
	}
	s.state.Open.Clear()
	for _, h := range headers {
		s.state.Open.Add(h)
	}
}

// ExpandAll opens every item of the sections View last drew that has
// children, walking the whole tree.
func (s *SourceList[K]) ExpandAll() {
	if s.Children == nil {
		return
	}
	var walk func(section string, items []K)
	walk = func(section string, items []K) {
		for _, it := range items {
			if kids := s.Children(it); kids != nil {
				s.state.Open.Add(sourceKey[K]{section: section, item: it})
				walk(section, kids)
			}
		}
	}
	for _, sec := range s.sections {
		walk(sec.ID, sec.Items)
	}
}

// StartRename chooses item in section and shows a field of its title in
// its row, when Rename is set and CanRename lets it.
func (s *SourceList[K]) StartRename(section string, item K) {
	if s.Rename == nil || s.CanRename != nil && !s.CanRename(item) {
		return
	}
	s.Select(section, item)
	s.renaming, s.renameKey = true, sourceKey[K]{section: section, item: item}
	s.renameFrom = s.title(s.renameKey)
	s.draft = s.renameFrom
}

// Renaming reports whether an item's name is being typed.
func (s *SourceList[K]) Renaming() bool { return s.renaming }

// title is the text of a row: a section's title, or an item's.
func (s *SourceList[K]) title(k sourceKey[K]) string {
	if k.header {
		for _, sec := range s.sections {
			if sec.ID == k.section {
				return sec.Title
			}
		}
		return ""
	}
	if s.Row == nil {
		return ""
	}
	return s.Row(k.item).Title
}

// titled reports whether the section with the ID has a header.
func (s *SourceList[K]) titled(id string) bool {
	for _, sec := range s.sections {
		if sec.ID == id {
			return sec.Title != ""
		}
	}
	return false
}

// children returns the rows under a row: a section's items, or an item's
// children in its section.
func (s *SourceList[K]) children(k sourceKey[K]) []sourceKey[K] {
	var items []K
	if k.header {
		for _, sec := range s.sections {
			if sec.ID == k.section {
				items = sec.Items
				break
			}
		}
		if items == nil {
			items = []K{}
		}
	} else if s.Children != nil {
		items = s.Children(k.item)
	}
	if items == nil {
		return nil
	}
	out := make([]sourceKey[K], len(items))
	for i, it := range items {
		out[i] = sourceKey[K]{section: k.section, item: it}
	}
	return out
}

// roots are the rows at the top: the sections' headers, or the items of
// sections without one.
func (s *SourceList[K]) roots() []sourceKey[K] {
	var roots []sourceKey[K]
	for _, sec := range s.sections {
		if sec.Title == "" {
			for _, it := range sec.Items {
				roots = append(roots, sourceKey[K]{section: sec.ID, item: it})
			}
			continue
		}
		h := sourceKey[K]{section: sec.ID, header: true}
		if !s.seen[sec.ID] {
			s.seen[sec.ID] = true
			s.state.Open.Add(h)
		}
		roots = append(roots, h)
	}
	return roots
}

// reveal opens the rows above the item asked for by Select, and returns
// the row it will show in, -1 when no section holds it.
func (s *SourceList[K]) reveal(roots []sourceKey[K]) (sourceKey[K], int) {
	want := *s.want
	// The path from a section's items down to the item.
	var path []sourceKey[K]
	var find func(section string, items []K) bool
	find = func(section string, items []K) bool {
		for _, it := range items {
			k := sourceKey[K]{section: section, item: it}
			if it == want.item {
				path = append(path, k)
				return true
			}
			if s.Children == nil {
				continue
			}
			if kids := s.Children(it); kids != nil {
				path = append(path, k)
				if find(section, kids) {
					return true
				}
				path = path[:len(path)-1]
			}
		}
		return false
	}
	found := false
	for _, sec := range s.sections {
		if (s.wantAny || sec.ID == want.section) && find(sec.ID, sec.Items) {
			found = true
			if sec.Title != "" {
				s.state.Open.Add(sourceKey[K]{section: sec.ID, header: true})
			}
			break
		}
	}
	if !found {
		return want, -1
	}
	for _, k := range path[:len(path)-1] {
		s.state.Open.Add(k)
	}
	target := path[len(path)-1]
	return target, s.rowOf(roots, target)
}

// rowOf returns the row that shows target as Outline will lay the rows
// out, -1 when none does.
func (s *SourceList[K]) rowOf(roots []sourceKey[K], target sourceKey[K]) int {
	row, at := 0, -1
	var walk func(keys []sourceKey[K]) bool
	walk = func(keys []sourceKey[K]) bool {
		for _, k := range keys {
			if k == target {
				at = row
				return true
			}
			row++
			if s.state.Open.Has(k) {
				if kids := s.children(k); kids != nil && walk(kids) {
					return true
				}
			}
		}
		return false
	}
	walk(roots)
	return at
}

// restore finds the row of an item whose choice lost its row: while its
// section is closed it stays chosen, and shows again as the section
// opens; an item no longer shown otherwise is no longer chosen.
func (s *SourceList[K]) restore(roots []sourceKey[K]) {
	st := &s.state
	if !s.hasSel || s.row >= 0 && s.row < st.Rows() && !st.Item(s.row).header || !s.SectionOpen(s.sel.section) {
		return
	}
	if s.row = s.rowOf(roots, s.sel); s.row < 0 {
		s.hasSel, s.sel = false, sourceKey[K]{}
	}
}

// View draws the list of the sections' trees, filling the height its
// parent gives it (Grow it in a column), and returns the outline.
func (s *SourceList[K]) View(c *ui.Context, sections []SourceSection[K]) ui.Element {
	t := c.Theme()
	s.sections = sections
	if s.seen == nil {
		s.seen = map[string]bool{}
		s.row = -1
	}
	roots := s.roots()
	st := &s.state
	if s.want != nil {
		if k, row := s.reveal(roots); row >= 0 {
			s.row, s.sel, s.hasSel = row, k, true
			st.List.ScrollIntoView(row)
		}
		s.want = nil
	} else {
		s.restore(roots)
	}
	st.List.Selected = &s.row
	st.List.Header = func(i int) bool { return st.Item(i).header }
	st.List.Label = func(i int) string {
		// A header is named by its title, inside the part that takes the
		// click; an item's name also tells its count and state.
		k := st.Item(i)
		if k.header || s.Row == nil {
			return ""
		}
		return s.Row(k.item).accessName()
	}
	st.List.DragRows = nil
	if s.DragItems != nil {
		st.List.DragRows = func(rows []int) any {
			var items []K
			for _, i := range rows {
				if k := st.Item(i); !k.header {
					items = append(items, k.item)
				}
			}
			if len(items) == 0 {
				return nil
			}
			return s.DragItems(items)
		}
	}
	st.Prefix = func(r ui.OutlineRow) { s.cur = r }
	out := ui.Outline(c, st, roots, s.children, func(k sourceKey[K]) {
		if k.header {
			s.header(c, t, k)
		} else {
			s.item(c, t, k)
		}
	}).Label(s.Label)
	// The choice, as the user moved it.
	if s.row >= 0 && s.row < st.Rows() && !st.Item(s.row).header {
		if k := st.Item(s.row); !s.hasSel || k != s.sel {
			s.sel, s.hasSel = k, true
			if s.OnSelect != nil {
				s.OnSelect(k.item)
			}
		}
	}
	if s.Rename != nil {
		out.OnShortcut(0, ui.KeyF2, func() {
			if s.hasSel {
				s.StartRename(s.sel.section, s.sel.item)
			}
		})
	}
	if s.OnActivate != nil {
		if out.Submitted() && s.hasSel && !s.renaming {
			s.OnActivate(s.sel.item, c.Modifiers())
		}
		for _, mods := range []ui.Modifiers{ui.Cmd, ui.Cmd | ui.Alt} {
			out.OnShortcut(mods, ui.KeyEnter, func() {
				if s.hasSel {
					s.OnActivate(s.sel.item, mods)
				}
			})
		}
	}
	return out
}

// header draws a section's header: its title, which hides and shows the
// section's items, with an arrow that shows as the pointer rests on it.
func (s *SourceList[K]) header(c *ui.Context, t *ui.Theme, k sourceKey[K]) {
	r := s.cur
	head := r.Arrow().Grow(1).Height(28).Row().Gap(SpaceS).Padding(0, SpaceS).AlignItems(ui.Center).
		Cursor(ui.CursorPointer).Background(Chrome(t))
	head.Children(func() {
		ui.Text(c, s.title(k)).FontSize(t.FontSize - 1).FontWeight(600).TextColor(t.TextMuted).Grow(1).SingleLine()
		chevron := "go-next"
		if r.Open {
			chevron = "pan-down"
		}
		icon := Icon(c, chevron).FontSize(11).TextColor(t.TextMuted)
		if !head.Hovered() {
			icon.Opacity(0)
		}
	})
}

// item draws an item's row: its indentation and arrow, icon, title (or
// the field renaming it), tag, spinner or alert, and count.
func (s *SourceList[K]) item(c *ui.Context, t *ui.Theme, k sourceKey[K]) {
	r := s.cur
	var row SourceRow
	if s.Row != nil {
		row = s.Row(k.item)
	}
	st := &s.state
	chosen := s.row >= 0 && s.row < st.Rows() && st.Item(s.row) == k
	depth := r.Depth
	if s.titled(k.section) {
		depth--
	}
	content := ui.Row(c).Grow(1).Height(28).Gap(6).Padding(0, SpaceS, 0, SpaceXS).AlignItems(ui.Center)
	if s.Menu != nil {
		content.ContextMenu(func(m *ui.Menu) { s.Menu(m, k.item) })
	}
	over := false
	for _, drop := range s.drops {
		if drop(c, content, k.item) {
			over = true
		}
	}
	if over {
		content.Radius(t.Radius).Background(t.Accent.Alpha(0.22)).Border(1, t.Accent)
	}
	if row.Dim {
		content.Opacity(0.55)
	}
	muted := t.TextMuted
	if chosen {
		muted = t.AccentText
	}
	content.Children(func() {
		if depth > 0 {
			ui.Box(c).Width(16 * float32(depth)).Shrink(0)
		}
		arrow := r.Arrow().Size(16, 16).Center()
		if r.Branch {
			arrow.Cursor(ui.CursorPointer).Children(func() {
				chevron := "go-next"
				if r.Open {
					chevron = "pan-down"
				}
				Icon(c, chevron).FontSize(11).TextColor(muted)
			})
		}
		icon := Icon(c, row.Icon)
		if !chosen {
			icon.TextColor(t.Accent)
		}
		if s.renaming && s.renameKey == k {
			s.renameField(c)
		} else {
			ui.Text(c, row.Title).Grow(1).MinWidth(0).SingleLine()
		}
		if row.Tag != "" {
			ui.Box(c).Padding(0, 4).Radius(4).Border(1, muted).Shrink(0).Children(func() {
				ui.Text(c, row.Tag).FontSize(10).TextColor(muted).SingleLine()
			})
		}
		if row.Busy {
			ui.Spinner(c).Size(14, 14).Label("Busy")
		}
		if row.Alert != "" {
			alert := Icon(c, "dialog-warning").Label(row.Alert).Tooltip(row.Alert)
			if !chosen {
				alert.TextColor(t.Warning)
			}
		}
		if row.Count != "" {
			if chosen {
				ui.Box(c).Padding(1, 7).Radius(999).Background(t.AccentText.Alpha(0.25)).Shrink(0).Children(func() {
					ui.Text(c, row.Count).FontSize(11.5).Bold().SingleLine()
				})
			} else {
				Pill(c, row.Count, row.CountTone)
			}
		}
	})
}

// renameField draws the field of the name being typed.
func (s *SourceList[K]) renameField(c *ui.Context) {
	in := ui.TextInput(c, &s.draft).Label("Name").AutoFocus().Grow(1).MinWidth(0).Height(24)
	end := func(keep bool) {
		if !s.renaming {
			return
		}
		s.renaming = false
		if name := s.draft; keep && name != "" && name != s.renameFrom {
			s.Rename(s.renameKey.item, name)
		}
	}
	in.OnShortcut(0, ui.KeyEscape, func() {
		end(false)
		s.state.List.Focus(c)
	})
	switch {
	case in.Submitted():
		end(true)
		s.state.List.Focus(c)
	case in.FocusLeft():
		end(true)
	}
}
