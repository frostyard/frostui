package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func overlaysPage(cx *core.Scope) core.Node {
	confirm := core.NewSignal(cx, false)
	form := core.NewSignal(cx, false)
	name := core.NewSignal(cx, "Receipts")
	closes := func(s *core.Signal[bool]) func(bool, base.Reason) {
		return func(open bool, _ base.Reason) { s.Set(open) }
	}
	nothing := func() {}
	return workspaceui.Stack(
		section("A confirmation: NSGo's alert dialog",
			workspaceui.Row(ui.Button{Text: "Delete message…", Variant: ui.ButtonSecondary, OnPress: func() { confirm.Set(true) }}),
			caption("Cancel is secondary and comes first; the destroying action is the danger variant.")),
		section("A form in a dialog",
			workspaceui.Row(ui.Button{Text: "New mailbox…", Variant: ui.ButtonSecondary, OnPress: func() { form.Set(true) }})),
		section("A menu with icons, shortcuts, a separator and a danger item",
			workspaceui.Row(ui.Menu{Label: "Message actions", Trigger: ui.Button{Text: "More", Icon: lucide.Ellipsis, Variant: ui.ButtonSecondary}, Children: []core.Node{
				ui.MenuItem{Text: "Mark as unread", Icon: lucide.MailOpen, Shortcut: "mod+shift+u", OnSelect: nothing},
				ui.MenuItem{Text: "Flag", Icon: lucide.Flag, Shortcut: "mod+shift+l", OnSelect: nothing},
				ui.Submenu{Text: "Move to", Children: []core.Node{ui.MenuItem{Text: "Archive", OnSelect: nothing}, ui.MenuItem{Text: "Receipts", OnSelect: nothing}}},
				ui.MenuSeparator{},
				ui.MenuItem{Text: "Show headers", Icon: lucide.Code, OnSelect: nothing},
				ui.MenuSeparator{},
				ui.MenuItem{Text: "Delete message", Icon: lucide.Trash, Variant: ui.MenuItemDanger, Shortcut: "Delete", OnSelect: nothing},
			}})),
		section("Banners: a tone's fill and line, an icon, a sentence, at most one button",
			workspaceui.Banner{Tone: workspaceui.Neutral, Icon: lucide.ImageOff, Text: "1 remote image is not shown.", Action: "Load", OnAction: nothing},
			workspaceui.Banner{Tone: workspaceui.Warning, Icon: lucide.TriangleAlert, Text: "This message may not be from skyline-travel.example.", Action: "Details", OnAction: nothing},
			workspaceui.Banner{Tone: workspaceui.Accent, Icon: lucide.Clock, Text: "Reminder tomorrow at 8:00 AM.", Action: "Remove", OnAction: nothing},
			workspaceui.Banner{Tone: workspaceui.Danger, Icon: lucide.CircleAlert, Text: "Sending failed: the server refused the message.", Action: "Retry", OnAction: nothing},
			workspaceui.Banner{Tone: workspaceui.Success, Icon: lucide.Check, Text: "You accepted this invitation."}),
		ui.AlertDialog{Open: confirm.Get, OnOpenChange: closes(confirm), Title: "Delete message?", Description: "Your trip to Lisbon: itinerary", StyleFunc: workspaceui.DialogPanel(cx),
			Actions: []ui.AlertAction{
				{Text: "Cancel", Variant: ui.ButtonSecondary, Role: core.AlertCancel},
				{Text: "Delete", Variant: ui.ButtonDanger, Default: true}}},
		ui.Dialog{Open: form.Get, OnOpenChange: closes(form), Title: "New mailbox", StyleFunc: workspaceui.DialogPanel(cx),
			Children: []core.Node{
				workspaceui.Stack(core.Text{Content: "Name", Style: ui.TypeCaption}, ui.Input{Label: "Name", Value: name.Get, OnChange: name.Set}),
				workspaceui.Stack(core.Text{Content: "Inside", Style: ui.TypeCaption}, ui.Select[string]{Label: "Inside", Value: func() string { return "top" }, OnChange: func(string) {}, Options: []ui.Option[string]{{Value: "top", Label: "ann.lee@example.com"}, {Value: "archive", Label: "Archive"}}}),
			},
			Footer: []core.Node{
				ui.Button{Text: "Cancel", Variant: ui.ButtonSecondary, OnPress: func() { form.Set(false) }},
				ui.Button{Text: "Create", OnPress: func() { form.Set(false) }}}},
	)
}
