package frostui

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// Command is one entry of the command palette: an action the app offers,
// or a place to go. A CommandTable holds the app's commands, binds their
// keys, builds menus of them and fills the palette.
type Command struct {
	// ID identifies the command; it must be unique.
	ID string
	// Title is what the palette lists and matches, as "Update everything".
	Title string
	// Group heads the commands of one kind, as "Actions" or "Go to".
	Group string
	// Icon is a freedesktop icon name.
	Icon string
	// Keys is the shortcut that also runs it, as {"Ctrl", "U"}, as the
	// palette shows it. Empty, the palette shows the first of Bindings.
	Keys []string
	// Hint is matched as well as the title, but not shown: other words for
	// the command.
	Hint string
	// Disabled commands are listed, dimmed, and do not run; Reason says
	// why.
	Disabled bool
	Reason   string
	// Run runs on the UI thread after the palette closes.
	Run func()
	// Bindings are the keys a CommandTable binds to run the command, the
	// first of which the palette, the menus and the shortcuts dialog show.
	// The palette alone binds none.
	Bindings []KeyBinding
	// Enabled, when set, says each time the command is read whether it
	// runs now, and why not: it decides Disabled and Reason (Resolve).
	Enabled func() (ok bool, reason string)
	// Checked, when set, makes the command's menu item a setting it
	// toggles, checked while Checked returns true.
	Checked func() bool
	// Unlisted commands are bound and in menus, but the palette does not
	// list them, as the command that opens the palette.
	Unlisted bool
}

// Resolve returns the command as it stands now: Disabled and Reason from
// Enabled when it is set, and Keys from the first of Bindings when Keys
// is empty.
func (cmd Command) Resolve() Command {
	if cmd.Enabled != nil {
		ok, reason := cmd.Enabled()
		cmd.Disabled, cmd.Reason = !ok, reason
		if ok {
			cmd.Reason = ""
		}
	}
	if len(cmd.Keys) == 0 && len(cmd.Bindings) > 0 {
		cmd.Keys = cmd.Bindings[0].Caps()
	}
	return cmd
}

// Match scores how well query matches text: every rune of query must
// appear in text in order, ignoring case. A higher score is a better
// match; runs of consecutive runes, the start of words and a match at the
// start count more. ok is false when query does not match.
func Match(query, text string) (score int, ok bool) {
	q := []rune(strings.ToLower(strings.TrimSpace(query)))
	if len(q) == 0 {
		return 0, true
	}
	t := []rune(strings.ToLower(text))
	qi, run := 0, 0
	for ti := 0; ti < len(t) && qi < len(q); ti++ {
		if t[ti] != q[qi] {
			run = 0
			continue
		}
		score++
		if ti == 0 {
			score += 8
		} else if !unicode.IsLetter(t[ti-1]) && !unicode.IsDigit(t[ti-1]) {
			score += 5
		}
		if run > 0 {
			score += 3 * run
		}
		run++
		qi++
	}
	if qi < len(q) {
		return 0, false
	}
	// shorter texts win among equals
	score -= len(t) / 16
	return score, true
}

// Filter returns the commands that match query, best first. With an empty
// query it returns every command in its order, grouped.
func Filter(cmds []Command, query string) []Command {
	if strings.TrimSpace(query) == "" {
		out := append([]Command(nil), cmds...)
		order := map[string]int{}
		for _, c := range cmds {
			if _, ok := order[c.Group]; !ok {
				order[c.Group] = len(order)
			}
		}
		sort.SliceStable(out, func(i, j int) bool { return order[out[i].Group] < order[out[j].Group] })
		return out
	}
	type scored struct {
		cmd   Command
		score int
	}
	var hits []scored
	for _, c := range cmds {
		best, ok := Match(query, c.Title)
		if s, hok := Match(query, c.Hint); hok && c.Hint != "" && (!ok || s-4 > best) {
			best, ok = s-4, true
		}
		if ok {
			hits = append(hits, scored{c, best})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]Command, len(hits))
	for i, h := range hits {
		out[i] = h.cmd
	}
	return out
}

// Rank returns the items whose text matches query, best first (Match),
// at most limit of them (0 for all); ties keep the items' order. With an
// empty query it returns the first limit items. It reads no UI state:
// a palette source ranks a large set on a worker with it.
func Rank[T any](items []T, query string, limit int, text func(T) string) []T {
	if strings.TrimSpace(query) == "" {
		if limit > 0 && len(items) > limit {
			return append([]T(nil), items[:limit]...)
		}
		return append([]T(nil), items...)
	}
	type scored struct {
		item  T
		score int
	}
	var hits []scored
	for _, it := range items {
		if s, ok := Match(query, text(it)); ok {
			hits = append(hits, scored{it, s})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]T, len(hits))
	for i, h := range hits {
		out[i] = h.item
	}
	return out
}

// PaletteCommandID is the ID of Palette.Command.
const PaletteCommandID = "frostui.palette"

// PaletteCommands is the ID of the palette's own source, the commands it
// is given: ShowSource(PaletteCommands) opens it on the commands alone.
const PaletteCommands = "commands"

// PaletteSource is one kind of thing the palette finds besides the
// commands: notes by name, mailboxes, messages. The palette asks a source
// lazily, only while it shows it, once for each query, and shows its
// results in sections after the commands, in the order of the sources.
type PaletteSource struct {
	// ID identifies the source, for ShowSource.
	ID string
	// Title heads the source's results whose items name no group of
	// their own, as "Notes".
	Title string
	// Placeholder is the field's text while the palette shows the source
	// alone, as "Find a note"; "Search" and the title when empty.
	Placeholder string
	// Limit is the most results the source shows while other sources
	// show too; 0 shows them all. Shown alone, it shows all it gives.
	Limit int
	// Query asks the source for the results of q.Text. The palette calls
	// it on the UI thread as it opens on the source and each time the
	// query changes. The source answers with q.Reply, at once or later
	// from a worker through the window's Update; the palette drops a
	// reply to a query the user typed past.
	Query func(q *PaletteQuery)
}

// PaletteQuery is a question the palette asks a source.
type PaletteQuery struct {
	// Text is the query, trimmed; "" as the palette opens, for the
	// source's recent items.
	Text string
	// Alone tells that the palette shows this source alone (ShowSource).
	Alone bool

	ctx    context.Context
	p      *Palette
	seq    uint64
	source string
}

// Context is canceled once the query is stale: the user typed past it,
// or the palette closed. A worker stops on it.
func (q *PaletteQuery) Context() context.Context {
	if q.ctx == nil {
		return context.Background()
	}
	return q.ctx
}

// Stale reports whether the palette no longer waits for this query's
// results. Call it on the UI thread.
func (q *PaletteQuery) Stale() bool {
	return q.p == nil || !q.p.Open || q.p.seq != q.seq
}

// Reply gives the source's results for the query, best first, in place
// of any it gave before (a source may answer in steps), and reports
// whether the palette took them: it drops those of a stale query. Call it
// on the UI thread, where the window's Update runs what a worker posts.
func (q *PaletteQuery) Reply(items []PaletteItem) bool {
	if q.Stale() {
		return false
	}
	q.p.results[q.source] = &sourceResults{items: items, done: true}
	return true
}

// PaletteItem is one result of a palette source.
type PaletteItem struct {
	// ID identifies the item among its source's.
	ID string
	// Title is what the palette lists; Detail follows it, muted, as a
	// note's folder or a message's sender.
	Title, Detail string
	// Icon is a freedesktop icon name.
	Icon string
	// Group heads the item and those after it with the same group;
	// empty, the source's Title heads it.
	Group string
	// Keys are key caps shown at the end.
	Keys []string
	// Disabled items are listed, dimmed, and do not run; Reason says why.
	Disabled bool
	Reason   string
	// Run runs on the UI thread after the palette closes, as Enter or a
	// click chooses the item.
	Run func()
	// RunWith, when set, runs in place of Run with the modifier keys held
	// as the item was chosen: Ctrl+Enter, Ctrl+Alt+Enter, Shift+Enter or
	// a click with them, as to open a note in a new tab or to create one.
	RunWith func(mods ui.Modifiers)
}

// sourceResults are what a source answered to the current query.
type sourceResults struct {
	items []PaletteItem
	done  bool
}

// Palette is the command palette: a dialog with a field that filters the
// app's commands, and the results of its Sources, which Up, Down and
// Enter choose and run.
type Palette struct {
	Open     bool
	Query    string
	Selected int
	// Sources are what the palette finds besides the commands, shown
	// after them in this order.
	Sources []PaletteSource

	list    ui.ListState
	pending func()
	// source is the one source shown alone, "" for all of them.
	source string
	// asked is the source and query last asked, fresh that nothing was
	// asked since the palette opened.
	asked   string
	fresh   bool
	seq     uint64
	cancel  context.CancelFunc
	results map[string]*sourceResults
	// selKey is the key of the row chosen, which the choice follows as
	// results come.
	selKey any
}

// Show opens the palette with an empty query, at its first command,
// selected: the list scrolls back to its top, so the command Enter runs
// is the one shown highlighted. It shows the commands and every source.
func (p *Palette) Show() { p.ShowSource("") }

// ShowSource opens the palette as Show does, on the source with the ID
// alone (PaletteCommands for the commands), or on all of them for "".
func (p *Palette) ShowSource(id string) {
	p.Open, p.Query = true, ""
	p.source, p.fresh = id, true
	p.results = map[string]*sourceResults{}
	p.drop()
	p.selectFirst()
}

// drop makes the query asked last stale: the palette no longer takes its
// results, and its context is canceled.
func (p *Palette) drop() {
	p.seq++
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}

// Command is the command that opens the palette, on PaletteBindings; the
// palette does not list it. A frame with a table adds it there; an app may
// add it first, with another Title or Group.
func (p *Palette) Command() Command {
	return Command{ID: PaletteCommandID, Title: "Command Palette", Group: "View", Icon: "system-search",
		Bindings: PaletteBindings(), Unlisted: true, Run: p.Show}
}

// ShownSource returns the ID of the source the palette shows alone, ""
// when it shows all of them.
func (p *Palette) ShownSource() string { return p.source }

// selectFirst selects the first command and scrolls the list to it, for
// a list that kept the offset of an earlier query or opening.
func (p *Palette) selectFirst() {
	p.Selected, p.selKey = 0, nil
	p.list.ScrollTo(0, ui.Start)
}

// RunPending runs the command chosen in the last frame, once the palette
// closed. The shell calls it before it builds the page, so a command that
// changes the page or asks a question shows in the same frame.
func (p *Palette) RunPending() {
	if run := p.pending; run != nil && !p.Open {
		p.pending = nil
		run()
	}
}

// paletteRow is one row the palette shows: a command or a source's item.
type paletteRow struct {
	key           any
	source        string
	heading       string
	title, detail string
	icon          string
	keys          []string
	disabled      bool
	reason        string
	run           func()
	runWith       func(ui.Modifiers)
}

// shows reports whether the palette shows the source with the ID.
func (p *Palette) shows(id string) bool { return p.source == "" || p.source == id }

// ask asks the sources shown for the results of the query, unless it
// asked them already.
func (p *Palette) ask() {
	q := strings.TrimSpace(p.Query)
	key := p.source + "\x00" + q
	if !p.fresh && key == p.asked {
		return
	}
	p.fresh, p.asked = false, key
	p.drop()
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	if p.results == nil {
		p.results = map[string]*sourceResults{}
	}
	for _, src := range p.Sources {
		if !p.shows(src.ID) || src.Query == nil {
			continue
		}
		// The results of the query before stay while the source works,
		// so the list does not flicker as the user types.
		if r := p.results[src.ID]; r != nil {
			r.done = false
		} else {
			p.results[src.ID] = &sourceResults{}
		}
		src.Query(&PaletteQuery{Text: q, Alone: p.source != "", ctx: ctx, p: p, seq: p.seq, source: src.ID})
	}
}

// rows lists what the palette shows for the query: the commands, then
// each source's results; and whether a source still works on it.
func (p *Palette) rows(cmds []Command) (rows []paletteRow, waiting bool) {
	if p.shows(PaletteCommands) {
		var listed []Command
		for _, cmd := range cmds {
			if !cmd.Unlisted {
				listed = append(listed, cmd.Resolve())
			}
		}
		for _, cmd := range Filter(listed, p.Query) {
			rows = append(rows, paletteRow{key: [2]string{PaletteCommands, cmd.ID}, source: PaletteCommands,
				heading: cmd.Group, title: cmd.Title, icon: cmd.Icon, keys: cmd.Keys,
				disabled: cmd.Disabled, reason: cmd.Reason, run: cmd.Run})
		}
	}
	for _, src := range p.Sources {
		if !p.shows(src.ID) {
			continue
		}
		r := p.results[src.ID]
		if r == nil {
			continue
		}
		if !r.done {
			waiting = true
		}
		items := r.items
		if p.source == "" && src.Limit > 0 && len(items) > src.Limit {
			items = items[:src.Limit]
		}
		for _, it := range items {
			heading := it.Group
			if heading == "" {
				heading = src.Title
			}
			rows = append(rows, paletteRow{key: [2]string{src.ID, it.ID}, source: src.ID, heading: heading,
				title: it.Title, detail: it.Detail, icon: it.Icon, keys: it.Keys,
				disabled: it.Disabled, reason: it.Reason, run: it.Run, runWith: it.RunWith})
		}
	}
	return rows, waiting
}

// placeholder is the field's text: for the source shown alone, or for
// all.
func (p *Palette) placeholder() (label, text string) {
	switch p.source {
	case "":
		return "Command", "Type a command or search"
	case PaletteCommands:
		return "Command", "Type a command"
	}
	for _, src := range p.Sources {
		if src.ID == p.source {
			text := src.Placeholder
			if text == "" {
				text = "Search " + src.Title
			}
			return src.Title, text
		}
	}
	return "Command", "Type a command or search"
}

// View draws the palette while it is open, and runs a chosen command
// after it closes (RunPending, when the shell has not already).
func (p *Palette) View(c *ui.Context, cmds []Command) {
	p.RunPending()
	if !p.Open {
		if p.cancel != nil {
			p.drop()
			p.asked = ""
		}
		return
	}
	t := c.Theme()
	p.ask()
	shown, waiting := p.rows(cmds)
	if p.selKey != nil {
		for i, r := range shown {
			if r.key == p.selKey {
				p.Selected = i
				break
			}
		}
	}
	if p.Selected >= len(shown) {
		p.Selected = len(shown) - 1
	}
	if p.Selected < 0 {
		p.Selected = 0
	}
	if p.Selected < len(shown) {
		p.selKey = shown[p.Selected].key
	}
	choose := func(i int, mods ui.Modifiers) {
		if i < 0 || i >= len(shown) || shown[i].disabled || shown[i].run == nil && shown[i].runWith == nil {
			return
		}
		r := shown[i]
		if r.runWith != nil {
			p.pending = func() { r.runWith(mods) }
		} else {
			p.pending = r.run
		}
		p.Open = false
	}
	move := func(i int) {
		p.Selected = i
		p.selKey = shown[i].key
		p.list.ScrollIntoView(i)
	}
	label, placeholder := p.placeholder()
	ui.Modal(c, &p.Open, func() {
		ui.Column(c).Width(620).Height(440).Gap(0).AlignItems(ui.Stretch).Children(func() {
			field := ui.SearchField(c, &p.Query).Label(label).Placeholder(placeholder).AutoFocus()
			if field.Changed() {
				p.selectFirst()
			}
			field.OnShortcut(0, ui.KeyDown, func() {
				if p.Selected < len(shown)-1 {
					move(p.Selected + 1)
				}
			})
			field.OnShortcut(0, ui.KeyUp, func() {
				if p.Selected > 0 {
					move(p.Selected - 1)
				}
			})
			for _, mods := range []ui.Modifiers{ui.Cmd, ui.Cmd | ui.Alt, ui.Alt} {
				field.OnShortcut(mods, ui.KeyEnter, func() { choose(p.Selected, mods) })
			}
			if field.Submitted() {
				choose(p.Selected, c.Modifiers()&ui.Shift)
			}
			ui.Box(c).Height(SpaceS)
			if len(shown) == 0 {
				switch {
				case waiting:
					Loading(c, "Searching…")
				case len(p.Sources) == 0 || p.source == PaletteCommands:
					Empty(c, "edit-find", "No command matches", "")
				default:
					Empty(c, "edit-find", "Nothing matches", "")
				}
				return
			}
			p.list.Key = func(i int) any { return shown[i].key }
			ui.List(c, &p.list, len(shown), func(i int) {
				r := shown[i]
				header := i == 0 || shown[i-1].heading != r.heading || shown[i-1].source != r.source
				ui.Column(c).AlignItems(ui.Stretch).Children(func() {
					if header && r.heading != "" {
						Label(c, r.heading).Padding(8, 10, 4, 10)
					}
					row := ui.ButtonBase(c).Row().Gap(SpaceM).Padding(8, 10).AlignItems(ui.Center).
						Cursor(ui.CursorPointer)
					if i == p.Selected {
						selectedRow(row, t)
					} else if row.Hovered() {
						row.TextColor(ColorsOf(t).HoverText)
					}
					if r.disabled {
						row.Opacity(0.5)
					}
					if row.Clicked() {
						p.Selected, p.selKey = i, r.key
						choose(i, row.ClickModifiers())
					}
					row.Children(func() {
						Icon(c, r.icon).TextColor(t.TextMuted)
						if r.detail == "" {
							ui.Text(c, r.title).FontSize(TextSize(c, 13.5)).Grow(1).SingleLine()
						} else {
							ui.Text(c, r.title).FontSize(TextSize(c, 13.5)).SingleLine()
							Muted(c, r.detail).Grow(1).MinWidth(0).SingleLine()
						}
						if r.disabled && r.reason != "" {
							Caption(c, r.reason)
						}
						if len(r.keys) > 0 {
							Shortcut(c, r.keys...)
						}
					})
				})
			}).Grow(1)
			ui.Row(c).Gap(SpaceM).Padding(10, 6, 0, 6).AlignItems(ui.Center).Children(func() {
				Shortcut(c, "↑", "↓")
				Caption(c, "Navigate")
				Shortcut(c, "Enter")
				Caption(c, "Run")
				Shortcut(c, "Esc")
				Caption(c, "Close")
				if waiting {
					ui.Spacer(c)
					ui.Spinner(c).Label("Searching")
					Caption(c, "Searching…")
				}
			})
		})
	})
}
