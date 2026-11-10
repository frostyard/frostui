package frostui

import "github.com/egoist/mygo/ui"

// ConnectionState is how an app stands with the daemon it is a client of.
type ConnectionState int

const (
	// Connected is a working connection.
	Connected ConnectionState = iota
	// Connecting is the first try, before the app has anything to show:
	// a PaneShell draws the connecting screen over the whole window.
	Connecting
	// Reconnecting is a lost connection the app tries to get back, while
	// it keeps showing what it had.
	Reconnecting
)

// Connection is an app's connection to its daemon, for the status bar's
// first item (ConnectionStatus) and the connecting screen
// (ConnectingScreen).
type Connection struct {
	State ConnectionState
	// Daemon names what the app connects to, as a program's name.
	Daemon string
	// Host is where the daemon runs, "" for this computer.
	Host string
	// Reason says why the last try failed, "" for none.
	Reason string
	// OnClick makes the status item a button.
	OnClick func()
	// OnRetry and OnDetails are the connecting screen's Retry and Details
	// buttons; nil draws none.
	OnRetry, OnDetails func()
}

// place is where the status item says the daemon is.
func (conn Connection) place() string {
	if conn.Host != "" {
		return conn.Host
	}
	return "Local"
}

// summary says what the connection is doing, without the reason.
func (conn Connection) summary() string {
	where := conn.Daemon
	if where == "" {
		where = "the service"
	}
	if conn.Host != "" {
		where += " on " + conn.Host
	}
	switch conn.State {
	case Connecting:
		return "Connecting to " + where + "…"
	case Reconnecting:
		return "Reconnecting to " + where + "…"
	}
	return "Connected to " + where
}

// ConnectionStatus draws the status bar's connection item: a dot with
// "Local" or the host. While reconnecting it takes the warning tone, with
// the reason in its tooltip. PaneFrame.Connection draws it first in the
// status bar.
func ConnectionStatus(c *ui.Context, conn Connection) ui.Element {
	t := c.Theme()
	dot, color := t.Success, t.TextMuted
	switch conn.State {
	case Connecting:
		dot = t.TextMuted
	case Reconnecting:
		dot, color = t.Warning, t.Warning
	}
	tooltip := conn.summary()
	if conn.Reason != "" && conn.State != Connected {
		tooltip += " " + conn.Reason
	}
	var row ui.Element
	if conn.OnClick != nil {
		row = ui.ButtonBase(c).Row().Cursor(ui.CursorPointer)
		if row.Hovered() {
			row.Background(t.SurfaceHover)
		}
		if row.Clicked() {
			conn.OnClick()
		}
	} else {
		row = ui.Row(c)
	}
	row.Gap(6).Padding(2, 4).Radius(4).AlignItems(ui.Center).Label("Connection: " + conn.place()).Tooltip(tooltip)
	return row.Children(func() {
		ui.Box(c).Size(8, 8).Radius(4).Background(dot)
		ui.Text(c, conn.place()).FontSize(12).TextColor(color).SingleLine()
	})
}

// ConnectingScreen fills the window while an app has no daemon: what it
// connects to, the reason of the last failure, Retry and Details. It is
// the window's drag region under a hidden title bar, so draw it as the
// whole view; a PaneShell draws it for the Connecting state.
func ConnectingScreen(c *ui.Context, conn Connection) ui.Element {
	return connectingWindow(c, conn, nil)
}

// connectingWindow is ConnectingScreen with trailing drawn at the end of
// its title bar's room.
func connectingWindow(c *ui.Context, conn Connection, trailing func()) ui.Element {
	t := c.Theme()
	tb := c.TitleBar()
	return ui.Column(c).Fill().AlignItems(ui.Stretch).Background(t.Background).Children(func() {
		ui.Row(c).Height(TopBarHeight).Padding(0, SpaceL+tb.Right, 0, SpaceL+tb.Left).AlignItems(ui.Center).DragWindow().Children(func() {
			ui.Spacer(c)
			if trailing != nil {
				trailing()
			}
		})
		ui.Column(c).Grow(1).Gap(SpaceM).Padding(SpaceXL).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
			ui.Spinner(c)
			ui.Column(c).Gap(SpaceS).AlignItems(ui.Center).MaxWidth(ToastWidth).Role(ui.RoleStatus).Children(func() {
				ui.Text(c, conn.summary()).FontSize(18).Bold().Wrap().TextAlign(ui.Center)
				if conn.Reason != "" {
					Muted(c, conn.Reason).Wrap().TextAlign(ui.Center)
				}
			})
			ui.Row(c).Gap(SpaceS).Padding(SpaceS, 0, 0, 0).Children(func() {
				if conn.OnRetry != nil && LabeledButton(c, "view-refresh", "Retry", true).Clicked() {
					conn.OnRetry()
				}
				if conn.OnDetails != nil && ui.Button(c, "Details").Clicked() {
					conn.OnDetails()
				}
			})
		})
	})
}
