package workspaceui

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// BranchLabel is the content of a sidebar group row. A nil Expanded reserves
// the chevron's space for a group that has no children.
type BranchLabel struct {
	Text     string
	Icon     func() *core.SVG
	Expanded func() bool
	Trailing core.Node
}

func (b BranchLabel) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	var arrow core.Node = core.View{Style: core.Style{Width: core.Pt(14)}}
	if b.Expanded != nil {
		arrow = ui.IconOf(func() *core.SVG {
			if b.Expanded() {
				return lucide.ChevronDown()
			}
			if cx.Env().RTL() {
				return lucide.ChevronLeft()
			}
			return lucide.ChevronRight()
		}, 14, colors.Text)
	}
	label := ui.Text(b.Text, 13, ui.Semibold, colors.Text)
	label.MaxLines, label.Style.FlexShrink = 1, 1
	label.Style.FlexGrow = 1
	parts := []core.Node{arrow, ui.Icon(b.Icon, 16, colors.Accent), label}
	if b.Trailing != nil {
		parts = append(parts, b.Trailing)
	}
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, FlexGrow: 1, FlexShrink: 1, Gap: 6},
		Children: parts}
}

// ChildLabel indents the entire row content and draws a connector back to
// its parent's icon. Last ends the guide at this child's midpoint.
type ChildLabel struct {
	Text     string
	Icon     func() *core.SVG
	Depth    int
	Last     bool
	Trailing core.Node
}

func (r ChildLabel) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	line := ui.Filled(func() core.Color { return ui.Mix(colors.Border(), colors.Text(), 0.25) })
	height := core.Pt(32)
	if r.Last {
		height = core.Pt(16)
	}
	guide := core.View{Style: core.Style{Width: core.Pt(9), Height: core.Pt(32), JustifyContent: core.JustifyCenter},
		Children: []core.Node{
			core.View{Style: core.Style{Position: core.Absolute, Top: core.Pt(0), Width: core.Pt(1), Height: height}, StyleFunc: line},
			core.View{Style: core.Style{Height: core.Pt(1)}, StyleFunc: line},
		}}
	label := ui.Text(r.Text, 13, ui.Regular, colors.Text)
	label.MaxLines, label.Style.FlexShrink = 1, 1
	label.Style.FlexGrow = 1
	parts := []core.Node{ui.Icon(r.Icon, 14, colors.Muted), label}
	if r.Depth > 0 {
		parts = append([]core.Node{guide}, parts...)
	}
	if r.Trailing != nil {
		parts = append(parts, r.Trailing)
	}
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 6, FlexGrow: 1, FlexShrink: 1}, StyleFunc: func() core.Style {
		if r.Depth == 0 {
			return core.Style{}
		}
		indent := core.Points(8 + 20*r.Depth)
		if cx.Env().RTL() {
			return core.Style{Padding: core.Edges{Right: indent}}
		}
		return core.Style{Padding: core.Edges{Left: indent}}
	}, Children: parts}
}
