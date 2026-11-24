package main

import (
	"strconv"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// These are prototypes of the kit's parts (plan, Phase 2). They live here
// until the look is accepted, so that the kit's API does not change under
// the apps while it is being decided.

// roleIcon is the icon of a mailbox by its role.
func roleIcon(role string) func() *core.SVG {
	switch role {
	case "inbox":
		return lucide.Inbox
	case "drafts":
		return lucide.FilePen
	case "sent":
		return lucide.Send
	case "junk":
		return lucide.ShieldAlert
	case "trash":
		return lucide.Trash
	case "archive":
		return lucide.Archive
	case "flag":
		return lucide.Flag
	}
	return lucide.Folder
}

// navLabel is the content of a navigation row: an icon, a name, and a count
// at the row's end in figures of one width. An unread count takes the
// text's color; any other is muted.
func navLabel(m mailbox) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		name := ui.Text(m.name, 13, ui.Regular, colors.Text)
		name.MaxLines, name.Style.FlexGrow, name.Style.FlexShrink = 1, 1, 1
		parts := []core.Node{ui.Icon(roleIcon(m.role), 16, colors.Accent), name}
		if m.count > 0 {
			ink, weight := colors.Muted, ui.Regular
			if m.unread {
				ink, weight = colors.Text, ui.Semibold
			}
			parts = append(parts, ui.Tabular(ui.Text(strconv.Itoa(m.count), 12, weight, ink)))
		}
		return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8, FlexGrow: 1, FlexShrink: 1}, Children: parts}
	})
}

// navItems are mailboxes as items of NSGo's sidebar.
func navItems(boxes []mailbox) []ui.SidebarItem {
	var items []ui.SidebarItem
	for _, m := range boxes {
		items = append(items, ui.SidebarItem{ID: m.id, TestID: "mailbox-" + m.id, Content: navLabel(m)})
	}
	return items
}

// summaryRow is the content of a message row: a gutter for the unread dot,
// the sender's avatar, then the sender with the date at the row's end, the
// subject, and the preview, with the row's marks under the date.
func summaryRow(m message, previewLines int) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		dot := core.View{Style: core.Style{Width: core.Pt(8), Height: core.Pt(8), CornerRadius: 4}}
		if m.unread {
			dot.StyleFunc = ui.Filled(colors.Accent)
		}
		weight := ui.Regular
		if m.unread {
			weight = ui.Semibold
		}
		sender := ui.Text(m.from, 13, ui.Semibold, colors.Text)
		sender.MaxLines, sender.Style.FlexGrow, sender.Style.FlexShrink = 1, 1, 1
		subject := ui.Text(m.subject, 13, weight, colors.Text)
		subject.MaxLines, subject.Style.FlexGrow, subject.Style.FlexShrink = 1, 1, 1
		var marks []core.Node
		if m.flagged {
			marks = append(marks, ui.Icon(lucide.Flag, 13, colors.Warn))
		}
		if m.attachment {
			marks = append(marks, ui.Icon(lucide.Paperclip, 13, colors.Muted))
		}
		if m.reminder {
			marks = append(marks, ui.Icon(lucide.Clock, 13, colors.Muted))
		}
		if m.thread > 1 {
			marks = append(marks, ui.Badge{Text: strconv.Itoa(m.thread)})
		}
		lines := []core.Node{
			core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8}, Children: []core.Node{sender, ui.Tabular(ui.Text(m.when, 12, ui.Regular, colors.Muted))}},
			core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 6}, Children: append([]core.Node{subject}, marks...)},
		}
		if previewLines > 0 {
			preview := ui.Text(m.preview, 12, ui.Regular, colors.Muted)
			preview.MaxLines = previewLines
			preview.Style.LineHeight = 16
			lines = append(lines, preview)
		}
		return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignStart, Gap: 8, Padding: core.Symmetric(8, 8), Width: core.Pct(100)}, Children: []core.Node{
			core.View{Style: core.Style{Width: core.Pt(8), Height: core.Pt(32), JustifyContent: core.JustifyCenter}, Children: []core.Node{dot}},
			ui.Avatar{Name: m.from, Size: 32},
			core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0), Gap: 2}, Children: lines},
		}}
	})
}

// softRows is the style of a list's rows: rounded, a hover fill, and the
// Frostyard soft fill when chosen, with the row's own text colors.
func softRows(cx *core.Scope) func() core.Style {
	colors, kit := ui.ColorsOf(cx), workspaceui.ColorsOf(cx)
	return func() core.Style {
		return core.Style{CornerRadius: ui.RadiusControl,
			Hover:    &core.Style{Background: colors.Hover()},
			Selected: &core.Style{Background: kit.Selected()}}
	}
}

// emptyState is what a pane shows with nothing in it: an icon, a line, and
// a hint.
func emptyState(icon func() *core.SVG, title, hint string) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
		return core.View{Style: ui.FillRest.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Gap: 8, Padding: core.All(24)}), Children: []core.Node{
			ui.Icon(icon, 32, colors.Faint),
			core.Text{Content: title, Style: ui.TypeTitle},
			core.Text{Content: hint, Style: ui.TypeCaption.Merge(core.Style{TextAlign: core.TextAlignCenter}), StyleFunc: ui.Colored(colors.Muted)},
		}}
	})
}
