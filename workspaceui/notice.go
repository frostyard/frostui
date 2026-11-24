package workspaceui

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// Notice displays one dismissible message above the workspace. Its overlay
// measures the card in normal flow, including wrapped text on narrow screens.
type Notice struct {
	Kind          ui.ToastType
	Title, Detail string
	OnDismiss     func()
}

func (n Notice) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	icon, tone := lucide.Info, colors.Info
	switch n.Kind {
	case ui.ToastError:
		icon, tone = lucide.CircleAlert, colors.Danger
	case ui.ToastWarning:
		icon, tone = lucide.TriangleAlert, colors.Warn
	case ui.ToastSuccess:
		icon, tone = lucide.CircleCheck, colors.Up
	}
	text := []core.Node{core.Text{Content: n.Title, Style: ui.TypeLabel}}
	label := n.Title
	if n.Detail != "" {
		label += ". " + n.Detail
		text = append(text, core.Text{Content: n.Detail, Selectable: core.SelectableOn, Style: ui.TypeCaption})
	}
	card := core.View{Role: core.RoleAlert, Label: label, LiveRegion: core.LiveRegionPolite, Style: core.Style{
		Direction: core.Row, AlignItems: core.AlignStart, Gap: 10, Width: core.Pt(380), MaxWidth: core.Pct(100),
		Padding: core.All(12), CornerRadius: ui.RadiusCard, BorderWidth: 1,
	}, StyleFunc: func() core.Style {
		accent := tone()
		return core.Style{Background: ui.Mix(colors.Raised(), accent, 0.085), BorderColor: ui.Mix(colors.Border(), accent, 0.48), Color: colors.Text(), Shadows: ui.FinOf(cx).Glass().Shadows}
	}, Children: []core.Node{
		ui.IconOf(icon, 20, tone),
		core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0), Gap: 4}, Children: text},
		ui.IconButton{Label: "Dismiss message", Icon: lucide.X, OnPress: n.OnDismiss},
	}}
	return core.Overlay{PointerEvents: core.PointerEventsBoxNone, Style: core.Style{
		Top: core.Pt(0), Left: core.Pt(0), Width: core.Pct(100), AlignItems: core.AlignEnd,
		Padding: core.Edges{Top: 56, Left: 12, Right: 12}, PaddingEnv: core.EnvEdges{Top: core.EnvSafeArea},
	}, Children: []core.Node{card}}
}
