package frostui

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"github.com/egoist/mygo"
)

// TrayItem is an app's entry of a Tray's menu, between Open and Quit.
type TrayItem struct {
	// ID names the item for MyGo's Menu.ItemByID. IDs that start with
	// "frostui.tray." belong to the kit.
	ID string
	// Label is what the menu shows.
	Label string
	// Disabled greys the item out.
	Disabled bool
	// Click runs on the UI thread when the item is chosen.
	Click func()
}

// TrayOptions configures NewTray.
type TrayOptions struct {
	// ID names the icon across launches (MyGo's TrayOptions.ID): the
	// panel keeps its settings for the icon by it.
	ID string
	// ToolTip is the hover text. A status line (SetStatus) follows it.
	// GNOME shows no tooltip: the status line carries what it says.
	ToolTip string
	// OpenLabel is the label of the menu's first item, which shows and
	// focuses Window, such as "Open Frostmail"; empty is "Open".
	OpenLabel string
	// Icon is a PNG image in color, 32 pixels or more on a side. Draw it
	// in color on Linux: GNOME's top bar is dark in both modes, and MyGo
	// whitens a template icon only while the desktop prefers dark.
	Icon []byte
	// Window is the window that Open and a click on the icon show and
	// focus, and that CloseToTray hides. It must not be nil.
	Window *mygo.Window
	// Items are the app's entries between Open and Quit.
	Items []TrayItem
}

// The IDs of the kit's own items of a Tray's menu.
const (
	trayStatusID = "frostui.tray.status"
	trayOpenID   = "frostui.tray.open"
	trayQuitID   = "frostui.tray.quit"
)

// trayDotColor is the dot of SetDot: the Frostyard dark theme's Danger,
// which reads on a dark and on a light panel.
var trayDotColor = color.NRGBA{R: 0xe0, G: 0x7a, B: 0x72, A: 0xff}

// trayIcon is what a Tray asks of MyGo's tray icon; a test gives a fake.
type trayIcon interface {
	SetIcon(png []byte, template bool) error
	SetToolTip(tip string)
	SetMenu(m *mygo.Menu)
	Destroy()
}

// Tray is an icon in the notification area or menu bar, with 1 menu
// shape for every app: a status line (SetStatus), Open, the app's items,
// and Quit. It belongs to the UI thread: make it in mygo.App.WhenReady,
// and call its methods there or in Window.Update.
//
// A click on the icon opens the menu where the platform does so: on
// Linux, where the panel (AppIndicator) draws it, and on macOS. The
// panel's activation (GNOME: a double click; KDE Plasma: a click) and a
// click on Windows show and focus the window; a right click on Windows
// opens the menu.
//
// A tray does not keep the app alive by itself: the app keeps its window
// open, or hidden by CloseToTray.
type Tray struct {
	opts   TrayOptions
	icon   trayIcon
	dotted []byte

	hosted func() bool
	show   func(*mygo.Window)
	hide   func(*mygo.Window)

	status   string
	dot      bool
	quitting bool
	offs     []func()
}

// NewTray adds the tray icon with its menu. Call it in
// mygo.App.WhenReady, after the window exists. It fails when the icon is
// not a PNG, when Window is nil, or when the platform has no tray: on
// Linux, MyGo needs libayatana-appindicator3. Where no panel hosts the
// icon (GNOME without the AppIndicator extension), it succeeds and no
// icon shows: CloseToTray asks TrayHosted at each close.
func NewTray(opts TrayOptions) (*Tray, error) {
	t, err := newTray(opts)
	if err != nil {
		return nil, err
	}
	native, err := mygo.NewTray(mygo.TrayOptions{
		ID: opts.ID, Icon: opts.Icon, ToolTip: t.toolTip(), Menu: t.menu(),
	})
	if err != nil {
		return nil, fmt.Errorf("frostui: tray: %w", err)
	}
	t.icon = native
	t.offs = append(t.offs,
		native.OnClick(t.open),
		mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { t.quitting = true }),
		mygo.App.OnQuitCanceled(func() { t.quitting = false }),
	)
	return t, nil
}

// newTray checks opts and makes a Tray with no icon yet, with MyGo's
// window calls and TrayHosted. A test replaces them and sets icon.
func newTray(opts TrayOptions) (*Tray, error) {
	if opts.Window == nil {
		return nil, errors.New("frostui: tray: TrayOptions.Window is nil")
	}
	dotted, err := trayDotted(opts.Icon)
	if err != nil {
		return nil, err
	}
	opts.Items = append([]TrayItem(nil), opts.Items...)
	return &Tray{
		opts:   opts,
		dotted: dotted,
		hosted: TrayHosted,
		show:   func(w *mygo.Window) { w.Show(); w.Focus() },
		hide:   func(w *mygo.Window) { w.Hide() },
	}, nil
}

// SetStatus sets the menu's first line, disabled, and the line after the
// tooltip, such as "3 unread". Empty removes it. Never make the dot the
// only cue: put what it means here too.
func (t *Tray) SetStatus(line string) {
	if line == t.status {
		return
	}
	t.status = line
	if t.icon != nil {
		t.icon.SetToolTip(t.toolTip())
		t.icon.SetMenu(t.menu())
	}
}

// SetDot draws a dot on the icon's top right corner, or takes it away. A
// tray icon has no badge of its own on any platform, so the dot is part
// of the image.
func (t *Tray) SetDot(on bool) {
	if on == t.dot {
		return
	}
	t.dot = on
	if t.icon == nil {
		return
	}
	img := t.opts.Icon
	if on {
		img = t.dotted
	}
	// The icon was checked in NewTray; a platform that cannot take it
	// keeps the image it has.
	_ = t.icon.SetIcon(img, false)
}

// SetItems replaces the app's items between Open and Quit.
func (t *Tray) SetItems(items []TrayItem) {
	t.opts.Items = append([]TrayItem(nil), items...)
	if t.icon != nil {
		t.icon.SetMenu(t.menu())
	}
}

// CloseToTray decides a window's close in its OnClose listener. enabled
// is the app's own close-to-tray setting: an opt-in, off by default, so a
// close quits until the user turns it on. When enabled, the window is the
// tray's, no quit runs, and a panel hosts the icon (TrayHosted), it
// prevents the close, hides the window and reports true. Else it leaves the close alone and reports false, so the window
// closes as it did without a tray. The app keeps its own OnClose and
// calls it there:
//
//	win.OnClose(func(e *mygo.CloseEvent) {
//		if tray.CloseToTray(e, settings.CloseToTray) {
//			return
//		}
//		// the close goes on
//	})
//
// A quit (the menu's Quit, App.Quit, SIGTERM on Linux) is never held: the
// tray hears it through App.OnBeforeQuit. Show the hidden window again in
// App.OnSecondInstance; Open and the icon's activation show it too.
func (t *Tray) CloseToTray(e *mygo.CloseEvent, enabled bool) bool {
	if e == nil || !enabled || t.icon == nil || t.quitting {
		return false
	}
	if e.Window != t.opts.Window || !t.hosted() {
		return false
	}
	e.PreventDefault()
	t.hide(e.Window)
	return true
}

// Destroy removes the icon and stops listening for the quit. CloseToTray
// reports false after it. Later calls do nothing.
func (t *Tray) Destroy() {
	if t.icon == nil {
		return
	}
	for _, off := range t.offs {
		off()
	}
	t.offs = nil
	t.icon.Destroy()
	t.icon = nil
}

// open shows and focuses the tray's window.
func (t *Tray) open() { t.show(t.opts.Window) }

// toolTip is the tooltip with the status line after it.
func (t *Tray) toolTip() string {
	switch {
	case t.status == "":
		return t.opts.ToolTip
	case t.opts.ToolTip == "":
		return t.status
	}
	return t.opts.ToolTip + " — " + t.status
}

// menu builds the menu: the status line and a separator when there is a
// status, Open, the app's items, a separator and Quit.
func (t *Tray) menu() *mygo.Menu {
	var items []*mygo.MenuItem
	if t.status != "" {
		items = append(items, &mygo.MenuItem{ID: trayStatusID, Label: t.status, Disabled: true}, mygo.Separator())
	}
	open := t.opts.OpenLabel
	if open == "" {
		open = "Open"
	}
	items = append(items, &mygo.MenuItem{ID: trayOpenID, Label: open, Click: func(*mygo.MenuItem, *mygo.Window) { t.open() }})
	for _, it := range t.opts.Items {
		mi := &mygo.MenuItem{ID: it.ID, Label: it.Label, Disabled: it.Disabled}
		if click := it.Click; click != nil {
			mi.Click = func(*mygo.MenuItem, *mygo.Window) { click() }
		}
		items = append(items, mi)
	}
	items = append(items, mygo.Separator(), &mygo.MenuItem{ID: trayQuitID, Role: mygo.RoleQuit})
	return mygo.NewMenu(items)
}

// trayDotted returns icon, a PNG, with the dot drawn on its top right
// corner: a disc of trayDotColor, about 2 fifths of the icon wide, cut
// out of the icon by a clear ring, so that it reads on any icon.
func trayDotted(icon []byte) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(icon))
	if err != nil {
		return nil, fmt.Errorf("frostui: tray: the icon is not a PNG: %w", err)
	}
	b := src.Bounds()
	img := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), src, b.Min, draw.Src)
	w, h := float64(b.Dx()), float64(b.Dy())
	r := math.Max(1.5, math.Min(w, h)/5)
	ring := math.Max(1, r/4)
	cx, cy := w-r, r
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			cut := clamp01(r + ring + 0.5 - d)
			if cut == 0 {
				continue
			}
			dot := clamp01(r + 0.5 - d)
			img.SetNRGBA(x, y, trayDotPixel(img.NRGBAAt(x, y), cut, dot))
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("frostui: tray: encode the icon: %w", err)
	}
	return out.Bytes(), nil
}

// trayDotPixel clears cut of a pixel of the icon and lays dot of the
// dot's color over it, both as fractions of the pixel.
func trayDotPixel(p color.NRGBA, cut, dot float64) color.NRGBA {
	a := float64(p.A) / 255 * (1 - cut)
	outA := dot + a*(1-dot)
	if outA == 0 {
		return color.NRGBA{}
	}
	mix := func(dc, pc uint8) uint8 {
		return uint8(math.Round((float64(dc)*dot + float64(pc)*a*(1-dot)) / outA))
	}
	return color.NRGBA{
		R: mix(trayDotColor.R, p.R), G: mix(trayDotColor.G, p.G), B: mix(trayDotColor.B, p.B),
		A: uint8(math.Round(outA * 255)),
	}
}

func clamp01(v float64) float64 { return math.Min(1, math.Max(0, v)) }

// TrayHosted reports whether a panel can show a tray icon now. On Linux
// it asks the session bus whether org.kde.StatusNotifierWatcher has an
// owner: GNOME has none without the AppIndicator extension, and MyGo's
// tray then shows nothing and reports no error. It reports false when
// the bus cannot be asked or does not answer within half a second.
// Elsewhere it reports true. On Linux it waits for a D-Bus round trip,
// and for the connection to the bus on the first call, half a second at
// most: a query that takes longer goes on in the background, and the
// calls after it wait on that same query. Call it on the UI thread only
// where that wait is acceptable, as at a close.
func TrayHosted() bool { return trayHosted() }
