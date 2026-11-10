package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// undoTime is how long a delete can be undone.
const undoTime = 6 * time.Second

// message is one message of the example.
type message struct {
	From, Subject, When, Body string
}

// demo is the example's state, which belongs to the UI thread as frostui's
// models do.
type demo struct {
	shell    frostui.PaneShell
	table    *frostui.CommandTable
	dialogs  frostui.Dialogs
	console  *frostui.Console
	conn     frostui.Connection
	mailbox  string
	messages []message
	selected int
	quit     func()
}

var mailboxes = []frostui.NavItem{
	{ID: "inbox", Title: "Inbox", Icon: "mail-unread"},
	{ID: "archive", Title: "Archive", Icon: "mail-archive"},
	{ID: "trash", Title: "Trash", Icon: "user-trash"},
}

// newDemo returns the example's state. post hands work to the UI thread
// (Window.Update in the program, a frostuitest.Loop in the tests).
func newDemo(post func(func())) *demo {
	con := frostui.NewConsole(2000)
	con.SetNotify(frostui.NewWaker(post, 60*time.Millisecond).Wake)
	d := &demo{console: con, table: frostui.NewCommandTable(), mailbox: "inbox", messages: []message{
		{"Ada Lovelace", "The analytical engine", "9:41", "The engine might compose elaborate pieces of music."},
		{"Grace Hopper", "A nanosecond of wire", "9:12", "Here is what a nanosecond looks like: 11.8 inches."},
		{"Alan Turing", "Can machines think?", "Yesterday", "I propose to consider the question."},
	}}
	d.conn = frostui.Connection{Daemon: "exampled", OnClick: d.toggleConnection, OnRetry: d.connected, OnDetails: d.showDetails}
	d.shell.Dialogs = &d.dialogs
	d.shell.Console.Title = "Activity"
	d.shell.Console.Subtitle = "what the example's service did"
	d.shell.Console.EmptyNote = "What the service does shows here as it happens."
	d.shell.Console.Offered = 1<<frostui.LevelDebug | 1<<frostui.LevelInfo | 1<<frostui.LevelWarn | 1<<frostui.LevelError
	con.Add(frostui.LevelInfo, "Work", "Inbox synced: 3 new, 12 changed (0.4 s)")
	con.Add(frostui.LevelInfo, "Home", "CardDAV pass: 2 changed")
	con.Add(frostui.LevelWarn, "Lists", "The server is slow to answer")
	return d
}

// connected ends the first try or a reconnect.
func (d *demo) connected() {
	d.conn.State, d.conn.Reason = frostui.Connected, ""
	d.console.Add(frostui.LevelInfo, "app", "Connected to exampled")
}

// toggleConnection drops the pretend connection, or brings it back.
func (d *demo) toggleConnection() {
	if d.conn.State != frostui.Connected {
		d.connected()
		return
	}
	d.conn.State, d.conn.Reason = frostui.Reconnecting, "The socket closed."
	d.console.Add(frostui.LevelWarn, "app", "Lost the connection to exampled")
}

func (d *demo) showDetails() {
	d.shell.Info.Show("Connection", func(c *ui.Context) {
		frostui.Body(c, "A real app shows the service's log here.").Wrap()
	})
}

func (d *demo) view(c *ui.Context) {
	conn := d.conn
	d.table.Add(d.commands(c)...)
	d.shell.View(c, frostui.PaneFrame{
		SidebarPane: "mailboxes",
		Panes: []frostui.Pane{
			{ID: "mailboxes", Label: "Mailboxes", Width: 220, Min: 160, Max: 320, Collapsible: true,
				Bar: func() { d.mailboxesBar(c) }, Content: func() { d.mailboxesPane(c) }},
			{ID: "messages", Label: "Messages", Width: 360, Min: 280, Max: 560,
				Bar: func() { d.messagesBar(c) }, Content: func() { d.messagesPane(c) }},
			{ID: "reader", Label: "Message", Min: 360,
				Bar: func() { d.readerBar(c) }, Content: func() { d.readerPane(c) }},
		},
		Menu:        d.menu,
		Table:       d.table,
		Connection:  &conn,
		Status:      []frostui.StatusItem{{Text: "Work: Inbox synced 9:41", Progress: -1}},
		StatusEnd:   []frostui.StatusItem{{Text: fmt.Sprintf("%d messages", len(d.messages)), Progress: -1}},
		ConsoleData: d.console,
	})
}

func (d *demo) mailboxesBar(c *ui.Context) {
	if frostui.IconButton(c, "mail-send-receive", "Get Mail").Clicked() {
		d.console.Add(frostui.LevelInfo, "app", "Asked for new mail")
	}
}

func (d *demo) messagesBar(c *ui.Context) {
	ui.Column(c).Children(func() {
		ui.Text(c, "Inbox").Bold().SingleLine()
		frostui.Caption(c, fmt.Sprintf("%d messages", len(d.messages))).SingleLine()
	})
	ui.Spacer(c)
	frostui.IconButton(c, "mail-message-new", "New Message")
}

func (d *demo) readerBar(c *ui.Context) {
	if frostui.IconButton(c, "edit-delete", "Delete").Disabled(len(d.messages) == 0).Clicked() {
		d.deleteSelected(c)
	}
	frostui.IconButton(c, "mail-archive", "Archive")
	frostui.IconButton(c, "mail-reply-sender", "Reply")
	frostui.IconButton(c, "mail-reply-all", "Reply All")
	frostui.IconButton(c, "mail-forward", "Forward")
}

// deleteSelected takes the message away and offers to put it back: it is
// gone for good when the toast ends without Undo.
func (d *demo) deleteSelected(c *ui.Context) {
	if len(d.messages) == 0 {
		return
	}
	at := d.selected
	gone := d.messages[at]
	d.messages = slices.Delete(slices.Clone(d.messages), at, at+1)
	d.selected = min(at, max(len(d.messages)-1, 0))
	frostui.ShowActionToast(c, frostui.ActionToast{
		Title: "Message deleted", Action: "Undo", Timeout: undoTime, Countdown: true,
		OnAction: func() {
			d.messages = slices.Insert(slices.Clone(d.messages), min(at, len(d.messages)), gone)
			d.selected = min(at, len(d.messages)-1)
		},
		OnExpire: func() { d.console.Add(frostui.LevelInfo, "app", "Deleted "+gone.Subject) },
	})
}

func (d *demo) menu(m *ui.Menu) {
	if m.Item("Command palette").Shortcut(ui.Cmd, ui.KeyK).Chosen() {
		d.shell.Palette.Show()
	}
	d.table.MenuItem(m, "about")
	if d.quit != nil {
		m.Separator()
		if m.Item("Quit").Shortcut(ui.Cmd, ui.KeyQ).Chosen() {
			d.quit()
		}
	}
}

func (d *demo) commands(c *ui.Context) []frostui.Command {
	return []frostui.Command{
		{ID: "delete", Title: "Delete the message", Group: "Message", Icon: "edit-delete",
			Bindings: []frostui.KeyBinding{{Key: ui.KeyDelete, Scope: frostui.OutsideText}},
			Disabled: len(d.messages) == 0, Reason: "no message", Run: func() { d.deleteSelected(c) }},
		{ID: "sidebar", Title: "Show or hide the mailboxes", Group: "View", Icon: "sidebar-show",
			Run: func() { d.shell.Panes.Toggle("mailboxes") }},
		{ID: "connection", Title: "Drop or restore the connection", Group: "App", Icon: "network-server", Run: d.toggleConnection},
		{ID: "about", Title: "About Three panes", Group: "App", Icon: "help-about", Run: d.showAbout},
	}
}

func (d *demo) showAbout() {
	d.shell.Info.Show("About Three panes", frostui.AboutBody(frostui.About{
		Name: "Three panes", Version: "example", Summary: "The frame without the rail of frostui, with a pretend mail reader.",
		Website: "https://github.com/frostyard/frostui", License: "MIT",
	}))
}

func (d *demo) mailboxesPane(c *ui.Context) {
	t := c.Theme()
	ui.Sidebar(c, &d.mailbox, func() {
		ui.SidebarSection(c, "Work", nil, func() {
			for _, item := range mailboxes {
				ui.SidebarItem(c.Key(item.ID), item.ID, frostui.IconSVG(item.Icon), item.Title)
			}
		})
	}).Grow(1).Padding(frostui.SpaceS, frostui.SpaceS).Background(frostui.Chrome(t))
}

func (d *demo) messagesPane(c *ui.Context) {
	t := c.Theme()
	if len(d.messages) == 0 {
		frostui.Empty(c, "mail-read", "No messages", "").Grow(1)
		return
	}
	ui.Scroll(c).Grow(1).Fill().Children(func() {
		for i, m := range d.messages {
			row := ui.ButtonBase(c.Key(m.Subject)).Column().Gap(2).Padding(frostui.SpaceM, frostui.SpaceL).
				AlignItems(ui.Stretch).Cursor(ui.CursorPointer).Label(m.From + ": " + m.Subject)
			switch {
			case i == d.selected:
				row.Background(t.Accent.Alpha(0.16))
			case row.Hovered():
				row.Background(t.SurfaceHover)
			}
			if row.Clicked() {
				d.selected = i
			}
			row.Children(func() {
				ui.Row(c).AlignItems(ui.Center).Children(func() {
					frostui.Strong(c, m.From).Grow(1).SingleLine()
					frostui.Caption(c, m.When).SingleLine()
				})
				frostui.Body(c, m.Subject).SingleLine()
				frostui.Muted(c, m.Body).SingleLine()
			})
			frostui.RowDivider(c.Key(m.Subject + " divider"))
		}
	})
}

func (d *demo) readerPane(c *ui.Context) {
	if len(d.messages) == 0 {
		frostui.Empty(c, "mail-read", "No message selected", "").Grow(1)
		return
	}
	m := d.messages[d.selected]
	ui.Column(c).Grow(1).Gap(frostui.SpaceM).Padding(frostui.SpaceXL).AlignItems(ui.Stretch).Children(func() {
		frostui.Title(c, m.Subject)
		frostui.Muted(c, m.From+" · "+m.When)
		frostui.RowDivider(c)
		frostui.Body(c, m.Body).Wrap()
	})
}
