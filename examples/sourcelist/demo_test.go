package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

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
