package main

import (
	"fmt"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui"
)

// swatch is a color of the palette shown with its name and its value.
func swatch(name string, c func() core.Color) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		return core.View{Style: core.Style{Width: core.Pt(150), Gap: 4}, Children: []core.Node{
			core.View{Style: core.Style{Height: core.Pt(44), CornerRadius: ui.RadiusControl, BorderWidth: 1}, StyleFunc: func() core.Style {
				return core.Style{Background: c(), BorderColor: colors.Border()}
			}},
			core.Text{Content: name, Style: ui.TypeCaption},
			core.Text{ContentFunc: func() string { return fmt.Sprintf("#%08x", uint32(c())) }, Style: ui.TypeCode, StyleFunc: ui.Colored(colors.Muted)},
		}}
	})
}

func colorsPage(cx *core.Scope) core.Node {
	c, k := ui.ColorsOf(cx), workspaceui.ColorsOf(cx)
	tone := func(name string, t workspaceui.Tone, fg func() core.Color) core.Node {
		return core.View{Style: core.Style{Padding: core.Symmetric(8, 12), CornerRadius: ui.RadiusControl, BorderWidth: 1, Width: core.Pt(150)}, StyleFunc: func() core.Style {
			return core.Style{Background: k.Fill(t), BorderColor: k.Line(t)}
		}, Children: []core.Node{core.Text{Content: name, Style: ui.TypeLabel, StyleFunc: ui.Colored(fg)}}}
	}
	return workspaceui.Stack(
		section("Grounds",
			workspaceui.Row(swatch("background", c.Background), swatch("chrome (inset)", k.Chrome), swatch("panel (raised)", k.Panel), swatch("surface", c.Surface), swatch("field", c.Field))),
		section("Text",
			workspaceui.Row(swatch("text", c.Text), swatch("muted", c.Muted), swatch("label (faint)", k.Label), swatch("selected text", k.SelectedText))),
		section("Accent and states",
			workspaceui.Row(swatch("accent", c.Accent), swatch("on accent", c.OnAccent), swatch("success", c.Up), swatch("warning", c.Warn), swatch("danger", c.Danger))),
		section("Lines and fills",
			workspaceui.Row(swatch("border", c.Border), swatch("hover", c.Hover), swatch("pressed", c.Pressed), swatch("selected row", k.Selected), swatch("selection", func() core.Color { return cx.Palette().Selection }))),
		section("Tones: a fill, its line and its text",
			workspaceui.Row(tone("Neutral", workspaceui.Neutral, c.Text), tone("Accent", workspaceui.Accent, c.Accent), tone("Success", workspaceui.Success, c.Up), tone("Warning", workspaceui.Warning, c.Warn), tone("Danger", workspaceui.Danger, c.Danger))),
		section("Type, NSGo's scale in Inter",
			core.Text{Content: "Heading 20 / 26 semibold", Style: ui.TypeHeading},
			core.Text{Content: "Title 15 / 20 semibold", Style: ui.TypeTitle},
			core.Text{Content: "Label 13 / 18 medium", Style: ui.TypeLabel},
			core.Text{Content: "Body 13 / 18. The load test finished overnight: 50,000 accounts, no errors.", Style: ui.TypeBody},
			core.Text{Content: "Caption 12 / 16. Yesterday at 6:30 AM", Style: ui.TypeCaption, StyleFunc: ui.Colored(c.Muted)},
			core.Text{Content: "Code 12 / 18  workspaceui.Apply(cx)", Style: ui.TypeCode}),
	)
}
