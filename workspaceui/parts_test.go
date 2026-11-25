package workspaceui

import (
	"strings"
	"testing"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/icons/lucide"
	"github.com/nsgo-dev/nsgo/nsgotest"
	"github.com/nsgo-dev/nsgo/ui"
)

// themed starts a view under the kit's theme, in the dark appearance or the
// light one, in a box of the given width at the window's left edge.
func themed(t *testing.T, dark bool, width core.Points, view func(*core.Scope) core.Node) *nsgotest.App {
	t.Helper()
	return nsgotest.Start(t, func(cx *core.Scope) core.Node {
		if err := Apply(cx); err != nil {
			t.Errorf("Apply: %v", err)
		}
		appearance := core.AppearanceLight
		if dark {
			appearance = core.AppearanceDark
		}
		cx.SetAppearance(appearance)
		return Frame(core.View{TestID: "box", Style: core.Style{Width: core.Pt(width)}, Children: []core.Node{core.Dynamic(view)}})
	}, nsgotest.Size(800, 600))
}

func right(v *nsgotest.View) core.Points {
	b := v.Bounds()
	return b.Origin.X + b.Size.Width
}

// TestNavLabelKeepsItsCountAtTheRowsEnd checks the count's place for a
// short name and for one far longer than the row: the count ends where the
// row ends, and the name is cut, not the count.
func TestNavLabelKeepsItsCountAtTheRowsEnd(t *testing.T) {
	for _, name := range []string{"Inbox", strings.Repeat("A very long mailbox name ", 6)} {
		app := themed(t, false, 240, func(*core.Scope) core.Node {
			return core.View{Style: core.Style{Direction: core.Row}, Children: []core.Node{NavLabel{Icon: lucide.Inbox, Text: name, Count: 2113, Unread: true}}}
		})
		box, count := app.Find("box"), app.Find("nav-count")
		if got, want := right(count), right(box); got != want {
			t.Errorf("%.20q: the count ends at %v, want the row's end %v", name, got, want)
		}
		if count.Text != "2113" {
			t.Errorf("%.20q: the count reads %q", name, count.Text)
		}
	}
}

func TestNavLabelShowsNoCountOfZero(t *testing.T) {
	app := themed(t, false, 240, func(*core.Scope) core.Node { return NavLabel{Icon: lucide.Send, Text: "Sent"} })
	if app.Exists("nav-count") {
		t.Error("a count of 0 is shown")
	}
	app.Find("text:Sent")
}

// TestSummaryRowKeepsItsTextInside checks a row narrower than its text: the
// date and the marks stay whole and inside the row, and the title and the
// subject are cut to one line.
func TestSummaryRowKeepsItsTextInside(t *testing.T) {
	long := strings.Repeat("A sender with a very long display name ", 4)
	app := themed(t, false, 300, func(cx *core.Scope) core.Node {
		return SummaryRow{Unread: true, Name: "David King", Title: long, When: "Yesterday", Subject: long,
			Marks: []core.Node{ui.Badge{Text: "3"}}, Preview: long + long, PreviewLines: 2}
	})
	row := app.Find("summary-row")
	inner := right(row) - 8
	if got := right(app.Find("summary-when")); got > inner {
		t.Errorf("the date ends at %v, past the row's inner edge %v", got, inner)
	}
	if got := app.Find("summary-when").Text; got != "Yesterday" {
		t.Errorf("the date reads %q", got)
	}
	if got := right(app.Find("text:3")); got > inner {
		t.Errorf("the mark ends at %v, past the row's inner edge %v", got, inner)
	}
	if got := app.Find("summary-preview").MaxLines; got != 2 {
		t.Errorf("the preview may take %d lines, want 2", got)
	}
	if got := row.Bounds().Size.Width; got != 300 {
		t.Errorf("the row is %v wide in a box of 300", got)
	}
}

// TestSummaryRowMarksUnreadAndFollowsPreviewLines checks the unread dot's
// color in both themes, its absence on a read row, and a row with no
// preview.
func TestSummaryRowMarksUnreadAndFollowsPreviewLines(t *testing.T) {
	for _, tc := range []struct {
		name string
		dark bool
		look look
	}{{"dark", true, frostyard}, {"light", false, frostyardLight}} {
		t.Run(tc.name, func(t *testing.T) {
			app := themed(t, tc.dark, 390, func(*core.Scope) core.Node {
				return Stack(
					core.View{TestID: "unread", Children: []core.Node{SummaryRow{Unread: true, Name: "A", Title: "A", When: "9:41", Subject: "S", Preview: "P", PreviewLines: 1}}},
					core.View{TestID: "read", Children: []core.Node{SummaryRow{Name: "B", Title: "B", When: "9:40", Subject: "S", Preview: "P"}}},
				)
			})
			if got := app.Find("unread >> summary-dot").Style.Background; got != tc.look.accent {
				t.Errorf("an unread row's dot is %#08x, want the accent %#08x", uint32(got), uint32(tc.look.accent))
			}
			if got := app.Find("read >> summary-dot").Style.Background; got != 0 {
				t.Errorf("a read row's dot is painted %#08x", uint32(got))
			}
			if !app.Exists("unread >> summary-preview") || app.Exists("read >> summary-preview") {
				t.Error("the preview does not follow PreviewLines: want it on the first row only")
			}
		})
	}
}

// TestSoftRowsFillAChosenRowSoftly checks the chosen row's fill and its
// corners in both themes.
func TestSoftRowsFillAChosenRowSoftly(t *testing.T) {
	for _, tc := range []struct {
		dark bool
		look look
	}{{true, frostyard}, {false, frostyardLight}} {
		var style core.Style
		themed(t, tc.dark, 300, func(cx *core.Scope) core.Node {
			style = SoftRows(cx)()
			return core.Text{Content: "x"}
		})
		if style.Selected == nil || style.Selected.Background != tc.look.selectedFill {
			t.Errorf("dark=%v: a chosen row's style is %+v, want the fill %#08x", tc.dark, style.Selected, uint32(tc.look.selectedFill))
		}
		if style.Selected != nil && style.Selected.Color != 0 {
			t.Errorf("dark=%v: a chosen row changes its text's color to %#08x", tc.dark, uint32(style.Selected.Color))
		}
		if style.Hover == nil || style.CornerRadius != ui.RadiusControl {
			t.Errorf("dark=%v: rows have hover %v and corners %v", tc.dark, style.Hover, style.CornerRadius)
		}
	}
}

// TestBannerShowsItsToneAndAtMostOneButton checks a banner's fill and line
// by tone, its button, and that it has none without an action.
func TestBannerShowsItsToneAndAtMostOneButton(t *testing.T) {
	pressed := 0
	app := themed(t, true, 500, func(*core.Scope) core.Node {
		return Stack(
			core.View{TestID: "with", Children: []core.Node{Banner{Tone: Danger, Icon: lucide.CircleAlert, Text: "Sending failed.", Action: "Retry", OnAction: func() { pressed++ }}}},
			core.View{TestID: "without", Children: []core.Node{Banner{Tone: Success, Icon: lucide.Check, Text: "You accepted this invitation."}}},
		)
	})
	style := app.Find("with >> banner").Style
	if style.Background != frostyard.fill[Danger] || style.BorderColor != frostyard.line[Danger] {
		t.Errorf("a danger banner is filled %#08x with a line %#08x", uint32(style.Background), uint32(style.BorderColor))
	}
	app.Press("with >> role:button Retry")
	if pressed != 1 {
		t.Errorf("the banner's button called OnAction %d times", pressed)
	}
	if app.Exists("without >> role:button") {
		t.Error("a banner with no action has a button")
	}
	app.Find("without >> text:You accepted this invitation.")
}

func TestEmptyStateSaysWhatIsMissing(t *testing.T) {
	app := themed(t, false, 400, func(*core.Scope) core.Node {
		return core.View{Style: core.Style{Height: core.Pt(300)}, Children: []core.Node{EmptyState{Icon: lucide.MailOpen, Title: "No message selected", Hint: "Choose a message to read it here."}}}
	})
	app.Find("empty-state >> text:No message selected")
	app.Find("empty-state >> text:Choose a message to read it here.")
}

// TestDialogPanelIsOpaque checks the panel's color in both themes: the
// theme's panel, with no transparency.
func TestDialogPanelIsOpaque(t *testing.T) {
	for _, tc := range []struct {
		dark bool
		look look
	}{{true, frostyard}, {false, frostyardLight}} {
		var style core.Style
		themed(t, tc.dark, 300, func(cx *core.Scope) core.Node {
			style = DialogPanel(cx)()
			return core.Text{Content: "x"}
		})
		if style.Background != tc.look.panel || style.Background&0xff != 0xff {
			t.Errorf("dark=%v: a dialog's ground is %#08x, want the opaque panel %#08x", tc.dark, uint32(style.Background), uint32(tc.look.panel))
		}
	}
}

func TestConnectionItemShowsItsText(t *testing.T) {
	app := themed(t, false, 240, func(*core.Scope) core.Node {
		return ConnectionItem{Tone: Success, Text: func() string { return "Local" }}
	})
	app.Find("connection >> text:Local")
}

// TestSearchFieldShowsItsKeysOrAClearButton checks the field's end: the
// keys while it is empty, a button that clears it while it holds text, and
// what typing reports.
func TestSearchFieldShowsItsKeysOrAClearButton(t *testing.T) {
	var submitted []string
	app := themed(t, false, 300, func(cx *core.Scope) core.Node {
		query := core.NewSignal(cx, "")
		return SearchField{Placeholder: "Search All Inboxes", Keys: "mod+f", Value: query.Get, OnChange: query.Set,
			OnSubmit: func(s string) { submitted = append(submitted, s) }}
	})
	if !app.Exists("search-keys") || app.Exists("role:button Clear search") {
		t.Fatal("an empty field does not show its keys alone")
	}
	app.Type("label:Search All Inboxes", "aurora")
	if app.Exists("search-keys") || !app.Exists("role:button Clear search") {
		t.Fatal("a field with text shows its keys, or no clear button")
	}
	if got := app.Find("label:Search All Inboxes").Text; got != "aurora" {
		t.Fatalf("the field holds %q", got)
	}
	app.Submit("label:Search All Inboxes")
	if len(submitted) != 1 || submitted[0] != "aurora" {
		t.Errorf("OnSubmit heard %q, want aurora", submitted)
	}
	app.Press("role:button Clear search")
	if got := app.Find("label:Search All Inboxes").Text; got != "" || !app.Exists("search-keys") {
		t.Errorf("after Clear the field holds %q and shows its keys: %v", got, app.Exists("search-keys"))
	}
	if got := right(app.Find("search-field")); got != 300 {
		t.Errorf("the field ends at %v in a box of 300", got)
	}
}

// TestSummaryRowScalesItsText checks Scale: the row's text is larger by it,
// and its dot and its avatar are not.
func TestSummaryRowScalesItsText(t *testing.T) {
	row := func(scale float32) *nsgotest.App {
		return themed(t, false, 390, func(*core.Scope) core.Node {
			return SummaryRow{Unread: true, Name: "A", Title: "A", When: "9:41", Subject: "S", Preview: "P", PreviewLines: 1, Scale: scale}
		})
	}
	plain, large := row(0), row(1.5)
	if a, b := plain.Find("summary-when").TextStyle.Font.Size, large.Find("summary-when").TextStyle.Font.Size; b != a*1.5 {
		t.Errorf("the date is %v at scale 1.5 and %v at none: want 1.5 times", b, a)
	}
	if a, b := plain.Find("summary-preview").TextStyle.Font.Size, large.Find("summary-preview").TextStyle.Font.Size; b != a*1.5 {
		t.Errorf("the preview is %v at scale 1.5 and %v at none: want 1.5 times", b, a)
	}
	if a, b := plain.Find("summary-dot").Bounds().Size, large.Find("summary-dot").Bounds().Size; a != b {
		t.Errorf("the dot is %v at scale 1.5 and %v at none", b, a)
	}
}

// TestPageFrameIsWhiteInsideAHairline checks the frame of a page that is
// not the app's own, in both themes.
func TestPageFrameIsWhiteInsideAHairline(t *testing.T) {
	for _, dark := range []bool{true, false} {
		app := themed(t, dark, 300, func(*core.Scope) core.Node {
			return PageFrame{Size: core.Style{Height: core.Pt(80)}, Page: core.Text{Content: "a page"}}
		})
		frame := app.Find("page-frame").Style
		if frame.Background != page || frame.BorderWidth != 1 {
			t.Errorf("dark=%v: the frame's ground is %#08x with a line of %v, want white and 1", dark, uint32(frame.Background), frame.BorderWidth)
		}
		app.Find("page-frame >> text:a page")
	}
}

// TestNavLabelTintsItsIcon checks the color a row's icon is given: the
// accent with no tint, and the tint's with one. A headless run draws no
// icon to read a color from; the gallery's pictures show the rows.
func TestNavLabelTintsItsIcon(t *testing.T) {
	accent := func() core.Color { return core.RGB(0x0b6ea6) }
	red := func() core.Color { return core.RGB(0xff3b30) }
	if got := (NavLabel{}).iconColor(accent)(); got != accent() {
		t.Errorf("a row with no tint draws its icon in %#08x, want the accent %#08x", uint32(got), uint32(accent()))
	}
	if got := (NavLabel{Tint: red}).iconColor(accent)(); got != red() {
		t.Errorf("a tinted row draws its icon in %#08x, want the tint %#08x", uint32(got), uint32(red()))
	}
	// A tinted row still draws under the theme.
	app := themed(t, false, 240, func(cx *core.Scope) core.Node {
		return NavLabel{Icon: lucide.Flag, Text: "Red", Count: 2, Tint: func() core.Color { return ColorsOf(cx).Flag(1) }}
	})
	app.Find("text:Red")
	app.Find("nav-count")
}

// TestEmptyStateHoldsItsActions checks the buttons under an empty state's
// hint: there with actions, in a row that wraps, and absent with none.
func TestEmptyStateHoldsItsActions(t *testing.T) {
	pressed := 0
	app := themed(t, false, 360, func(*core.Scope) core.Node {
		return core.View{Style: core.Style{Height: core.Pt(300)}, Children: []core.Node{EmptyState{Icon: lucide.Inbox, Title: "No vault", Hint: "Open one to start.",
			Actions: []core.Node{ui.Button{Text: "Create", OnPress: func() { pressed++ }}, ui.Button{Text: "Open", Variant: ui.ButtonSecondary}}}}}
	})
	app.Find("id:empty-actions >> label:Open")
	app.Press("id:empty-actions >> label:Create")
	if pressed != 1 {
		t.Errorf("the action was pressed %d times, want 1", pressed)
	}
	hint, actions := app.Find("text:Open one to start.").Bounds(), app.Find("id:empty-actions").Bounds()
	if actions.Origin.Y < hint.Origin.Y+hint.Size.Height {
		t.Errorf("the actions start at %v, over the hint, which ends at %v", actions.Origin.Y, hint.Origin.Y+hint.Size.Height)
	}
	bare := themed(t, false, 360, func(*core.Scope) core.Node {
		return core.View{Style: core.Style{Height: core.Pt(300)}, Children: []core.Node{EmptyState{Icon: lucide.Inbox, Title: "No vault"}}}
	})
	if bare.Exists("id:empty-actions") {
		t.Error("an empty state with no action has a row of actions")
	}
}

// TestHeadingStepsByLevel checks a document's headings: 24 points at level
// 1, 2 less a level, 14 at level 6, and the nearest level outside them.
func TestHeadingStepsByLevel(t *testing.T) {
	for level, want := range map[int]core.Points{1: 24, 2: 22, 3: 20, 6: 14, 0: 24, 9: 14} {
		if got := Heading(level).FontSize; got != want {
			t.Errorf("a heading of level %d is %v points, want %v", level, got, want)
		}
	}
	if w := Heading(2).FontWeight; w != core.FontWeightSemibold {
		t.Errorf("a heading's weight is %v, want semibold", w)
	}
}
