package workspaceui

import (
	"math"
	"strings"
	"testing"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
	"github.com/nsgo-dev/nsgo/ui"
)

// rgba unpacks a color (0xRRGGBBAA) into its channels, 0 to 1.
func rgba(c core.Color) (r, g, b, a float64) {
	return float64(c>>24&0xff) / 255, float64(c>>16&0xff) / 255, float64(c>>8&0xff) / 255, float64(c&0xff) / 255
}

// over is fg laid on the opaque ground bg.
func over(fg, bg core.Color) core.Color {
	fr, fgG, fb, fa := rgba(fg)
	br, bgG, bb, _ := rgba(bg)
	mix := func(f, b float64) uint32 { return uint32(math.Round((f*fa + b*(1-fa)) * 255)) }
	return core.Color(mix(fr, br)<<24 | mix(fgG, bgG)<<16 | mix(fb, bb)<<8 | 0xff)
}

// luminance is WCAG 2's relative luminance of an opaque color.
func luminance(c core.Color) float64 {
	lin := func(v float64) float64 {
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	r, g, b, _ := rgba(c)
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrast is WCAG 2's contrast ratio of fg on bg, both laid over ground
// first when they are not opaque.
func contrast(fg, bg, ground core.Color) float64 {
	bg = over(bg, ground)
	a, b := luminance(over(fg, bg)), luminance(bg)
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05)
}

func role(t *testing.T, p core.Palette, name string) core.Color {
	t.Helper()
	c, ok := p.Color(name)
	if !ok {
		t.Fatalf("the palette has no role %q", name)
	}
	return c
}

// TestThemeHoldsTheSpecsColors pins the palettes to the values of
// docs/specs/frostyard-theme.md, by NSGo's role names and the kit's.
func TestThemeHoldsTheSpecsColors(t *testing.T) {
	theme := Theme()
	for _, tc := range []struct {
		name        string
		dark, light core.Color
	}{
		{"background", core.RGB(0x06111d), core.RGB(0xf5f9fc)},
		{"surface", core.RGB(0x0d2a41), core.RGB(0xdfeaf2)},
		{"border", core.RGBA(0xa2deff, 0.18), core.RGBA(0x06111d, 0.14)},
		{"text", core.RGB(0xedf8ff), core.RGB(0x06111d)},
		{"textMuted", core.RGB(0xb6d0df), core.RGB(0x3f5a6b)},
		{"accent", core.RGB(0x47b8ef), core.RGB(0x0b6ea6)},
		{"accentHover", core.RGB(0x68c5f2), core.RGB(0x0a6091)},
		{"accentPressed", core.RGB(0x3a97c5), core.RGB(0x0a527d)},
		{"accentText", core.RGB(0x02131f), core.RGB(0xffffff)},
		{"focus", core.RGB(0x47b8ef), core.RGB(0x0b6ea6)},
		{"success", core.RGB(0x7edafa), core.RGB(0x076377)},
		{"warning", core.RGB(0xe0b15c), core.RGB(0x8a5300)},
		{"danger", core.RGB(0xe07a72), core.RGB(0xb0322b)},
		{"selection", core.RGBA(0x47b8ef, 0.35), core.RGBA(0x0b6ea6, 0.25)},
		{"scrollbar", core.RGBA(0xa2deff, 0.35), core.RGBA(0x06111d, 0.32)},
		{"inset", core.RGB(0x081b2d), core.RGB(0xe8f0f6)},
		{"raised", core.RGB(0x0d263b), core.RGB(0xffffff)},
		{roleChrome, core.RGB(0x081b2d), core.RGB(0xe8f0f6)},
		{rolePanel, core.RGB(0x0d263b), core.RGB(0xffffff)},
		{roleSelected, core.RGBA(0x47b8ef, 0.08), core.RGBA(0x0b6ea6, 0.10)},
		{roleSelectedText, core.RGB(0xaee9ff), core.RGB(0x064a72)},
		{roleLabel, core.RGB(0x82b6d0), core.RGB(0x3b6680)},
		{roleFill + "neutral", core.RGBA(0xa2deff, 0.10), core.RGBA(0x06111d, 0.06)},
		{roleFill + "danger", core.RGB(0x3a181e), core.RGB(0xf9e2df)},
		{roleLine + "warning", core.RGBA(0xe0b15c, 0.33), core.RGBA(0x8a5300, 0.4)},
	} {
		if got := role(t, theme.Dark, tc.name); got != tc.dark {
			t.Errorf("Frostyard %s = %#08x, want %#08x", tc.name, uint32(got), uint32(tc.dark))
		}
		if got := role(t, theme.Light, tc.name); got != tc.light {
			t.Errorf("Frostyard light %s = %#08x, want %#08x", tc.name, uint32(got), uint32(tc.light))
		}
	}
	// The hover and pressed faces are overlays; over Surface they are the
	// spec's opaque values.
	for _, tc := range []struct {
		name          string
		p             core.Palette
		face, surface core.Color
		want          core.Color
	}{
		{"Frostyard hover", theme.Dark, theme.Dark.SurfaceHover, theme.Dark.Surface, core.RGB(0x1a364c)},
		{"Frostyard pressed", theme.Dark, theme.Dark.SurfacePressed, theme.Dark.Surface, core.RGB(0x284358)},
		{"Frostyard light hover", theme.Light, theme.Light.SurfaceHover, theme.Light.Surface, core.RGB(0xd2dde5)},
		{"Frostyard light pressed", theme.Light, theme.Light.SurfacePressed, theme.Light.Surface, core.RGB(0xc5d0d8)},
	} {
		got := over(tc.face, tc.surface)
		gr, gg, gb, _ := rgba(got)
		wr, wg, wb, _ := rgba(tc.want)
		if d := max(math.Abs(gr-wr), math.Abs(gg-wg), math.Abs(gb-wb)) * 255; d > 1 {
			t.Errorf("%s over the surface = %#08x, want %#08x within 1 step", tc.name, uint32(got), uint32(tc.want))
		}
	}
	if theme.Sans != FontFamily {
		t.Errorf("the theme's sans family = %q, want %q", theme.Sans, FontFamily)
	}
	if want := core.DefaultTheme().Radius; theme.Radius != want {
		t.Errorf("the theme's radius = %v, want NSGo's %v", theme.Radius, want)
	}
}

// TestThemeSetsEveryRoleNSGoNames fails when NSGo's default theme names a
// role the Frostyard palettes leave at NSGo's gray or blue: a role NSGo adds
// in a later version shows up here, to be given its Frostyard color. The
// chart colors are the only ones left as NSGo's.
func TestThemeSetsEveryRoleNSGoNames(t *testing.T) {
	theme, plain := Theme(), core.DefaultTheme()
	for _, appearance := range []struct {
		name        string
		ours, nsgos core.Palette
	}{{"dark", theme.Dark, plain.Dark}, {"light", theme.Light, plain.Light}} {
		for _, name := range appearance.nsgos.ColorNames() {
			if strings.HasPrefix(name, "chart-") {
				continue
			}
			ours, theirs := role(t, appearance.ours, name), role(t, appearance.nsgos, name)
			// White stays white in Frostyard light: its panel, its fields and
			// the text on its accent.
			if ours == theirs && ours != core.RGB(0xffffff) {
				t.Errorf("%s role %q is NSGo's default %#08x: give it a Frostyard color in palette", appearance.name, name, uint32(theirs))
			}
		}
	}
}

// TestThemeContrast holds the pairs a window draws to WCAG 2's 4.5:1 in
// both themes: text and the muted, label and tone colors on each ground,
// the text on the accent's three faces, and text on a chosen row.
func TestThemeContrast(t *testing.T) {
	theme := Theme()
	for _, appearance := range []struct {
		name string
		p    core.Palette
	}{{"Frostyard", theme.Dark}, {"Frostyard light", theme.Light}} {
		p := appearance.p
		get := func(name string) core.Color { return role(t, p, name) }
		grounds := []string{"background", "inset", "raised", "surface", "field"}
		type pair struct {
			what       string
			fg, bg, on core.Color
		}
		var pairs []pair
		for _, ground := range grounds {
			g := get(ground)
			for _, fg := range []string{"text", "textMuted", "faint", "accent", "success", "warning", "danger", roleSelectedText} {
				pairs = append(pairs, pair{fg + " on " + ground, get(fg), g, g})
			}
			pairs = append(pairs,
				pair{"text on a hovered row over " + ground, p.Text, p.SurfaceHover, g},
				pair{"text on a pressed row over " + ground, p.Text, p.SurfacePressed, g},
				pair{"text on a chosen row over " + ground, p.Text, get(roleSelected), g},
				pair{"muted text on a chosen row over " + ground, p.TextMuted, get(roleSelected), g},
			)
		}
		for _, face := range []string{"accent", "accentHover", "accentPressed"} {
			pairs = append(pairs, pair{"accent text on " + face, p.AccentText, get(face), p.Background})
		}
		pairs = append(pairs, pair{"the tooltip's text on its ground", get("onLight"), get("light"), p.Background})
		for tone := range toneCount {
			fg := []core.Color{p.Text, p.Accent, p.Success, p.Warning, p.Danger}[tone]
			pairs = append(pairs, pair{toneNames[tone] + " tone on its fill over the panel", fg, get(roleFill + toneNames[tone]), get(rolePanel)})
		}
		for _, pr := range pairs {
			if ratio := contrast(pr.fg, pr.bg, pr.on); ratio < 4.5 {
				t.Errorf("%s: %s is %.2f:1, under 4.5:1", appearance.name, pr.what, ratio)
			}
		}
		if len(pairs) < 60 {
			t.Errorf("%s: only %d pairs measured", appearance.name, len(pairs))
		}
	}
}

// TestColorsOfFollowsTheAppearance checks that Apply reaches the views: a
// control draws the theme's accent, the kit's colors read the palette shown,
// and a text in a Frame takes the kit's font family.
func TestColorsOfFollowsTheAppearance(t *testing.T) {
	for _, tc := range []struct {
		name       string
		appearance core.Appearance
		look       look
	}{{"dark", core.AppearanceDark, frostyard}, {"light", core.AppearanceLight, frostyardLight}} {
		t.Run(tc.name, func(t *testing.T) {
			var got struct{ chrome, panel, selected, label, fill, line, accent core.Color }
			app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
				if err := Apply(cx); err != nil {
					t.Errorf("Apply: %v", err)
				}
				cx.SetAppearance(tc.appearance)
				return Frame(core.Dynamic(func(cx *core.Scope) core.Node {
					c := ColorsOf(cx)
					got.chrome, got.panel, got.selected, got.label = c.Chrome(), c.Panel(), c.Selected(), c.Label()
					got.fill, got.line, got.accent = c.Fill(Danger), c.Line(Warning), ui.ColorsOf(cx).Accent()
					return core.Text{Content: "Sample", TestID: "sample"}
				}))
			}, nsgotest.Size(400, 300))
			want := tc.look
			if got.chrome != want.chrome || got.panel != want.panel || got.selected != want.selectedFill || got.label != want.label {
				t.Errorf("chrome %#08x panel %#08x selected %#08x label %#08x, want %#08x %#08x %#08x %#08x",
					uint32(got.chrome), uint32(got.panel), uint32(got.selected), uint32(got.label),
					uint32(want.chrome), uint32(want.panel), uint32(want.selectedFill), uint32(want.label))
			}
			if got.fill != want.fill[Danger] || got.line != want.line[Warning] || got.accent != want.accent {
				t.Errorf("danger fill %#08x warning line %#08x accent %#08x, want %#08x %#08x %#08x",
					uint32(got.fill), uint32(got.line), uint32(got.accent), uint32(want.fill[Danger]), uint32(want.line[Warning]), uint32(want.accent))
			}
			if family := app.Find("sample").TextStyle.Font.Family; family != FontFamily {
				t.Errorf("a text in a Frame is set in %q, want %q", family, FontFamily)
			}
		})
	}
}

// TestAppearanceChoiceShowsAndReportsTheChoice presses Dark: the app turns
// dark at once and hears the choice to save.
func TestAppearanceChoiceShowsAndReportsTheChoice(t *testing.T) {
	var saved []core.Appearance
	var ground core.Color
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		if err := Apply(cx); err != nil {
			t.Errorf("Apply: %v", err)
		}
		chosen := core.NewSignal(cx, core.AppearanceLight)
		cx.SetAppearance(core.AppearanceLight)
		return Frame(
			AppearanceChoice{Value: chosen.Get, OnChange: func(a core.Appearance) { chosen.Set(a); saved = append(saved, a) }},
			core.Dynamic(func(cx *core.Scope) core.Node {
				ground = ui.ColorsOf(cx).Background()
				return core.Text{Content: "Sample"}
			}),
		)
	}, nsgotest.Size(400, 300))
	if ground != frostyardLight.background {
		t.Fatalf("the ground starts %#08x, want Frostyard light's %#08x", uint32(ground), uint32(frostyardLight.background))
	}
	app.Press("role:tab Dark")
	if len(saved) != 1 || saved[0] != core.AppearanceDark {
		t.Errorf("OnChange heard %v, want one call with the dark appearance", saved)
	}
	if ground != frostyard.background {
		t.Errorf("after Dark the ground is %#08x, want Frostyard's %#08x", uint32(ground), uint32(frostyard.background))
	}
	if app.Appearance != core.AppearanceDark {
		t.Errorf("the app asked for appearance %v, want dark", app.Appearance)
	}
}

// TestAToneOutOfRangeIsNeutral keeps Fill and Line from reading a role that
// no palette has.
func TestAToneOutOfRangeIsNeutral(t *testing.T) {
	if got := toneName(toneCount + 3); got != "neutral" {
		t.Errorf("toneName past the last tone = %q, want neutral", got)
	}
}
