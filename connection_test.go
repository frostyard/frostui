package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func connectionTester(conn *Connection) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Children(func() { ConnectionStatus(c, *conn) })
	}, 400, 100)
}

func TestConnectionStatusNamesThePlace(t *testing.T) {
	conn := Connection{Daemon: "exampled"}
	tt := connectionTester(&conn)
	if !tt.HasText("Local") {
		t.Errorf("no Local in %q", tt.Texts())
	}
	a, ok := tt.Access("Connection: Local")
	if !ok || a.Description != "Connected to exampled" {
		t.Errorf("access %+v, %v", a, ok)
	}
	conn.Host = "mail.example.org"
	tt.Frame()
	if !tt.HasText("mail.example.org") || tt.HasText("Local") {
		t.Errorf("with a host: %q", tt.Texts())
	}
}

func TestConnectionStatusWarnsWithTheReason(t *testing.T) {
	conn := Connection{State: Reconnecting, Daemon: "exampled", Reason: "The socket closed."}
	tt := connectionTester(&conn)
	a, ok := tt.Access("Connection: Local")
	if !ok || a.Description != "Reconnecting to exampled… The socket closed." {
		t.Errorf("access %+v, %v", a, ok)
	}
}

func TestConnectionStatusClicks(t *testing.T) {
	clicks := 0
	conn := Connection{OnClick: func() { clicks++ }}
	tt := connectionTester(&conn)
	if a, ok := tt.AccessByRole(ui.RoleButton, "Connection: Local"); !ok || !a.Focusable {
		t.Fatalf("not a button that takes the focus: %+v, %v", a, ok)
	}
	if err := tt.Click("Connection: Local"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	if clicks != 2 {
		t.Errorf("a click and Enter ran it %d times", clicks)
	}
}

func TestConnectingScreenOffersRetryAndDetails(t *testing.T) {
	retries, details := 0, 0
	conn := Connection{State: Connecting, Daemon: "exampled", Reason: "No such file.",
		OnRetry: func() { retries++ }, OnDetails: func() { details++ }}
	tt := ui.NewTester(func(c *ui.Context) { ConnectingScreen(c, conn) }, 800, 600)
	for _, want := range []string{"Connecting to exampled…", "No such file.", "Retry", "Details"} {
		if !tt.HasText(want) {
			t.Errorf("no %q in %q", want, tt.Texts())
		}
	}
	if _, ok := tt.AccessByRole(ui.RoleStatus, ""); !ok {
		t.Error("the screen's text is no status for screen readers")
	}
	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Details"); err != nil {
		t.Fatal(err)
	}
	if retries != 1 || details != 1 {
		t.Errorf("retries %d, details %d", retries, details)
	}
}

func TestConnectingScreenDrawsOnlyWhatItCanDo(t *testing.T) {
	conn := Connection{State: Reconnecting, Daemon: "exampled", Host: "hub.example.org"}
	tt := ui.NewTester(func(c *ui.Context) { ConnectingScreen(c, conn) }, 800, 600)
	if !tt.HasText("Reconnecting to exampled on hub.example.org…") || tt.HasText("Retry") || tt.HasText("Details") {
		t.Errorf("%q", tt.Texts())
	}
}
