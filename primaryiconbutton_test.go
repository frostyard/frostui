package frostui

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// These paint snapshots and input checks share one test: an IconButton
// delegation must fail the rendered face and foreground assertions in every
// theme, even though it preserves the underlying button mechanics.
func TestPrimaryIconButton(t *testing.T) {
	previousIcons := icons
	SetIcons(&Icons{cache: map[string]*ui.SVG{}}) // Deterministic embedded fallback, no desktop icons.
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
						if PrimaryIconButton(c, "window-close", "Close item").Width(72).Height(44).Clicked() {
							clicks++
						}
						if PrimaryIconButton(c, "window-close", "Unavailable").Size(72, 44).Disabled(true).Clicked() {
							clicks++
						}
						PrimaryButton(c, "").Label("Face reference").Size(72, 44)
					})
				}, 240, 240, ui.TesterOptions{Dark: dark, Preferences: ui.Preferences{HighContrast: high}})
				tt.Frame()
				on, ok := tt.AccessByRole(ui.RoleButton, "Close item")
				off, ok2 := tt.AccessByRole(ui.RoleButton, "Unavailable")
				ref, ok3 := tt.AccessByRole(ui.RoleButton, "Face reference")
				if !ok || !ok2 || !ok3 {
					t.Fatalf("missing name/role: %v %v %v", ok, ok2, ok3)
				}
				b := on.Bounds
				if b.W != 72 || b.H != 44 {
					t.Errorf("requested dimensions: %v", b)
				}
				if !off.Disabled {
					t.Error("disabled semantics missing")
				}
				k := ColorsOf(th)
				lead, trail := faceEnds(tt, b)
				refLead, refTrail := faceEnds(tt, ref.Bounds)
				checkFace := func(state string) {
					t.Helper()
					l, r := faceEnds(tt, b)
					if colorDistance(l, refLead) > 2 || colorDistance(r, refTrail) > 2 {
						t.Errorf("%s: face differs from PrimaryButton: %v, %v", state, l, r)
					}
					// High contrast brings the gradient ends closer together;
					// the wide-button helper's fixed distance threshold does not
					// apply to this compact target. Check direction and the actual
					// PrimaryButton paint instead.
					if l == r || colorDistance(l, k.GradientFrom) >= colorDistance(l, k.GradientTo) ||
						colorDistance(r, k.GradientTo) >= colorDistance(r, k.GradientFrom) {
						t.Errorf("%s: gradient direction: %v, %v", state, l, r)
					}
				}
				checkFace("rest")
				// Every pixel outside the central icon matches a blank PrimaryButton:
				// no visible label, border or alternative fill can hide behind semantics.
				for y := float32(0); y < b.H; y++ {
					for x := float32(0); x < b.W; x++ {
						if x >= b.W/2-10 && x < b.W/2+10 && y >= b.H/2-10 && y < b.H/2+10 {
							continue
						}
						if colorDistance(pixel(tt, b.X+x, b.Y+y), pixel(tt, ref.Bounds.X+x, ref.Bounds.Y+y)) > 2 {
							t.Fatalf("icon-only face differs from blank PrimaryButton at (%v,%v)", x, y)
						}
					}
				}
				checkInk := func(r ui.Rect, want ui.Color) ui.Color {
					t.Helper()
					ink, distance := ui.Color{}, 1000
					for y := r.Y + r.H/2 - 7; y < r.Y+r.H/2+7; y++ {
						for x := r.X + r.W/2 - 7; x < r.X+r.W/2+7; x++ {
							p := pixel(tt, x, y)
							if d := colorDistance(p, want); d < distance {
								ink, distance = p, d
							}
						}
					}
					if distance > 12 {
						t.Errorf("rendered icon ink %v, want near on-gradient %v", ink, want)
					}
					return ink
				}
				ink := checkInk(b, k.OnGradient)
				min := 4.5
				if high {
					min = 7
				}
				for _, bg := range []ui.Color{lead, trail} {
					if ratio := contrast(ink, bg, th.Background); ratio < min {
						t.Errorf("rendered contrast %.2f:1, want %.1f:1", ratio, min)
					}
				}
				offLead, offTrail := faceEnds(tt, off.Bounds)
				for _, p := range [][2]ui.Color{{lead, offLead}, {trail, offTrail}} {
					if want := p[0].Alpha(0.5).Over(th.Background); colorDistance(p[1], want) > 6 {
						t.Errorf("disabled face %v, want %v", p[1], want)
					}
				}
				// A blank reference supplies the gradient beneath the icon without ink.
				centerFace := pixel(tt, ref.Bounds.X+ref.Bounds.W/2, ref.Bounds.Y+ref.Bounds.H/2).Alpha(0.5).Over(th.Background)
				checkInk(off.Bounds, k.OnGradient.Alpha(0.5).Over(centerFace))
				labelCount := func() int {
					n := 0
					for _, text := range tt.Texts() {
						if text == "Close item" {
							n++
						}
					}
					return n
				}
				before := labelCount()
				x, y := b.X+b.W/2, b.Y+b.H/2
				tt.Move(x, y)
				checkFace("hover")
				tt.Advance(time.Second)
				if labelCount() <= before {
					t.Error("label not displayed as tooltip after simulated delay")
				}
				tt.Press(x, y)
				checkFace("pressed")
				tt.Release(x, y)
				if clicks != 1 {
					t.Errorf("pointer activation: %d clicks", clicks)
				}
				tt.Move(220, 220)
				tt.Key(0, ui.KeyTab)
				if !tt.Focused("Face reference") {
					t.Error("Tab failed to skip disabled button")
				}
				tt.Key(0, ui.KeyTab)
				if !tt.Focused("Close item") {
					t.Fatal("Tab failed to focus primary icon button")
				}
				if got, want := pixel(tt, b.X-2, b.Y+b.H/2), th.Focus.Over(th.Background); colorDistance(got, want) > 3 {
					t.Errorf("focus ring %v, want %v", got, want)
				}
				tt.Key(0, ui.KeyEnter)
				tt.Key(0, ui.KeySpace)
				if clicks != 3 {
					t.Errorf("Enter/Space activation: %d clicks, want 3", clicks)
				}
				tt.ClickAt(off.Bounds.X+off.Bounds.W/2, off.Bounds.Y+off.Bounds.H/2)
				if clicks != 3 {
					t.Errorf("disabled activation: %d clicks", clicks)
				}
			})
		}
	}
}
