package frostui

import "github.com/egoist/mygo/ui"

// Metrics of the kit, in DIPs. Pages use them instead of numbers of their
// own so that every frostyard app has the same rhythm.
const (
	SpaceXS = 4
	SpaceS  = 8
	SpaceM  = 12
	SpaceL  = 16
	SpaceXL = 24

	// RadiusCard rounds cards and panels; RadiusChip rounds chips, pills,
	// fields, banners and small tiles. The brand is nearly square
	// (ADR-0011): buttons take the theme's Radius, 4.
	RadiusCard = 0
	RadiusChip = 2

	// RailWidth is the width of the left rail.
	RailWidth = 232
	// TopBarHeight is the height of the top bar.
	TopBarHeight = 52
)

// Tone is the meaning a color carries.
type Tone int

const (
	// Neutral is the default: muted text on the surface.
	Neutral Tone = iota
	// Accent is the desktop's accent.
	Accent
	// Success is what went well or is on.
	Success
	// Warning is what needs attention.
	Warning
	// Danger is a failure or a destructive action.
	Danger
)

// ToneColor is the foreground color of tone.
func ToneColor(t *ui.Theme, tone Tone) ui.Color {
	switch tone {
	case Accent:
		return t.Accent
	case Success:
		return t.Success
	case Warning:
		return t.Warning
	case Danger:
		return t.Danger
	}
	return t.TextMuted
}

// ToneFill is the soft background behind tone's foreground, as a pill's.
func ToneFill(t *ui.Theme, tone Tone) ui.Color { return ColorsOf(t).Fill(tone) }

// Adapt sets the kit's theme for this frame and returns it: Frostyard when
// MyGo picked a dark theme for the desktop, Frostyard light otherwise
// (ADR-0011). The desktop's high contrast and text size still apply, but
// not its accent: the accent is the brand's, and it stays in windows
// without the focus. A theme that is already Frostyard (an app chose one
// itself, or Adapt ran earlier in the frame) is kept. MyGo resets the theme
// at each build pass, so call it first in the root view; Shell.View and
// PaneShell.View do. The first call registers the kit's font.
func Adapt(c *ui.Context) *ui.Theme {
	registerInter()
	t := c.Theme()
	if _, ok := frostyardColors(t); ok {
		return t
	}
	p := c.Preferences()
	p.Accent = ui.Color{}
	ft := FrostyardTheme(t.Dark).WithPreferences(p)
	c.SetTheme(ft)
	return ft
}

// Chrome is the color of the top bar, rail, pane bars and status bar.
func Chrome(t *ui.Theme) ui.Color { return ColorsOf(t).Chrome }

// Panel is the face of cards.
func Panel(t *ui.Theme) ui.Color { return ColorsOf(t).Panel }
