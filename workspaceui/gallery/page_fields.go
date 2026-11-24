package main

import (
	"time"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func fieldsPage(cx *core.Scope) core.Node {
	name := core.NewSignal(cx, "Ann Lee")
	size := core.NewSignal(cx, "default")
	notify := core.NewSignal(cx, true)
	tray := core.NewSignal(cx, false)
	photos := core.NewSignal(cx, true)
	day := core.NewSignal(cx, base.DateValue{Start: time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)})
	at := core.NewSignal(cx, base.TimeOfDay{Hour: 9, Minute: 30})
	to := core.NewSignalFunc(cx, []string{"david.king@example.com", "iris@example.com"}, func(_, _ []string) bool { return false })
	width := core.Style{Width: core.Pt(260)}
	return workspaceui.Stack(
		section("Text fields",
			workspaceui.Row(
				ui.Input{Label: "Name", Placeholder: "Name", Value: name.Get, OnChange: name.Set, Style: width},
				ui.Input{Label: "Search", Placeholder: "Search All Inboxes", Style: width},
				ui.Input{Label: "Disabled", Placeholder: "Disabled", Disabled: func() bool { return true }, Style: width})),
		section("A choice from a list, a date and a time",
			workspaceui.Row(
				ui.Select[string]{Label: "Message text size", Value: size.Get, OnChange: size.Set, Style: width, Options: []ui.Option[string]{
					{Value: "small", Label: "Small"}, {Value: "default", Label: "Default"}, {Value: "large", Label: "Large"}}},
				ui.DatePicker{Label: "Date", Value: day.Get, OnChange: day.Set, Style: core.Style{Width: core.Pt(180)}},
				ui.TimeField{Label: "Time", Value: at.Get, OnChange: at.Set}),
			caption("Rule 11: a date is a date picker and a time a time field, never text typed as 2026-11-26 09:30.")),
		section("Recipients: tags in a field",
			ui.TagInput{Label: "To", Placeholder: "To", Tags: to.Get, OnChange: to.Set, Style: core.Style{Width: core.Pt(540)}}),
		section("On and off, and items of a list",
			workspaceui.Row(
				ui.Switch{Label: "Notifications", Checked: notify.Get, OnChange: notify.Set},
				core.Text{Content: "Notifications", Style: ui.TypeBody},
				ui.Switch{Label: "Close to tray", Checked: tray.Get, OnChange: tray.Set},
				core.Text{Content: "Close to tray", Style: ui.TypeBody}),
			workspaceui.Row(
				ui.Checkbox{Text: "Contact photos", Checked: photos.Get, OnChange: photos.Set},
				ui.Checkbox{Text: "Buy stamps", Checked: func() bool { return false }, OnChange: func(bool) {}},
				ui.Checkbox{Text: "Disabled", Disabled: func() bool { return true }, Checked: func() bool { return true }}),
			caption("A setting that is on or off is a switch. A checkbox is an item of a list, or a task's done mark.")),
		section("Settings rows: NSGo's form section",
			ui.FormSection{Title: "General", Footer: "The appearance is this device's.", Children: []core.Node{
				ui.FormRow{Icon: lucide.SunMoon, Title: "Appearance", Detail: "Light, dark, or as the system", Control: workspaceui.AppearanceChoice{}},
				ui.FormRow{Icon: lucide.Bell, Title: "New message notifications", Control: ui.Switch{Label: "New message notifications", Checked: notify.Get, OnChange: notify.Set}},
				ui.FormRow{Icon: lucide.ALargeSmall, Title: "Message text size", Control: ui.Select[string]{Label: "Message text size", Value: size.Get, OnChange: size.Set, Style: core.Style{Width: core.Pt(140)}, Options: []ui.Option[string]{
					{Value: "small", Label: "Small"}, {Value: "default", Label: "Default"}, {Value: "large", Label: "Large"}}}},
			}}),
	)
}
