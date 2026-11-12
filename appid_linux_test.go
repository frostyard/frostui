package frostui

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

func TestCStringCopiesTheText(t *testing.T) {
	b := cString("abc")
	if b == nil || *b != 'a' {
		t.Fatal("cString does not copy the text")
	}
}

// glibPrgname gives the test GLib's g_get_prgname and g_set_prgname, or
// skips it on a host without GLib.
func glibPrgname(t *testing.T) (get func() string, set func(string)) {
	t.Helper()
	lib, err := purego.Dlopen("libglib-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("no GLib: %v", err)
	}
	var getPrgname func() *byte
	var setPrgname func(*byte)
	purego.RegisterLibFunc(&getPrgname, lib, "g_get_prgname")
	purego.RegisterLibFunc(&setPrgname, lib, "g_set_prgname")
	get = func() string {
		p := getPrgname()
		if p == nil {
			return ""
		}
		n := 0
		for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
			n++
		}
		return string(unsafe.Slice(p, n))
	}
	return get, func(s string) { setPrgname(cString(s)) }
}

// MyGo names the program after its executable in App.Run, after SetAppID
// (egoist/mygo#1640). The name SetAppID gave must come back once MyGo is
// ready, before the app opens a window.
func TestSetAppIDNamesTheProgramAgainWhenMyGoIsReady(t *testing.T) {
	get, set := glibPrgname(t)
	var ready []func()
	saved := whenReady
	whenReady = func(fn func()) { ready = append(ready, fn) }
	t.Cleanup(func() { whenReady = saved })

	const id = "org.example.AppIDTest"
	if err := setAppID(id, ""); err != nil {
		t.Fatalf("setAppID: %v", err)
	}
	if got := get(); got != id {
		t.Fatalf("program name after setAppID = %q, want %q", got, id)
	}
	if len(ready) != 1 {
		t.Fatalf("setAppID registered %d callbacks for when MyGo is ready, want 1", len(ready))
	}

	set("frostui.test") // what MyGo's App.Run does
	ready[0]()
	if got := get(); got != id {
		t.Errorf("program name once MyGo is ready = %q, want %q", got, id)
	}
}

func TestSetAppIDRejectsAnEmptyIDAndWaitsForNothing(t *testing.T) {
	glibPrgname(t)
	var ready []func()
	saved := whenReady
	whenReady = func(fn func()) { ready = append(ready, fn) }
	t.Cleanup(func() { whenReady = saved })

	if err := setAppID("", ""); err == nil {
		t.Fatal("setAppID accepted an empty ID")
	}
	if len(ready) != 0 {
		t.Errorf("setAppID with an empty ID registered %d callbacks, want 0", len(ready))
	}
}
