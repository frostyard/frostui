package workspaceui

import (
	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// FileExplorer gives a directory tree and its preview separate scroll areas.
// Compact windows show one pane at a time and keep the tree within reach.
type FileExplorer struct {
	Header, Tree, Preview core.Node
	HasPreview            func() bool
	BackToTree            func()
	// TreeWidth is the tree's width as the explorer opens, and
	// OnTreeWidthChange hears a drag of its edge change it: an app keeps
	// the width with it.
	TreeWidth         core.Points
	OnTreeWidthChange func(core.Points)
}

func (e FileExplorer) Render(cx *core.Scope) core.Node {
	compact := core.NewMemo(cx, func() bool { return cx.Env().Window().Width < 620 })
	width := core.NewSignal(cx, e.TreeWidth)
	resize := func(w core.Points) {
		if w == width.GetUntracked() {
			return
		}
		width.Set(w)
		if e.OnTreeWidthChange != nil {
			e.OnTreeWidthChange(w)
		}
	}
	panes := core.Dynamic(func(*core.Scope) core.Node {
		if compact.Get() {
			if !e.HasPreview() {
				return e.Tree
			}
			return core.View{Style: ui.FillRest, Children: []core.Node{ui.Button{Text: "File tree", Label: "File tree", Icon: lucide.ChevronLeft, Variant: ui.ButtonGhost, OnPress: e.BackToTree}, e.Preview}}
		}
		return ui.ResizablePanelGroup{Style: ui.FillRest, HandleLabel: "Resize file tree", Panels: []base.ResizablePanel{
			{Size: width.Get, OnResize: resize, MinSize: 110, MaxSize: 440, Children: []core.Node{e.Tree}},
			{MinSize: 170, Children: []core.Node{e.Preview}},
		}}
	})
	return core.View{Style: ui.FillRest, Children: []core.Node{e.Header, panes}}
}
