package frostui

import "github.com/egoist/mygo/ui"

// The sizes of panes and of the separators between them, in DIPs.
const (
	paneDefaultWidth = 240
	paneMinWidth     = 40
	paneStep         = 16
	// gripThickness is how wide a separator takes the pointer. It draws a
	// line 1 DIP thick and takes only that much room: the rest lies over
	// the panes on either side.
	gripThickness = 7
)

// Pane is one pane of a Panes split.
type Pane struct {
	// ID names the pane in PaneLayout and in Panes' methods. It must be
	// unique among the panes.
	ID string
	// Label names the pane and its separator for assistive technology,
	// which sees the pane as a region. Empty leaves the pane unnamed and
	// calls the separator "Divider".
	Label string
	// Width is the pane's width until the user or a restored PaneLayout
	// gives it another. The first pane whose Width is 0 takes the room
	// left, and when no pane does, the last one takes it. Another pane
	// without a width starts 240 DIPs wide.
	Width float32
	// Min and Max bound the pane's width. A Min of 0 is 40 DIPs; a Max of 0
	// is no limit. The pane that takes the room left uses only Min.
	Min, Max float32
	// Collapsible lets the pane hide, through Panes.Toggle and
	// Panes.SetCollapsed; a window too narrow for every pane's Min hides
	// such panes too, the last first. The pane that takes the room left
	// never hides. A hidden pane is not built, so keep what it must remember
	// in the app's state.
	Collapsible bool
	// Bar draws the pane's segment of a PaneShell's top bar, which is as
	// wide as the pane; nil leaves the segment empty. Panes.View does not
	// draw it.
	Bar func()
	// Content draws the pane, which is full height and does not scroll.
	Content func()
}

// PaneLayout is what the user changed in a Panes split: the width given to
// each pane and the panes collapsed, by ID. An app saves it and restores it
// to bring the layout back; it holds plain values for encoding/json, and
// the kit stores nothing.
type PaneLayout struct {
	Widths    map[string]float32 `json:"widths,omitempty"`
	Collapsed map[string]bool    `json:"collapsed,omitempty"`
}

// Panes lays out panes side by side, with a separator between each two that
// the user drags, or moves with the arrow keys, Home and End once it has the
// focus. A double click puts its pane back at its Width. A separator
// resizes the pane on its far side from the pane that takes the room left.
// Keep one Panes in the app's state and call View every frame.
type Panes struct {
	// Layout is the state an app saves and restores.
	Layout PaneLayout
	// Step is how far an arrow key moves a separator, in DIPs; 0 is 16.
	Step float32

	hidden map[string]bool // the panes the last frame did not show
	prefer string          // the pane the app showed last, hidden for room last
	total  float32         // the width of the last layout
	window float32         // the window's width in the last build
}

// View draws panes and returns the row that holds them; give it a size,
// as Fill or Grow.
func (p *Panes) View(c *ui.Context, panes []Pane) ui.Element {
	return p.view(c, panes, nil)
}

// Shown reports whether the pane shows: it is not collapsed, and the last
// frame had room for it.
func (p *Panes) Shown(id string) bool {
	if p.hidden != nil {
		return !p.hidden[id]
	}
	return !p.Layout.Collapsed[id]
}

// SetCollapsed hides or shows a Collapsible pane. A pane shown this way is
// the last to hide for want of room.
func (p *Panes) SetCollapsed(id string, collapsed bool) {
	if p.hidden == nil {
		p.hidden = map[string]bool{}
	}
	if collapsed {
		if p.Layout.Collapsed == nil {
			p.Layout.Collapsed = map[string]bool{}
		}
		p.Layout.Collapsed[id] = true
		p.hidden[id] = true
		if p.prefer == id {
			p.prefer = ""
		}
		return
	}
	delete(p.Layout.Collapsed, id)
	delete(p.hidden, id)
	p.prefer = id
}

// Toggle hides a pane that shows and shows one that does not.
func (p *Panes) Toggle(id string) { p.SetCollapsed(id, p.Shown(id)) }

// flexPane is the index of the pane that takes the room left.
func flexPane(panes []Pane) int {
	for i, pn := range panes {
		if pn.Width == 0 {
			return i
		}
	}
	return len(panes) - 1
}

func paneMin(pn Pane) float32 {
	if pn.Min > 0 {
		return pn.Min
	}
	return paneMinWidth
}

// paneWanted is the width the pane has when there is room for it.
func (p *Panes) paneWanted(pn Pane) float32 {
	w := p.Layout.Widths[pn.ID]
	if w <= 0 {
		w = pn.Width
	}
	if w <= 0 {
		w = paneDefaultWidth
	}
	if pn.Max > 0 {
		w = min(w, pn.Max)
	}
	return max(w, paneMin(pn))
}

// arrange decides which panes show and how wide each is, in total DIPs, or
// with total 0 as if there were room for all.
func (p *Panes) arrange(panes []Pane, flex int, total float32) (widths []float32, shown []bool) {
	shown = make([]bool, len(panes))
	for i, pn := range panes {
		shown[i] = i == flex || !pn.Collapsible || !p.Layout.Collapsed[pn.ID]
	}
	if total > 0 {
		p.squeeze(panes, flex, total, shown)
	}
	widths = make([]float32, len(panes))
	var fixed, slack, lines float32
	for i, pn := range panes {
		if !shown[i] {
			continue
		}
		lines++
		if i == flex {
			continue
		}
		widths[i] = p.paneWanted(pn)
		fixed += widths[i]
		slack += widths[i] - paneMin(pn)
	}
	lines--
	if total <= 0 {
		return widths, shown
	}
	// Too narrow for the widths wanted: the panes give way in proportion
	// to what they have above their minimums.
	if excess := fixed + lines + paneMin(panes[flex]) - total; excess > 0 && slack > 0 {
		k := min(excess/slack, 1)
		fixed = 0
		for i, pn := range panes {
			if shown[i] && i != flex {
				widths[i] -= (widths[i] - paneMin(pn)) * k
				fixed += widths[i]
			}
		}
	}
	widths[flex] = max(total-fixed-lines, paneMin(panes[flex]))
	return widths, shown
}

// squeeze hides collapsible panes, the last first and the one the app
// showed last after them, until the minimums of those left fit in total.
func (p *Panes) squeeze(panes []Pane, flex int, total float32, shown []bool) {
	fits := func() bool {
		var need, lines float32
		for i, pn := range panes {
			if shown[i] {
				need += paneMin(pn)
				lines++
			}
		}
		return need+lines-1 <= total
	}
	hide := func(preferred bool) {
		for i := len(panes) - 1; i >= 0 && !fits(); i-- {
			if shown[i] && i != flex && panes[i].Collapsible && (panes[i].ID == p.prefer) == preferred {
				shown[i] = false
			}
		}
	}
	hide(false)
	hide(true)
}

// view draws the panes; bar, when set, draws each one's top-bar segment,
// told whether it is the first or the last shown. The segments are in the
// panes' columns, so they are always as wide as the panes under them.
func (p *Panes) view(c *ui.Context, panes []Pane, bar func(pn Pane, first, last bool)) ui.Element {
	row := ui.Row(c).AlignItems(ui.Stretch).Clip()
	if len(panes) == 0 {
		return row
	}
	// The room is what the last frame laid out: when it changed, or the
	// window did, one more frame makes the widths match the new room.
	total := row.LayoutBounds().W
	window, _ := c.Size()
	if total != p.total || window != p.window {
		p.total, p.window = total, window
		c.AnimationFrame()
	}
	flex := flexPane(panes)
	widths, shown := p.arrange(panes, flex, total)
	p.hidden = map[string]bool{}
	first, last := -1, -1
	for i, pn := range panes {
		if !shown[i] {
			p.hidden[pn.ID] = true
			continue
		}
		if first < 0 {
			first = i
		}
		last = i
	}
	row.Children(func() {
		prev := -1
		for i, pn := range panes {
			if !shown[i] {
				continue
			}
			if prev >= 0 {
				p.separator(c, panes, widths, flex, total, prev, i)
			}
			prev = i
			col := ui.Column(c.Key(pn.ID)).AlignItems(ui.Stretch).Clip()
			if i == flex {
				col.Grow(1).Shrink(1).Basis(0).MinWidth(paneMin(pn))
			} else {
				col.Width(widths[i]).Shrink(1).MinWidth(paneMin(pn))
			}
			col.Children(func() {
				if bar != nil {
					bar(pn, i == first, i == last)
				}
				body := ui.Column(c).Grow(1).MinHeight(0).AlignItems(ui.Stretch).Clip()
				if pn.Label != "" {
					body.Role(ui.RoleRegion).Label(pn.Label)
				}
				if pn.Content != nil {
					body.Children(pn.Content)
				}
			})
		}
	})
	return row
}

// separator draws the line between the shown panes left and right, which
// resizes the one on its far side from the pane that takes the room left.
func (p *Panes) separator(c *ui.Context, panes []Pane, widths []float32, flex int, total float32, left, right int) {
	target, sign := right, float32(-1)
	if left < flex {
		target, sign = left, 1
	}
	if c.Direction() == ui.RightToLeft {
		sign = -sign
	}
	pn := panes[target]
	lo, now := paneMin(pn), widths[target]
	hi := now
	if total > 0 {
		hi = now + max(widths[flex]-paneMin(panes[flex]), 0)
	}
	if pn.Max > 0 {
		hi = min(hi, pn.Max)
	}
	hi = max(hi, now)
	label := pn.Label
	if label == "" {
		label = "Divider"
	}
	step := p.Step
	if step <= 0 {
		step = paneStep
	}
	id := pn.ID
	drawGrip(c.Key([2]any{"separator", id}), grip{
		vertical: true, label: label, lo: lo, hi: hi, now: now, step: step, sign: sign,
		set: func(v float32) {
			if p.Layout.Widths == nil {
				p.Layout.Widths = map[string]float32{}
			}
			p.Layout.Widths[id] = v
		},
		reset: func() { delete(p.Layout.Widths, id) },
	})
}

// grip is a line between two areas that the user moves to resize one: the
// value is that area's size.
type grip struct {
	// vertical is a line running up and down, between areas side by side.
	vertical bool
	label    string
	// lo, hi and now are the value's range and the value as shown; step
	// is how far an arrow key moves the line.
	lo, hi, now, step float32
	// sign is +1 when moving the line right or down grows the value, -1
	// when it shrinks it.
	sign float32
	set  func(v float32)
	// reset puts the value back at its default, on a double click.
	reset func()
}

// drawGrip draws g, a separator for assistive technology, with its value.
// It takes one DIP of room and the pointer a few DIPs either side, above
// the areas beside it.
func drawGrip(c *ui.Context, g grip) ui.Element {
	t := c.Theme()
	const overhang = (gripThickness - 1) / 2
	e := ui.Box(c).Shrink(0).Focusable().Role(ui.RoleSeparator).Label(g.label).
		Range(float64(g.lo), float64(g.hi), float64(g.now)).Step(float64(g.step)).ZIndex(1)
	back, forth := ui.KeyUp, ui.KeyDown
	if g.vertical {
		e.Width(gripThickness).MarginX(-overhang).Orientation(ui.Vertical).Cursor(ui.CursorResizeEW)
		back, forth = ui.KeyLeft, ui.KeyRight
	} else {
		e.Height(gripThickness).MarginY(-overhang).Orientation(ui.Horizontal).Cursor(ui.CursorResizeNS)
	}
	set := func(v float32) { g.set(min(max(v, g.lo), g.hi)) }
	dx, dy, held := e.Dragged()
	move := dy
	if g.vertical {
		move = dx
	}
	if held && move != 0 {
		set(g.now + g.sign*move)
	}
	if e.Shortcut(0, back) {
		set(g.now - g.sign*g.step)
	}
	if e.Shortcut(0, forth) {
		set(g.now + g.sign*g.step)
	}
	if e.Shortcut(0, ui.KeyHome) {
		set(g.lo)
	}
	if e.Shortcut(0, ui.KeyEnd) {
		set(g.hi)
	}
	if g.reset != nil && e.DoubleClicked() {
		g.reset()
	}
	color := t.Border
	if held || e.Hovered() || e.Focused() {
		color = t.Accent
	}
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		if g.vertical {
			p.Fill(ui.Rect{X: r.X + overhang, Y: r.Y, W: 1, H: r.H}, color, 0)
		} else {
			p.Fill(ui.Rect{X: r.X, Y: r.Y + overhang, W: r.W, H: 1}, color, 0)
		}
	})
	return e
}
