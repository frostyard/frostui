package workspaceui

import (
	"strconv"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui"
)

// NavLabel is the content of a row of NSGo's sidebar (SidebarItem.Content):
// an icon, a name, and a count at the row's end in figures of one width. A
// count of 0 is not shown. An Unread count takes the text's color and a
// heavier weight; any other is muted.
type NavLabel struct {
	Icon   func() *core.SVG
	Text   string
	Count  int
	Unread bool
}

func (n NavLabel) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	name := ui.Text(n.Text, 13, ui.Regular, colors.Text)
	name.MaxLines, name.Style.FlexGrow, name.Style.FlexShrink = 1, 1, 1
	name.Style.MinWidth = core.Pt(0)
	parts := []core.Node{ui.Icon(n.Icon, 16, colors.Accent), name}
	if n.Count > 0 {
		ink, weight := colors.Muted, ui.Regular
		if n.Unread {
			ink, weight = colors.Text, ui.Semibold
		}
		count := ui.Tabular(ui.Text(strconv.Itoa(n.Count), 12, weight, ink))
		count.TestID = "nav-count"
		parts = append(parts, count)
	}
	return core.View{Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8, FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0)}, Children: parts}
}

// SummaryRow is the content of a row of a list of things with a sender or
// an owner, such as messages: a gutter for the unread dot, an avatar, then
// the title with When at the row's end, the subject with the row's marks at
// its end, and up to PreviewLines lines of preview. The row keeps its text
// inside its width: a long title or subject is cut, not the date or a mark.
type SummaryRow struct {
	// Unread shows the dot and sets the subject in a heavier weight.
	Unread bool
	// Name is whose avatar it is: Image, or the initials of Name.
	Name  string
	Image core.ImageSource
	// Title is the first line, usually the sender; When ends it.
	Title, When string
	Subject     string
	// Marks end the subject's line: small icons and a count.
	Marks []core.Node
	// Preview shows on up to PreviewLines lines; 0 leaves it out.
	Preview      string
	PreviewLines int
}

func (r SummaryRow) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	dot := core.View{TestID: "summary-dot", Style: core.Style{Width: core.Pt(8), Height: core.Pt(8), CornerRadius: 4}}
	weight := ui.Regular
	if r.Unread {
		dot.StyleFunc = ui.Filled(colors.Accent)
		weight = ui.Semibold
	}
	cut := func(t core.Text) core.Text {
		t.MaxLines, t.Style.FlexGrow, t.Style.FlexShrink, t.Style.MinWidth = 1, 1, 1, core.Pt(0)
		return t
	}
	when := ui.Tabular(ui.Text(r.When, 12, ui.Regular, colors.Muted))
	when.TestID = "summary-when"
	line := core.Style{Direction: core.Row, AlignItems: core.AlignCenter}
	lines := []core.Node{
		core.View{Style: line.Merge(core.Style{Gap: 8}), Children: []core.Node{cut(ui.Text(r.Title, 13, ui.Semibold, colors.Text)), when}},
		core.View{Style: line.Merge(core.Style{Gap: 6}), Children: append([]core.Node{cut(ui.Text(r.Subject, 13, weight, colors.Text))}, r.Marks...)},
	}
	if r.PreviewLines > 0 && r.Preview != "" {
		preview := ui.Text(r.Preview, 12, ui.Regular, colors.Muted)
		preview.MaxLines, preview.Style.LineHeight, preview.TestID = r.PreviewLines, 16, "summary-preview"
		lines = append(lines, preview)
	}
	return core.View{TestID: "summary-row", Style: core.Style{Direction: core.Row, AlignItems: core.AlignStart, Gap: 8, Padding: core.All(8), Width: core.Pct(100)}, Children: []core.Node{
		core.View{Style: core.Style{Width: core.Pt(8), Height: core.Pt(32), JustifyContent: core.JustifyCenter}, Children: []core.Node{dot}},
		ui.Avatar{Name: r.Name, Image: r.Image, Size: 32},
		core.View{Style: core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0), Gap: 2}, Children: lines},
	}}
}

// SoftRows is the style of the rows of NSGo's list (List.RowStyleFunc):
// rounded, a hover fill under the pointer, and the Frostyard soft fill when
// chosen. A chosen row keeps its own text colors (ADR-0014).
func SoftRows(cx *core.Scope) func() core.Style {
	colors, kit := ui.ColorsOf(cx), ColorsOf(cx)
	return func() core.Style {
		return core.Style{CornerRadius: ui.RadiusControl,
			Hover:    &core.Style{Background: colors.Hover()},
			Selected: &core.Style{Background: kit.Selected()}}
	}
}

// Banner is a notice inside a pane: a tone's fill and line, an icon in the
// tone's color, a sentence, and at most one button at its end. It is for
// what a pane has to say about its content (blocked images, a failed send);
// Notice is for the app's own messages over the window.
type Banner struct {
	Tone Tone
	Icon func() *core.SVG
	Text string
	// Action names the button; empty leaves it out.
	Action   string
	OnAction func()
}

func (b Banner) Render(cx *core.Scope) core.Node {
	colors, kit := ui.ColorsOf(cx), ColorsOf(cx)
	ink := [toneCount]func() core.Color{Neutral: colors.Muted, Accent: colors.Accent, Success: colors.Up, Warning: colors.Warn, Danger: colors.Danger}
	tone := b.Tone
	if tone >= toneCount {
		tone = Neutral
	}
	text := core.Text{Content: b.Text, Style: ui.TypeBody.Merge(core.Style{FlexGrow: 1, FlexShrink: 1, MinWidth: core.Pt(0)})}
	children := []core.Node{ui.Icon(b.Icon, 16, ink[tone]), text}
	if b.Action != "" {
		children = append(children, ui.Button{Text: b.Action, Variant: ui.ButtonSecondary, Size: ui.SizeSmall, OnPress: b.OnAction})
	}
	return core.View{Role: core.RoleStatus, Label: b.Text, TestID: "banner", Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 10, Padding: core.Symmetric(8, 12), CornerRadius: ui.RadiusControl, BorderWidth: 1}, StyleFunc: func() core.Style {
		return core.Style{Background: kit.Fill(tone), BorderColor: kit.Line(tone)}
	}, Children: children}
}

// EmptyState is what a pane shows with nothing in it: an icon, a line that
// says so, and a hint of what to do.
type EmptyState struct {
	Icon        func() *core.SVG
	Title, Hint string
}

func (e EmptyState) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	children := []core.Node{ui.Icon(e.Icon, 32, colors.Faint), core.Text{Content: e.Title, Style: ui.TypeTitle}}
	if e.Hint != "" {
		children = append(children, core.Text{Content: e.Hint, Style: ui.TypeCaption.Merge(core.Style{TextAlign: core.TextAlignCenter}), StyleFunc: ui.Colored(colors.Muted)})
	}
	return core.View{TestID: "empty-state", Style: ui.FillRest.Merge(core.Style{AlignItems: core.AlignCenter, JustifyContent: core.JustifyCenter, Gap: 8, Padding: core.All(24)}), Children: children}
}

// DialogPanel is the ground of a dialog (Dialog.StyleFunc and
// AlertDialog.StyleFunc): Frostyard's panel, opaque. NSGo's own is the
// page's color at 92% over a blur, and a platform that draws no blur, as
// Linux, shows the view under the dialog through it.
func DialogPanel(cx *core.Scope) func() core.Style {
	kit := ColorsOf(cx)
	return func() core.Style { return core.Style{Background: kit.Panel()} }
}

// ConnectionItem is the item at the foot of a navigation that says where
// the app is connected: a dot in the tone's color and a line of text.
// Success is a live connection, Warning one being made again.
type ConnectionItem struct {
	Tone Tone
	Text func() string
}

func (c ConnectionItem) Render(cx *core.Scope) core.Node {
	colors := ui.ColorsOf(cx)
	ink := [toneCount]func() core.Color{Neutral: colors.Muted, Accent: colors.Accent, Success: colors.Up, Warning: colors.Warn, Danger: colors.Danger}
	tone := c.Tone
	if tone >= toneCount {
		tone = Neutral
	}
	text := core.Text{ContentFunc: c.Text, MaxLines: 1, Style: ui.TypeCaption.Merge(core.Style{FlexShrink: 1, MinWidth: core.Pt(0)}), StyleFunc: ui.Colored(colors.Muted)}
	return core.View{TestID: "connection", Style: core.Style{Direction: core.Row, AlignItems: core.AlignCenter, Gap: 8, Padding: core.Symmetric(8, 12), FlexShrink: 1, MinWidth: core.Pt(0)}, Children: []core.Node{
		core.View{Style: core.Style{Width: core.Pt(8), Height: core.Pt(8), CornerRadius: 4}, StyleFunc: ui.Filled(ink[tone])},
		text,
	}}
}
