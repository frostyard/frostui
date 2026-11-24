package main

import (
	"slices"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// folders is a small tree of mailboxes and notes.
var folders = map[any][]any{
	"Archive":  {"2025", "2026", "Receipts"},
	"2026":     {"Travel", "Taxes"},
	"Projects": {"Aurora", "Onboarding"},
}

// searchField is a search box: NSGo's input with a search icon before it
// and, while it is empty, the keys that focus it at its end. A prototype of
// the kit's part: NSGo has no search field.
func searchField(placeholder, keys string, value *core.Signal[string]) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		hint := core.Show{When: func() bool { return value.Get() == "" }, Then: func(*core.Scope) core.Node {
			return core.View{Style: core.Style{Padding: core.Symmetric(1, 5), CornerRadius: ui.RadiusSmall, BorderWidth: 1}, StyleFunc: func() core.Style {
				return core.Style{Background: colors.Hover(), BorderColor: colors.Border()}
			}, Children: []core.Node{ui.Text(cx.KeyLabel(keys), 11, ui.Medium, colors.Muted)}}
		}}
		return core.View{TestID: "search-field", Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 6, Padding: core.Edges{Left: 8, Right: 6}, Height: core.Pt(ui.ButtonHeight), CornerRadius: ui.RadiusControl, BorderWidth: 1}, StyleFunc: func() core.Style {
			return core.Style{Background: colors.Field(), BorderColor: colors.FieldBorder(), FocusWithin: &core.Style{BorderColor: colors.Accent()}}
		}, Children: []core.Node{
			ui.Icon(lucide.Search, 14, colors.Muted),
			base.Input{Label: placeholder, Placeholder: placeholder, Value: value.Get, OnChange: value.Set, Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0), FontSize: 13, BorderWidth: 0}, StyleFunc: func() core.Style {
				return core.Style{Background: core.Transparent, Color: colors.Text(), PlaceholderColor: colors.Muted()}
			}},
			hint,
		}}
	})
}

func morePage(cx *core.Scope) core.Node {
	open := core.NewSignalFunc(cx, []string{"Archive"}, func(_, _ []string) bool { return false })
	selection := ui.NewSelection(cx, false)
	query, typed := core.NewSignal(cx, ""), core.NewSignal(cx, "aurora")
	page := core.NewSignal(cx, 2)
	details := core.NewSignal(cx, true)
	palette := core.NewSignal(cx, false)
	toaster := ui.NewToaster(cx)
	colors := ui.ColorsOf(cx)
	nothing := func() {}
	return workspaceui.Stack(
		section("A search field: an icon, and the keys that focus it while it is empty",
			workspaceui.Row(
				core.View{Style: core.Style{Width: core.Pt(280)}, Children: []core.Node{searchField("Search All Inboxes", "mod+f", query)}},
				core.View{Style: core.Style{Width: core.Pt(280)}, Children: []core.Node{searchField("Search All Inboxes", "mod+f", typed)}})),
		section("A tree of folders",
			core.View{Style: core.Style{Width: core.Pt(260), Height: core.Pt(210), CornerRadius: ui.RadiusCard, Padding: core.All(6)}, StyleFunc: ui.Filled(workspaceui.ColorsOf(cx).Chrome), Children: []core.Node{
				ui.Tree{Label: "Folders", Selection: selection, ToggleLabel: "Expand", Style: ui.FillRest,
					Roots:    func() []any { return []any{"Archive", "Projects", "Notes"} },
					Children: func(k any) []any { return folders[k] },
					Expanded: func(k any) bool { return slices.Contains(open.Get(), k.(string)) },
					OnExpand: func(k any, on bool) {
						next := slices.DeleteFunc(slices.Clone(open.GetUntracked()), func(s string) bool { return s == k })
						if on {
							next = append(next, k.(string))
						}
						open.Set(next)
					},
					RowLabel: func(k any) string { return k.(string) },
					Row: func(cx *core.Scope, item ui.TreeItem) core.Node {
						return ui.HStack{Gap: 6, Children: []core.Node{
							ui.Icon(lucide.Folder, 15, ui.ColorsOf(cx).Accent),
							core.Text{ContentFunc: func() string { return item.Key().(string) }, Style: ui.TypeBody},
						}}
					}}}}),
		section("A disclosure and paging",
			ui.Collapsible{Text: "Details", Open: details.Get, OnOpenChange: details.Set,
				Content: core.Text{Content: "Received: from mail.example.com by mx.example.net; Tue, 24 Nov 2026 06:30:11 +0000", Style: ui.TypeCode}},
			ui.Pagination{Label: "Pages", Page: page.Get, TotalPages: func() int { return 9 }, OnChange: page.Set}),
		section("Detail beside the content: a popover and a hover card",
			workspaceui.Row(
				ui.Popover{Label: "Remind me", Style: core.Style{Background: workspaceui.ColorsOf(cx).Panel()}, Trigger: ui.Button{Text: "Remind me", Icon: lucide.Clock, Variant: ui.ButtonSecondary}, Children: []core.Node{
					workspaceui.Stack(core.Text{Content: "Remind me", Style: ui.TypeLabel},
						ui.Button{Text: "Later today", Variant: ui.ButtonGhost, OnPress: nothing},
						ui.Button{Text: "Tomorrow at 8:00 AM", Variant: ui.ButtonGhost, OnPress: nothing},
						ui.Button{Text: "Next week", Variant: ui.ButtonGhost, OnPress: nothing})}},
				ui.HoverCard{Trigger: core.Text{Content: "David King", Style: ui.TypeLabel, StyleFunc: ui.Colored(colors.Accent)}, Children: []core.Node{
					workspaceui.Row(ui.Avatar{Name: "David King", Size: 40},
						workspaceui.Stack(core.Text{Content: "David King", Style: ui.TypeLabel}, core.Text{Content: "david.king@example.com", Style: ui.TypeCaption, StyleFunc: ui.Colored(colors.Muted)}))}})),
		section("What was done, with a way back: a toast; and the command palette",
			workspaceui.Row(
				ui.Button{Text: "Archive a message", Variant: ui.ButtonSecondary, OnPress: func() {
					toaster.Show(ui.ToastContent{Title: "Message archived", Description: "Your trip to Lisbon: itinerary", Action: "Undo", OnAction: nothing})
				}},
				ui.Button{Text: "Fail to send", Variant: ui.ButtonSecondary, OnPress: func() {
					toaster.Show(ui.ToastContent{Type: ui.ToastError, Title: "Sending failed", Description: "The server refused the message."})
				}},
				ui.Button{Text: "Open the palette", Variant: ui.ButtonSecondary, OnPress: func() { palette.Set(true) }})),
		ui.Toasts{Toaster: toaster, Placement: ui.ToastBottomEnd, StyleFunc: workspaceui.DialogPanel(cx)},
		ui.CommandDialog{Open: palette.Get, OnOpenChange: func(on bool, _ base.Reason) { palette.Set(on) }, Style: core.Style{Background: workspaceui.ColorsOf(cx).Panel()}, Command: ui.Command{
			Placeholder: "Search or run a command", SearchLabel: "Commands", Empty: "No matching command", OnSelect: func(string) { palette.Set(false) },
			Items: func() []base.CommandItem {
				return []base.CommandItem{
					{ID: "compose", Label: "New message", Group: "Mail", Shortcut: "mod+n", Icon: lucide.SquarePen()},
					{ID: "reply", Label: "Reply", Group: "Mail", Shortcut: "mod+r", Icon: lucide.Reply()},
					{ID: "archive", Label: "Archive", Group: "Mail", Shortcut: "mod+shift+a", Icon: lucide.Archive()},
					{ID: "calendar", Label: "Go to Calendar", Group: "Go", Shortcut: "mod+2", Icon: lucide.Calendar()},
					{ID: "settings", Label: "Settings", Group: "Go", Shortcut: "mod+,", Icon: lucide.Settings()},
				}
			}}},
	)
}
