package frostui

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestLaunchDesktopApp(t *testing.T) {
	if os.Getenv("FROSTUI_LAUNCH_CHILD") == "1" {
		if err := LaunchDesktopApp("org.example.Unlisted"); err != nil {
			t.Fatal(err)
		}
		if err := LaunchDesktopApp("org.example.Unlisted.desktop"); err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{"org.example.Missing", "org.example.Deleted"} {
			if err := LaunchDesktopApp(id); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("%s: got %v, want not found", id, err)
			}
		}
		if err := LaunchDesktopApp("org.example.BadDirectory"); err == nil || !strings.Contains(err.Error(), "directory") {
			t.Fatalf("launch failure lost GIO detail: %v", err)
		}
		if err := LaunchDesktopApp("bad\x00id"); err == nil {
			t.Fatal("NUL in ID succeeded")
		}
		return
	}
	dir := t.TempDir()
	apps := filepath.Join(dir, "applications")
	if err := os.MkdirAll(apps, 0o755); err != nil {
		t.Fatal(err)
	}
	fifo := filepath.Join(dir, "launched")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "fixture")
	content := "#!/bin/sh\nprintf 'launched:%s\\n' \"$1\" > " + fifo + "\nprintf 'should-not-inherit-stdout\\n'\nprintf 'should-not-inherit-stderr\\n' >&2\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	entry := "[Desktop Entry]\nType=Application\nName=Unlisted\nExec=" + script + " expected\nNoDisplay=true\n"
	if err := os.WriteFile(filepath.Join(apps, "org.example.Unlisted.desktop"), []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, extra := range map[string]string{
		"Deleted":      "Hidden=true\n",
		"BadDirectory": "Path=/nonexistent/frostui-launch-test-directory\n",
	} {
		data := "[Desktop Entry]\nType=Application\nName=" + name + "\nExec=" + script + "\n" + extra
		if err := os.WriteFile(filepath.Join(apps, "org.example."+name+".desktop"), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := make(chan string, 1)
	go func() {
		var observed string
		for range 2 {
			f, err := os.Open(fifo)
			if err != nil {
				read <- err.Error()
				return
			}
			data, err := io.ReadAll(io.LimitReader(f, 100))
			_ = f.Close()
			if err != nil {
				read <- err.Error()
				return
			}
			observed += string(data)
		}
		read <- observed
	}()
	cmd := exec.Command(os.Args[0], "-test.run=^TestLaunchDesktopApp$")
	for _, pair := range os.Environ() {
		if !strings.HasPrefix(pair, "DISPLAY=") && !strings.HasPrefix(pair, "WAYLAND_DISPLAY=") {
			cmd.Env = append(cmd.Env, pair)
		}
	}
	cmd.Env = append(cmd.Env, "FROSTUI_LAUNCH_CHILD=1", "XDG_DATA_HOME="+dir, "XDG_DATA_DIRS="+dir, "XDG_RUNTIME_DIR="+dir, "DBUS_SESSION_BUS_ADDRESS=unix:path="+filepath.Join(dir, "no-bus"), "PATH="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child: %v\n%s", err, out)
	}
	if strings.Contains(string(out), "should-not-inherit") {
		t.Fatalf("launched app inherited the test's output streams: %s", out)
	}
	select {
	case got := <-read:
		if got != "launched:expected\nlaunched:expected\n" {
			t.Fatalf("application launch wrote %q to FIFO", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("desktop application did not signal launch")
	}
}
