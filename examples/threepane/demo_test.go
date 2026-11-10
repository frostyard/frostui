package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
	"github.com/frostyard/frostui/frostuitest"
)

// hasAll fails t for each text tt does not show.
func hasAll(t *testing.T, tt *ui.Tester, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
}

func newTester(t *testing.T) (*demo, *ui.Tester) {
	t.Helper()
	loop := frostuitest.NewLoop()
	d := newDemo(loop.Post)
	return d, ui.NewTester(d.view, 1280, 800)
}

// TestThreePaneSnapshot draws the frame with the console open and a delete
// to undo. With FROSTUI_SNAPSHOT_DIR set it writes the frame, light and
// dark, as PNGs there; docs/images/threepane.png is the light one.
func TestThreePaneSnapshot(t *testing.T) {
	t.Setenv("FROSTUI_ICON_THEME", "Adwaita")
	frostui.SetIcons(frostui.NewIcons(nil))
	t.Cleanup(func() { frostui.SetIcons(nil) })
	for _, dark := range []bool{false, true} {
		loop := frostuitest.NewLoop()
		d := newDemo(loop.Post)
		d.shell.Console.Open = true
		tt := ui.NewTesterWith(d.view, 1280, 800, ui.TesterOptions{Dark: dark})
		tt.Frame()
		if err := tt.Click("Delete"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		hasAll(t, tt, "Inbox", "Message deleted", "Undo", "Grace Hopper", "Local", "Activity",
			"Inbox synced: 3 new, 12 changed (0.4 s)", "Work: Inbox synced 9:41", "2 messages")
		dir := os.Getenv("FROSTUI_SNAPSHOT_DIR")
		if dir == "" {
			continue
		}
		name := "threepane.png"
		if dark {
			name = "threepane-dark.png"
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", f.Name())
	}
}

func TestThreePaneUndoPutsTheMessageBack(t *testing.T) {
	d, tt := newTester(t)
	tt.Frame()
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(d.messages) != 2 {
		t.Fatalf("%d messages after the delete", len(d.messages))
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(d.messages) != 3 || d.messages[0].From != "Ada Lovelace" {
		t.Errorf("after Undo: %+v", d.messages)
	}
}

func TestThreePaneDeleteEndsWhenTheToastDoes(t *testing.T) {
	d, tt := newTester(t)
	tt.Frame()
	if err := tt.Click("Delete"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(undoTime + time.Second)
	if len(d.messages) != 2 || len(d.console.Lines(nil, "Deleted The analytical engine")) != 1 {
		t.Errorf("messages %d, lines %v", len(d.messages), d.console.Lines(nil, "Deleted"))
	}
}

func TestThreePaneConnects(t *testing.T) {
	d, tt := newTester(t)
	d.conn.State = frostui.Connecting
	tt.Frame()
	hasAll(t, tt, "Connecting to exampled…", "Retry", "Details")
	if err := tt.Click("Retry"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	hasAll(t, tt, "Grace Hopper", "Local")
	if err := tt.Click("Connection: Local"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if d.conn.State != frostui.Reconnecting {
		t.Errorf("state %v after the click", d.conn.State)
	}
	if a, ok := tt.Access("Connection: Local"); !ok || a.Description != "Reconnecting to exampled… The socket closed." {
		t.Errorf("access %+v, %v", a, ok)
	}
}

func TestThreePanePanesAndConsole(t *testing.T) {
	d, tt := newTester(t)
	tt.Frame()
	if err := tt.Click("Sidebar"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if d.shell.Panes.Shown("mailboxes") || tt.HasText("Trash") {
		t.Errorf("the sidebar shows: %q", tt.Texts())
	}
	tt.Key(ui.Cmd, ui.KeyJ)
	tt.Frame()
	hasAll(t, tt, "Inbox synced: 3 new, 12 changed (0.4 s)", "what the example's service did")
	if tt.HasText("Dry run") {
		t.Error("the console offers levels the example has no lines of")
	}
}
