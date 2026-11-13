package frostui

import (
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDiffLineNumbersAndMetadata(t *testing.T) {
	patch := "diff --git a/example b/example\n--- a/example\n+++ b/example\n@@ -4,2 +8,3 @@\n unchanged\n---removed text\n+++added text\n+second addition\n\\ No newline at end of file\n@@ -0,0 +1 @@\n+new file\nBinary files a/image and b/image differ\n"
	lines := parseDiff(patch)
	for _, tc := range []struct {
		index, old, next int
		kind             byte
	}{
		{2, 0, 0, 0}, {3, 0, 0, '@'}, {4, 4, 8, 0}, {5, 5, 0, '-'}, {6, 0, 9, '+'}, {7, 0, 10, '+'}, {8, 0, 0, 0}, {10, 0, 1, '+'}, {11, 0, 0, 0},
	} {
		line := lines[tc.index]
		if line.old != tc.old || line.next != tc.next || line.kind != tc.kind {
			t.Errorf("line %d: %+v", tc.index, line)
		}
	}
	if len(lines) != 12 || lines[5].text != "---removed text" {
		t.Fatal("diff text changed")
	}
	if len(parseDiff("")) != 0 {
		t.Fatal("empty diff has a row")
	}
}

func TestDiffViewWrapsAndReplacesPatch(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			var view DiffView
			patch := "--- a/config.go\n+++ b/config.go\n@@ -1,2 +1,2 @@\n package example\n-const greeting = \"hello\"\n+const greeting = \"Hello, world\"\n"
			view.SetPatch(patch)
			first := &view.lines[0]
			view.SetPatch(patch)
			if &view.lines[0] != first {
				t.Fatal("unchanged patch was reparsed")
			}
			tt := ui.NewTesterWith(func(c *ui.Context) { Adapt(c); view.View(c).Fill() }, 700, 350, ui.TesterOptions{Dark: dark})
			tt.Frame()
			if !tt.HasText(`+const greeting = "Hello, world"`) {
				t.Fatal("added line absent")
			}
			if dir := os.Getenv("FROSTUI_SNAPSHOT_DIR"); dir != "" {
				f, err := os.Create(filepath.Join(dir, "diff-"+map[bool]string{false: "light", true: "dark"}[dark]+".png"))
				if err != nil {
					t.Fatal(err)
				}
				if err := png.Encode(f, tt.Image()); err != nil {
					t.Fatal(err)
				}
				if err := f.Close(); err != nil {
					t.Fatal(err)
				}
			}
			long := "+" + strings.Repeat("a long line ", 30)
			view.SetPatch("@@ -0,0 +1 @@\n" + long + "\n")
			tt.Frame()
			if !tt.HasText(long) || tt.HasText("package example") {
				t.Fatal("replacement lost or retained stale text")
			}
			view.SetPatch("")
			tt.Frame()
			if tt.HasText(long) {
				t.Fatal("empty patch retained old diff")
			}
		})
	}
}
