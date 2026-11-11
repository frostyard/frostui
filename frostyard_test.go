package frostui

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// luminance is a color's relative luminance, as WCAG 2 defines it.
func luminance(c ui.Color) float64 {
	ch := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(c.R) + 0.7152*ch(c.G) + 0.0722*ch(c.B)
}

// contrast is the WCAG 2 contrast ratio of fg drawn on bg, each composited
// over the opaque base first when it is translucent.
func contrast(fg, bg, base ui.Color) float64 {
	if bg.A < 255 {
		bg = bg.Over(base)
	}
	if fg.A < 255 {
		fg = fg.Over(bg)
	}
	a, b := luminance(fg), luminance(bg)
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05)
}

// pair is a foreground that must stand out from a ground by min.
type pair struct {
	name     string
	fg, bg   ui.Color
	min      float64
	overBase ui.Color // what a translucent ground lies on
}

// frostyardPairs lists every pair docs/specs/frostyard-theme.md lists.
func frostyardPairs(t *ui.Theme) []pair {
	k := ColorsOf(t)
	var ps []pair
	add := func(name string, fg, bg ui.Color, min float64, base ui.Color) {
		ps = append(ps, pair{name, fg, bg, min, base})
	}
	grounds := []struct {
		name string
		c    ui.Color
	}{
		{"background", t.Background}, {"chrome", k.Chrome}, {"panel", k.Panel},
		{"panel gradient end", k.PanelTo}, {"surface", t.Surface},
		{"surface hover", t.SurfaceHover}, {"surface pressed", t.SurfacePressed},
	}
	for _, g := range grounds {
		add("text on "+g.name, t.Text, g.c, 4.5, t.Background)
		add("muted text on "+g.name, t.TextMuted, g.c, 4.5, t.Background)
	}
	for _, base := range []struct {
		name string
		c    ui.Color
	}{{"background", t.Background}, {"chrome", k.Chrome}} {
		add("text on selected fill over "+base.name, t.Text, k.SelectedFill, 4.5, base.c)
		add("muted text on selected fill over "+base.name, t.TextMuted, k.SelectedFill, 4.5, base.c)
		add("selected text over "+base.name, k.SelectedText, k.SelectedFill, 4.5, base.c)
	}
	add("text on selection", t.Text, t.Selection, 4.5, t.Background)
	for _, face := range []struct {
		name string
		c    ui.Color
	}{{"accent", t.Accent}, {"accent hover", t.AccentHover}, {"accent pressed", t.AccentPressed}} {
		add("accent text on "+face.name, t.AccentText, face.c, 4.5, t.Background)
	}
	add("on-gradient text at its start", k.OnGradient, k.GradientFrom, 4.5, t.Background)
	add("on-gradient text at its end", k.OnGradient, k.GradientTo, 4.5, t.Background)
	for _, g := range []struct {
		name string
		c    ui.Color
	}{{"background", t.Background}, {"chrome", k.Chrome}, {"panel", k.Panel}} {
		add("hover text on "+g.name, k.HoverText, g.c, 4.5, t.Background)
		add("label on "+g.name, k.Label, g.c, 4.5, t.Background)
		add("tag text on "+g.name, k.TagText, g.c, 4.5, t.Background)
		add("focus ring on "+g.name, t.Focus, g.c, 3, t.Background)
		for _, tone := range []Tone{Accent, Success, Warning, Danger} {
			add(fmt.Sprintf("tone %d on %s", tone, g.name), ToneColor(t, tone), g.c, 4.5, t.Background)
		}
	}
	for _, tone := range []Tone{Neutral, Accent, Success, Warning, Danger} {
		fg := ToneColor(t, tone)
		if tone == Neutral {
			fg = k.TagText
		}
		add(fmt.Sprintf("tone %d on its fill over panel", tone), fg, k.Fill(tone), 4.5, k.Panel)
	}
	add("text on code", t.Text, k.Code, 4.5, t.Background)
	return ps
}

// TestFrostyardContrast holds both themes, with and without the desktop's
// high contrast, to WCAG 2: 4.5:1 for text, 3:1 for the focus ring
// (ADR-0011; frostyard/frostui#16).
func TestFrostyardContrast(t *testing.T) {
	for _, dark := range []bool{true, false} {
		for _, high := range []bool{false, true} {
			th := FrostyardTheme(dark)
			if high {
				th = th.WithPreferences(ui.Preferences{HighContrast: true})
			}
			for _, p := range frostyardPairs(th) {
				if r := contrast(p.fg, p.bg, p.overBase); r < p.min {
					t.Errorf("dark=%v high contrast=%v: %s is %.2f:1, under %.1f:1", dark, high, p.name, r, p.min)
				}
			}
		}
	}
}

// TestColorsOfFindsFrostyardThroughPreferences looks the theme colors up
// after the desktop's preferences and a recolored accent, as Adapt and
// SoftSelection leave the theme.
func TestColorsOfFindsFrostyardThroughPreferences(t *testing.T) {
	for _, dark := range []bool{true, false} {
		want := frostyardLight
		if dark {
			want = frostyardDark
		}
		th := FrostyardTheme(dark).WithPreferences(ui.Preferences{HighContrast: true, TextScale: 1.5})
		if got := ColorsOf(th); got != want {
			t.Errorf("dark=%v: after preferences, ColorsOf gives %+v", dark, got)
		}
		if got := ColorsOf(SoftSelection(th)); got != want {
			t.Errorf("dark=%v: in SoftSelection's scope, ColorsOf gives %+v", dark, got)
		}
	}
}

// TestColorsOfDerivesOtherThemes keeps the kit's former surfaces for a
// theme that is not Frostyard: MyGo's own, or an app's.
func TestColorsOfDerivesOtherThemes(t *testing.T) {
	for _, th := range []*ui.Theme{ui.LightTheme(), ui.DarkTheme()} {
		k := ColorsOf(th)
		if want := th.Background.Mix(th.Surface, 0.55); k.Chrome != want || Chrome(th) != want {
			t.Errorf("dark=%v: chrome %v, want %v", th.Dark, k.Chrome, want)
		}
		if want := th.Background.Mix(th.Surface, 0.7); k.Panel != want || Panel(th) != want {
			t.Errorf("dark=%v: panel %v, want %v", th.Dark, k.Panel, want)
		}
		if got, want := ToneFill(th, Neutral), th.Border.Alpha(0.6); got != want {
			t.Errorf("dark=%v: neutral fill %v, want %v", th.Dark, got, want)
		}
		if got, want := ToneFill(th, Danger), th.Danger.Alpha(0.16); got != want {
			t.Errorf("dark=%v: danger fill %v, want %v", th.Dark, got, want)
		}
	}
}

// TestAdaptSetsFrostyard draws a frame for a dark and a light desktop with
// an accent of its own, high contrast and a larger text size: the frame
// takes Frostyard or Frostyard light, keeps the brand's accent, and
// applies the rest.
func TestAdaptSetsFrostyard(t *testing.T) {
	desktopAccent := ui.Hex("#e01b24")
	for _, dark := range []bool{true, false} {
		var got, again *ui.Theme
		var body float32
		prefs := ui.Preferences{Accent: desktopAccent, HighContrast: true, TextScale: 1.25}
		tt := ui.NewTesterWith(func(c *ui.Context) {
			got = Adapt(c)
			again = Adapt(c)
			body = TextSize(c, 13.5)
			Body(c, "Frostyard")
		}, 400, 200, ui.TesterOptions{Dark: dark, Preferences: prefs})
		tt.Frame()
		want := FrostyardTheme(dark)
		if got == nil || got.Background != want.Background || got.Dark != dark {
			t.Fatalf("dark=%v: Adapt set %+v", dark, got)
		}
		if again != got {
			t.Errorf("dark=%v: a second Adapt in the frame set another theme", dark)
		}
		if got.Accent != want.Accent || got.InactiveAccent != want.Accent {
			t.Errorf("dark=%v: accent %v inactive %v, want the brand's %v", dark, got.Accent, got.InactiveAccent, want.Accent)
		}
		if got.Focus.A != 255 {
			t.Errorf("dark=%v: high contrast did not apply: focus %v", dark, got.Focus)
		}
		if got.FontSize != want.FontSize*1.25 {
			t.Errorf("dark=%v: font size %v, want %v", dark, got.FontSize, want.FontSize*1.25)
		}
		if math.Abs(float64(body-13.5*1.25)) > 0.01 {
			t.Errorf("dark=%v: Body draws at %v, want %v", dark, body, 13.5*1.25)
		}
		if got.Font != FontFamily {
			t.Errorf("dark=%v: font %q, want %q", dark, got.Font, FontFamily)
		}
	}
}

// TestAdaptRegistersInter adds the kit's font to the text system, with
// its weights: bold text is wider than regular.
func TestAdaptRegistersInter(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		ui.Column(c).AlignItems(ui.Start).Children(func() {
			ui.Text(c, "Frostyard frostui").Font(FontFamily).FontWeight(400).Label("regular")
			ui.Text(c, "Frostyard frostui").Font(FontFamily).FontWeight(700).Label("bold")
		})
	}, 400, 200)
	tt.Frame()
	if interErr != nil && strings.Contains(interErr.Error(), "no text engine") {
		t.Skipf("this system's text engine adds no fonts: %v", interErr)
	}
	var names []string
	for _, f := range ui.FontFamilies() {
		names = append(names, f.Name)
	}
	if !slices.Contains(names, FontFamily) {
		t.Fatalf("no %q among the text system's families after Adapt", FontFamily)
	}
	regular, ok1 := tt.Find("regular")
	bold, ok2 := tt.Find("bold")
	if !ok1 || !ok2 {
		t.Fatalf("texts not drawn: regular %v, bold %v", ok1, ok2)
	}
	if bold.W <= regular.W {
		t.Errorf("bold text is %v wide, regular %v: the font's weights did not apply", bold.W, regular.W)
	}
}

// TestLabelKeepsItsName names a label as written, though it draws it in
// capitals, so a screen reader reads the word.
func TestLabelKeepsItsName(t *testing.T) {
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		Label(c, "Mailboxes")
	}, 300, 100)
	tt.Frame()
	if _, ok := tt.Find("Mailboxes"); !ok {
		t.Errorf("texts %q: the label is not named Mailboxes", tt.Texts())
	}
}
