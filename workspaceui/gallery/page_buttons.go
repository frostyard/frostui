package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func buttonsPage(cx *core.Scope) core.Node {
	nothing := func() {}
	off := func() bool { return true }
	filter := core.NewSignal(cx, "all")
	view := core.NewSignal(cx, "week")
	todo := core.NewSignal(cx, true)
	return workspaceui.Stack(
		section("Button variants",
			workspaceui.Row(
				ui.Button{Text: "Primary", OnPress: nothing},
				ui.Button{Text: "Secondary", Variant: ui.ButtonSecondary, OnPress: nothing},
				ui.Button{Text: "Outline", Variant: ui.ButtonOutline, OnPress: nothing},
				ui.Button{Text: "Ghost", Variant: ui.ButtonGhost, OnPress: nothing},
				ui.Button{Text: "Tinted", Variant: ui.ButtonTinted, OnPress: nothing},
				ui.Button{Text: "Danger", Variant: ui.ButtonDanger, OnPress: nothing}),
			caption("Rule 3: one filled button for a window region; Cancel is secondary; a destroying action is danger.")),
		section("With an icon, disabled, and the three sizes",
			workspaceui.Row(
				ui.Button{Text: "Compose", Icon: lucide.SquarePen, OnPress: nothing},
				ui.Button{Text: "Disabled", Disabled: off, OnPress: nothing},
				ui.Button{Text: "Disabled", Variant: ui.ButtonSecondary, Disabled: off, OnPress: nothing},
				ui.Button{Text: "Small", Size: ui.SizeSmall, Variant: ui.ButtonSecondary, OnPress: nothing},
				ui.Button{Text: "Default", Variant: ui.ButtonSecondary, OnPress: nothing},
				ui.Button{Text: "Large", Size: ui.SizeLarge, Variant: ui.ButtonSecondary, OnPress: nothing})),
		section("Icon buttons: ghost, raised, filled, primary, and one selected",
			workspaceui.Row(
				ui.IconButton{Label: "Archive", Icon: lucide.Archive, OnPress: nothing},
				ui.IconButton{Label: "Archive", Icon: lucide.Archive, Variant: ui.IconButtonRaised, OnPress: nothing},
				ui.IconButton{Label: "Archive", Icon: lucide.Archive, Variant: ui.IconButtonFilled, OnPress: nothing},
				ui.IconButton{Label: "Compose", Icon: lucide.SquarePen, Variant: ui.IconButtonPrimary, OnPress: nothing},
				ui.IconButton{Label: "Flag", Icon: lucide.Flag, Selected: func() bool { return true }, OnPress: nothing},
				ui.IconButton{Label: "Delete", Icon: lucide.Trash, Disabled: off, OnPress: nothing})),
		section("A toolbar: groups of icon buttons with a separator",
			ui.Toolbar{Label: "Message", Children: []core.Node{
				ui.ToolbarGroup{Label: "File", Children: []core.Node{
					ui.IconButton{Label: "Archive", Icon: lucide.Archive, OnPress: nothing},
					ui.IconButton{Label: "Delete", Icon: lucide.Trash, OnPress: nothing},
					ui.IconButton{Label: "Junk", Icon: lucide.ShieldAlert, OnPress: nothing}}},
				ui.ToolbarSeparator{},
				ui.ToolbarGroup{Label: "Answer", Children: []core.Node{
					ui.IconButton{Label: "Reply", Icon: lucide.Reply, OnPress: nothing},
					ui.IconButton{Label: "Reply all", Icon: lucide.ReplyAll, OnPress: nothing},
					ui.IconButton{Label: "Forward", Icon: lucide.Forward, OnPress: nothing}}},
				ui.ToolbarSeparator{},
				ui.ToggleButton{Label: "To-Do Bar", Icon: lucide.PanelRight, On: todo.Get, OnChange: todo.Set},
			}}),
		section("A choice among a few: segmented tabs",
			workspaceui.Row(
				ui.Tabs[string]{Variant: ui.TabsSegmented, Label: "Filter", Value: filter.Get, OnChange: filter.Set, Options: []ui.Option[string]{
					{Value: "all", Label: "All"}, {Value: "unread", Label: "Unread"}, {Value: "flagged", Label: "Flagged"}}},
				ui.Tabs[string]{Variant: ui.TabsSegmented, Label: "View", Value: view.Get, OnChange: view.Set, Options: []ui.Option[string]{
					{Value: "day", Label: "Day"}, {Value: "week", Label: "Week"}, {Value: "month", Label: "Month"}}},
				ui.Tabs[string]{Variant: ui.TabsSegmented, Label: "Module", Value: func() string { return "mail" }, OnChange: func(string) {}, Options: []ui.Option[string]{
					{Value: "mail", Label: "Mail", Icon: lucide.Mail}, {Value: "calendar", Label: "Calendar", Icon: lucide.Calendar},
					{Value: "people", Label: "People", Icon: lucide.Users}, {Value: "tasks", Label: "Tasks", Icon: lucide.ListChecks}}})),
	)
}
