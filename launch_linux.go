package frostui

import (
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

var desktopGIO struct {
	once    sync.Once
	err     error
	newInfo func(*byte) uintptr
	launch  func(uintptr, uintptr, uintptr, uint32, uintptr, uintptr, uintptr, uintptr, **gioError) bool
	unref   func(uintptr)
	freeErr func(*gioError)
}

// gioError is GError's domain, code and NUL-terminated message pointer.
type gioError struct {
	_       uint32 // domain
	_       int32  // code
	message *byte
}

func loadDesktopGIO() error {
	desktopGIO.once.Do(func() {
		lib, err := purego.Dlopen("libgio-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			desktopGIO.err = err
			return
		}
		for _, symbol := range []struct {
			name string
			into any
		}{
			{"g_desktop_app_info_new", &desktopGIO.newInfo},
			{"g_desktop_app_info_launch_uris_as_manager", &desktopGIO.launch},
			{"g_object_unref", &desktopGIO.unref},
			{"g_error_free", &desktopGIO.freeErr},
		} {
			address, lookupErr := purego.Dlsym(lib, symbol.name)
			if lookupErr != nil {
				desktopGIO.err = fmt.Errorf("GIO symbol %s: %w", symbol.name, lookupErr)
				return
			}
			purego.RegisterFunc(symbol.into, address)
		}
	})
	return desktopGIO.err
}

func launchDesktopApp(id string) error {
	if id == "" || strings.ContainsRune(id, '\x00') {
		return fmt.Errorf("frostui: invalid desktop application ID %q", id)
	}
	if err := loadDesktopGIO(); err != nil {
		return fmt.Errorf("frostui: load GIO: %w", err)
	}
	info := desktopGIO.newInfo(cString(strings.TrimSuffix(id, ".desktop") + ".desktop"))
	if info == 0 {
		return fmt.Errorf("frostui: desktop application %q not found or not launchable: %w", id, fs.ErrNotExist)
	}
	defer desktopGIO.unref(info)
	// GLib's spawn flags: search PATH and discard both output streams, as
	// the previous gtk-launch process did. An app must not hold our pipes open.
	const discardOutput = 1<<2 | 1<<3 | 1<<4
	var gerr *gioError
	if !desktopGIO.launch(info, 0, 0, discardOutput, 0, 0, 0, 0, &gerr) {
		if gerr == nil {
			return fmt.Errorf("frostui: launch desktop application %q failed", id)
		}
		var text []byte
		for p := gerr.message; p != nil && *p != 0; p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1)) {
			text = append(text, *p)
		}
		desktopGIO.freeErr(gerr)
		return fmt.Errorf("frostui: launch desktop application %q: %s", id, text)
	}
	return nil
}
