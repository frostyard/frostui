package frostui

import (
	"fmt"
	"math"
	"runtime"
	"testing"

	"github.com/egoist/mygo/ui"
)

// infoVisible requires visible geometry: Find also returns clipped elements.
func infoVisible(t *testing.T, tt *ui.Tester, label string, width, height int) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok || r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 ||
		r.X+r.W > float32(width) || r.Y+r.H > float32(height) {
		t.Fatalf("%q is not visible inside %dx%d: found=%v, bounds=%+v", label, width, height, ok, r)
	}
	return r
}

func infoSameHeight(t *testing.T, label string, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 0.01 {
		t.Errorf("%s height = %v, want full height %v", label, got, want)
	}
}

func infoEntries() []ShortcutEntry {
	const count, perGroup = 60, 10
	entries := make([]ShortcutEntry, count)
	for n := range entries {
		entries[n] = ShortcutEntry{
			Group: fmt.Sprintf("Group %d", n/perGroup), Title: fmt.Sprintf("Command %02d", n),
			Keys: []string{"Ctrl", fmt.Sprintf("Key %02d", n)},
		}
	}
	return entries
}

func TestInfoSmallBodiesStayContentSized(t *testing.T) {
	const width, height = 1280, 800
	for _, dark := range []bool{false, true} {
		for _, tc := range []struct {
			name    string
			body    func(*ui.Context)
			visible []string
		}{
			{name: "empty", body: func(*ui.Context) {}},
			{name: "single", body: ShortcutsBody([]ShortcutEntry{{Group: "General", Title: "One command", Keys: []string{"Ctrl", "K"}}}),
				visible: []string{"General", "One command", "Ctrl", "K"}},
			{name: "about", body: AboutBody(About{Name: "Example", Version: "v1.0.0", Summary: "A short summary.", License: "MIT"}),
				visible: []string{"Example", "v1.0.0", "A short summary.", "License", "MIT"}},
		} {
			t.Run(fmt.Sprintf("%s/dark=%v", tc.name, dark), func(t *testing.T) {
				// Measure the body's natural height independently. This uses the
				// current text engine on native and js/wasm, not native font metrics.
				reference := ui.NewTesterWith(func(c *ui.Context) {
					Adapt(c)
					ui.Column(c).Width(460).Gap(SpaceM).AlignItems(ui.Stretch).Label("Reference body").Children(func() { tc.body(c) })
				}, width, height, ui.TesterOptions{Dark: dark})
				body, ok := reference.Find("Reference body")
				if !ok {
					t.Fatal("missing natural-height reference")
				}
				var info Info
				info.Show("Info size test", tc.body)
				tt := ui.NewTesterWith(func(c *ui.Context) { Adapt(c); info.View(c) }, width, height, ui.TesterOptions{Dark: dark})
				dialog := infoVisible(t, tt, "Info size test", width, height)
				close := infoVisible(t, tt, "Close", width, height)
				infoSameHeight(t, "content-sized dialog", dialog.H, close.H+SpaceM+body.H)
				for _, label := range tc.visible {
					r := infoVisible(t, tt, label, width, height)
					full := infoVisible(t, reference, label, width, height)
					infoSameHeight(t, label, r.H, full.H)
				}
				if err := tt.Click("Close"); err != nil || info.Open {
					t.Fatalf("Close: err=%v, open=%v", err, info.Open)
				}
			})
		}
	}
}

func TestInfoNilBodyKeepsExistingBehavior(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(fmt.Sprintf("dark=%v", dark), func(t *testing.T) {
			var info Info
			info.Show("Nil body", nil)
			tt := ui.NewTesterWith(func(c *ui.Context) { Adapt(c); info.View(c) }, 640, 480, ui.TesterOptions{Dark: dark})
			if !info.Open || len(tt.Texts()) != 0 {
				t.Fatalf("nil body must draw nothing without closing: open=%v, texts=%q", info.Open, tt.Texts())
			}
		})
	}
}

func TestInfoLongBodyScrollsAndResizes(t *testing.T) {
	entries := infoEntries()
	for _, dark := range []bool{false, true} {
		for _, initial := range [][2]int{{1280, 800}, {640, 480}} {
			t.Run(fmt.Sprintf("%dx%d/dark=%v", initial[0], initial[1], dark), func(t *testing.T) {
				var info Info
				info.Show("Long info", ShortcutsBody(entries))
				tt := ui.NewTesterWith(func(c *ui.Context) { Adapt(c); info.View(c) }, initial[0], initial[1], ui.TesterOptions{Dark: dark})
				// Include both supplied viewport sizes and a width below 460.
				// Every resize uses this same Info and Tester.
				var distance float32
				for _, size := range [][2]int{initial, {640, 480}, {360, 480}, {1280, 800}, initial} {
					tt.SetSize(size[0], size[1])
					width, height := size[0], size[1]
					dialog := infoVisible(t, tt, "Long info", width, height)
					close := infoVisible(t, tt, "Close", width, height)
					first := infoVisible(t, tt, entries[0].Title, width, height)
					second := infoVisible(t, tt, entries[1].Title, width, height)
					key := infoVisible(t, tt, entries[0].Keys[1], width, height)
					group := infoVisible(t, tt, entries[0].Group, width, height)
					// MyGo v0.10.0 Find retains fully clipped labels with zero height.
					last, ok := tt.Find(entries[len(entries)-1].Title)
					if !ok || last.H != 0 {
						t.Fatalf("long body's last row must initially be clipped: found=%v, bounds=%+v", ok, last)
					}
					groups := 0
					for n, entry := range entries {
						if n == 0 || entry.Group != entries[n-1].Group {
							groups++
						}
					}
					// Bound the full content by row stride plus every group header.
					// Precise Scroll applies the distance immediately, without sleeps.
					distance = float32(len(entries))*(second.Y-first.Y) + float32(groups)*(group.H+SpaceM)
					tt.Scroll(first.X+first.W/2, first.Y+first.H/2, 0, distance)
					last = infoVisible(t, tt, entries[len(entries)-1].Title, width, height)
					lastKey := infoVisible(t, tt, entries[len(entries)-1].Keys[1], width, height)
					infoSameHeight(t, "last title", last.H, first.H)
					infoSameHeight(t, "last key", lastKey.H, key.H)
					if last.Y < close.Y+close.H || last.Y+last.H > dialog.Y+dialog.H {
						t.Errorf("last row leaves the body: last=%+v, close=%+v, dialog=%+v", last, close, dialog)
					}
					if after := infoVisible(t, tt, "Close", width, height); after != close {
						t.Errorf("scroll moved Close: before=%+v, after=%+v", close, after)
					}
					tt.Scroll(last.X+last.W/2, last.Y+last.H/2, 0, -distance)
					if returned := infoVisible(t, tt, entries[0].Title, width, height); returned != first {
						t.Errorf("scroll did not restore first row: before=%+v, after=%+v", first, returned)
					}
					infoSameHeight(t, "first key on return", infoVisible(t, tt, entries[0].Keys[1], width, height).H, key.H)
				}
				// Close must still take a click with the body at its end.
				first := infoVisible(t, tt, entries[0].Title, initial[0], initial[1])
				tt.Scroll(first.X+first.W/2, first.Y+first.H/2, 0, distance)
				infoVisible(t, tt, entries[len(entries)-1].Title, initial[0], initial[1])
				infoVisible(t, tt, "Close", initial[0], initial[1])
				if err := tt.Click("Close"); err != nil || info.Open {
					t.Fatalf("Close after scrolling and resizing: err=%v, open=%v", err, info.Open)
				}
			})
		}
	}
}

func TestInfoDismissalAndFocus(t *testing.T) {
	const width, height = 640, 480
	for _, dark := range []bool{false, true} {
		for _, dismissal := range []string{"Close", "Escape", "backdrop"} {
			t.Run(fmt.Sprintf("%s/dark=%v", dismissal, dark), func(t *testing.T) {
				var info Info
				tt := ui.NewTesterWith(func(c *ui.Context) {
					Adapt(c)
					ui.Column(c).Children(func() {
						if ui.Button(c, "Open info").Clicked() {
							info.Show("Focus info", func(c *ui.Context) {
								ui.Button(c, "First action")
								ui.Button(c, "Last action")
							})
						}
						ui.Button(c, "Behind dialog")
					})
					info.View(c)
				}, width, height, ui.TesterOptions{Dark: dark})
				if err := tt.Click("Open info"); err != nil {
					t.Fatal(err)
				}
				infoVisible(t, tt, "Close", width, height)
				if !tt.Focused("Close") {
					t.Fatal("opening did not focus Close")
				}
				for _, label := range []string{"First action", "Last action", "Close"} {
					tt.Key(0, ui.KeyTab)
					if !tt.Focused(label) {
						t.Fatalf("Tab must remain in dialog and focus %q", label)
					}
				}
				for _, label := range []string{"Last action", "First action", "Close"} {
					tt.Key(ui.Shift, ui.KeyTab)
					if !tt.Focused(label) {
						t.Fatalf("Shift+Tab must remain in dialog and focus %q", label)
					}
				}
				switch dismissal {
				case "Close":
					if err := tt.Click("Close"); err != nil {
						t.Fatal(err)
					}
				case "Escape":
					// Keyboard focus on Close opens its tooltip. MyGo gives
					// that overlay Escape first; test dialog dismissal from
					// a body button without a tooltip instead.
					tt.Key(0, ui.KeyTab)
					if !tt.Focused("First action") {
						t.Fatal("Escape setup did not focus the body button")
					}
					tt.Key(0, ui.KeyEscape)
				case "backdrop":
					tt.ClickAt(width-1, height-1)
				}
				if info.Open || !tt.Focused("Open info") {
					t.Fatalf("dismissal must close and return focus: open=%v, opener focused=%v", info.Open, tt.Focused("Open info"))
				}
				if _, ok := tt.Find("Close"); ok {
					t.Error("dismissed Info still draws Close")
				}
			})
		}
	}
}

func TestInfoBodyScrollsWithLongTitleInShortViewport(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("Node's MyGo text stub does not wrap titles; short-header geometry needs a native text renderer")
	}
	const width, tall, short = 240, 480, 200
	const title = "Keyboard shortcuts for navigating between windows, switching workspaces, and managing open documents"
	const firstLabel, lastLabel = "Dialog body", "End of body"
	const middleRows = 20
	var info Info
	info.Show(title, func(c *ui.Context) {
		ui.Text(c, firstLabel)
		for n := range middleRows {
			ui.Text(c, fmt.Sprintf("Body row %d", n))
		}
		ui.Text(c, lastLabel)
	})
	tt := ui.NewTester(func(c *ui.Context) { Adapt(c); info.View(c) }, width, tall)
	dialog := infoVisible(t, tt, title, width, tall)
	close := infoVisible(t, tt, "Close", width, tall)
	first := infoVisible(t, tt, firstLabel, width, tall)
	if headerH := first.Y - dialog.Y - SpaceM; headerH <= close.H {
		t.Fatalf("expected the wrapped header to be taller than Close at %dx%d: header=%v, Close=%v", width, tall, headerH, close.H)
	}

	tt.SetSize(width, short)
	firstShort := infoVisible(t, tt, firstLabel, width, short)
	infoSameHeight(t, firstLabel, firstShort.H, first.H)
	dialog = infoVisible(t, tt, title, width, short)
	closeShort := infoVisible(t, tt, "Close", width, short)
	if closeShort.W != close.W || closeShort.H != close.H {
		t.Fatalf("resize clipped Close: before=%+v, after=%+v", close, closeShort)
	}
	if end, ok := tt.Find(lastLabel); !ok || end.H != 0 {
		t.Fatalf("overflowing body end must start clipped: found=%v, bounds=%+v", ok, end)
	}
	// Scroll past the height of the rows and their gaps to reach the end.
	distance := (middleRows + 2) * (first.H + SpaceM)
	tt.Scroll(firstShort.X+firstShort.W/2, firstShort.Y+firstShort.H/2, 0, distance)
	end := infoVisible(t, tt, lastLabel, width, short)
	infoSameHeight(t, lastLabel, end.H, first.H)
	if end.Y < closeShort.Y+closeShort.H || end.Y+end.H > dialog.Y+dialog.H {
		t.Errorf("scrolled row outside body: row=%+v, Close=%+v, dialog=%+v", end, closeShort, dialog)
	}
	if after := infoVisible(t, tt, "Close", width, short); after != closeShort {
		t.Errorf("body scroll moved Close: before=%+v, after=%+v", closeShort, after)
	}
	if err := tt.Click("Close"); err != nil || info.Open {
		t.Fatalf("Close at %dx%d after scrolling: err=%v, open=%v", width, short, err, info.Open)
	}
}

func TestInfoCloseWithLongTitleInNarrowViewport(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("Node's MyGo text stub does not wrap titles; narrow-header geometry needs a native text renderer")
	}
	const width, height = 240, 480
	const title = "Keyboard shortcuts for navigating between windows, switching workspaces, and managing open documents"
	var info Info
	info.Show(title, func(c *ui.Context) { ui.Text(c, "Dialog body") })
	tt := ui.NewTester(func(c *ui.Context) { Adapt(c); info.View(c) }, width, height)
	infoVisible(t, tt, "Close", width, height)
	if err := tt.Click("Close"); err != nil || info.Open {
		t.Fatalf("Close with a long title: err=%v, open=%v", err, info.Open)
	}
}
