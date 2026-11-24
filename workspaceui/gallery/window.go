package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// moduleTab is a module of the app as a tab of an icon alone: its name is
// what assistive technology reads and what its tooltip shows.
func moduleTab(id, name string, icon func() *core.SVG) ui.Option[string] {
	return ui.Option[string]{Value: id, Label: name, Content: core.Dynamic(func(cx *core.Scope) core.Node {
		return ui.Tooltip{Text: name, Trigger: ui.Icon(icon, 16, ui.ColorsOf(cx).Text)}
	})}
}

// standardWindow is the window of plan 0016's standard, drawn with fixture
// mail: the toolbar's segments, the navigation with counts, the list of
// summary rows and the reader. Nothing in it does anything but choose.
func standardWindow(cx *core.Scope, gallery func()) core.Node {
	box := core.NewSignal(cx, "all-inboxes")
	module := core.NewSignal(cx, "mail")
	filter := core.NewSignal(cx, "all")
	todo := core.NewSignal(cx, false)
	selection := ui.NewSelection(cx, true)
	selection.Set(3)
	opened := func() *message {
		if i, ok := selection.Cursor().(int); ok && i >= 0 && i < len(inbox) && selection.Len() == 1 {
			return &inbox[i]
		}
		return nil
	}
	nothing := func() {}
	noMessage := func() bool { return opened() == nil }
	action := func(label string, icon func() *core.SVG) core.Node {
		return ui.IconButton{Label: label, Icon: icon, Disabled: noMessage, OnPress: nothing}
	}
	colors := ui.ColorsOf(cx)
	list := core.View{Style: ui.FillRest.Merge(core.Style{Gap: 8, Padding: core.Edges{Top: 8, Left: 8, Right: 8}}), Children: []core.Node{
		ui.Input{Label: "Search All Inboxes", Placeholder: "Search All Inboxes"},
		core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8}, Children: []core.Node{
			ui.Tabs[string]{Variant: ui.TabsSegmented, Size: ui.SizeSmall, Label: "Filter", Value: filter.Get, OnChange: filter.Set, Options: []ui.Option[string]{
				{Value: "all", Label: "All"}, {Value: "unread", Label: "Unread"}, {Value: "flagged", Label: "Flagged"}}},
			ui.Spacer{},
			ui.Menu{Label: "Sort", Trigger: ui.Button{Label: "Sort", Icon: lucide.ArrowDownUp, Variant: ui.ButtonGhost, Size: ui.SizeSmall}, Children: []core.Node{
				ui.MenuItem{Text: "Date", Checked: func() bool { return true }, OnSelect: nothing}, ui.MenuItem{Text: "Sender", OnSelect: nothing}, ui.MenuItem{Text: "Subject", OnSelect: nothing},
				ui.MenuSeparator{}, ui.MenuItem{Text: "Show conversations", Checked: func() bool { return true }, OnSelect: nothing}}}}},
		ui.List{Label: "Messages", TestID: "messages", Style: ui.FillRest, EstimatedSize: 84, Gap: 2, Selection: selection, TapSelects: true,
			Count:        func() int { return len(inbox) },
			RowLabel:     func(i int) string { return inbox[i].subject },
			RowStyleFunc: softRows(cx),
			Row:          func(_ *core.Scope, row core.ListRow) core.Node { return summaryRow(inbox[row.Index()], 2) }},
	}}
	reader := core.Dynamic(func(cx *core.Scope) core.Node {
		m := opened()
		if m == nil {
			if selection.Len() > 1 {
				return emptyState(lucide.Mails, "Several messages selected", "Archive, delete or move them from the toolbar.")
			}
			return emptyState(lucide.MailOpen, "No message selected", "Choose a message to read it here.")
		}
		return core.Scroll{Style: ui.FillRest, Children: []core.Node{core.View{Style: core.Style{Padding: core.All(16), Gap: 12}, Children: []core.Node{
			core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignStart, Gap: 12}, Children: []core.Node{
				ui.Avatar{Name: m.from, Size: 40},
				core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, Gap: 2}, Children: []core.Node{
					core.Text{Content: m.from, Style: ui.TypeLabel.Merge(core.Style{FontWeight: ui.Semibold})},
					core.Text{Content: "To: Ann Lee", Style: ui.TypeCaption, StyleFunc: ui.Colored(colors.Muted)}}},
				ui.Tabular(ui.Text("Tue, Nov 24, 2026 at "+m.when, 12, ui.Regular, colors.Muted))}},
			core.Text{Content: m.subject, Style: ui.TypeHeading},
			banner(workspaceui.Neutral, lucide.ImageOff, "1 remote image is not shown.", "Load"),
			core.View{Style: core.Style{Padding: core.All(24), CornerRadius: ui.RadiusInset, BorderWidth: 1, Gap: 10}, StyleFunc: func() core.Style {
				return core.Style{Background: core.RGB(0xffffff), BorderColor: colors.Border(), Color: core.RGB(0x1f2933)}
			}, Children: []core.Node{
				core.Text{Content: m.subject, Style: ui.TypeTitle},
				core.Text{Content: m.preview + " The message's own page stays white inside a hairline, whatever the theme: it is the sender's page, not the app's.", Style: ui.TypeBody},
			}},
		}}}}
	})
	split := ui.ResizablePanelGroup{Style: ui.FillRest, HandleLabel: "Resize message list", Panels: []base.ResizablePanel{
		{Size: func() core.Points { return 390 }, MinSize: 260, Children: []core.Node{list}},
		{MinSize: 320, Children: []core.Node{reader}},
	}}
	return workspaceui.Shell{
		Brand:    "FrostMail",
		Selected: box.Get,
		Navigate: box.Set,
		Sections: func() []ui.SidebarSection {
			return []ui.SidebarSection{
				{Title: "Favorites", Items: navItems(favorites)},
				{Title: "ann.lee@example.com", Collapsible: true, Items: navItems(account)},
				{Title: "Gallery", Items: []ui.SidebarItem{{ID: "gallery", Text: "Controls and colors", Icon: lucide.Palette}}},
			}
		},
		Disclose: func(id string) bool {
			if id == "gallery" {
				gallery()
				return true
			}
			return false
		},
		SidebarHeader: []core.Node{core.View{Style: core.Style{Padding: core.Symmetric(4, 8)}, Children: []core.Node{
			ui.Tabs[string]{Variant: ui.TabsSegmented, Equal: true, Label: "Module", Value: module.Get, OnChange: module.Set, Options: []ui.Option[string]{
				moduleTab("mail", "Mail", lucide.Mail), moduleTab("calendar", "Calendar", lucide.Calendar),
				moduleTab("people", "People", lucide.Users), moduleTab("tasks", "Tasks", lucide.ListChecks)}}}}},
		SidebarFooter: []core.Node{core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Width: core.Pct(100), Padding: core.Edges{Right: 6}}, Children: []core.Node{
			connection("Local"), ui.Spacer{}, ui.IconButton{Label: "Settings", Icon: lucide.Settings, OnPress: nothing}}}},
		Header: []core.Node{core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 10, Width: core.Pt(378)}, Children: []core.Node{
			core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1}, Children: []core.Node{
				core.Text{Content: "All Inboxes", Style: ui.TypeTitle},
				core.Text{Content: "3 unread", Style: ui.TypeCaption, StyleFunc: ui.Colored(colors.Muted)}}},
			ui.Button{Text: "Compose", Icon: lucide.SquarePen, OnPress: nothing}}}},
		Actions: []core.Node{
			ui.Toolbar{Label: "Message", Children: []core.Node{
				ui.ToolbarGroup{Label: "File", Children: []core.Node{action("Archive", lucide.Archive), action("Delete", lucide.Trash), action("Junk", lucide.ShieldAlert)}},
				ui.ToolbarSeparator{},
				ui.ToolbarGroup{Label: "Answer", Children: []core.Node{action("Reply", lucide.Reply), action("Reply all", lucide.ReplyAll), action("Forward", lucide.Forward)}},
				ui.ToolbarSeparator{},
				ui.ToolbarGroup{Label: "Mark", Children: []core.Node{action("Flag", lucide.Flag), action("Move", lucide.FolderInput), action("More", lucide.Ellipsis)}},
			}},
			ui.SidebarButton{Text: "Search or run a command", Icon: lucide.Search, Shortcut: "mod+k", Variant: ui.SidebarButtonField, OnPress: nothing},
			ui.ToggleButton{Label: "To-Do Bar", Icon: lucide.PanelRight, On: todo.Get, OnChange: todo.Set},
		},
		Content: []core.Node{split},
	}
}
