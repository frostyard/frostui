package frostui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"
)

// appIDRunEnv makes the test binary a helper that runs MyGo's App.Run
// (runAppIDHelper) instead of the tests. Its value is the app ID the
// helper gives SetAppID, or "-" for none.
const appIDRunEnv = "FROSTUI_TEST_APPID_RUN"

// appIDRunName is the display name the helper gives SetAppID.
const appIDRunName = "App ID Run"

func TestMain(m *testing.M) {
	if id, ok := os.LookupEnv(appIDRunEnv); ok {
		os.Exit(runAppIDHelper(id))
	}
	os.Exit(m.Run())
}

// runAppIDHelper names the program as an app does (SetAppID, unless id is
// "-"), runs MyGo's App.Run on the main goroutine, as TestMain does, and
// prints what GLib then has as the program and application names. With
// no display, Run names the program and fails at gtk_init_check, before
// any window or event loop. It prints the error of Run on its first line.
func runAppIDHelper(id string) int {
	if id != "-" {
		if err := SetAppID(id, appIDRunName); err != nil {
			fmt.Printf("setappid: %v\n", err)
			return 1
		}
	}
	err := mygo.App.Run()
	if err == nil {
		err = errors.New("<nil>")
	}
	get, getApplicationName, ok := glibNames()
	if !ok {
		fmt.Println("run: no GLib")
		return 1
	}
	fmt.Printf("run: %v\nprgname: %s\napplication: %s\n", err, get(), getApplicationName())
	return 0
}

func TestCStringCopiesTheText(t *testing.T) {
	b := cString("abc")
	if b == nil || *b != 'a' {
		t.Fatal("cString does not copy the text")
	}
}

// glibNames gives GLib's g_get_prgname and g_get_application_name, and
// false on a host without GLib.
func glibNames() (prgname, applicationName func() string, ok bool) {
	lib, err := purego.Dlopen("libglib-2.0.so.0", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return nil, nil, false
	}
	var getPrgname, getApplicationName func() *byte
	purego.RegisterLibFunc(&getPrgname, lib, "g_get_prgname")
	purego.RegisterLibFunc(&getApplicationName, lib, "g_get_application_name")
	return func() string { return goString(getPrgname()) },
		func() string { return goString(getApplicationName()) }, true
}

// goString copies the NUL-terminated C string at p, "" for nil.
func goString(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}

func TestSetAppIDNamesTheProgramInGLib(t *testing.T) {
	get, _, ok := glibNames()
	if !ok {
		t.Skip("no GLib")
	}
	const id = "org.example.AppIDTest"
	if err := setAppID(id, ""); err != nil {
		t.Fatalf("setAppID: %v", err)
	}
	if got := get(); got != id {
		t.Fatalf("program name after setAppID = %q, want %q", got, id)
	}
}

func TestSetAppIDRejectsAnEmptyID(t *testing.T) {
	if _, _, ok := glibNames(); !ok {
		t.Skip("no GLib")
	}
	if err := setAppID("", ""); err == nil {
		t.Fatal("setAppID accepted an empty ID")
	}
}

// runAppIDHelperProcess runs the test binary as the helper of
// runAppIDHelper, with no display, and returns what it printed by name
// (run, prgname, application). It skips the test when MyGo cannot load
// GTK 3, as Run then fails before it names the program.
func runAppIDHelperProcess(t *testing.T, id string) map[string]string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe)
	// No display: MyGo names the program, then gtk_init_check fails, and
	// Run returns before it opens a window or runs the event loop.
	env := []string{appIDRunEnv + "=" + id}
	for _, kv := range os.Environ() {
		switch strings.SplitN(kv, "=", 2)[0] {
		case "DISPLAY", "WAYLAND_DISPLAY", appIDRunEnv:
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper: %v\n%s", err, out)
	}
	got := map[string]string{}
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if k, v, ok := strings.Cut(line, ": "); ok {
			got[k] = v
		}
	}
	if run := got["run"]; !strings.Contains(run, "cannot open display") {
		t.Skipf("MyGo's App.Run did not reach gtk_init_check (no GTK 3?): %s", run)
	}
	return got
}

// MyGo's App.Run names the program after its executable when the app has
// not named it: the helper proves that Run reached that step, so the test
// below does not pass for a Run that named nothing.
func TestMyGoRunNamesAnUnnamedProgramAfterItsExecutable(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	got := runAppIDHelperProcess(t, "-")
	if want := filepath.Base(exe); got["prgname"] != want {
		t.Errorf("program name after App.Run = %q, want the executable's %q", got["prgname"], want)
	}
}

// The name SetAppID gives before App.Run must hold through Run, which
// names the program only while it has no name (egoist/mygo#1761; MyGo
// v0.9.0 named it after the executable, #1640): GTK gives the windows
// that name as their Wayland app ID and X11 WM_CLASS instance.
func TestSetAppIDNameHoldsThroughMyGoRun(t *testing.T) {
	const id = "org.example.AppIDRun"
	got := runAppIDHelperProcess(t, id)
	if got["prgname"] != id {
		t.Errorf("program name after App.Run = %q, want %q", got["prgname"], id)
	}
	if got["application"] != appIDRunName {
		t.Errorf("application name after App.Run = %q, want %q", got["application"], appIDRunName)
	}
}
