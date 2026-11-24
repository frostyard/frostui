package workspaceui

import (
	_ "embed"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui"
)

// FontFamily is the family the kit sets its text in: Inter, then the
// system's sans-serif face. NSGo finds a registered font by the family its
// file names, and the variable file Apply registers names "Inter Variable";
// "Inter" is a system's own static faces.
const FontFamily = `"Inter Variable", Inter, sans-serif`

// Inter 4.001's upright variable file, under the SIL Open Font License 1.1
// (fonts/OFL.txt).
//
//go:embed fonts/InterVariable.ttf
var interVariable []byte

// Theme returns the Frostyard look as an NSGo theme: Frostyard for the dark
// appearance and Frostyard light for the light one, with the colors of
// docs/specs/frostyard-theme.md. It keeps NSGo's corners. Each call returns
// a theme of its own.
func Theme() *core.Theme {
	t := core.DefaultTheme()
	t.Sans = FontFamily
	t.Dark = palette(t.Dark, frostyard)
	t.Light = palette(t.Light, frostyardLight)
	return t
}

// Apply gives cx, and every view built from it, the Frostyard theme, and
// registers Inter with the platform. An app calls it once at its root,
// before it builds a view, and puts its views in a Frame. The theme is set
// even when it returns an error: the text then falls back to the system's
// sans-serif face.
func Apply(cx *core.Scope) error {
	cx.SetTheme(Theme())
	return cx.RegisterFont(interVariable)
}

// Frame is the root view of a window. It fills the window, paints the
// theme's ground under everything in it, and sets the text's color and the
// kit's font family, which the texts inside it inherit.
func Frame(children ...core.Node) core.Node { return frame{children} }

type frame struct{ children []core.Node }

func (f frame) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	return core.View{TestID: "frame", Style: ui.FillRest.Merge(core.Style{FontFamily: FontFamily}), StyleFunc: func() core.Style {
		return core.Style{Background: colors.Background(), Color: colors.Text()}
	}, Children: f.children}
}

// look holds one appearance's colors, by the names of the themes' spec.
type look struct {
	background, surface, text, textMuted          core.Color
	border, scrollbar                             core.Color
	accent, accentHover, accentPressed, accentTxt core.Color
	success, warning, danger                      core.Color
	selection                                     core.Color
	chrome, panel, code                           core.Color
	selectedFill, selectedText, label, knob       core.Color
	fill, line                                    [toneCount]core.Color
}

var frostyard = look{
	background: core.RGB(0x06111d), surface: core.RGB(0x0d2a41),
	text: core.RGB(0xedf8ff), textMuted: core.RGB(0xb6d0df),
	border: core.RGBA(0xa2deff, 0.18), scrollbar: core.RGBA(0xa2deff, 0.35),
	accent: core.RGB(0x47b8ef), accentHover: core.RGB(0x68c5f2), accentPressed: core.RGB(0x3a97c5), accentTxt: core.RGB(0x02131f),
	success: core.RGB(0x7edafa), warning: core.RGB(0xe0b15c), danger: core.RGB(0xe07a72),
	selection: core.RGBA(0x47b8ef, 0.35),
	chrome:    core.RGB(0x081b2d), panel: core.RGB(0x0d263b), code: core.RGB(0x061826),
	selectedFill: core.RGBA(0x47b8ef, 0.08), selectedText: core.RGB(0xaee9ff), label: core.RGB(0x82b6d0), knob: core.RGB(0xedf8ff),
	fill: [toneCount]core.Color{
		Neutral: core.RGBA(0xa2deff, 0.10), Accent: core.RGBA(0x47b8ef, 0.08),
		Success: core.RGB(0x0e3247), Warning: core.RGBA(0x332914, 0.2), Danger: core.RGB(0x3a181e),
	},
	line: [toneCount]core.Color{
		Neutral: core.RGBA(0xa6dffa, 0.22), Accent: core.RGBA(0x47b8ef, 0.3),
		Success: core.RGBA(0x47b8ef, 0.3), Warning: core.RGBA(0xe0b15c, 0.33), Danger: core.RGBA(0xe07a72, 0.33),
	},
}

var frostyardLight = look{
	background: core.RGB(0xf5f9fc), surface: core.RGB(0xdfeaf2),
	text: core.RGB(0x06111d), textMuted: core.RGB(0x3f5a6b),
	border: core.RGBA(0x06111d, 0.14), scrollbar: core.RGBA(0x06111d, 0.32),
	accent: core.RGB(0x0b6ea6), accentHover: core.RGB(0x0a6091), accentPressed: core.RGB(0x0a527d), accentTxt: core.RGB(0xffffff),
	success: core.RGB(0x076377), warning: core.RGB(0x8a5300), danger: core.RGB(0xb0322b),
	selection: core.RGBA(0x0b6ea6, 0.25),
	chrome:    core.RGB(0xe8f0f6), panel: core.RGB(0xffffff), code: core.RGB(0xffffff),
	selectedFill: core.RGBA(0x0b6ea6, 0.10), selectedText: core.RGB(0x064a72), label: core.RGB(0x3b6680), knob: core.RGB(0xffffff),
	fill: [toneCount]core.Color{
		Neutral: core.RGBA(0x06111d, 0.06), Accent: core.RGBA(0x0b6ea6, 0.08),
		Success: core.RGB(0xddf1f6), Warning: core.RGB(0xf8ecd6), Danger: core.RGB(0xf9e2df),
	},
	line: [toneCount]core.Color{
		Neutral: core.RGBA(0x06111d, 0.22), Accent: core.RGBA(0x0b6ea6, 0.4),
		Success: core.RGBA(0x076377, 0.4), Warning: core.RGBA(0x8a5300, 0.4), Danger: core.RGBA(0xb0322b, 0.4),
	},
}

// palette lays l over p, NSGo's default palette of the same appearance:
// NSGo's roles, the named roles package ui's controls read, shadcn/ui's
// tokens that classes name, and the kit's own roles (ColorsOf).
func palette(p core.Palette, l look) core.Palette {
	p.Background, p.Surface = l.background, l.surface
	// Hover and pressed faces are the text's color laid over the ground, so
	// they show on the chrome and on a panel as on a surface. Over Surface
	// they are the spec's SurfaceHover and SurfacePressed.
	p.SurfaceHover, p.SurfacePressed = ui.Alpha(l.text, 0.06), ui.Alpha(l.text, 0.12)
	p.Border, p.Text, p.TextMuted = l.border, l.text, l.textMuted
	p.Accent, p.AccentHover, p.AccentPressed, p.AccentText = l.accent, l.accentHover, l.accentPressed, l.accentTxt
	p.Danger, p.Warning, p.Success = l.danger, l.warning, l.success
	p.Selection, p.Focus, p.Scrollbar = l.selection, l.accent, l.scrollbar

	roles := map[string]core.Color{
		// Package ui's roles.
		"inset": l.chrome, "raised": l.panel, "field": l.code, "fieldBorder": l.line[Neutral],
		"line": l.border, "faint": l.label, "accentFill": l.accent,
		"light": l.text, "onLight": l.background, "track": ui.Alpha(l.text, 0.25), "knob": l.knob,
		"code": l.text, "info": l.accent, "down": l.danger,
		// shadcn/ui's tokens.
		"foreground": l.text, "card": l.panel, "card-foreground": l.text, "popover": l.panel, "popover-foreground": l.text,
		"primary": l.accent, "primary-foreground": l.accentTxt, "secondary": l.surface, "secondary-foreground": l.text,
		"muted": l.surface, "muted-foreground": l.textMuted, "accent-foreground": l.accentTxt, "destructive": l.danger,
		"input": l.line[Neutral], "ring": l.accent,
		"sidebar": l.chrome, "sidebar-foreground": l.text, "sidebar-primary": l.accent, "sidebar-primary-foreground": l.accentTxt,
		"sidebar-accent": l.selectedFill, "sidebar-accent-foreground": l.text, "sidebar-border": l.border, "sidebar-ring": l.accent,
		// The kit's roles.
		roleChrome: l.chrome, rolePanel: l.panel, roleSelected: l.selectedFill, roleSelectedText: l.selectedText, roleLabel: l.label,
	}
	for tone := range toneCount {
		roles[roleFill+toneNames[tone]] = l.fill[tone]
		roles[roleLine+toneNames[tone]] = l.line[tone]
	}
	for name, c := range roles {
		p.SetColor(name, c)
	}
	return p
}
