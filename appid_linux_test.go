package frostui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
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

// appIDHelperEnv is the helper's environment: environ without what could
// let GTK open a display or make App.Run stop early, with GDK held to X11
// with no DISPLAY, so that gtk_init_check fails also where Wayland would
// find $XDG_RUNTIME_DIR/wayland-0 with no WAYLAND_DISPLAY. Every MYGO_
// variable goes too: MYGO_GENERATE makes Run return before it names the
// program.
func appIDHelperEnv(environ []string, id string) []string {
	env := []string{appIDRunEnv + "=" + id, "GDK_BACKEND=x11"}
	for _, kv := range environ {
		k, _, _ := strings.Cut(kv, "=")
		switch {
		case k == "DISPLAY", k == "WAYLAND_DISPLAY", k == "WAYLAND_SOCKET", k == "GDK_BACKEND", k == appIDRunEnv,
			strings.HasPrefix(k, "MYGO_"):
			continue
		}
		env = append(env, kv)
	}
	return env
}

// errNoGTK is why parseAppIDHelper skips: MyGo could not load GTK 3 or
// another library it opens, so Run failed before it named the program.
var errNoGTK = errors.New("MyGo's App.Run could not load its libraries (no GTK 3?)")

// parseAppIDHelper reads what runAppIDHelper printed by name (run,
// prgname, application). Run must have failed at gtk_init_check, after it
// named the program; a library it could not load gives errNoGTK, and any
// other result is an error, so that the name checks never pass unrun.
func parseAppIDHelper(out string) (map[string]string, error) {
	got := map[string]string{}
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		if k, v, ok := strings.Cut(line, ": "); ok {
			got[k] = v
		}
	}
	run, ok := got["run"]
	switch {
	case !ok:
		return nil, fmt.Errorf("the helper printed no result of App.Run:\n%s", out)
	case strings.Contains(run, "cannot open shared object file"):
		return nil, fmt.Errorf("%w: %s", errNoGTK, run)
	case !strings.Contains(run, "cannot open display"):
		return nil, fmt.Errorf("App.Run = %s, want the error of gtk_init_check with no display", run)
	}
	for _, k := range []string{"prgname", "application"} {
		if _, ok := got[k]; !ok {
			return nil, fmt.Errorf("the helper printed no %s:\n%s", k, out)
		}
	}
	return got, nil
}

// runAppIDHelperProcess runs the test binary as the helper of
// runAppIDHelper, with no display (appIDHelperEnv), and returns what it
// printed (parseAppIDHelper). It skips the test only when MyGo cannot
// load GTK 3, and not in CI (CI=true), which installs it. A deadline
// ends a helper that hangs, as one that opened a display would in MyGo's
// event loop.
func runAppIDHelperProcess(t *testing.T, id string) map[string]string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe)
	cmd.Env = appIDHelperEnv(os.Environ(), id)
	cmd.WaitDelay = time.Second
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper: %v\n%s", err, out)
	}
	got, err := parseAppIDHelper(string(out))
	if errors.Is(err, errNoGTK) && os.Getenv("CI") != "true" {
		// CI installs GTK 3 (.github/workflows/ci.yml), so there it fails.
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestAppIDHelperEnvHasNoDisplayAndNoMyGoSettings(t *testing.T) {
	environ := []string{
		"PATH=/usr/bin", "DISPLAY=:0", "WAYLAND_DISPLAY=wayland-0", "WAYLAND_SOCKET=3", "GDK_BACKEND=wayland",
		"MYGO_GENERATE=/tmp/client.ts", "MYGO_E2E=1", appIDRunEnv + "=org.example.Old", "XDG_RUNTIME_DIR=/run/user/1000",
	}
	env := appIDHelperEnv(environ, "org.example.New")
	want := []string{appIDRunEnv + "=org.example.New", "GDK_BACKEND=x11", "PATH=/usr/bin", "XDG_RUNTIME_DIR=/run/user/1000"}
	if !slices.Equal(env, want) {
		t.Errorf("appIDHelperEnv = %q, want %q", env, want)
	}
}

func TestParseAppIDHelperFailsUnlessRunReachedTheDisplay(t *testing.T) {
	const names = "prgname: org.example.X\napplication: X\n"
	for _, tc := range []struct {
		name, out string
		noGTK     bool
		ok        bool
	}{
		{name: "display", out: "run: mygo: cannot open display (is DISPLAY or WAYLAND_DISPLAY set?)\n" + names, ok: true},
		{name: "no GTK", out: "run: libgtk-3.so.0: cannot open shared object file: No such file or directory\n" + names, noGTK: true},
		{name: "no result", out: names},
		{name: "empty", out: ""},
		{name: "success", out: "run: <nil>\n" + names},
		{name: "not main", out: "run: mygo: App.Run must be called from the main goroutine\n" + names},
		{name: "no names", out: "run: mygo: cannot open display\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAppIDHelper(tc.out)
			switch {
			case tc.ok && (err != nil || got["prgname"] != "org.example.X"):
				t.Errorf("parseAppIDHelper = %v, %v, want the names", got, err)
			case tc.noGTK && !errors.Is(err, errNoGTK):
				t.Errorf("parseAppIDHelper error = %v, want errNoGTK", err)
			case !tc.ok && !tc.noGTK && (err == nil || errors.Is(err, errNoGTK)):
				t.Errorf("parseAppIDHelper error = %v, want a failure, not a skip", err)
			}
		})
	}
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
