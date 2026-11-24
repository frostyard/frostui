package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// settingsPage is NSGo's settings view under the theme: categories at the
// side, groups of rows, and its own search.
func settingsPage(cx *core.Scope) core.Node {
	category, query := core.NewSignal(cx, "general"), core.NewSignal(cx, "")
	notify, tray, photos := core.NewSignal(cx, true), core.NewSignal(cx, false), core.NewSignal(cx, true)
	size := core.NewSignal(cx, "default")
	toggle := func(label string, s *core.Signal[bool]) core.Node {
		return ui.Switch{Label: label, Checked: s.Get, OnChange: s.Set}
	}
	return core.View{Style: core.Style{Height: core.Pt(620), CornerRadius: ui.RadiusCard, BorderWidth: 1}, StyleFunc: func() core.Style {
		return core.Style{BorderColor: ui.ColorsOf(cx).Border()}
	}, Children: []core.Node{ui.Settings{Title: "Settings", Category: category, Query: query, SearchPlaceholder: "Search settings", Categories: []ui.SettingsCategory{
		{ID: "general", Title: "General", Icon: lucide.Settings, Groups: []ui.SettingsGroup{
			{ID: "look", Title: "Look", Items: []ui.SettingsItem{
				{ID: "appearance", Title: "Appearance", Description: "Light, dark, or as the system. This device only.", Control: workspaceui.AppearanceChoice{}},
				{ID: "text", Title: "Message text size", Control: ui.Select[string]{Label: "Message text size", Value: size.Get, OnChange: size.Set, Style: core.Style{Width: core.Pt(140)}, Options: []ui.Option[string]{
					{Value: "small", Label: "Small"}, {Value: "default", Label: "Default"}, {Value: "large", Label: "Large"}}}},
				{ID: "photos", Title: "Contact photos", Description: "Show a sender's picture in the list.", Control: toggle("Contact photos", photos)},
			}},
			{ID: "window", Title: "Window", Items: []ui.SettingsItem{
				{ID: "tray", Title: "Close to the tray", Description: "Keep running when the last window closes.", Control: toggle("Close to the tray", tray)},
			}},
		}},
		{ID: "notifications", Title: "Notifications", Icon: lucide.Bell, Groups: []ui.SettingsGroup{
			{ID: "new", Title: "New mail", Items: []ui.SettingsItem{
				{ID: "notify", Title: "New message notifications", Control: toggle("New message notifications", notify)},
			}},
		}},
		{ID: "accounts", Title: "Accounts", Icon: lucide.Users, Groups: []ui.SettingsGroup{
			{ID: "list", Title: "ann.lee@example.com", Actions: []core.Node{ui.Button{Text: "Add account", Variant: ui.ButtonSecondary, Size: ui.SizeSmall, OnPress: func() {}}}, Items: []ui.SettingsItem{
				{ID: "sync", Title: "Mail, Contacts, Calendar", Description: "Synced a minute ago", Control: ui.Badge{Text: "Online", Variant: ui.BadgeSuccess}},
				{ID: "remove", Title: "Remove this account", Description: "Its mail stays on the server.", Control: ui.Button{Text: "Remove…", Variant: ui.ButtonDanger, Size: ui.SizeSmall, OnPress: func() {}}},
			}},
		}},
	}}}}
}
