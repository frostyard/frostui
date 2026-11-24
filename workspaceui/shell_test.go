package workspaceui

import (
	"testing"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
	"github.com/nsgo-dev/nsgo/ui"
)

func testShell(widths PaneWidths, navigated *[]string) func(*core.Scope) core.Node {
	return func(cx *core.Scope) core.Node {
		selected := core.NewSignal(cx, "inbox")
		return Shell{
			Brand: "Example",
			Sections: func() []ui.SidebarSection {
				return []ui.SidebarSection{{Title: "Places", Items: []ui.SidebarItem{
					{ID: "inbox", Text: "Inbox"}, {ID: "archive", Text: "Archive"},
				}}}
			},
			Selected: selected.Get,
			Navigate: func(id string) { selected.Set(id); *navigated = append(*navigated, id) },
			Header:   []core.Node{core.Text{ContentFunc: selected.Get, TestID: "title"}},
			Content:  []core.Node{core.Text{Content: "Body", TestID: "body"}},
			Widths:   widths,
		}
	}
}

// TestShellKeepsItsNavigationWidth pins the widths apps save and restore:
// the default, a saved width, and the clamp of a saved width out of range.
func TestShellKeepsItsNavigationWidth(t *testing.T) {
	for _, tc := range []struct {
		name  string
		saved core.Points
		want  core.Points
	}{
		{"default", 0, 246},
		{"saved", 260, 260},
		{"too narrow", 40, 180},
		{"too wide", 900, 420},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var navigated []string
			app := nsgotest.Start(t, testShell(PaneWidths{Navigation: tc.saved}, &navigated), nsgotest.Size(1200, 800))
			if got := app.Find("sidebar-panel").Frame.Size.Width; got != tc.want {
				t.Errorf("navigation width = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestShellNavigates checks that a wide shell shows the navigation beside
// the content and reports the chosen destination.
func TestShellNavigates(t *testing.T) {
	var navigated []string
	app := nsgotest.Start(t, testShell(PaneWidths{}, &navigated), nsgotest.Size(1200, 800))
	app.Find("body")
	app.Press("sidebar-panel >> text:Archive")
	if len(navigated) != 1 || navigated[0] != "archive" {
		t.Fatalf("navigated %q, want archive", navigated)
	}
	if got := app.Find("title").Text; got != "archive" {
		t.Fatalf("the header shows %q, want the chosen destination", got)
	}
}

// TestCompactShellClosesNavigationAfterAChoice checks the narrow layout:
// the content alone, the navigation on its button, and the content again
// once a destination is chosen.
func TestCompactShellClosesNavigationAfterAChoice(t *testing.T) {
	var navigated []string
	app := nsgotest.Start(t, testShell(PaneWidths{}, &navigated), nsgotest.Size(600, 800))
	app.Find("body")
	if app.Exists("text:Archive") {
		t.Fatal("a compact shell shows its navigation before its button is pressed")
	}
	app.Press("role:button Navigation")
	app.Press("text:Archive")
	if len(navigated) != 1 || navigated[0] != "archive" {
		t.Fatalf("navigated %q, want archive", navigated)
	}
	app.Find("body")
	if app.Exists("text:Archive") {
		t.Fatal("the navigation stays open after a choice")
	}
}
