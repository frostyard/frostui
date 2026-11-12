package frostui

import "sync"

var setAppIDOnce sync.Once

// SetAppID names the program for the desktop before MyGo opens the
// display. Call it before mygo.App.Run; only the first call counts, as
// GLib allows, on every system.
//
// On Linux, GTK takes the Wayland app ID and the X11 WM_CLASS from GLib's
// program name, which MyGo sets to the executable's name when App.Run
// starts. SetAppID sets it (g_set_prgname) to id, the reverse-DNS ID of
// the app's .desktop file, so the desktop shows the app's icon and name
// for its windows, and sets GLib's application name
// (g_set_application_name) to displayName, the human name, when it is not
// empty. It sets the program name again from mygo.App.WhenReady, after
// MyGo has named it, so call it before the WhenReady that opens the app's
// first window. It loads GLib with purego, so a host without GLib gets an
// error and the program runs on unnamed.
//
// On other systems MyGo does not use GTK and the system names the app
// from its bundle (macOS), its executable (Windows) or its page (a
// browser), so SetAppID only checks that id is not empty.
func SetAppID(id, displayName string) error {
	var err error
	setAppIDOnce.Do(func() { err = setAppID(id, displayName) })
	return err
}
