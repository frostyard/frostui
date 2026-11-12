package frostui

import (
	"errors"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"
)

// whenReady runs fn once MyGo has opened the display. It is MyGo's
// App.WhenReady; tests replace it.
var whenReady = mygo.App.WhenReady

// setAppID is SetAppID on Linux, where MyGo draws on GTK 3: it names the
// program through GLib, then names it again once MyGo is ready.
func setAppID(id, displayName string) error {
	var lib uintptr
	var err error
	for _, name := range []string{"libglib-2.0.so.0", "libglib-2.0.so"} {
		if lib, err = purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL); err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	var setPrgname, setApplicationName func(*byte)
	purego.RegisterLibFunc(&setPrgname, lib, "g_set_prgname")
	purego.RegisterLibFunc(&setApplicationName, lib, "g_set_application_name")
	if id == "" {
		return errors.New("frostui: empty app ID")
	}
	setPrgname(cString(id))
	if displayName != "" {
		setApplicationName(cString(displayName))
	}
	// MyGo since v0.9.0 names the program after its executable in App.Run,
	// before it opens the display (egoist/mygo#1640), and GLib keeps the
	// last name it is given, so the name set above is lost and GTK names
	// the windows after the executable. WhenReady runs after that, and
	// before the callbacks an app registers after this call, which open its
	// windows: GTK reads the name when it creates a window (the Wayland app
	// ID, the instance part of X11's WM_CLASS), so naming the program again
	// here gives the windows the app's ID.
	whenReady(func() { setPrgname(cString(id)) })
	return nil
}

// cString returns a NUL-terminated copy of s. GLib copies the strings it
// is given, so the Go memory may go after the call.
func cString(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}
