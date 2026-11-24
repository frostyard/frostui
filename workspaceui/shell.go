// Package workspaceui composes NSGo's themed controls into a reusable workspace.
// Applications supply destinations, content and actions. All colors follow the current theme.
package workspaceui

import (
	"runtime"
	"time"

	"github.com/nsgo-dev/nsgo/base"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/ui"
)

// Shell shows a resizable navigation sidebar and a content pane. On compact
// windows the navigation can take the full width and closes after a selection.
type Shell struct {
	Brand                                                  string
	Header, Actions, SidebarHeader, SidebarFooter, Content []core.Node
	Sections                                               func() []ui.SidebarSection
	Selected                                               func() string
	Navigate                                               func(string)
	// Disclose handles branch rows without navigating away or closing the compact rail.
	Disclose       func(id string) bool
	Inspector      *Inspector
	Auxiliary      *Inspector
	Widths         PaneWidths
	OnWidthsChange func(PaneWidths)
}

// Inspector supplies contextual content for the shell's optional right panel.
// Available controls which destinations offer it. Content is built on demand.
type Inspector struct {
	Icon      func() *core.SVG
	Title     string
	Available func() bool
	Content   func(*core.Scope) core.Node
}

type PaneWidths struct{ Navigation, Inspector, Auxiliary core.Points }

func paneWidth(value, initial, lo, hi core.Points) core.Points {
	if value <= 0 {
		return initial
	}
	return min(hi, max(lo, value))
}

func (s Shell) Render(cx *core.Scope) core.Node {
	compact := core.NewMemo(cx, func() bool { return cx.Env().Window().Width < 720 }).Get
	open := core.NewSignal(cx, !compact())
	cx.Effect(func() { open.Set(!compact()) })
	width := core.NewSignal(cx, paneWidth(s.Widths.Navigation, 246, 180, 420))
	rightWidth := core.NewSignal(cx, paneWidth(s.Widths.Inspector, 300, 240, 420))
	auxWidth := core.NewSignal(cx, paneWidth(s.Widths.Auxiliary, 510, 340, 900))
	var cancel func() bool
	save := func() {
		cancel = nil
		if s.OnWidthsChange != nil {
			s.OnWidthsChange(PaneWidths{Navigation: width.GetUntracked(), Inspector: rightWidth.GetUntracked(), Auxiliary: auxWidth.GetUntracked()})
		}
	}
	resize := func(signal *core.Signal[core.Points]) func(core.Points) {
		return func(w core.Points) {
			if w == signal.GetUntracked() {
				return
			}
			signal.Set(w)
			if s.OnWidthsChange == nil {
				return
			}
			if cancel != nil {
				cancel()
			}
			cancel = cx.AfterFunc(200*time.Millisecond, save)
		}
	}
	cx.Cleanup(func() {
		if cancel != nil {
			cancel()
			save()
		}
	})
	auxOpen := core.NewSignal(cx, false)
	rightOpen := core.NewSignal(cx, false)
	available := core.NewMemo(cx, func() bool { return s.Inspector != nil && (s.Inspector.Available == nil || s.Inspector.Available()) })
	sideBySide := core.NewMemo(cx, func() bool { return cx.Env().Window().Width >= 1000 })
	bothRails := core.NewMemo(cx, func() bool { return cx.Env().Window().Width >= 1550 })
	// When the window gets too narrow for both inspectors, keep the file
	// preview open beside its conversation.
	cx.Effect(func() {
		if !bothRails.Get() && rightOpen.Get() && auxOpen.Get() {
			rightOpen.Set(false)
		}
	})
	var lastDestination string
	cx.Effect(func() {
		destination := ""
		if s.Selected != nil {
			destination = s.Selected()
		}
		// Opening a related session from a full-width inspector should show
		// its conversation. A docked inspector can follow the selection.
		if !available.Get() || (destination != lastDestination && !sideBySide.GetUntracked()) {
			rightOpen.Set(false)
			auxOpen.Set(false)
		}
		lastDestination = destination
	})
	cx.OnBackPress(func() bool {
		if auxOpen.Get() {
			auxOpen.Set(false)
			return true
		}
		if !rightOpen.Get() {
			return false
		}
		rightOpen.Set(false)
		return true
	})
	// Keep the conversation, draft, and scroll position when a narrow window
	// displays the inspector in the space normally occupied by the content.
	center := core.NewHolder(cx, func(*core.Scope) core.Node { return core.View{Style: ui.FillRest, Children: s.Content} })
	nav := ui.Sidebar{SectionsFunc: s.Sections, Selected: s.Selected, OnSelect: func(id string) {
		if s.Disclose != nil && s.Disclose(id) {
			return
		}
		s.Navigate(id)
		if compact() {
			open.Set(false)
		}
	}}
	return core.Dynamic(func(*core.Scope) core.Node {
		actions := s.Actions
		content := center.Place()
		hasPanel := available.Get()
		docked := hasPanel && sideBySide.Get()
		setDetails := func(open bool) {
			if open && !bothRails.GetUntracked() {
				auxOpen.Set(false)
			}
			rightOpen.Set(open)
		}
		if hasPanel && s.Auxiliary != nil && (s.Auxiliary.Available == nil || s.Auxiliary.Available()) {
			aux := s.Auxiliary
			actions = append(append([]core.Node{}, actions...), ui.IconButton{Label: aux.Title, Icon: aux.Icon, Selected: auxOpen.Get, Tint: func() core.Color {
				colors := ui.ColorsOf(cx)
				if auxOpen.Get() {
					return colors.Accent()
				}
				return colors.Text()
			}, OnPress: func() {
				if compact() {
					open.Set(false)
				}
				if !bothRails.GetUntracked() {
					rightOpen.Set(false)
				}
				auxOpen.Set(!auxOpen.Get())
			}})
			content = core.Show{When: auxOpen.Get, Then: func(*core.Scope) core.Node {
				panel := core.Dynamic(aux.Content)
				if !docked {
					return panel
				}
				return ui.ResizablePanelGroup{Style: ui.FillRest, HandleLabel: "Resize " + aux.Title, Panels: []base.ResizablePanel{
					{MinSize: 280, Children: []core.Node{center.Place()}},
					{Size: auxWidth.Get, MinSize: 340, MaxSize: 900, OnResize: resize(auxWidth), Children: []core.Node{panel}},
				}}
			}, Else: func(*core.Scope) core.Node { return center.Place() }}
		}
		if hasPanel && !docked {
			actions = append(append([]core.Node{}, actions...), ui.IconButton{Label: s.Inspector.Title, Icon: lucide.PanelRight, OnPress: func() {
				if compact() {
					open.Set(false)
				}
				setDetails(!rightOpen.Get())
			}})
			mainContent := content
			content = core.Show{When: rightOpen.Get, Then: func(*core.Scope) core.Node { return core.Dynamic(s.Inspector.Content) }, Else: func(*core.Scope) core.Node { return mainContent }}
		}
		if compact() {
			buttons := []core.Node{ui.Button{Label: "Navigation", Icon: lucide.PanelLeft, Variant: ui.ButtonGhost, OnPress: func() { rightOpen.Set(false); auxOpen.Set(false); open.Set(!open.Get()) }}}
			heading := core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0)}, Children: s.Header}
			buttons = append(buttons, heading)
			buttons = append(buttons, actions...)
			navContent := append([]core.Node{}, s.SidebarHeader...)
			navContent = append(navContent, nav)
			navContent = append(navContent, s.SidebarFooter...)
			// The one pane paints its own ground: NSGo's shell does it in a wide
			// window, and without one a dark appearance shows its light text on
			// the window's white.
			colors := ui.ColorsOf(cx)
			return core.View{TestID: "compact-shell", Style: ui.FillRest, StyleFunc: func() core.Style {
				return core.Style{Background: colors.Background(), Color: colors.Text()}
			}, Children: []core.Node{core.View{Style: core.Style{Direction: core.Row, Gap: 4, AlignItems: core.AlignCenter}, Children: buttons}, core.Show{When: open.Get, Then: func(*core.Scope) core.Node { return core.Scroll{Style: ui.FillRest, Children: navContent} }, Else: func(*core.Scope) core.Node { return content }}}}
		}
		shell := ui.SidebarShell{
			Variant: ui.SidebarShellInset, DefaultWidth: 246, MinWidth: 180, MaxWidth: 420,
			Open: open.Get, OnOpenChange: open.Set, Width: width.Get, OnWidthChange: resize(width),
			ToggleLabel: "Navigation", ResizeLabel: "Resize navigation",
			SidebarActions: s.brand(cx),
			Sidebar:        append(append([]core.Node{}, s.SidebarHeader...), nav), SidebarFooter: s.SidebarFooter,
			Header: s.Header, Actions: actions, Content: []core.Node{content},
		}
		if docked {
			shell.Right = []core.Node{core.Dynamic(s.Inspector.Content)}
			shell.RightHeader = []core.Node{core.Text{Content: s.Inspector.Title, Style: ui.TypeTitle}}
			shell.RightToggleLabel = s.Inspector.Title
			shell.RightResizeLabel = "Resize right panel"
			shell.RightOpen, shell.OnRightOpenChange = rightOpen.Get, setDetails
			shell.RightWidth, shell.OnRightWidthChange = rightWidth.Get, resize(rightWidth)
			shell.RightDefaultWidth, shell.RightMinWidth, shell.RightMaxWidth = 300, 240, 420
		}
		// GTK drops its resize borders along with a hidden title bar.
		// Render the standard shell, then restore system decorations on Linux.
		node := shell.Render(cx)
		if systemTitleBar(cx) {
			desktop.SetTitleBar(cx, desktop.TitleBarDefault, 0)
		}
		return node
	})
}

// systemTitleBar reports whether the window keeps the system's title bar,
// which names the app: on Linux, in a desktop window (ADR-0014).
func systemTitleBar(cx *core.Scope) bool {
	return runtime.GOOS == "linux" && desktop.IsAvailable(cx)
}

// brand is the app's name at the head of the navigation, where no system
// title bar shows it.
func (s Shell) brand(cx *core.Scope) []core.Node {
	if s.Brand == "" || systemTitleBar(cx) {
		return nil
	}
	return []core.Node{core.Text{Content: s.Brand, Style: ui.TypeTitle}}
}

// Page gives settings, forms and detail screens a readable measure on a wide window.
func Page(children ...core.Node) core.Node {
	return core.Scroll{Style: ui.FillRest, Children: []core.Node{core.View{
		Style: core.Style{Width: core.Pct(100), MaxWidth: core.Pt(920), AlignSelf: core.AlignCenter, Padding: core.All(20), Gap: 14}, Children: children,
	}}}
}

func Row(children ...core.Node) core.Node {
	return core.View{Style: core.Style{Direction: core.Row, Wrap: core.Wrap, AlignItems: core.AlignCenter, Gap: 8}, Children: children}
}

func Stack(children ...core.Node) core.Node {
	return core.View{Style: core.Style{Gap: 8}, Children: children}
}
