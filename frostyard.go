package frostui

import "github.com/egoist/mygo/ui"

// The kit's look is frostyard's brand, in a dark theme (Frostyard) and a
// light one (Frostyard light): ADR-0011. docs/specs/frostyard-theme.md
// lists every value below, and frostyard_test.go measures their contrast.

// ThemeColors holds the colors of the kit's roles that ui.Theme has no field
// for. ColorsOf returns the one that goes with a theme.
type ThemeColors struct {
	// Chrome paints the top bar, the rail, pane bars, the status bar, the
	// console and source lists.
	Chrome ui.Color
	// Panel paints cards, dialogs, the palette and toasts; a card's face
	// is the gradient from PanelFrom to PanelTo.
	Panel, PanelFrom, PanelTo ui.Color
	// SelectedFill and SelectedText are the selected rail item's and
	// source-list row's face and text; the row takes a 1px accent rule on
	// its leading edge.
	SelectedFill, SelectedText ui.Color
	// HoverText is the text and icon color of a hovered button, row or
	// link, whose face does not change.
	HoverText ui.Color
	// GradientFrom and GradientTo are the primary button's face, at 135
	// degrees, and OnGradient its text.
	GradientFrom, GradientTo, OnGradient ui.Color
	// Label colors uppercase section titles, TagText a neutral pill's text,
	// Code the face of fields and the console, and LineStrong the outline
	// of what floats over the window, as a toast.
	Label, TagText, Code, LineStrong ui.Color

	fills, lines [toneCount]ui.Color
}

// toneCount is the number of tones, Neutral to Danger.
const toneCount = int(Danger) + 1

// Fill is the soft face behind tone's foreground, as a banner's.
func (p ThemeColors) Fill(tone Tone) ui.Color { return p.fills[clampTone(tone)] }

// Line is the 1px outline of a pill or badge in tone.
func (p ThemeColors) Line(tone Tone) ui.Color { return p.lines[clampTone(tone)] }

func clampTone(tone Tone) Tone {
	if tone < Neutral || int(tone) >= toneCount {
		return Neutral
	}
	return tone
}

// FrostyardTheme returns the kit's dark theme (Frostyard) or its light one
// (Frostyard light), before the desktop's preferences. Adapt sets the one
// the desktop asks for each frame; an app with its own light or dark
// setting passes its choice here, then the desktop's preferences without
// the accent:
//
//	p := c.Preferences()
//	p.Accent = ui.Color{}
//	c.SetTheme(frostui.FrostyardTheme(dark).WithPreferences(p))
func FrostyardTheme(dark bool) *ui.Theme {
	base := ui.LightTheme()
	t := &ui.Theme{
		Dark:            dark,
		FocusRingWidth:  2,
		FocusRingOffset: 1,
		Radius:          4,
		Spacing:         4,
		FontSize:        base.FontSize,
		Font:            FontFamily,
		TooltipDelay:    base.TooltipDelay,
	}
	if dark {
		t.Background = ui.Hex("#06111d")
		t.Surface = ui.Hex("#0d2a41")
		t.Border = ui.RGBA(162, 222, 255, 0.18)
		t.Text = ui.Hex("#edf8ff")
		t.TextMuted = ui.Hex("#b6d0df")
		t.Accent = ui.Hex("#47b8ef")
		t.AccentText = ui.Hex("#02131f")
		t.Success = ui.Hex("#7edafa")
		t.Warning = ui.Hex("#e0b15c")
		t.Danger = ui.Hex("#e07a72")
		t.Selection = ui.RGBA(71, 184, 239, 0.35)
		t.Focus = ui.Hex("#47b8ef")
		t.Scrollbar = ui.RGBA(162, 222, 255, 0.35)
	} else {
		t.Background = ui.Hex("#f5f9fc")
		t.Surface = ui.Hex("#dfeaf2")
		t.Border = ui.RGBA(6, 17, 29, 0.14)
		t.Text = ui.Hex("#06111d")
		t.TextMuted = ui.Hex("#3f5a6b")
		t.Accent = ui.Hex("#0b6ea6")
		t.AccentText = ui.Hex("#ffffff")
		t.Success = ui.Hex("#076377")
		t.Warning = ui.Hex("#8a5300")
		t.Danger = ui.Hex("#b0322b")
		t.Selection = ui.RGBA(11, 110, 166, 0.25)
		t.Focus = ui.Hex("#0b6ea6")
		t.Scrollbar = ui.RGBA(6, 17, 29, 0.32)
	}
	// The brand changes no face on hover: the kit's own controls shift
	// their text to the theme colors' HoverText. MyGo's widgets (menus, lists,
	// buttons drawn with ui.Button) only change their face, so theirs
	// move a little toward the text, and the accent's too.
	t.SurfaceHover = t.Surface.Mix(t.Text, 0.06)
	t.SurfacePressed = t.Surface.Mix(t.Text, 0.12)
	if dark {
		t.AccentHover = t.Accent.Mix(t.Text, 0.2)
		t.AccentPressed = t.Accent.Mix(t.Background, 0.2)
	} else {
		t.AccentHover = t.Accent.Mix(t.Text, 0.15)
		t.AccentPressed = t.Accent.Mix(t.Text, 0.3)
	}
	t.InactiveAccent = t.Accent
	return t
}

var (
	frostyardDark = ThemeColors{
		Chrome:       ui.Hex("#081b2d"),
		Panel:        ui.Hex("#0d263b"),
		PanelFrom:    ui.Hex("#0f3048"),
		PanelTo:      ui.Hex("#091c2d"),
		SelectedFill: ui.RGBA(71, 184, 239, 0.08),
		SelectedText: ui.Hex("#aee9ff"),
		HoverText:    ui.Hex("#aee9ff"),
		GradientFrom: ui.Hex("#5ac9f7"),
		GradientTo:   ui.Hex("#1584ca"),
		OnGradient:   ui.Hex("#02131f"),
		Label:        ui.Hex("#82b6d0"),
		TagText:      ui.Hex("#b9e9fb"),
		Code:         ui.Hex("#061826"),
		LineStrong:   ui.Hex("#2c6380"),
		fills: [toneCount]ui.Color{
			Neutral: ui.RGBA(162, 222, 255, 0.10),
			Accent:  ui.RGBA(71, 184, 239, 0.08),
			Success: ui.Hex("#0e3247"),
			Warning: ui.RGBA(0x33, 0x29, 0x14, 0.2),
			Danger:  ui.Hex("#3a181e"),
		},
		lines: [toneCount]ui.Color{
			Neutral: ui.RGBA(166, 223, 250, 0.22),
			Accent:  ui.RGBA(71, 184, 239, 0.3),
			Success: ui.RGBA(71, 184, 239, 0.3),
			Warning: ui.RGBA(224, 177, 92, 0.33),
			Danger:  ui.RGBA(224, 122, 114, 0.33),
		},
	}
	frostyardLight = ThemeColors{
		Chrome:       ui.Hex("#e8f0f6"),
		Panel:        ui.Hex("#ffffff"),
		PanelFrom:    ui.Hex("#ffffff"),
		PanelTo:      ui.Hex("#eef5fa"),
		SelectedFill: ui.RGBA(11, 110, 166, 0.10),
		SelectedText: ui.Hex("#064a72"),
		HoverText:    ui.Hex("#064a72"),
		GradientFrom: ui.Hex("#5ac9f7"),
		GradientTo:   ui.Hex("#1584ca"),
		OnGradient:   ui.Hex("#02131f"),
		Label:        ui.Hex("#3b6680"),
		TagText:      ui.Hex("#24485e"),
		Code:         ui.Hex("#ffffff"),
		LineStrong:   ui.Hex("#2c6380"),
		fills: [toneCount]ui.Color{
			Neutral: ui.RGBA(6, 17, 29, 0.06),
			Accent:  ui.RGBA(11, 110, 166, 0.08),
			Success: ui.Hex("#ddf1f6"),
			Warning: ui.Hex("#f8ecd6"),
			Danger:  ui.Hex("#f9e2df"),
		},
		lines: [toneCount]ui.Color{
			Neutral: ui.RGBA(6, 17, 29, 0.22),
			Accent:  ui.RGBA(11, 110, 166, 0.4),
			Success: ui.RGBA(7, 99, 119, 0.4),
			Warning: ui.RGBA(138, 83, 0, 0.4),
			Danger:  ui.RGBA(176, 50, 43, 0.4),
		},
	}
)

// frostyardColors returns the colors of a Frostyard theme, found by its
// background, which neither the desktop's preferences nor a scope that
// recolors the accent (SoftSelection) changes.
func frostyardColors(t *ui.Theme) (ThemeColors, bool) {
	switch {
	case t.Dark && t.Background == ui.Hex("#06111d"):
		return frostyardDark, true
	case !t.Dark && t.Background == ui.Hex("#f5f9fc"):
		return frostyardLight, true
	}
	return ThemeColors{}, false
}

// ColorsOf returns the theme colors that go with t: the Frostyard colors of
// a Frostyard theme, and for any other theme (MyGo's own, or an app's) one
// derived from it as the kit drew before ADR-0011.
func ColorsOf(t *ui.Theme) ThemeColors {
	if p, ok := frostyardColors(t); ok {
		return p
	}
	chrome := t.Background.Mix(t.Surface, 0.55)
	panel := t.Background.Mix(t.Surface, 0.7)
	p := ThemeColors{
		Chrome: chrome, Panel: panel, PanelFrom: panel, PanelTo: panel,
		SelectedFill: t.Accent, SelectedText: t.AccentText, HoverText: t.Text,
		GradientFrom: t.Accent, GradientTo: t.Accent, OnGradient: t.AccentText,
		Label: t.TextMuted, TagText: t.TextMuted, Code: t.Background, LineStrong: t.Border,
	}
	for tone := Neutral; int(tone) < toneCount; tone++ {
		if tone == Neutral {
			p.fills[tone] = t.Border.Alpha(0.6)
			p.lines[tone] = t.Border
			continue
		}
		p.fills[tone] = ToneColor(t, tone).Alpha(0.16)
		p.lines[tone] = ToneColor(t, tone).Alpha(0.35)
	}
	return p
}

// ToneLine is the 1px outline of a pill or badge in tone.
func ToneLine(t *ui.Theme, tone Tone) ui.Color { return ColorsOf(t).Line(tone) }
