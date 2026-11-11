package frostui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// The kit's text sizes are given at the platform's usual text size (14 on
// Linux and Windows, 13 on macOS) and scale with the theme's FontSize, so
// the desktop's text size reaches every style (ADR-0011).

// TextSize returns size, in DIPs at the platform's usual text size, scaled
// to the theme's: what a kit style of that size draws at.
func TextSize(c *ui.Context, size float32) float32 {
	base := ui.LightTheme().FontSize
	t := c.Theme()
	if base <= 0 || t.FontSize <= 0 {
		return size
	}
	return size * t.FontSize / base
}

// Title is a page title, tightly spaced as the brand's.
func Title(c *ui.Context, s string) ui.Element {
	size := TextSize(c, 24)
	return ui.Text(c, s).FontSize(size).Bold().LetterSpacing(-0.05 * size)
}

// Heading heads a card or section.
func Heading(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 14.5)).Bold()
}

// Strong is text that stands out in a row: a name.
func Strong(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 13.5)).Bold()
}

// Body is ordinary text.
func Body(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 13.5))
}

// Muted is secondary text.
func Muted(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 12.5)).TextColor(c.Theme().TextMuted)
}

// Caption is small secondary text, as a label above a value.
func Caption(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 11.5)).TextColor(c.Theme().TextMuted)
}

// Mono is text in the monospace font: versions, IDs, commands.
func Mono(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).Font("monospace").FontSize(TextSize(c, 12.5))
}

// Toned is text in a tone's color.
func Toned(c *ui.Context, s string, tone Tone) ui.Element {
	return ui.Text(c, s).FontSize(TextSize(c, 12.5)).TextColor(ToneColor(c.Theme(), tone))
}

// Label is an uppercase monospace label with letter spacing, in the
// theme colors' Label: a section title in a rail or source list. It
// shows s in capitals and keeps s as its accessible name.
func Label(c *ui.Context, s string) ui.Element {
	size := TextSize(c, 9.5)
	return ui.Text(c, strings.ToUpper(s)).Font("monospace").FontSize(size).FontWeight(700).
		LetterSpacing(size * 0.14).TextColor(ColorsOf(c.Theme()).Label).Label(s)
}
