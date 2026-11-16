package frostui

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// trayWatcherName is the D-Bus name of the StatusNotifierWatcher, which
// a panel owns while it can show AppIndicator icons.
const trayWatcherName = "org.kde.StatusNotifierWatcher"

// trayNameHasOwner asks the session bus whether a name has an owner. A
// test replaces it.
var trayNameHasOwner = gioNameHasOwner

func trayHosted() bool {
	owned, err := trayNameHasOwner(trayWatcherName)
	return err == nil && owned
}

var busGIO struct {
	once      sync.Once
	err       error
	busGet    func(int32, uintptr, **gioError) uintptr
	call      func(uintptr, *byte, *byte, *byte, *byte, uintptr, *byte, int32, int32, uintptr, **gioError) uintptr
	newString func(*byte) uintptr
	newTuple  func(*uintptr, uintptr) uintptr
	child     func(uintptr, uintptr) uintptr
	boolean   func(uintptr) int32
	unrefVar  func(uintptr)
	unref     func(uintptr)
	freeErr   func(*gioError)
}

func loadBusGIO() error {
	busGIO.once.Do(func() {
		lib, err := purego.Dlopen("libgio-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			busGIO.err = err
			return
		}
		for _, symbol := range []struct {
			name string
			into any
		}{
			{"g_bus_get_sync", &busGIO.busGet},
			{"g_dbus_connection_call_sync", &busGIO.call},
			{"g_variant_new_string", &busGIO.newString},
			{"g_variant_new_tuple", &busGIO.newTuple},
			{"g_variant_get_child_value", &busGIO.child},
			{"g_variant_get_boolean", &busGIO.boolean},
			{"g_variant_unref", &busGIO.unrefVar},
			{"g_object_unref", &busGIO.unref},
			{"g_error_free", &busGIO.freeErr},
		} {
			address, lookupErr := purego.Dlsym(lib, symbol.name)
			if lookupErr != nil {
				busGIO.err = fmt.Errorf("GIO symbol %s: %w", symbol.name, lookupErr)
				return
			}
			purego.RegisterFunc(symbol.into, address)
		}
	})
	return busGIO.err
}

// gioNameHasOwner calls org.freedesktop.DBus.NameHasOwner on the session
// bus through GIO, with a timeout of half a second.
func gioNameHasOwner(name string) (bool, error) {
	if name == "" || strings.ContainsRune(name, '\x00') {
		return false, fmt.Errorf("frostui: invalid D-Bus name %q", name)
	}
	if err := loadBusGIO(); err != nil {
		return false, fmt.Errorf("frostui: load GIO: %w", err)
	}
	const sessionBus = 2 // G_BUS_TYPE_SESSION
	var gerr *gioError
	conn := busGIO.busGet(sessionBus, 0, &gerr)
	if conn == 0 {
		return false, gioFailure("connect to the session bus", gerr)
	}
	defer busGIO.unref(conn)
	// The tuple takes the floating string, and the call the floating
	// tuple: nothing here to free.
	args := [1]uintptr{busGIO.newString(cString(name))}
	params := busGIO.newTuple(&args[0], 1)
	const timeoutMS = 500
	reply := busGIO.call(conn, cString("org.freedesktop.DBus"), cString("/org/freedesktop/DBus"),
		cString("org.freedesktop.DBus"), cString("NameHasOwner"), params, cString("(b)"), 0, timeoutMS, 0, &gerr)
	if reply == 0 {
		return false, gioFailure("NameHasOwner "+name, gerr)
	}
	defer busGIO.unrefVar(reply)
	value := busGIO.child(reply, 0)
	defer busGIO.unrefVar(value)
	return busGIO.boolean(value) != 0, nil
}

// gioFailure turns a GError into an error and frees it.
func gioFailure(what string, gerr *gioError) error {
	if gerr == nil {
		return fmt.Errorf("frostui: %s failed", what)
	}
	var text []byte
	for p := gerr.message; p != nil && *p != 0; p = (*byte)(unsafe.Add(unsafe.Pointer(p), 1)) {
		text = append(text, *p)
	}
	busGIO.freeErr(gerr)
	return fmt.Errorf("frostui: %s: %s", what, text)
}
