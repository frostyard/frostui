package main

import (
	"fmt"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
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

// tokensPage shows what an app takes from the kit that is not a part: the
// flags' colors, the ground and the frame of a page that is not the app's
// own, and the spacing steps.
func tokensPage(cx *core.Scope) core.Node {
	k := workspaceui.ColorsOf(cx)
	return workspaceui.Stack(
		section("Flags: 7 colors, each a little lighter in the dark theme",
			workspaceui.Row(flagSwatches(k)...)),
		section("A page that is not the app's own",
			workspaceui.Row(swatch("page", k.Page)),
			core.View{Style: core.Style{Height: core.Pt(96)}, Children: []core.Node{workspaceui.PageFrame{Size: ui.FillRest, Page: core.View{
				Style:    ui.FillRest.Merge(core.Style{Padding: core.All(workspaceui.SpaceL)}),
				Children: []core.Node{core.Text{Content: "A message's page keeps its sender's colors, inside a hairline, in both themes.", Style: ui.TypeBody, StyleFunc: ui.Colored(func() core.Color { return frostText })}}}}}}),
		section("A row at an app's largest text size: its text is 1.3 times the usual, and the avatar keeps its size",
			core.View{TestID: "scaled-row", Style: core.Style{Width: core.Pt(390)}, Children: []core.Node{workspaceui.SummaryRow{Unread: true, Name: "David King", Title: "David King", When: "8:41 AM",
				Subject: "Re: Aurora launch checklist", Preview: "The load test finished overnight: 50,000 accounts, no errors.", PreviewLines: 2, Scale: 1.3}}}),
		section("Spacing: the steps a padding or a gap takes",
			workspaceui.Row(step("SpaceXS 4", workspaceui.SpaceXS), step("SpaceS 8", workspaceui.SpaceS), step("SpaceM 12", workspaceui.SpaceM), step("SpaceL 16", workspaceui.SpaceL), step("SpaceXL 24", workspaceui.SpaceXL))),
	)
}

// frostText is the ink of the gallery's sample page: a page is its
// sender's, so its text is not the theme's.
var frostText = core.RGB(0x06111d)

// flagSwatches are the 7 flags as a flag in its color over its name.
func flagSwatches(k workspaceui.Colors) []core.Node {
	names := []string{"Red", "Orange", "Yellow", "Green", "Blue", "Purple", "Gray"}
	var out []core.Node
	for n, name := range names {
		out = append(out, core.View{TestID: "flag-" + name, Style: core.Style{Width: core.Pt(84), Gap: workspaceui.SpaceXS, AlignItems: core.AlignCenter}, Children: []core.Node{
			ui.Icon(lucide.Flag, 20, func() core.Color { return k.Flag(n + 1) }),
			core.Text{Content: name, Style: ui.TypeCaption},
		}})
	}
	return out
}

// step is a spacing step drawn as a bar of its width.
func step(name string, w core.Points) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		return core.View{Style: core.Style{Width: core.Pt(120), Gap: workspaceui.SpaceXS}, Children: []core.Node{
			core.View{Style: core.Style{Width: core.Pt(w), Height: core.Pt(16), CornerRadius: ui.RadiusSmall}, StyleFunc: ui.Filled(colors.Accent)},
			core.Text{Content: name, Style: ui.TypeCaption},
		}}
	})
}
