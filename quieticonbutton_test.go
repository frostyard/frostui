package frostui

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Paint assertions are the component snapshots: no binary baselines. Keep
// behavior and paint in one test so delegating to IconButton fails assertions,
// rather than just preserving the shared button mechanics.
func TestQuietIconButton(t *testing.T) {
	previousIcons := icons
	SetIcons(&Icons{cache: map[string]*ui.SVG{}}) // Embedded fallback only, independent of the desktop.
	t.Cleanup(func() { SetIcons(previousIcons) })
	if IconSVG("window-close") == nil {
		t.Fatal("embedded window-close icon is missing")
	}
	for _, dark := range []bool{false, true} {
		for _, high := range []bool{false, true} {
			t.Run(fmt.Sprintf("dark=%v/high=%v", dark, high), func(t *testing.T) {
				var th *ui.Theme
				clicks := 0
				tt := ui.NewTesterWith(func(c *ui.Context) {
					th = Adapt(c)
					ui.Column(c).Fill().Background(th.Background).Padding(20).Gap(20).AlignItems(ui.Start).Children(func() {
						if QuietIconButton(c, "window-close", "Copy item").Width(72).Height(44).Clicked() {
							clicks++
						}
						if QuietIconButton(c, "window-close", "Unavailable").Size(72, 44).Disabled(true).Clicked() {
							clicks++
						}
						IconButton(c, "window-close", "Ghost").Size(72, 44)
					})
				}, 240, 240, ui.TesterOptions{Dark: dark, Preferences: ui.Preferences{HighContrast: high}})
				tt.Frame()
				on, ok := tt.AccessByRole(ui.RoleButton, "Copy item")
				off, ok2 := tt.AccessByRole(ui.RoleButton, "Unavailable")
				ghost, ok3 := tt.AccessByRole(ui.RoleButton, "Ghost")
				if !ok || !ok2 || !ok3 {
					t.Fatalf("missing button semantics: %v %v %v", ok, ok2, ok3)
				}
				b := on.Bounds
				if b.W != 72 || b.H != 44 {
					t.Errorf("chained dimensions: %v", b)
				}
				if !off.Disabled {
					t.Error("disabled button not reported disabled")
				}
				assertColor := func(name string, got, want ui.Color) {
					t.Helper()
					if colorDistance(got, want) > 3 {
						t.Errorf("%s: paint %v, want %v", name, got, want)
					}
				}
				face := func(r ui.Rect) ui.Color { return pixel(tt, r.X+5, r.Y+r.H/2) }
				assertColor("rest face", face(b), th.Background)
				assertColor("rest border", pixel(tt, b.X, b.Y+b.H/2), th.Background)
				if colorDistance(pixel(tt, ghost.Bounds.X, ghost.Bounds.Y+ghost.Bounds.H/2), th.Background) < 5 {
					t.Error("ghost control has no contrasting resting border")
				}
				// Locate actual SVG ink, not a text-engine-dependent glyph. It must
				// retain Text's tested contrast on each painted face, including AAA.
				checkInk := func(name string, bg ui.Color, disabled bool) {
					t.Helper()
					want := th.Text
					if disabled {
						want = want.Alpha(0.5).Over(bg)
					}
					ink, distance := ui.Color{}, 1000
					for y := b.Y + b.H/2 - 7; y < b.Y+b.H/2+7; y++ {
						for x := b.X + b.W/2 - 7; x < b.X+b.W/2+7; x++ {
							p := pixel(tt, x, y)
							if d := colorDistance(p, want); d < distance {
								ink, distance = p, d
							}
						}
					}
					if distance > 12 {
						t.Errorf("%s: SVG ink %v not near %v", name, ink, want)
					}
					if !disabled {
						min := 4.5
						if high {
							min = 7
						}
						if ratio := contrast(ink, bg, th.Background); ratio < min {
							t.Errorf("%s: painted icon contrast %.2f:1, want %.1f:1", name, ratio, min)
						}
					}
				}
				checkInk("rest", face(b), false)
				x, y := b.X+b.W/2, b.Y+b.H/2
				tt.Move(x, y)
				assertColor("hover", face(b), th.SurfaceHover)
				if colorDistance(face(b), th.Background) < 5 {
					t.Error("hover has no visible treatment")
				}
				checkInk("hover", face(b), false)
				labelCount := func() int {
					n := 0
					for _, text := range tt.Texts() {
						if text == "Copy item" {
							n++
						}
					}
					return n
				}
				beforeTooltip := labelCount()
				tt.Advance(time.Second)
				if labelCount() <= beforeTooltip {
					t.Error("required label not used as tooltip")
				}
				tt.Press(x, y)
				assertColor("pressed", face(b), th.SurfacePressed)
				if colorDistance(face(b), th.SurfaceHover) < 5 {
					t.Error("press does not differ visibly from hover")
				}
				checkInk("pressed", face(b), false)
				tt.Release(x, y)
				if clicks != 1 {
					t.Errorf("pointer activation: clicks=%d", clicks)
				}
				tt.Move(220, 220)
				tt.Key(0, ui.KeyTab) // Pointer focus is not keyboard focus; cycle back.
				tt.Key(0, ui.KeyTab)
				if !tt.Focused("Copy item") {
					t.Fatal("Tab did not focus the button (or disabled button was not skipped)")
				}
				ring := pixel(tt, b.X-2, b.Y+b.H/2)
				assertColor("keyboard focus ring", ring, th.Focus.Over(th.Background))
				tt.Key(0, ui.KeyEnter)
				tt.Key(0, ui.KeySpace)
				if clicks != 3 {
					t.Errorf("Enter and Space activation: clicks=%d, want 3", clicks)
				}
				tt.Move(off.Bounds.X+off.Bounds.W/2, off.Bounds.Y+off.Bounds.H/2)
				assertColor("disabled hover face", face(off.Bounds), th.Background)
				tt.Press(off.Bounds.X+off.Bounds.W/2, off.Bounds.Y+off.Bounds.H/2)
				assertColor("disabled pressed face", face(off.Bounds), th.Background)
				tt.Release(off.Bounds.X+off.Bounds.W/2, off.Bounds.Y+off.Bounds.H/2)
				if clicks != 3 {
					t.Errorf("disabled activation: clicks=%d", clicks)
				}
				b = off.Bounds
				checkInk("disabled", face(b), true)
			})
		}
	}
}
