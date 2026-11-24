package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func rowsPage(cx *core.Scope) core.Node {
	chosen := core.NewSignal(cx, "inbox")
	selection := ui.NewSelection(cx, true)
	selection.Set(3)
	lines := core.NewSignal(cx, 2)
	return workspaceui.Stack(
		section("Navigation rows: NSGo's sidebar, with a role's icon and a count at the row's end",
			core.View{Style: core.Style{Width: core.Pt(246), CornerRadius: ui.RadiusCard, Padding: core.All(6)}, StyleFunc: ui.Filled(workspaceui.ColorsOf(cx).Chrome), Children: []core.Node{
				ui.Sidebar{Selected: chosen.Get, OnSelect: chosen.Set, Sections: []ui.SidebarSection{
					{Title: "Favorites", Items: navItems(favorites[:2])},
					{Title: "ann.lee@example.com", Collapsible: true, Items: navItems(account)}}}}}),
		section("Message rows: unread, a thread, a reminder, a flag, an attachment; the fourth is chosen",
			workspaceui.Row(core.Text{Content: "Preview lines", Style: ui.TypeCaption}, ui.Tabs[int]{Variant: ui.TabsSegmented, Size: ui.SizeSmall, Label: "Preview lines", Value: lines.Get, OnChange: lines.Set,
				Options: []ui.Option[int]{{Value: 0, Label: "0"}, {Value: 1, Label: "1"}, {Value: 2, Label: "2"}}}),
			core.View{Style: core.Style{Width: core.Pt(390), Height: core.Pt(560)}, Children: []core.Node{core.Dynamic(func(cx *core.Scope) core.Node {
				n := lines.Get()
				return ui.List{Label: "Messages", Style: ui.FillRest, EstimatedSize: 84, Gap: 2, Selection: selection, TapSelects: true,
					Count:        func() int { return len(inbox) },
					RowStyleFunc: softRows(cx),
					Row:          func(_ *core.Scope, row core.ListRow) core.Node { return summaryRow(inbox[row.Index()], n) }}
			})}}),
		section("Marks: avatars, badges and a loading row",
			workspaceui.Row(ui.Avatar{Name: "David King", Size: 24}, ui.Avatar{Name: "Skyline Travel", Size: 32}, ui.Avatar{Name: "Iris Young", Size: 40},
				ui.Badge{Text: "3"}, ui.Badge{Text: "Draft", Variant: ui.BadgeAccent}, ui.Badge{Text: "Sent", Variant: ui.BadgeSuccess}, ui.Badge{Text: "Offline", Variant: ui.BadgeWarn}, ui.Badge{Text: "Failed", Variant: ui.BadgeDanger}),
			core.View{Style: core.Style{Width: core.Pt(390)}, Children: []core.Node{ui.Skeleton{Lines: 3, Avatar: true, Label: "Loading messages"}}}),
		section("An empty pane",
			core.View{Style: core.Style{Height: core.Pt(180), CornerRadius: ui.RadiusCard, BorderWidth: 1}, StyleFunc: func() core.Style {
				return core.Style{BorderColor: ui.ColorsOf(cx).Border()}
			}, Children: []core.Node{emptyState(lucide.MailOpen, "No message selected", "Choose a message to read it here.")}}),
	)
}
