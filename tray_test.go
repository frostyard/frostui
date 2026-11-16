package frostui

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"slices"
	"testing"

	"github.com/egoist/mygo"
)

// fakeTrayIcon records what a Tray asks of MyGo's tray icon.
type fakeTrayIcon struct {
	menus     []*mygo.Menu
	tips      []string
	icons     [][]byte
	destroyed int
	iconErr   error
}

func (f *fakeTrayIcon) SetIcon(png []byte, _ bool) error {
	f.icons = append(f.icons, png)
	return f.iconErr
}
func (f *fakeTrayIcon) SetToolTip(tip string) { f.tips = append(f.tips, tip) }
func (f *fakeTrayIcon) SetMenu(m *mygo.Menu)  { f.menus = append(f.menus, m) }
func (f *fakeTrayIcon) Destroy()              { f.destroyed++ }

// trayIconPNG is a 32 by 32 PNG, opaque blue.
func trayIconPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for i := range img.Pix {
		if i%4 == 2 || i%4 == 3 {
			img.Pix[i] = 0xff
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// testTray is a Tray on a fake icon whose window calls are recorded, and
// whose watcher check answers hosted.
type testTray struct {
	*Tray
	icon          *fakeTrayIcon
	shown, hidden []*mygo.Window
	hostedAsks    int
	hostedAnswer  bool
}

func newTestTray(t *testing.T, opts TrayOptions) *testTray {
	t.Helper()
	if opts.Window == nil {
		opts.Window = &mygo.Window{}
	}
	if opts.Icon == nil {
		opts.Icon = trayIconPNG(t)
	}
	tray, err := newTray(opts)
	if err != nil {
		t.Fatal(err)
	}
	tt := &testTray{Tray: tray, icon: &fakeTrayIcon{}, hostedAnswer: true}
	tray.icon = tt.icon
	tray.show = func(w *mygo.Window) { tt.shown = append(tt.shown, w) }
	tray.hide = func(w *mygo.Window) { tt.hidden = append(tt.hidden, w) }
	tray.hosted = func() bool { tt.hostedAsks++; return tt.hostedAnswer }
	return tt
}

// menuShape describes a menu as its items' IDs, a separator as "-", the
// Quit role as "quit", with a disabled item marked "(off)".
func menuShape(m *mygo.Menu) []string {
	var out []string
	for _, it := range m.Items() {
		switch {
		case it.Type == mygo.MenuItemSeparator:
			out = append(out, "-")
		case it.Role == mygo.RoleQuit:
			out = append(out, "quit")
		case it.Disabled:
			out = append(out, it.ID+"(off)")
		default:
			out = append(out, it.ID)
		}
	}
	return out
}

func TestTrayMenuShape(t *testing.T) {
	var clicked []string
	tt := newTestTray(t, TrayOptions{OpenLabel: "Open Mail", Items: []TrayItem{
		{ID: "new", Label: "New message", Click: func() { clicked = append(clicked, "new") }},
		{ID: "later", Label: "Later", Disabled: true},
	}})

	menu := tt.menu()
	if got, want := menuShape(menu), []string{trayOpenID, "new", "later(off)", "-", "quit"}; !slices.Equal(got, want) {
		t.Fatalf("menu without a status = %q, want %q", got, want)
	}
	if got := menu.ItemByID(trayOpenID).Label; got != "Open Mail" {
		t.Errorf("Open label = %q, want %q", got, "Open Mail")
	}
	if got := menu.ItemByID("new").Label; got != "New message" {
		t.Errorf("item label = %q, want %q", got, "New message")
	}
	menu.ItemByID(trayOpenID).Click(nil, nil)
	if len(tt.shown) != 1 || tt.shown[0] != tt.opts.Window {
		t.Errorf("Open showed %v, want the tray's window once", tt.shown)
	}
	menu.ItemByID("new").Click(nil, nil)
	if !slices.Equal(clicked, []string{"new"}) {
		t.Errorf("clicks = %q, want [new]", clicked)
	}
	if menu.ItemByID("later").Click != nil {
		t.Error("an item with no Click has a Click in the menu")
	}

	tt.SetStatus("3 unread")
	if n := len(tt.icon.menus); n != 1 {
		t.Fatalf("SetStatus set %d menus, want 1", n)
	}
	menu = tt.icon.menus[0]
	if got, want := menuShape(menu), []string{trayStatusID + "(off)", "-", trayOpenID, "new", "later(off)", "-", "quit"}; !slices.Equal(got, want) {
		t.Fatalf("menu with a status = %q, want %q", got, want)
	}
	if got := menu.ItemByID(trayStatusID).Label; got != "3 unread" {
		t.Errorf("status line = %q, want %q", got, "3 unread")
	}

	tt.SetItems(nil)
	if got, want := menuShape(tt.icon.menus[1]), []string{trayStatusID + "(off)", "-", trayOpenID, "-", "quit"}; !slices.Equal(got, want) {
		t.Errorf("menu with no items = %q, want %q", got, want)
	}
	tt.SetStatus("")
	if got, want := menuShape(tt.icon.menus[2]), []string{trayOpenID, "-", "quit"}; !slices.Equal(got, want) {
		t.Errorf("menu after the status went = %q, want %q", got, want)
	}
}

func TestTrayOpenLabelDefaultsToOpen(t *testing.T) {
	tt := newTestTray(t, TrayOptions{})
	if got := tt.menu().ItemByID(trayOpenID).Label; got != "Open" {
		t.Errorf("Open label = %q, want %q", got, "Open")
	}
}

func TestTrayKeepsItsOwnCopyOfItems(t *testing.T) {
	items := []TrayItem{{ID: "a", Label: "A"}}
	tt := newTestTray(t, TrayOptions{Items: items})
	items[0].ID = "changed"
	if got := menuShape(tt.menu()); !slices.Contains(got, "a") {
		t.Errorf("menu = %q after the caller changed its slice, want item a kept", got)
	}
	more := []TrayItem{{ID: "b", Label: "B"}}
	tt.SetItems(more)
	more[0].ID = "changed"
	if got := menuShape(tt.menu()); !slices.Contains(got, "b") {
		t.Errorf("menu = %q after the caller changed its slice, want item b kept", got)
	}
}

func TestTraySetStatusSetsTheToolTipOnlyOnAChange(t *testing.T) {
	tt := newTestTray(t, TrayOptions{ToolTip: "Mail"})
	tt.SetStatus("3 unread")
	tt.SetStatus("3 unread")
	if got, want := tt.icon.tips, []string{"Mail — 3 unread"}; !slices.Equal(got, want) {
		t.Errorf("tooltips = %q, want %q", got, want)
	}
	if n := len(tt.icon.menus); n != 1 {
		t.Errorf("menus set = %d, want 1: an unchanged status sets nothing", n)
	}
	tt.SetStatus("")
	if got := tt.icon.tips[len(tt.icon.tips)-1]; got != "Mail" {
		t.Errorf("tooltip with no status = %q, want %q", got, "Mail")
	}
	bare := newTestTray(t, TrayOptions{})
	bare.SetStatus("Idle")
	if got := bare.icon.tips; !slices.Equal(got, []string{"Idle"}) {
		t.Errorf("tooltip with no ToolTip = %q, want [Idle]", got)
	}
}

// TestTrayCloseToTray is the close table: a close hides the window only
// when the setting is on, the window is the tray's, no quit runs, the
// tray is not destroyed and a panel hosts the icon.
func TestTrayCloseToTray(t *testing.T) {
	cases := []struct {
		name         string
		enabled      bool
		hosted       bool
		quitting     bool
		otherWindow  bool
		destroyed    bool
		want         bool
		wantAskHosts bool
	}{
		{name: "on, hosted, not quitting", enabled: true, hosted: true, want: true, wantAskHosts: true},
		{name: "on, not hosted", enabled: true, hosted: false, want: false, wantAskHosts: true},
		{name: "on, hosted, quitting", enabled: true, hosted: true, quitting: true, want: false},
		{name: "on, not hosted, quitting", enabled: true, quitting: true, want: false},
		{name: "off, hosted", enabled: false, hosted: true, want: false},
		{name: "off, not hosted", enabled: false, want: false},
		{name: "off, hosted, quitting", enabled: false, hosted: true, quitting: true, want: false},
		{name: "on, hosted, another window", enabled: true, hosted: true, otherWindow: true, want: false},
		{name: "on, hosted, destroyed", enabled: true, hosted: true, destroyed: true, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tt := newTestTray(t, TrayOptions{})
			tt.hostedAnswer = tc.hosted
			tt.quitting = tc.quitting
			if tc.destroyed {
				tt.Destroy()
			}
			e := &mygo.CloseEvent{Window: tt.opts.Window}
			if tc.otherWindow {
				e.Window = &mygo.Window{}
			}
			got := tt.CloseToTray(e, tc.enabled)
			if got != tc.want {
				t.Errorf("CloseToTray = %v, want %v", got, tc.want)
			}
			if e.DefaultPrevented() != tc.want {
				t.Errorf("close prevented = %v, want %v", e.DefaultPrevented(), tc.want)
			}
			if wantHidden := tc.want; (len(tt.hidden) == 1) != wantHidden || len(tt.hidden) > 1 {
				t.Errorf("hidden = %v, want hidden %v", tt.hidden, wantHidden)
			} else if wantHidden && tt.hidden[0] != tt.opts.Window {
				t.Error("CloseToTray hid another window")
			}
			if (tt.hostedAsks > 0) != tc.wantAskHosts {
				t.Errorf("watcher asked %d times, want asked %v", tt.hostedAsks, tc.wantAskHosts)
			}
		})
	}
}

func TestTrayCloseToTrayAsksTheWatcherAtEachClose(t *testing.T) {
	tt := newTestTray(t, TrayOptions{})
	e := &mygo.CloseEvent{Window: tt.opts.Window}
	if !tt.CloseToTray(e, true) {
		t.Fatal("first close: CloseToTray = false, want true while hosted")
	}
	tt.hostedAnswer = false // the panel went away
	e = &mygo.CloseEvent{Window: tt.opts.Window}
	if tt.CloseToTray(e, true) {
		t.Error("second close: CloseToTray = true, want false once no panel hosts the icon")
	}
	if tt.hostedAsks != 2 {
		t.Errorf("watcher asked %d times, want 2", tt.hostedAsks)
	}
}

func TestTrayCloseToTrayIgnoresANilEvent(t *testing.T) {
	tt := newTestTray(t, TrayOptions{})
	if tt.CloseToTray(nil, true) {
		t.Error("CloseToTray(nil) = true, want false")
	}
}

func TestTrayDestroyRunsOnce(t *testing.T) {
	tt := newTestTray(t, TrayOptions{})
	offs := 0
	tt.offs = []func(){func() { offs++ }, func() { offs++ }}
	tt.Destroy()
	tt.Destroy()
	if tt.icon.destroyed != 1 || offs != 2 {
		t.Errorf("destroyed %d times, %d listeners removed; want 1 and 2", tt.icon.destroyed, offs)
	}
	tt.SetStatus("after")
	tt.SetDot(true)
	tt.SetItems(nil)
	if len(tt.icon.menus) != 0 || len(tt.icon.icons) != 0 || len(tt.icon.tips) != 0 {
		t.Error("a destroyed tray still changed its icon")
	}
}

func TestNewTrayRejectsBadOptions(t *testing.T) {
	if _, err := newTray(TrayOptions{Icon: trayIconPNG(t)}); err == nil {
		t.Error("newTray with no Window succeeded, want an error")
	}
	if _, err := newTray(TrayOptions{Window: &mygo.Window{}, Icon: []byte("not a png")}); err == nil {
		t.Error("newTray with an icon that is not a PNG succeeded, want an error")
	}
	if _, err := NewTray(TrayOptions{Icon: trayIconPNG(t)}); err == nil {
		t.Error("NewTray with no Window succeeded, want an error")
	}
}

func TestTraySetDot(t *testing.T) {
	icon := trayIconPNG(t)
	tt := newTestTray(t, TrayOptions{Icon: icon})
	tt.SetDot(false) // no change: no icon set
	tt.SetDot(true)
	tt.SetDot(true)
	if len(tt.icon.icons) != 1 {
		t.Fatalf("icons set = %d, want 1", len(tt.icon.icons))
	}
	dotted := tt.icon.icons[0]
	if bytes.Equal(dotted, icon) {
		t.Fatal("SetDot(true) set the plain icon")
	}
	img, err := png.Decode(bytes.NewReader(dotted))
	if err != nil {
		t.Fatalf("the dotted icon is not a PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != 32 || b.Dy() != 32 {
		t.Fatalf("dotted icon is %v, want 32 by 32", b)
	}
	blue := color.NRGBA{B: 0xff, A: 0xff}
	nrgba := func(x, y int) color.NRGBA { return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) }
	// The dot's radius is 32/5 = 6.4 pixels, its center (25.6, 6.4); the
	// clear ring 1.6 pixels outside it.
	if got := nrgba(25, 6); got != trayDotColor {
		t.Errorf("dot center = %v, want %v", got, trayDotColor)
	}
	if got := nrgba(18, 6); got.A != 0 {
		t.Errorf("ring left of the dot = %v, want clear", got)
	}
	for _, p := range []image.Point{{0, 0}, {0, 31}, {31, 31}, {10, 20}} {
		if got := nrgba(p.X, p.Y); got != blue {
			t.Errorf("icon at %v = %v, want it unchanged (%v)", p, got, blue)
		}
	}

	tt.SetDot(false)
	if got := tt.icon.icons[len(tt.icon.icons)-1]; !bytes.Equal(got, icon) {
		t.Error("SetDot(false) did not give back the plain icon")
	}
}

func TestTraySetDotKeepsTheStateWhenThePlatformRefuses(t *testing.T) {
	tt := newTestTray(t, TrayOptions{})
	tt.icon.iconErr = errors.New("refused")
	tt.SetDot(true)
	if !tt.dot {
		t.Error("the dot state did not follow SetDot")
	}
}

func TestTrayDotOnASmallIcon(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	out, err := trayDotted(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(out)); err != nil {
		t.Errorf("dotted small icon is not a PNG: %v", err)
	}
}
