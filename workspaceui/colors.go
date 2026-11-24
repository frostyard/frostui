package workspaceui

import "github.com/nsgo-dev/nsgo/core"

// Tone is the meaning a fill or a line carries: none, the accent, or one of
// the three states.
type Tone uint8

const (
	Neutral Tone = iota
	Accent
	Success
	Warning
	Danger
	toneCount
)

var toneNames = [toneCount]string{Neutral: "neutral", Accent: "accent", Success: "success", Warning: "warning", Danger: "danger"}

// The names of the kit's roles in a palette's Colors. A class can name one
// too, as in "bg-frost-chrome".
const (
	roleChrome       = "frost-chrome"
	rolePanel        = "frost-panel"
	roleSelected     = "frost-selected"
	roleSelectedText = "frost-selected-text"
	roleLabel        = "frost-label"
	roleFill         = "frost-fill-"
	roleLine         = "frost-line-"
)

// Colors reads the Frostyard colors NSGo has no role for from cx's palette,
// in the appearance shown. Read them inside a StyleFunc, so that a view
// follows a change of appearance. Under a theme other than Theme they are
// the zero color.
type Colors struct{ cx *core.Scope }

// ColorsOf returns the kit's colors for cx. NSGo's own roles are
// ui.ColorsOf's.
func ColorsOf(cx *core.Scope) Colors { return Colors{cx} }

func (c Colors) role(name string) core.Color { return c.cx.Palette().Colors[name] }

// Chrome is the ground of the window around its content: the navigation
// and the toolbar.
func (c Colors) Chrome() core.Color { return c.role(roleChrome) }

// Panel is the ground of a card, a menu and a dialog.
func (c Colors) Panel() core.Color { return c.role(rolePanel) }

// Selected is the fill of a chosen row. The row keeps its own text colors.
func (c Colors) Selected() core.Color { return c.role(roleSelected) }

// SelectedText is the color of a chosen item that changes its text, such as
// a tab.
func (c Colors) SelectedText() core.Color { return c.role(roleSelectedText) }

// Label is the color of a section's caption.
func (c Colors) Label() core.Color { return c.role(roleLabel) }

// Fill is the ground of a banner, a tag or a badge of tone t.
func (c Colors) Fill(t Tone) core.Color { return c.role(roleFill + toneName(t)) }

// Line is the hairline around a fill of tone t.
func (c Colors) Line(t Tone) core.Color { return c.role(roleLine + toneName(t)) }

func toneName(t Tone) string {
	if t >= toneCount {
		t = Neutral
	}
	return toneNames[t]
}
