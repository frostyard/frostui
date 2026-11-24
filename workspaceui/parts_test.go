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
