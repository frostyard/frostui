package frostui

import (
	"fmt"
	"math"
	"os"
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

// pair is a foreground that must stand out from a ground: 7:1 for text
// under the kit's high contrast, 4.5:1 for text otherwise, and 3:1 for the
// focus ring, which is no text (WCAG 2).
type pair struct {
	name     string
	fg, bg   ui.Color
	text     bool
	overBase ui.Color // what a translucent ground lies on
}

// toneName is tone's name in the spec's list of pairs.
func toneName(tone Tone) string {
	switch tone {
	case Accent:
		return "accent"
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Danger:
		return "danger"
	}
	return "neutral"
}

// frostyardPairs lists every pair docs/specs/frostyard-theme.md lists.
func frostyardPairs(t *ui.Theme) []pair {
	k := ColorsOf(t)
	var ps []pair
	add := func(name string, fg, bg, base ui.Color) {
		ps = append(ps, pair{name, fg, bg, true, base})
	}
	addRing := func(name string, fg, bg, base ui.Color) {
		ps = append(ps, pair{name, fg, bg, false, base})
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
		add("text on "+g.name, t.Text, g.c, t.Background)
		add("muted text on "+g.name, t.TextMuted, g.c, t.Background)
	}
	for _, base := range []struct {
		name string
		c    ui.Color
	}{{"background", t.Background}, {"chrome", k.Chrome}} {
		add("text on selected fill over "+base.name, t.Text, k.SelectedFill, base.c)
		add("muted text on selected fill over "+base.name, t.TextMuted, k.SelectedFill, base.c)
		add("selected text over "+base.name, k.SelectedText, k.SelectedFill, base.c)
	}
	add("text on selection", t.Text, t.Selection, t.Background)
	for _, face := range []struct {
		name string
		c    ui.Color
	}{{"accent", t.Accent}, {"accent hover", t.AccentHover}, {"accent pressed", t.AccentPressed}} {
		add("accent text on "+face.name, t.AccentText, face.c, t.Background)
	}
	add("on-gradient text at its start", k.OnGradient, k.GradientFrom, t.Background)
	add("on-gradient text at its end", k.OnGradient, k.GradientTo, t.Background)
	for _, g := range []struct {
		name string
		c    ui.Color
	}{{"background", t.Background}, {"chrome", k.Chrome}, {"panel", k.Panel}} {
		add("hover text on "+g.name, k.HoverText, g.c, t.Background)
		add("label on "+g.name, k.Label, g.c, t.Background)
		add("tag text on "+g.name, k.TagText, g.c, t.Background)
		addRing("focus ring on "+g.name, t.Focus, g.c, t.Background)
		for _, tone := range []Tone{Accent, Success, Warning, Danger} {
			add(toneName(tone)+" tone on "+g.name, ToneColor(t, tone), g.c, t.Background)
		}
	}
	for _, tone := range []Tone{Neutral, Accent, Success, Warning, Danger} {
		name := toneName(tone) + " tone on its fill over panel"
		fg := ToneColor(t, tone)
		if tone == Neutral {
			name = "neutral tag text on its fill over panel"
			fg = k.TagText
		}
		add(name, fg, k.Fill(tone), k.Panel)
	}
	add("text on code", t.Text, k.Code, t.Background)
	return ps
}

// checkContrast holds ps to WCAG 2: textMin for text, 3:1 for the focus
// ring.
func checkContrast(t *testing.T, label string, ps []pair, textMin float64) {
	t.Helper()
	for _, p := range ps {
		min := 3.0
		if p.text {
			min = textMin
		}
		if r := contrast(p.fg, p.bg, p.overBase); r < min {
			t.Errorf("%s: %s is %.2f:1, under %.1f:1", label, p.name, r, min)
		}
	}
}

// TestFrostyardContrast holds both themes to WCAG 2 in 3 cases: the brand,
// MyGo's mix alone (the desktop's high contrast), and the kit's high
// contrast. Text holds 4.5:1 in the first 2, and 7:1 (AAA) in the last; the
// focus ring holds 3:1 in every case (ADR-0011, ADR-0013;
// frostyard/frostui#16, frostyard/frostui#32).
func TestFrostyardContrast(t *testing.T) {
	for _, dark := range []bool{true, false} {
		brand := FrostyardTheme(dark)
		mix := brand.WithPreferences(ui.Preferences{HighContrast: true})
		for _, tt := range []struct {
			name  string
			theme *ui.Theme
			min   float64
		}{
			{"brand", brand, 4.5},
			{"MyGo's mix", mix, 4.5},
			{"the kit's high contrast", frostyardHighContrast(mix), 7},
		} {
			checkContrast(t, fmt.Sprintf("dark=%v, %s", dark, tt.name), frostyardPairs(tt.theme), tt.min)
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
// takes Frostyard or Frostyard light, keeps the brand's accent, gives it the
// kit's high-contrast values, and applies the rest (ADR-0013,
// frostyard/frostui#32).
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
		p := prefs
		p.Accent = ui.Color{}
		high := frostyardHighContrast(FrostyardTheme(dark).WithPreferences(p))
		if got == nil || got.Background != want.Background || got.Dark != dark {
			t.Fatalf("dark=%v: Adapt set %+v", dark, got)
		}
		if again != got {
			t.Errorf("dark=%v: a second Adapt in the frame set another theme", dark)
		}
		if got.Accent != high.Accent || got.InactiveAccent != high.Accent {
			t.Errorf("dark=%v: accent %v inactive %v, want the kit's high-contrast %v", dark, got.Accent, got.InactiveAccent, high.Accent)
		}
		if got.Accent == desktopAccent || got.Accent == want.Accent {
			t.Errorf("dark=%v: high contrast kept the accent %v", dark, got.Accent)
		}
		wantColors := frostyardDarkHigh.colors
		if !dark {
			wantColors = frostyardLightHigh.colors
		}
		if colors := ColorsOf(got); colors != wantColors {
			t.Errorf("dark=%v: high contrast gives the theme colors %+v", dark, colors)
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

// TestAdaptAppliesHighContrast draws frames with the desktop's high
// contrast on and off, for both themes: with it on, Adapt sets the kit's
// high-contrast values, over a Frostyard theme the app set itself too, and
// a second Adapt returns the same theme; with it off, the brand's values
// stay (ADR-0013, frostyard/frostui#32).
func TestAdaptAppliesHighContrast(t *testing.T) {
	for _, dark := range []bool{true, false} {
		p := ui.Preferences{HighContrast: true}
		want := frostyardLightHigh
		brand := frostyardLight
		if dark {
			want, brand = frostyardDarkHigh, frostyardDark
		}
		wantTheme := frostyardHighContrast(FrostyardTheme(dark).WithPreferences(p))

		// (a) Adapt's own theme.
		var got, again *ui.Theme
		tt := ui.NewTesterWith(func(c *ui.Context) {
			got = Adapt(c)
			again = Adapt(c)
		}, 400, 200, ui.TesterOptions{Dark: dark, Preferences: p})
		tt.Frame()
		if got == nil {
			t.Fatalf("dark=%v: Adapt set no theme", dark)
		}
		if got.Text != wantTheme.Text || got.TextMuted != wantTheme.TextMuted ||
			got.Accent != wantTheme.Accent || got.Danger != wantTheme.Danger {
			t.Errorf("dark=%v: high contrast gives text %v muted %v accent %v danger %v, want %+v",
				dark, got.Text, got.TextMuted, got.Accent, got.Danger, wantTheme)
		}
		if again != got {
			t.Errorf("dark=%v: a second Adapt set another theme", dark)
		}
		if colors := ColorsOf(got); colors != want.colors {
			t.Errorf("dark=%v: high contrast gives the theme colors %+v", dark, colors)
		}
		if colors := ColorsOf(SoftSelection(got)); colors != want.colors {
			t.Errorf("dark=%v: in SoftSelection's scope, the theme colors are %+v", dark, colors)
		}

		// (b) A Frostyard theme the app sets itself.
		var first, second, appSet *ui.Theme
		tt = ui.NewTesterWith(func(c *ui.Context) {
			appSet = FrostyardTheme(dark).WithPreferences(p)
			c.SetTheme(appSet)
			first = Adapt(c)
			second = Adapt(c)
		}, 400, 200, ui.TesterOptions{Dark: dark, Preferences: p})
		tt.Frame()
		if first == nil || first == appSet {
			t.Fatalf("dark=%v: Adapt kept the app's theme", dark)
		}
		if first.Text != wantTheme.Text || first.TextMuted != wantTheme.TextMuted ||
			first.Accent != wantTheme.Accent || first.Danger != wantTheme.Danger {
			t.Errorf("dark=%v: over the app's theme, high contrast gives text %v muted %v accent %v danger %v",
				dark, first.Text, first.TextMuted, first.Accent, first.Danger)
		}
		if second != first {
			t.Errorf("dark=%v: a second Adapt over the app's theme set another theme", dark)
		}
		if colors := ColorsOf(first); colors != want.colors {
			t.Errorf("dark=%v: over the app's theme, the theme colors are %+v", dark, colors)
		}

		// (c) Without high contrast, nothing changes.
		var off *ui.Theme
		tt = ui.NewTesterWith(func(c *ui.Context) {
			off = Adapt(c)
		}, 400, 200, ui.TesterOptions{Dark: dark})
		tt.Frame()
		if off == nil || off.Text != FrostyardTheme(dark).Text {
			t.Fatalf("dark=%v: without high contrast, text is %v", dark, off)
		}
		if got := ColorsOf(off); got != brand {
			t.Errorf("dark=%v: without high contrast, the theme colors are %+v", dark, got)
		}
	}
}

// TestSpecListsHighContrastRatios keeps the ratio table of
// docs/specs/frostyard-theme.md in step with the kit's colors: the spec
// holds the table this builds, character for character. Its failure prints
// the block to paste there (ADR-0013, frostyard/frostui#32).
func TestSpecListsHighContrastRatios(t *testing.T) {
	darkMix := FrostyardTheme(true).WithPreferences(ui.Preferences{HighContrast: true})
	lightMix := FrostyardTheme(false).WithPreferences(ui.Preferences{HighContrast: true})
	lists := [4][]pair{
		frostyardPairs(darkMix),
		frostyardPairs(frostyardHighContrast(darkMix)),
		frostyardPairs(lightMix),
		frostyardPairs(frostyardHighContrast(lightMix)),
	}
	names := lists[0]
	for i, list := range lists[1:] {
		if len(list) != len(names) {
			t.Fatalf("column %d has %d pairs, the first has %d", i+1, len(list), len(names))
		}
	}
	var b strings.Builder
	b.WriteString("| Pair | Frostyard, MyGo's mix | Frostyard, the kit's | Frostyard light, MyGo's mix | Frostyard light, the kit's |\n")
	b.WriteString("| --- | ---: | ---: | ---: | ---: |\n")
	for i, p := range names {
		fmt.Fprintf(&b, "| %s", p.name)
		for _, list := range lists {
			q := list[i]
			if q.name != p.name {
				t.Fatalf("pair %d of a column is %q, the first column's is %q", i, q.name, p.name)
			}
			fmt.Fprintf(&b, " | %.2f:1", contrast(q.fg, q.bg, q.overBase))
		}
		b.WriteString(" |\n")
	}
	table := b.String()
	spec, err := os.ReadFile("docs/specs/frostyard-theme.md")
	if err != nil {
		t.Fatalf("reading the spec: %v", err)
	}
	if !strings.Contains(string(spec), table) {
		t.Errorf("docs/specs/frostyard-theme.md does not hold this table:\n%s", table)
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
