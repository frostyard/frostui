package main

import (
	"strconv"

	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

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

// navItems are mailboxes as items of NSGo's sidebar, each with the kit's
// label: a role's icon, the name, and the count at the row's end.
func navItems(boxes []mailbox) []ui.SidebarItem {
	var items []ui.SidebarItem
	for _, m := range boxes {
		items = append(items, ui.SidebarItem{ID: m.id, TestID: "mailbox-" + m.id,
			Content: workspaceui.NavLabel{Icon: roleIcon(m.role), Text: m.name, Count: m.count, Unread: m.unread}})
	}
	return items
}

// summaryRow is a fixture message as the kit's summary row. Its marks are
// NSGo's icons and badge: a flag in the warning tone, a paperclip, a
// reminder's clock, and a thread's count.
func summaryRow(m message, previewLines int) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		colors := ui.ColorsOf(cx)
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
		return workspaceui.SummaryRow{Unread: m.unread, Name: m.from, Title: m.from, When: m.when, Subject: m.subject,
			Marks: marks, Preview: m.preview, PreviewLines: previewLines}
	})
}
