package main

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
	"github.com/frostyard/frostui/frostuitest"
)

func hasAll(t *testing.T, tt *ui.Tester, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
}

// TestSourceListSnapshot draws the window: the tree with its sections,
// counts, a spinner and an alert, and the open note. With
// FROSTUI_SNAPSHOT_DIR set it writes the frame, light and dark, as PNGs
// there; docs/images/sourcelist.png is the light one.
func TestSourceListSnapshot(t *testing.T) {
	t.Setenv("FROSTUI_ICON_THEME", "Adwaita")
	setIcons()
	t.Cleanup(func() { frostui.SetIcons(nil) })
	for _, dark := range []bool{false, true} {
		d := newDemo(frostuitest.NewLoop().Post)
		tt := ui.NewTesterWith(d.view, 1000, 680, ui.TesterOptions{Dark: dark})
		tt.Frame()
		hasAll(t, tt, "Pinned", "Vault", "Projects", "Garden plan", "Journal", "Archive", "13 words")
		dir := os.Getenv("FROSTUI_SNAPSHOT_DIR")
		if dir == "" {
			continue
		}
		name := "sourcelist.png"
		if dark {
			name = "sourcelist-dark.png"
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

func TestPinnedHeaderExplainsItsRule(t *testing.T) {
	const tip = "Notes pinned from a note's menu or with Pin Note"
	d := newDemo(frostuitest.NewLoop().Post)
	tt := ui.NewTester(d.view, 1000, 680)
	if tt.HasText(tip) {
		t.Fatal("the tooltip shows before the pointer rests on the header")
	}
	at, ok := tt.Find("Pinned")
	if !ok {
		t.Fatalf("no Pinned header in %q", tt.Texts())
	}
	tt.Move(at.X+at.W/2, at.Y+at.H/2)
	tt.Advance(2 * time.Second)
	tt.Frame()
	hasAll(t, tt, tip)
}

func clickRole(t *testing.T, tt *ui.Tester, role ui.Role, name string) {
	t.Helper()
	a, ok := tt.AccessByRole(role, name)
	if !ok {
		t.Fatalf("no element of role %v named %q", role, name)
	}
	tt.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
}

func TestEditorKeepsCtrlKAndThePaletteOpensElsewhere(t *testing.T) {
	d := newDemo(frostuitest.NewLoop().Post)
	tt := ui.NewTester(d.view, 1000, 680)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyK)
	if !d.palette.Open {
		t.Fatal("Ctrl K did not open the palette outside the editor")
	}
	tt.Key(0, ui.KeyEscape)
	clickRole(t, tt, ui.RoleTextField, "Note")
	before := d.current.body
	tt.Key(ui.Cmd, ui.KeyK)
	if d.palette.Open {
		t.Fatal("Ctrl K opened the palette in the editor")
	}
	if d.current.body == before {
		t.Error("the editor did not take Ctrl K")
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
	if !d.palette.Open {
		t.Error("Ctrl+Shift+P did not open the palette in the editor")
	}
}

func TestQuickSwitcherFindsNotesOnAWorker(t *testing.T) {
	loop := frostuitest.NewLoop()
	d := newDemo(loop.Post)
	tt := ui.NewTester(d.view, 1000, 680)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyO)
	if d.palette.ShownSource() != "notes" {
		t.Fatalf("source = %q", d.palette.ShownSource())
	}
	// Archive and Journal are closed, so their notes show in the palette
	// alone.
	loop.RunUntil(t, func() bool { tt.Frame(); return tt.HasText("Old ideas") })
	tt.Type("mond")
	loop.RunUntil(t, func() bool { tt.Frame(); return !tt.HasText("Old ideas") && tt.HasText("Monday") })
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if d.current == nil || d.current.title != "Monday" {
		t.Errorf("opened %+v", d.current)
	}
	if !d.list.IsOpen("vault", folderPrefix+"Journal") {
		t.Error("the tree did not open the way to the note")
	}
}

func TestCommandsFromKeysAndTheMenu(t *testing.T) {
	d := newDemo(frostuitest.NewLoop().Post)
	tt := ui.NewTester(d.view, 1000, 680)
	tt.Frame()
	n := len(d.notes)
	tt.Key(ui.Cmd, ui.KeyN)
	tt.Frame()
	if len(d.notes) != n+1 || d.current.title != "Untitled" {
		t.Fatalf("Ctrl+N: %d notes, current %+v", len(d.notes), d.current)
	}
	if err := tt.Click("Main menu"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Settings"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	hasAll(t, tt, "Editor", "Appearance", "Spell check")
	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	tt.Type("hub")
	tt.Frame()
	hasAll(t, tt, "hub.example.org", "Sync in the background")
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyF1)
	hasAll(t, tt, "Keyboard shortcuts", "New Note", "Quick Switcher")
}

// waitForLinkSignal uses a deadline only to bound a broken worker or command.
func waitForLinkSignal(t *testing.T, signal <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(frostuitest.Timeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func exampleLinkCommand(t *testing.T, d *demo) frostui.Command {
	t.Helper()
	for _, cmd := range d.table.Commands() {
		if cmd.Title == "Open Example Link" {
			return cmd
		}
	}
	t.Fatal("Open Example Link is missing from the command table")
	return frostui.Command{}
}

func TestExampleLinkFromMainMenu(t *testing.T) {
	loop := frostuitest.NewLoop()
	posted := make(chan struct{}, 1)
	d := newDemo(func(fn func()) { loop.Post(fn); posted <- struct{}{} })
	target := make(chan string, 1)
	d.opener.Launch = func(raw string) error { target <- raw; return nil }
	tt := ui.NewTester(d.view, 1000, 680)
	tt.Frame()
	if d.linkStatus != "" {
		t.Fatalf("status before invocation = %q", d.linkStatus)
	}
	cmd := exampleLinkCommand(t, d)
	if cmd.Disabled || len(cmd.Bindings) != 0 {
		t.Fatalf("initial command is disabled or has a key binding: %+v", cmd)
	}
	if err := tt.Click("Main menu"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Open Example Link"); err != nil {
		t.Fatal(err)
	}
	waitForLinkSignal(t, posted, "the link completion post")
	if raw := <-target; raw != "https://example.org" {
		t.Fatalf("launched %q", raw)
	}
	if !d.linkPending || !exampleLinkCommand(t, d).Disabled {
		t.Fatal("worker completion changed UI state before the post was flushed")
	}
	tt.Frame()
	hasAll(t, tt, "Opening https://example.org…")
	loop.Flush()
	tt.Frame()
	if d.linkPending || exampleLinkCommand(t, d).Disabled {
		t.Fatal("the command did not become enabled after completion")
	}
	hasAll(t, tt, "Sent https://example.org to the system handler")
	tt.Key(ui.Cmd, ui.KeyN)
	tt.Frame()
	hasAll(t, tt, "Sent https://example.org to the system handler")
	// The command-only palette avoids unrelated note-source worker posts.
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Type("Open Example Link")
	tt.Frame()
	hasAll(t, tt, "Open Example Link")
}

func TestExampleLinkRefusesURLs(t *testing.T) {
	for _, tc := range []struct{ name, raw, reason string }{
		{"empty", "", "it has no scheme"},
		{"javascript", "javascript:alert(1)", "the scheme javascript is not allowed"},
		{"hostless HTTPS", "https:///missing-host", "it has no host"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loop := frostuitest.NewLoop()
			posted := make(chan struct{}, 1)
			d := newDemo(func(fn func()) { loop.Post(fn); posted <- struct{}{} })
			var calls atomic.Int32
			d.opener.Launch = func(string) error { calls.Add(1); return nil }
			tt := ui.NewTester(d.view, 1000, 680)
			d.openLink(tc.raw)
			waitForLinkSignal(t, posted, "the refusal post")
			if !d.linkPending {
				t.Fatal("refusal changed pending state before the UI post was flushed")
			}
			loop.Flush()
			tt.Frame()
			if calls.Load() != 0 {
				t.Fatal("a refused URL reached Launch")
			}
			want := fmt.Sprintf("Could not open %s: frostui: refused to open %q: %s: %v", tc.raw, tc.raw, tc.reason, frostui.ErrNotAllowed)
			hasAll(t, tt, want)
			if d.linkPending || exampleLinkCommand(t, d).Disabled {
				t.Fatal("a refusal left the command pending or disabled")
			}
		})
	}
}

func TestExampleLinkShowsLauncherError(t *testing.T) {
	loop := frostuitest.NewLoop()
	posted := make(chan struct{}, 1)
	d := newDemo(func(fn func()) { loop.Post(fn); posted <- struct{}{} })
	cause := errors.New("no system handler")
	d.opener.Launch = func(string) error { return cause }
	tt := ui.NewTester(d.view, 1000, 680)
	exampleLinkCommand(t, d).Run()
	waitForLinkSignal(t, posted, "the launch error post")
	if !d.linkPending {
		t.Fatal("launch error changed UI state before the post was flushed")
	}
	loop.Flush()
	tt.Frame()
	hasAll(t, tt, "Could not open https://example.org: "+cause.Error())
	if d.linkPending || exampleLinkCommand(t, d).Disabled {
		t.Fatal("a launch error left the command pending or disabled")
	}
}

func TestExampleLinkKeepsUIResponsive(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	done := make(chan struct{})
	stop := make(chan struct{})
	result := make(chan error, 1)
	t.Cleanup(func() {
		close(stop)
		unblock()
		waitForLinkSignal(t, done, "the UI test to stop")
	})
	// This worker owns all UI state and reports errors without failing the
	// test itself. Cleanup can release a regressed synchronous launcher.
	run := func() error {
		loop := frostuitest.NewLoop()
		posted := make(chan struct{}, 3)
		d := newDemo(func(fn func()) { loop.Post(fn); posted <- struct{}{} })
		command := func() frostui.Command {
			for _, cmd := range d.table.Commands() {
				if cmd.Title == "Open Example Link" {
					return cmd
				}
			}
			return frostui.Command{}
		}
		entered := make(chan struct{}, 3)
		var calls atomic.Int32
		d.opener.Launch = func(string) error {
			calls.Add(1)
			entered <- struct{}{}
			<-release
			return nil
		}
		tt := ui.NewTester(d.view, 1000, 680)
		tt.Frame()
		cmd := command()
		if cmd.Run == nil {
			return errors.New("Open Example Link has no command action")
		}
		cmd.Run()
		select {
		case <-entered:
		case <-stop:
			return errors.New("the UI test stopped before Launch started")
		}
		if !d.linkPending || !command().Disabled {
			return errors.New("the open command is not disabled while Launch blocks")
		}
		status := d.linkStatus
		d.openLink("https://duplicate.example.org")
		command().Run()
		if d.linkStatus != status || calls.Load() != 1 {
			return errors.New("a duplicate open changed the pending request")
		}
		before := len(d.notes)
		tt.Key(ui.Cmd, ui.KeyN)
		tt.Frame()
		if len(d.notes) != before+1 || d.current.title != "Untitled" {
			return errors.New("New Note did not run while Launch was blocked")
		}
		if !d.linkPending || !strings.Contains(d.linkStatus, "https://example.org") {
			return errors.New("the pending link status did not survive New Note")
		}
		unblock()
		// No UI post is flushed here. The test goroutine's watchdog
		// bounds this wait, and cleanup can stop it after a failure.
		select {
		case <-posted:
		case <-stop:
			return errors.New("the UI test stopped before the launch completion post")
		}
		if !d.linkPending || !command().Disabled {
			return errors.New("the worker changed UI state before its post was flushed")
		}
		loop.Flush()
		tt.Frame()
		if d.linkPending || command().Disabled || calls.Load() != 1 {
			return errors.New("completion did not enable the command or a duplicate reached Launch")
		}
		if !tt.HasText("Sent https://example.org to the system handler") {
			return fmt.Errorf("missing successful launch status in %q", tt.Texts())
		}
		return nil
	}
	go func() {
		defer close(done)
		result <- run()
	}()
	waitForLinkSignal(t, done, "the UI to remain responsive during Launch")
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

func TestNotesMoveAndRenameInTheTree(t *testing.T) {
	d := newDemo(frostuitest.NewLoop().Post)
	tt := ui.NewTester(d.view, 1000, 680)
	tt.Frame()
	src, ok := tt.Find("Reading list")
	dst, ok2 := tt.Find("Journal")
	if !ok || !ok2 {
		t.Fatalf("texts: %q", tt.Texts())
	}
	sx, sy, dx, dy := src.X+src.W/2, src.Y+src.H/2, dst.X+dst.W/2, dst.Y+dst.H/2
	tt.Press(sx, sy)
	tt.Move(sx+8, sy+8)
	tt.Move(dx, dy)
	tt.Frame()
	tt.Release(dx, dy)
	tt.Frame()
	if n := d.notes[1]; n.folder != "Journal" {
		t.Errorf("Reading list is in %q", n.folder)
	}
	if err := tt.RightClick("Garden plan"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Rename"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !d.list.Renaming() {
		t.Fatal("Rename did not start renaming")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Orchard")
	tt.Key(0, ui.KeyEnter)
	if d.notes[0].title != "Orchard" {
		t.Errorf("title = %q", d.notes[0].title)
	}
}
