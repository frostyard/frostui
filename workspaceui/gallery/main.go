// The gallery shows the workspaceui kit and NSGo's controls in the Frostyard
// themes: a standard window built from them, and a page for each family of
// controls. It is what the look is signed off on (docs/plans/workspaceui.md,
// Phase 2), and its picture tours keep a picture of every page.
package main

import (
	"github.com/frostyard/frostui/workspaceui"
	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

func main() {
	nsgo.Run(func(app *core.Scope) { nsgo.SetMainView(app, root) }, nsgo.WindowSize(1280, 820))
}

// page is one destination of the gallery.
type page struct {
	id, title string
	icon      func() *core.SVG
	view      func(*core.Scope) core.Node
}

var pages = []page{
	{"colors", "Colors and type", lucide.Palette, colorsPage},
	{"buttons", "Buttons and toolbars", lucide.MousePointerClick, buttonsPage},
	{"fields", "Fields and forms", lucide.TextCursorInput, fieldsPage},
	{"rows", "Rows and marks", lucide.List, rowsPage},
	{"overlays", "Dialogs and menus", lucide.PanelTop, overlaysPage},
}

func root(cx *core.Scope) core.Node {
	cx.SetTitle("Frostyard Gallery")
	_ = workspaceui.Apply(cx)
	appearance := core.NewSignal(cx, core.AppearanceSystem)
	// "window" is the standard window, which takes the whole frame; every
	// other destination is a page inside the gallery's own shell.
	current := core.NewSignal(cx, "window")
	return workspaceui.Frame(core.Dynamic(func(cx *core.Scope) core.Node {
		if current.Get() == "window" {
			return standardWindow(cx, func() { current.Set(pages[0].id) })
		}
		items := []ui.SidebarItem{{ID: "window", Text: "The standard window", Icon: lucide.AppWindow}}
		for _, p := range pages {
			items = append(items, ui.SidebarItem{ID: p.id, Text: p.title, Icon: p.icon})
		}
		return workspaceui.Shell{
			Brand:    "Gallery",
			Sections: func() []ui.SidebarSection { return []ui.SidebarSection{{Title: "Pages", Items: items}} },
			Selected: current.Get,
			Navigate: current.Set,
			Header: []core.Node{core.Text{ContentFunc: func() string {
				for _, p := range pages {
					if p.id == current.Get() {
						return p.title
					}
				}
				return ""
			}, Style: ui.TypeTitle}},
			Actions:       []core.Node{workspaceui.AppearanceChoice{Value: appearance.Get, OnChange: appearance.Set}},
			SidebarFooter: []core.Node{workspaceui.ConnectionItem{Tone: workspaceui.Success, Text: func() string { return "Local" }}},
			Content: []core.Node{core.Dynamic(func(cx *core.Scope) core.Node {
				for _, p := range pages {
					if p.id == current.Get() {
						return workspaceui.Page(p.view(cx))
					}
				}
				return nil
			})},
		}
	}))
}

// section is a titled group of a page.
func section(title string, children ...core.Node) core.Node {
	return workspaceui.Stack(append([]core.Node{core.Text{Content: title, Style: ui.TypeLabel}}, children...)...)
}

// caption is a small muted note under an example.
func caption(text string) core.Node {
	return core.Dynamic(func(cx *core.Scope) core.Node {
		return core.Text{Content: text, Style: ui.TypeCaption, StyleFunc: ui.Colored(ui.ColorsOf(cx).Muted)}
	})
}
