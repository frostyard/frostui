package workspaceui

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
	"github.com/nsgo-dev/nsgo/ui"
)

// onDesktop is a desktop's environment, where an app's window is its own.
var onDesktop = nsgotest.Environment(core.Environment{OS: "linux", Idiom: core.IdiomDesktop, PixelRatio: 1})

func keepRoot(cx *core.Scope) core.Node {
	KeepWindow(cx, 480, 360)
	return Frame(core.Text{Content: "kept"})
}

// TestKeepWindowKeepsAndShowsTheWindow checks the kept window: a start
// asks for the size the device kept, a change of the size is kept, and a
// window under the least size is not.
func TestKeepWindowKeepsAndShowsTheWindow(t *testing.T) {
	app := nsgotest.Start(t, keepRoot, nsgotest.Size(1000, 700), onDesktop, nsgotest.Storage(map[string]string{windowKey: `{"Width":900,"Height":640}`}))
	app.Find("text:kept")
	asked := slices.ContainsFunc(app.Window, func(r nsgotest.WindowRequest) bool {
		return r.Action == nsgotest.WindowSetFrame && r.Frame.Width == 900 && r.Frame.Height == 640
	})
	if !asked {
		t.Errorf("the app did not ask for its kept window, 900 by 640: %+v", app.Window)
	}
	kept := func() (k keptWindow) {
		text, _ := app.Storage().GetItem(windowKey)
		_ = json.Unmarshal([]byte(text), &k)
		return k
	}
	app.Resize(1100, 760)
	app.Advance(time.Second)
	app.WaitUntil(3*time.Second, func() bool { k := kept(); return k.Width == 1100 && k.Height == 760 })
	app.Resize(300, 200)
	app.Advance(time.Second)
	if k := kept(); k.Width != 1100 || k.Height != 760 {
		t.Errorf("a window 300 by 200 replaced the kept size: %+v", k)
	}
}

// TestKeepWindowLeavesASmallKeptSize checks that a kept size under the
// least one is not asked for.
func TestKeepWindowLeavesASmallKeptSize(t *testing.T) {
	app := nsgotest.Start(t, keepRoot, nsgotest.Size(1000, 700), onDesktop, nsgotest.Storage(map[string]string{windowKey: `{"Width":200,"Height":100}`}))
	app.Find("text:kept")
	if slices.ContainsFunc(app.Window, func(r nsgotest.WindowRequest) bool { return r.Action == nsgotest.WindowSetFrame }) {
		t.Errorf("the app asked for a window 200 by 100: %+v", app.Window)
	}
}

// TestFileExplorerTellsItsTreesWidth checks that the explorer opens its
// tree at the width given and tells a drag of its edge.
func TestFileExplorerTellsItsTreesWidth(t *testing.T) {
	var told []core.Points
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		return Frame(FileExplorer{TreeWidth: 200, OnTreeWidthChange: func(w core.Points) { told = append(told, w) },
			Tree: core.View{TestID: "tree", Style: ui.FillRest}, Preview: core.View{TestID: "preview", Style: ui.FillRest}, HasPreview: func() bool { return true }})
	}, nsgotest.Size(900, 500), onDesktop)
	if width := app.Find("id:tree").Bounds().Size.Width; width != 200 {
		t.Errorf("the tree opens %v wide, want 200", width)
	}
	app.Drag("label:Resize file tree", core.Point{X: 40})
	app.Settle()
	if len(told) == 0 || told[len(told)-1] <= 200 {
		t.Errorf("a drag of the tree's edge told %v, want a width over 200", told)
	}
	if width := app.Find("id:tree").Bounds().Size.Width; width <= 200 {
		t.Errorf("the tree is %v wide after the drag, want over 200", width)
	}
}

// TestInspectorOpensAsKeptAndTellsItsChanges checks the shell's right
// panel: open as the shell opens when asked, and its toggle told.
func TestInspectorOpensAsKeptAndTellsItsChanges(t *testing.T) {
	var told []bool
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		return Frame(Shell{Brand: "App", Sections: func() []ui.SidebarSection { return nil }, Selected: func() string { return "" }, Navigate: func(string) {},
			Content: []core.Node{core.Text{Content: "main"}},
			Inspector: &Inspector{Title: "Details", Open: true, OnOpenChange: func(open bool) { told = append(told, open) },
				Content: func(*core.Scope) core.Node { return core.Text{Content: "inspector content"} }}})
	}, nsgotest.Size(1200, 700), onDesktop)
	app.WaitFor("text:inspector content", 3*time.Second)
	if len(told) != 0 {
		t.Errorf("the panel's first state was told as a change: %v", told)
	}
	app.Press("label:Details")
	app.WaitForGone("text:inspector content", 3*time.Second)
	if len(told) != 1 || told[0] {
		t.Errorf("closing the panel told %v, want [false]", told)
	}
}

// TestInspectorKeptOpenStaysShutInANarrowWindow checks that a panel kept
// open does not open over the content of a window too narrow to dock it.
func TestInspectorKeptOpenStaysShutInANarrowWindow(t *testing.T) {
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		return Frame(Shell{Brand: "App", Sections: func() []ui.SidebarSection { return nil }, Selected: func() string { return "" }, Navigate: func(string) {},
			Content:   []core.Node{core.Text{Content: "main"}},
			Inspector: &Inspector{Title: "Details", Open: true, Content: func(*core.Scope) core.Node { return core.Text{Content: "inspector content"} }}})
	}, nsgotest.Size(800, 700), onDesktop)
	app.WaitFor("text:main", 3*time.Second)
	if found := app.Query("text:inspector content"); len(found) != 0 {
		t.Errorf("a window 800 wide opened the panel over its content")
	}
}

// TestKeepWindowKeepsTheWindowsOwnSize checks that the size kept is the
// window's, which a start asks for, where the platform reports it: a
// window with a frame of its own is larger than its content, and keeping
// the content's size made each start's window smaller (seen on Windows:
// 16 narrower and 9 shorter).
func TestKeepWindowKeepsTheWindowsOwnSize(t *testing.T) {
	framed := nsgotest.Environment(core.Environment{OS: "windows", Idiom: core.IdiomDesktop, PixelRatio: 1,
		CanPlaceWindows: true, WindowFrame: core.WindowFrame{Width: 1016, Height: 709, X: 60, Y: 20, Move: true}})
	app := nsgotest.Start(t, keepRoot, nsgotest.Size(1000, 700), framed)
	app.Find("text:kept")
	// The content changes, and the platform's frame says what the window is.
	app.Resize(1001, 700)
	app.Advance(time.Second)
	var k keptWindow
	app.WaitUntil(3*time.Second, func() bool {
		text, ok := app.Storage().GetItem(windowKey)
		return ok && json.Unmarshal([]byte(text), &k) == nil
	})
	if want := (keptWindow{Width: 1016, Height: 709, X: 60, Y: 20, Placed: true}); k != want {
		t.Errorf("kept %+v, want the window's frame %+v, not its content's size", k, want)
	}
}
