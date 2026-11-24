package workspaceui

import (
	"strconv"

	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// Section groups cards under a full-width disclosure. Its content is built
// only while open, so a folded archive costs a single header.
type Section struct {
	Title         string
	Count         func() int
	InitiallyOpen bool
	Content       func(*core.Scope) core.Node
}

func (s Section) Render(cx *core.Scope) core.Node {
	open := core.NewSignal(cx, s.InitiallyOpen)
	colors := ui.ColorsOf(cx)
	chevron := func() *core.SVG {
		if open.Get() {
			return lucide.ChevronDown()
		}
		if cx.Env().RTL() {
			return lucide.ChevronLeft()
		}
		return lucide.ChevronRight()
	}
	header := base.Button{Label: s.Title, Expanded: open.Get,
		OnPress: func(core.PressEvent) { open.Set(!open.Get()) },
		Style:   core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 10, Padding: core.Symmetric(8, 4), CornerRadius: ui.RadiusControl},
		StyleFunc: func() core.Style {
			return core.Style{FocusVisible: ui.FocusRing(cx, -2), Hover: &core.Style{Background: colors.Hover()}}
		},
		Children: []core.Node{
			ui.IconOf(chevron, 18, colors.Muted),
			core.Text{Content: s.Title, Style: ui.TypeHeading.Merge(core.Style{FlexShrink: 1})},
			ui.Badge{TextFunc: func() string { return strconv.Itoa(s.Count()) }},
		},
	}
	return Stack(base.AccordionHeader{Level: 2, Children: []core.Node{header}}, core.Show{When: open.Get, Then: s.Content})
}

// Card keeps one item's content and actions together on a themed surface.
type Card struct{ Children []core.Node }

func (c Card) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	return core.View{Style: core.Style{Padding: core.All(16), Gap: 8, CornerRadius: 12, BorderWidth: 1},
		StyleFunc: func() core.Style {
			return core.Style{Background: colors.Raised(), BorderColor: ui.Mix(colors.Border(), colors.Text(), 0.12)}
		}, Children: c.Children}
}

// InspectorGroup lays out labeled values and their actions on one surface.
func InspectorGroup(label string, children ...core.Node) core.Node {
	return Stack(core.Text{Content: label, Style: ui.TypeLabel}, Card{Children: children})
}

func InspectorRow(label string, value func() string, actions ...core.Node) core.Node {
	return Stack(core.Text{Content: label, Style: ui.TypeCaption}, Row(append([]core.Node{
		core.Text{ContentFunc: value, Selectable: core.SelectableOn, Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0)}},
	}, actions...)...))
}
