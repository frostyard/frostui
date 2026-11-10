package main

import (
	"context"
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"testing"

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

// fakeFetch syncs at once, except the Work account's Spam folder, which
// fails, and the folder named block, which runs until release closes.
func fakeFetch(block string, release <-chan struct{}) fetchFunc {
	return func(ctx context.Context, p *frostui.Progress, account, folder string) (int, error) {
		switch {
		case account == "Work" && folder == "Spam":
			return 0, errors.New("the server closed the connection")
		case account+"/"+folder == block:
			p.Set(0.4, folder+": 2 of 5")
			p.Log(frostui.LevelOutput, "%s/%s: fetched block 2 of 5", account, folder)
			select {
			case <-release:
				return 7, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}
		p.Log(frostui.LevelOutput, "%s/%s: fetched 5 blocks", account, folder)
		return 4, nil
	}
}

// midSync returns the example in the middle of a sync of every account:
// five folders ended (Work's Spam failed) and the Lists inbox runs until
// the test ends.
func midSync(t *testing.T) *demo {
	t.Helper()
	loop := frostuitest.NewLoop()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	d := newDemo(loop.Post, fakeFetch("Lists/Inbox", release))
	d.shell.Console.Open = true
	d.sync(d.accounts...)
	loop.RunUntil(t, func() bool {
		cur := d.jobs.Current()
		return cur != nil && cur.Source == "Lists" && cur.Fraction > 0
	})
	return d
}

// TestCommandCenterSnapshot draws the shell in the middle of a sync: a job
// running in the rail and the status bar, the pipeline with done, failed
// and running folders, and the console open. With FROSTUI_SNAPSHOT_DIR
// set it writes the frame, light and dark, as PNGs there; the README's
// picture is docs/images/commandcenter.png, the light one.
func TestCommandCenterSnapshot(t *testing.T) {
	// The icons come from Adwaita where it is installed, so a picture
	// does not depend on the desktop it was taken on.
	t.Setenv("FROSTUI_ICON_THEME", "Adwaita")
	setIcons()
	t.Cleanup(func() { frostui.SetIcons(nil) })
	for _, dark := range []bool{false, true} {
		// A ListState belongs to one window, so each Tester draws its own
		// example.
		d := midSync(t)
		tt := ui.NewTesterWith(d.view, 1280, 800, ui.TesterOptions{Dark: dark})
		hasAll(t, tt, "Command center", "Overview", "Accounts", "Settings", "Sync run", "Job 6 of 6",
			"Syncing 6 of 6", "Personal", "Work", "Spam", "Lists", "every job the example runs, as it runs",
			"Syncing Work · Spam failed: the server closed the connection", "Lists/Inbox: fetched block 2 of 5")
		for _, step := range []string{"Inbox, done", "Spam, failed", "Inbox, running"} {
			if _, ok := tt.Find(step); !ok {
				t.Errorf("no pipeline step named %q", step)
			}
		}
		dir := os.Getenv("FROSTUI_SNAPSHOT_DIR")
		if dir == "" {
			continue
		}
		name := "commandcenter.png"
		if dark {
			name = "commandcenter-dark.png"
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

// TestCommandCenterRunsItsCommands drives the example as a user would: the
// rail, an account's Sync button, the palette and a confirmation.
func TestCommandCenterRunsItsCommands(t *testing.T) {
	loop := frostuitest.NewLoop()
	d := newDemo(loop.Post, fakeFetch("", nil))
	tt := ui.NewTester(d.view, 1280, 800)
	if err := tt.Click("Accounts"); err != nil {
		t.Fatal(err)
	}
	hasAll(t, tt, "imaps://mail.example.org", "Never synced")
	if err := tt.Click("Sync Personal"); err != nil {
		t.Fatal(err)
	}
	loop.RunUntil(t, func() bool { return !d.jobs.Busy() })
	tt.Frame()
	hasAll(t, tt, "Synced")
	if d.accounts[0].Messages != 12 || !d.accounts[1].Synced.IsZero() {
		t.Errorf("Personal has %d messages; Work synced %v: only Personal's three folders should have run", d.accounts[0].Messages, d.accounts[1].Synced)
	}

	if err := tt.Click("Search or run a command"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Clear the console"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	hasAll(t, tt, "Clear the console?")
	if err := tt.Click("Clear"); err != nil {
		t.Fatal(err)
	}
	if n := len(d.console.Lines(nil, "")); n != 0 {
		t.Errorf("the console kept %d lines after Clear", n)
	}

	d.stopOnFailure = true
	d.sync(d.accounts[1])
	loop.RunUntil(t, func() bool { return !d.jobs.Busy() })
	if s := frostui.RunSummary(d.lastRun); s != "1 done · 1 failed" {
		t.Errorf("Work's run = %q", s)
	}
	if !d.shell.Console.Open {
		t.Error("a failure did not open the console")
	}
}

// TestStopAtTheFirstFailureDropsTheRest checks the setting: the folders
// queued after a failure are canceled, not run.
func TestStopAtTheFirstFailureDropsTheRest(t *testing.T) {
	loop := frostuitest.NewLoop()
	d := newDemo(loop.Post, fakeFetch("", nil))
	d.stopOnFailure = true
	d.accounts[1].Folders = []string{"Spam", "Inbox"}
	d.sync(d.accounts[1], d.accounts[2])
	loop.RunUntil(t, func() bool { return !d.jobs.Busy() })
	if s := frostui.RunSummary(d.lastRun); s != "1 failed · 2 canceled" {
		t.Errorf("run = %q, want the failure and the two folders after it canceled", s)
	}
}
