package frostui

import (
	"errors"

	"github.com/ebitengine/purego"
)

// setAppID is SetAppID on Linux, where MyGo draws on GTK 3: it names the
// program through GLib.
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
	return nil
}

// cString returns a NUL-terminated copy of s. GLib copies the strings it
// is given, so the Go memory may go after the call.
func cString(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}
