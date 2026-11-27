package main

import (
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
)

// A tour runs once in each theme, as a test of its own: on a target each
// test is a process, and on the web target only a process's first app
// writes its pictures.

// TestMain starts NSGo's test host, so the tours also run in a real window
// (make gallery-pictures), where each Screenshot is a PNG.
func TestMain(m *testing.M) { nsgotest.Main(m) }

func start(t *testing.T, dark bool, width, height core.Points) *nsgotest.App {
	t.Helper()
	return nsgotest.Start(t, root, nsgotest.Size(width, height),
		nsgotest.Environment(core.Environment{OS: "linux", Idiom: core.IdiomDesktop, PixelRatio: 1, Dark: dark}))
}

func look(dark bool) string {
	if dark {
		return "dark"
	}
	return "light"
}

// The standard window with a message open, with none, and narrow.
func TestPicturesOfTheWindow(t *testing.T)       { picturesOfTheWindow(t, false) }
func TestPicturesOfTheWindowInDark(t *testing.T) { picturesOfTheWindow(t, true) }

func picturesOfTheWindow(t *testing.T, dark bool) {
	a := start(t, dark, 1280, 820)
	a.WaitFor("label:Your trip to Lisbon: itinerary", 5*time.Second)
	a.Screenshot("window-" + look(dark))
	// The filter is NSGo's segmented tabs, and the toolbar its toolbar:
	// a control of another kind fails here (rule 11).
	a.Find("role:tab Unread")
	a.Find("role:toolbar")
}

func TestPicturesOfANarrowWindow(t *testing.T)       { picturesOfANarrowWindow(t, false) }
func TestPicturesOfANarrowWindowInDark(t *testing.T) { picturesOfANarrowWindow(t, true) }

func picturesOfANarrowWindow(t *testing.T, dark bool) {
	a := start(t, dark, 620, 820)
	a.WaitFor("label:Your trip to Lisbon: itinerary", 5*time.Second)
	a.Screenshot("window-narrow-" + look(dark))
}

// Each page of controls, in both themes.
func TestPicturesOfThePages(t *testing.T)       { picturesOfThePages(t, false) }
func TestPicturesOfThePagesInDark(t *testing.T) { picturesOfThePages(t, true) }

func picturesOfThePages(t *testing.T, dark bool) {
	a := start(t, dark, 1280, 1100)
	a.Press("text:Controls and colors")
	for _, p := range pages {
		a.Press("sidebar-panel >> text:" + p.title)
		a.Settle()
		a.Screenshot("page-" + p.id + "-" + look(dark))
	}
}

// The open states: a confirmation, a form in a dialog, and a menu.
func TestPicturesOfTheOverlays(t *testing.T)       { picturesOfTheOverlays(t, false) }
func TestPicturesOfTheOverlaysInDark(t *testing.T) { picturesOfTheOverlays(t, true) }

func picturesOfTheOverlays(t *testing.T, dark bool) {
	a := start(t, dark, 1280, 820)
	a.Press("text:Controls and colors")
	a.Press("sidebar-panel >> text:Dialogs and menus")
	a.Press("text:Delete message…")
	a.WaitFor("role:alertdialog", 5*time.Second)
	a.Screenshot("overlay-confirm-" + look(dark))
	a.Press("role:button Cancel")
	a.Press("text:New mailbox…")
	a.WaitFor("label:Name", 5*time.Second)
	a.Screenshot("overlay-form-" + look(dark))
	a.Press("role:button Cancel")
	a.Press("role:button More")
	a.WaitFor("text:Show headers", 5*time.Second)
	a.Screenshot("overlay-menu-" + look(dark))
}

// More open states: a toast with its way back, a popover, and the palette.
func TestPicturesOfMoreOverlays(t *testing.T)       { picturesOfMoreOverlays(t, false) }
func TestPicturesOfMoreOverlaysInDark(t *testing.T) { picturesOfMoreOverlays(t, true) }

func picturesOfMoreOverlays(t *testing.T, dark bool) {
	a := start(t, dark, 1280, 820)
	a.Press("text:Controls and colors")
	a.Press("sidebar-panel >> text:Search, trees and popovers")
	a.Press("role:button Archive a message")
	a.WaitFor("text:Message archived", 5*time.Second)
	a.Screenshot("overlay-toast-" + look(dark))
	a.Press("role:button Remind me")
	a.WaitFor("text:Tomorrow at 8:00 AM", 5*time.Second)
	a.Screenshot("overlay-popover-" + look(dark))
	a.Key("Escape")
	a.Press("role:button Open the palette")
	a.WaitFor("text:Go to Calendar", 5*time.Second)
	a.Screenshot("overlay-palette-" + look(dark))
}
