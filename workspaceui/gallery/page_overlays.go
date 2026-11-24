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
			banner(workspaceui.Neutral, lucide.ImageOff, "1 remote image is not shown.", "Load"),
			banner(workspaceui.Warning, lucide.TriangleAlert, "This message may not be from skyline-travel.example.", "Details"),
			banner(workspaceui.Accent, lucide.Clock, "Reminder tomorrow at 8:00 AM.", "Remove"),
			banner(workspaceui.Danger, lucide.CircleAlert, "Sending failed: the server refused the message.", "Retry"),
			banner(workspaceui.Success, lucide.Check, "You accepted this invitation.", "")),
		ui.AlertDialog{Open: confirm.Get, OnOpenChange: closes(confirm), Title: "Delete message?", Description: "Your trip to Lisbon: itinerary", StyleFunc: panel(cx),
			Actions: []ui.AlertAction{
				{Text: "Cancel", Variant: ui.ButtonSecondary, Role: core.AlertCancel},
				{Text: "Delete", Variant: ui.ButtonDanger, Default: true}}},
		ui.Dialog{Open: form.Get, OnOpenChange: closes(form), Title: "New mailbox", StyleFunc: panel(cx),
			Children: []core.Node{
				workspaceui.Stack(core.Text{Content: "Name", Style: ui.TypeCaption}, ui.Input{Label: "Name", Value: name.Get, OnChange: name.Set}),
				workspaceui.Stack(core.Text{Content: "Inside", Style: ui.TypeCaption}, ui.Select[string]{Label: "Inside", Value: func() string { return "top" }, OnChange: func(string) {}, Options: []ui.Option[string]{{Value: "top", Label: "ann.lee@example.com"}, {Value: "archive", Label: "Archive"}}}),
			},
			Footer: []core.Node{
				ui.Button{Text: "Cancel", Variant: ui.ButtonSecondary, OnPress: func() { form.Set(false) }},
				ui.Button{Text: "Create", OnPress: func() { form.Set(false) }}}},
	)
}

// panel is the ground of a dialog: Frostyard's panel, opaque. NSGo's own is
// the page's color at 92%, over a blur that Linux does not draw, so the view
// under a dialog shows through it there.
func panel(cx *core.Scope) func() core.Style {
	kit := workspaceui.ColorsOf(cx)
	return func() core.Style { return core.Style{Background: kit.Panel()} }
}

// banner is a notice inside a pane: a tone's fill and line, an icon, a
// sentence and at most one button at its end. A prototype of the kit's part.
func banner(tone workspaceui.Tone, icon func() *core.SVG, text, action string) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors, kit := ui.ColorsOf(cx), workspaceui.ColorsOf(cx)
		ink := map[workspaceui.Tone]func() core.Color{workspaceui.Neutral: colors.Muted, workspaceui.Accent: colors.Accent, workspaceui.Success: colors.Up, workspaceui.Warning: colors.Warn, workspaceui.Danger: colors.Danger}[tone]
		children := []core.Node{ui.Icon(icon, 16, ink), core.Text{Content: text, Style: ui.TypeBody.Merge(core.Style{FlexGrow: 1, FlexShrink: 1})}}
		if action != "" {
			children = append(children, ui.Button{Text: action, Variant: ui.ButtonSecondary, Size: ui.SizeSmall, OnPress: func() {}})
		}
		return core.View{Role: core.RoleStatus, Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 10, Padding: core.Symmetric(8, 12), CornerRadius: ui.RadiusControl, BorderWidth: 1}, StyleFunc: func() core.Style {
			return core.Style{Background: kit.Fill(tone), BorderColor: kit.Line(tone)}
		}, Children: children}
	})
}
