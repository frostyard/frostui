package frostui

import (
	"encoding/json"
	"testing"

	"github.com/egoist/mygo/ui"
)

// threePanes is a sidebar, a list and a reader that takes the room left.
func threePanes(c *ui.Context) []Pane {
	text := func(s string) func() { return func() { ui.Text(c, s) } }
	return []Pane{
		{ID: "side", Label: "Sidebar", Width: 200, Min: 160, Max: 320, Collapsible: true, Content: text("side body")},
		{ID: "list", Label: "List", Width: 300, Min: 200, Max: 500, Content: text("list body")},
		{ID: "read", Label: "Reader", Min: 300, Content: text("read body")},
	}
}

// panesTester draws threePanes in a window width wide.
func panesTester(p *Panes, width int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) { p.View(c, threePanes(c)).Fill() }, width, 400)
}

func x(t *testing.T, tt *ui.Tester, s string) float32 {
	t.Helper()
	r, ok := tt.Find(s)
	if !ok {
		t.Fatalf("no %q in %q", s, tt.Texts())
	}
	return r.X
}

func drag(tt *ui.Tester, fromX, fromY, toX, toY float32) {
	tt.Press(fromX, fromY)
	tt.Move(toX, toY)
	tt.Release(toX, toY)
}

func TestPanesLayOutAtTheirWidths(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	if got := x(t, tt, "side body"); got != 0 {
		t.Errorf("side starts at %v", got)
	}
	if got := x(t, tt, "list body"); got != 201 {
		t.Errorf("list starts at %v, want 201", got)
	}
	if got := x(t, tt, "read body"); got != 502 {
		t.Errorf("reader starts at %v, want 502", got)
	}
}

func TestSeparatorDragsWithinItsBounds(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	drag(tt, 200, 100, 250, 100)
	if got := p.Layout.Widths["side"]; got != 250 {
		t.Errorf("dragged 50 right: side is %v", got)
	}
	if got := x(t, tt, "list body"); got != 251 {
		t.Errorf("the list follows to %v", got)
	}
	drag(tt, 250, 100, 900, 100)
	if got := p.Layout.Widths["side"]; got != 320 {
		t.Errorf("dragged past Max: side is %v", got)
	}
	drag(tt, 320, 100, 0, 100)
	if got := p.Layout.Widths["side"]; got != 160 {
		t.Errorf("dragged past Min: side is %v", got)
	}
}

// TestSeparatorKeepsTheFlexiblePaneItsMinimum stops a drag where the pane
// taking the room left reaches its Min, below the dragged pane's Max.
func TestSeparatorKeepsTheFlexiblePaneItsMinimum(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 900) // reader: 900-200-300-2 = 398
	drag(tt, 501, 100, 880, 100)
	if got := p.Layout.Widths["list"]; got != 398 {
		t.Errorf("list is %v: the reader was to keep its 300", got)
	}
}

func TestSeparatorIsASeparatorWithAValue(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	tt.Frame()
	a, ok := tt.AccessByRole(ui.RoleSeparator, "Sidebar")
	if !ok {
		t.Fatal("no separator named Sidebar")
	}
	if a.Min != 160 || a.Max != 320 || a.Now != 200 || a.Step != 16 {
		t.Errorf("range %v..%v now %v step %v", a.Min, a.Max, a.Now, a.Step)
	}
	if !a.Focusable || !a.Vertical {
		t.Errorf("focusable %v, vertical %v", a.Focusable, a.Vertical)
	}
	if _, ok := tt.AccessByRole(ui.RoleRegion, "Reader"); !ok {
		t.Error("the reader is not a region")
	}
}

func TestSeparatorMovesWithTheKeyboard(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	tt.Key(0, ui.KeyTab)
	if !tt.Focused("Sidebar") {
		t.Fatal("Tab did not reach the first separator")
	}
	want := func(key ui.Key, w float32) {
		t.Helper()
		tt.Key(0, key)
		if got := p.Layout.Widths["side"]; got != w {
			t.Errorf("key %v: side is %v, want %v", key, got, w)
		}
		if a, _ := tt.AccessByRole(ui.RoleSeparator, "Sidebar"); a.Now != float64(w) {
			t.Errorf("key %v: the separator reads %v, want %v", key, a.Now, w)
		}
	}
	want(ui.KeyRight, 216)
	want(ui.KeyLeft, 200)
	want(ui.KeyHome, 160)
	want(ui.KeyEnd, 320)
	want(ui.KeyRight, 320)
}

func TestSeparatorAfterTheFlexiblePaneMovesTheWayItPoints(t *testing.T) {
	var p Panes
	tt := ui.NewTester(func(c *ui.Context) {
		p.View(c, []Pane{
			{ID: "work", Min: 100, Content: func() { ui.Text(c, "work body") }},
			{ID: "right", Label: "Right", Width: 200, Min: 100, Content: func() { ui.Text(c, "right body") }},
		}).Fill()
	}, 800, 300)
	drag(tt, 599, 100, 549, 100)
	if got := p.Layout.Widths["right"]; got != 250 {
		t.Errorf("dragged 50 left: the right pane is %v", got)
	}
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyLeft)
	if got := p.Layout.Widths["right"]; got != 266 {
		t.Errorf("Left: the right pane is %v, want it to grow", got)
	}
}

func TestSeparatorRightToLeft(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	tt.SetDirection(ui.RightToLeft)
	tt.Frame()
	// The sidebar is on the right, so its line is 200 from the right edge.
	drag(tt, 799, 100, 749, 100)
	if got := p.Layout.Widths["side"]; got != 250 {
		t.Errorf("dragged 50 left: side is %v", got)
	}
}

func TestSeparatorDoubleClickResets(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	drag(tt, 200, 100, 250, 100)
	tt.ClickAt(250, 100)
	tt.ClickAt(250, 100)
	if _, ok := p.Layout.Widths["side"]; ok {
		t.Errorf("double click kept %v", p.Layout.Widths["side"])
	}
	if got := x(t, tt, "list body"); got != 201 {
		t.Errorf("the list is back at %v, want 201", got)
	}
}

func TestPanesCollapseAndShowAgain(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	p.Toggle("side")
	tt.Frame()
	if tt.HasText("side body") || p.Shown("side") || !p.Layout.Collapsed["side"] {
		t.Errorf("collapsed: %q, shown %v, layout %v", tt.Texts(), p.Shown("side"), p.Layout)
	}
	if got := x(t, tt, "list body"); got != 0 {
		t.Errorf("the list did not close up: %v", got)
	}
	if _, ok := tt.AccessByRole(ui.RoleSeparator, "Sidebar"); ok {
		t.Error("a collapsed pane keeps its separator")
	}
	p.Toggle("side")
	tt.Frame()
	if !tt.HasText("side body") || !p.Shown("side") || len(p.Layout.Collapsed) != 0 {
		t.Errorf("shown again: %q, layout %v", tt.Texts(), p.Layout)
	}
}

func TestPanesGiveWayWhenNarrow(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	tt.SetSize(700, 400) // 200+300+300+2 wanted: the panes give way
	if got := x(t, tt, "read body"); got < 399.5 || got > 400.5 {
		t.Errorf("reader starts at %v, want it to keep its minimum", got)
	}
	if !p.Shown("side") {
		t.Error("the sidebar hid while the minimums still fit")
	}
	tt.SetSize(600, 400) // the minimums are 660: the collapsible pane hides
	if p.Shown("side") || tt.HasText("side body") {
		t.Errorf("no room: %q", tt.Texts())
	}
	if p.Layout.Collapsed["side"] {
		t.Error("hiding for room changed the saved layout")
	}
	tt.SetSize(1000, 400)
	if !p.Shown("side") {
		t.Error("with room again the pane did not come back")
	}
}

// TestPanesShowTheOneTheAppAskedFor hides another pane first when the app
// shows one that lacked room.
func TestPanesShowTheOneTheAppAskedFor(t *testing.T) {
	var p Panes
	view := func(c *ui.Context) {
		p.View(c, []Pane{
			{ID: "left", Width: 200, Min: 200, Collapsible: true, Content: func() { ui.Text(c, "left body") }},
			{ID: "mid", Min: 200, Content: func() { ui.Text(c, "mid body") }},
			{ID: "right", Width: 200, Min: 200, Collapsible: true, Content: func() { ui.Text(c, "right body") }},
		}).Fill()
	}
	tt := ui.NewTester(view, 500, 300)
	tt.Frame()
	if !p.Shown("left") || p.Shown("right") {
		t.Fatalf("the last pane should hide first: left %v, right %v", p.Shown("left"), p.Shown("right"))
	}
	p.Toggle("right")
	tt.Frame()
	if p.Shown("left") || !p.Shown("right") {
		t.Errorf("asked for the right pane: left %v, right %v", p.Shown("left"), p.Shown("right"))
	}
}

func TestPaneLayoutRoundTrips(t *testing.T) {
	var p Panes
	tt := panesTester(&p, 1000)
	drag(tt, 200, 100, 250, 100)
	p.SetCollapsed("list", false)
	p.SetCollapsed("side", true)
	data, err := json.Marshal(p.Layout)
	if err != nil {
		t.Fatal(err)
	}
	var restored Panes
	if err := json.Unmarshal(data, &restored.Layout); err != nil {
		t.Fatal(err)
	}
	if restored.Layout.Widths["side"] != 250 || !restored.Layout.Collapsed["side"] {
		t.Errorf("restored %+v from %s", restored.Layout, data)
	}
	restored.SetCollapsed("side", false)
	tt2 := panesTester(&restored, 1000)
	if got := x(t, tt2, "list body"); got != 251 {
		t.Errorf("restored: the list starts at %v, want 251", got)
	}
}

func TestPaneWidthsOutOfBoundsAreClamped(t *testing.T) {
	p := Panes{Layout: PaneLayout{Widths: map[string]float32{"side": 9000, "list": -5}}}
	tt := panesTester(&p, 1200)
	if got := x(t, tt, "list body"); got != 321 {
		t.Errorf("list starts at %v, want the sidebar at its Max", got)
	}
	if got := x(t, tt, "read body"); got != 622 {
		t.Errorf("reader starts at %v, want the list at its default", got)
	}
}
