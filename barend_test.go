package frostui

import (
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The window and the window controls of frostyard/frostui#11.
const barTestWidth, barTestHeight = 1200, 700

var barTestControls = ui.TitleBar{Height: TopBarHeight, Left: 70, Right: 110}

// TestPaneShellBarEndStaysInTheWindow guards frostyard/frostui#11, B1 of
// the cross-review of #4: the bar's end sat in the last pane's segment, and
// a right pane 200 DIPs wide pushed the main menu out of the window and the
// palette field under the window controls. There the field is now compact
// and the end reaches over the pane before; at 480 it is a narrower field.
// Right to left, the end is at the left, by the controls there.
func TestPaneShellBarEndStaysInTheWindow(t *testing.T) {
	for _, tc := range []struct {
		side    float32
		rtl     bool
		compact bool
	}{
		{side: 200, compact: true},
		{side: 480},
		{side: 200, rtl: true, compact: true},
	} {
		var s PaneShell
		tt := ui.NewTester(func(c *ui.Context) {
			s.View(c, PaneFrame{
				SidebarPane: "side",
				Panes: []Pane{
					{ID: "main", Label: "Main", Bar: func() { ui.Text(c, "main bar") }, Content: func() { ui.Text(c, "main body") }},
					{ID: "side", Label: "Side", Width: tc.side, Min: tc.side, Max: tc.side, Collapsible: true,
						Bar: func() { ui.Text(c, "side bar") }, Content: func() { ui.Text(c, "side body") }},
				},
				End:         func() { IconButton(c, "view-refresh", "App item") },
				Menu:        func(m *ui.Menu) { m.Item("About") },
				ConsoleData: NewConsole(16),
			})
		}, barTestWidth, barTestHeight)
		controls := barTestControls
		if tc.rtl {
			tt.SetDirection(ui.RightToLeft)
			controls.Left, controls.Right = controls.Right, controls.Left
		}
		tt.SetTitleBar(controls)
		tt.Frame()
		if r := rect(t, tt, "side body"); r.W != tc.side {
			t.Fatalf("side %v: the pane is %v wide", tc.side, r.W)
		}
		barEndClear(t, tt, barTestWidth, controls, paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		barApart(t, tt, barTestWidth, "Sidebar", "main bar", "side bar", paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		paletteCompact(t, tt, barTestWidth, tc.compact, segmentFieldLeast)
		barEndWorks(t, tt, barTestWidth, &s.Palette, &s.Console.Open)
	}
}

// TestPaneShellPlainBarEndStaysInTheWindow is the guard of the same
// overflow in a bar not split: its trailing side shrank below the trailing
// controls, which ran under the window controls, and the field covered
// the sidebar toggle.
func TestPaneShellPlainBarEndStaysInTheWindow(t *testing.T) {
	for _, tc := range []struct {
		width   int
		compact bool
	}{{width: 800}, {width: 560, compact: true}} {
		var s PaneShell
		var side Panes
		tt := ui.NewTester(func(c *ui.Context) {
			s.View(c, PaneFrame{
				SidebarPane: "left", SidebarPanes: &side,
				Start:       func() { ui.Text(c, "Brand") },
				End:         func() { IconButton(c, "view-refresh", "App item") },
				Menu:        func(m *ui.Menu) { m.Item("About") },
				ConsoleData: NewConsole(16),
				Content:     func() { side.View(c, shellPanesNoBar(c)).Grow(1) },
			})
		}, tc.width, barTestHeight)
		tt.SetTitleBar(barTestControls)
		tt.Frame()
		barEndClear(t, tt, tc.width, barTestControls, "Sidebar", paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		barApart(t, tt, tc.width, "Sidebar", "Brand", paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		paletteCompact(t, tt, tc.width, tc.compact, paletteFieldLeast)
		barEndWorks(t, tt, tc.width, &s.Palette, &s.Console.Open)
	}
}

// TestShellBarEndStaysInTheWindow is the guard of the same overflow in
// the command center's bar: the field did not give way, and the trailing
// controls ran under the window controls and out of the window.
func TestShellBarEndStaysInTheWindow(t *testing.T) {
	for _, tc := range []struct {
		width   int
		compact bool
	}{{width: 900}, {width: 760, compact: true}} {
		var s Shell
		s.Brand.Name = "Example"
		tt := ui.NewTester(func(c *ui.Context) {
			s.View(c, Frame{
				Actions:     func() { IconButton(c, "view-refresh", "App item") },
				Menu:        func(m *ui.Menu) { m.Item("About") },
				ConsoleData: NewConsole(16),
				Content:     func() {},
			})
		}, tc.width, barTestHeight)
		tt.SetTitleBar(barTestControls)
		tt.Frame()
		barEndClear(t, tt, tc.width, barTestControls, "Example", paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		barApart(t, tt, tc.width, "Example", paletteLabel, "App item", "Console (Ctrl J)", "Main menu")
		paletteCompact(t, tt, tc.width, tc.compact, paletteFieldLeast)
		barEndWorks(t, tt, tc.width, &s.Palette, &s.Console.Open)
	}
}

// barEndClear fails t unless each named element shows, inside the window
// and clear of its controls.
func barEndClear(t *testing.T, tt *ui.Tester, w int, controls ui.TitleBar, names ...string) {
	t.Helper()
	for _, name := range names {
		r := rect(t, tt, name)
		if r.W <= 0 || r.H <= 0 || r.X < controls.Left || r.X+r.W > float32(w)-controls.Right {
			t.Errorf("window %v: %q at %+v, not between the window controls (%v to %v)", w, name, r, controls.Left, float32(w)-controls.Right)
		}
	}
}

// barApart fails t when two of the named elements that show overlap.
func barApart(t *testing.T, tt *ui.Tester, w int, names ...string) {
	t.Helper()
	for i, a := range names {
		ra := rect(t, tt, a)
		for _, b := range names[i+1:] {
			rb := rect(t, tt, b)
			if ra.W > 0 && rb.W > 0 && ra.X < rb.X+rb.W && rb.X < ra.X+ra.W && ra.Y < rb.Y+rb.H && rb.Y < ra.Y+ra.H {
				t.Errorf("window %v: %q at %+v overlaps %q at %+v", w, a, ra, b, rb)
			}
		}
	}
}

// paletteCompact fails t unless the palette's control is the compact
// button when compact says so, with the field's name, tooltip and key, and
// otherwise the field, no narrower than least.
func paletteCompact(t *testing.T, tt *ui.Tester, w int, compact bool, least float32) {
	t.Helper()
	r := rect(t, tt, paletteLabel)
	_, key := tt.Find("Ctrl")
	if !compact {
		if r.W < least || !key {
			t.Errorf("window %v: the field is %v wide (least %v), key shown %v", w, r.W, least, key)
		}
		return
	}
	a, ok := tt.AccessByRole(ui.RoleButton, paletteLabel)
	if r.W >= least || key || !ok || a.Description != paletteTooltip || a.KeyShortcuts != "Ctrl+K" {
		t.Errorf("window %v: not the compact field: %v wide, key shown %v, %+v", w, r.W, key, a)
	}
}

// barEndWorks clicks the main menu, the palette's control and the
// console's button, and presses Ctrl K.
func barEndWorks(t *testing.T, tt *ui.Tester, w int, p *Palette, consoleOpen *bool) {
	t.Helper()
	if err := tt.Click("Main menu"); err != nil {
		t.Fatalf("window %v: %v", w, err)
	}
	if !slices.Contains(tt.Menu(), "About") {
		t.Errorf("window %v: the main menu shows %v", w, tt.Menu())
	}
	tt.CloseMenu()
	if err := tt.Click(paletteLabel); err != nil {
		t.Fatalf("window %v: %v", w, err)
	}
	if !p.Open {
		t.Errorf("window %v: a click did not open the palette", w)
	}
	p.Open = false
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !p.Open {
		t.Errorf("window %v: Ctrl K did not open the palette", w)
	}
	p.Open = false
	tt.Frame()
	if err := tt.Click("Console (Ctrl J)"); err != nil {
		t.Fatalf("window %v: %v", w, err)
	}
	if !*consoleOpen {
		t.Errorf("window %v: the button did not open the console", w)
	}
}
