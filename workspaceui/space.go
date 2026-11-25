package workspaceui

import "github.com/nsgo-dev/nsgo/core"

// The kit's spacing: the steps an app's paddings and gaps take, so that no
// app writes one as a number (the standard for an app, rule 2). They are
// the steps the kit's own parts and NSGo's controls stand on.
const (
	// SpaceXS is between a mark and what it marks, and inside a small cell.
	SpaceXS core.Points = 4
	// SpaceS is a pane's inset from its edges, and between a bar's
	// controls.
	SpaceS core.Points = 8
	// SpaceM is between a pane's blocks.
	SpaceM core.Points = 12
	// SpaceL is a reading pane's inset.
	SpaceL core.Points = 16
	// SpaceXL is a page's inset.
	SpaceXL core.Points = 24
)

// indentStep is how far a row of a tree stands in from its parent.
const indentStep core.Points = 12

// Indent is the leading inset of a row depth levels inside a tree or an
// outline: 0 at the top.
func Indent(depth int) core.Points { return core.Points(max(depth, 0)) * indentStep }

// Heading is the type of a heading in a document an app draws, such as a
// note's: level 1 the largest, 24 points, then 2 less a level to level 6,
// 14, each semibold. A level outside 1 to 6 is the nearest of them. An
// app's own chrome takes NSGo's type styles, not these.
func Heading(level int) core.Style {
	level = max(1, min(6, level))
	size := core.Points(26 - 2*level)
	return core.Style{FontSize: size, LineHeight: size + 6, FontWeight: core.FontWeightSemibold}
}
