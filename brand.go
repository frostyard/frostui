package frostui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// The kit's components in the brand (ADR-0011): cards on the panel
// gradient, the primary face on the brand gradient, the selected row on a
// tint with an accent rule, and outlined pills. Each helper takes its
// colors from ColorsOf, so a theme that is not Frostyard gets the kit's
// former look where ColorsOf derives it.

// panelFace fills e as a card's face: the panel gradient, at 160 degrees
// to 62% of the way, over the panel color.
func panelFace(e ui.Element, t *ui.Theme) ui.Element {
	k := ColorsOf(t)
	e = e.Background(k.Panel)
	if k.PanelFrom != k.PanelTo {
		e = e.LinearGradient(ui.LinearGradient{From: k.PanelFrom, To: k.PanelTo, Angle: 160, End: 0.62})
	}
	return e
}

// primaryFace fills e as the primary button's face, the brand gradient at
// 135 degrees, with its text color.
func primaryFace(e ui.Element, t *ui.Theme) ui.Element {
	k := ColorsOf(t)
	return e.LinearGradient(ui.LinearGradient{From: k.GradientFrom, To: k.GradientTo, Angle: 135}).TextColor(k.OnGradient)
}

// selectedRow draws e as the selected row of a rail, source list or the
// palette: the selected fill, a 1px accent rule on its leading edge and
// the selected text.
func selectedRow(e ui.Element, t *ui.Theme) ui.Element {
	k := ColorsOf(t)
	return e.Background(k.SelectedFill).BorderWidth(0, 0, 0, 1).BorderColor(t.Accent).TextColor(k.SelectedText)
}

// railTheme is t for a rail's items, whose chosen item MyGo draws with the
// accent's text color: the selected text, so that its icon and title read
// on the selected fill.
func railTheme(t *ui.Theme) *ui.Theme {
	r := *t
	r.AccentText = ColorsOf(t).SelectedText
	return &r
}

// railItem is a rail's item for id: square, the selected row when it is
// selected, and the hover text while the pointer is over it.
func railItem(c *ui.Context, t *ui.Theme, id, selected, icon, title string) ui.Element {
	it := ui.SidebarItem(c.Key(id), id, IconSVG(icon), title).Radius(0)
	if id == selected {
		selectedRow(it, t)
	} else if it.Hovered() {
		it.TextColor(ColorsOf(t).HoverText)
	}
	return it
}

// pillText is a pill's label: the text in capitals, monospace, bold and
// letter-spaced, named as written.
func pillText(c *ui.Context, text string, color ui.Color) ui.Element {
	size := TextSize(c, 9)
	return ui.Text(c, strings.ToUpper(text)).Font("monospace").FontSize(size).FontWeight(700).
		LetterSpacing(size * 0.08).TextColor(color).SingleLine().Label(text)
}

// outlinedPill is a pill: a 1px outline in line around lead, if any, and
// text in fg.
func outlinedPill(c *ui.Context, text string, fg, line ui.Color, lead func()) ui.Element {
	return ui.Row(c).Gap(5).Padding(3, 6).Radius(RadiusChip).Border(1, line).AlignItems(ui.Center).Children(func() {
		if lead != nil {
			lead()
		}
		pillText(c, text, fg)
	})
}

// pillColor is the text color of a pill in tone: the tag text for Neutral.
func pillColor(t *ui.Theme, tone Tone) ui.Color {
	if tone == Neutral {
		return ColorsOf(t).TagText
	}
	return ToneColor(t, tone)
}

// selectedPill is a pill on a selected row, in the selected text, whatever
// its tone: a tone's color can be too faint on the selected fill
// (frostyard/frostui#15).
func selectedPill(c *ui.Context, text string, t *ui.Theme) ui.Element {
	fg := ColorsOf(t).SelectedText
	return outlinedPill(c, text, fg, fg.Alpha(0.4), nil)
}

// ghost makes b a ghost button: no face of its own, so the hairline MyGo
// draws is all of it, and its text in the hover text while the pointer is
// over it. Selected, called after, still gives it the accent's face.
func ghost(c *ui.Context, b ui.Element) ui.Element {
	b.Background(ui.Transparent)
	if b.Hovered() {
		b.TextColor(ColorsOf(c.Theme()).HoverText)
	}
	return b
}

// choiceTheme is t for a list whose chosen row MyGo paints with the accent
// and the accent's text: the selected fill, over the chrome, and the
// selected text.
func choiceTheme(t *ui.Theme) *ui.Theme {
	k := ColorsOf(t)
	r := *t
	r.Accent = k.SelectedFill.Over(k.Chrome)
	r.InactiveAccent = r.Accent
	r.AccentText = k.SelectedText
	return &r
}

// railLabel heads a group of a rail's items: the brand's uppercase label,
// with room above it unless it is first.
func railLabel(c *ui.Context, title string, first bool) ui.Element {
	top := float32(16)
	if first {
		top = 4
	}
	return Label(c, title).Padding(top, 10, 5, 10).Role(ui.RoleHeading)
}
