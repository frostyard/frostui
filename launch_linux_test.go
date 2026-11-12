package frostui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchDesktopApp(t *testing.T) {
	if os.Getenv("FROSTUI_LAUNCH_CHILD") == "1" {
		if err := LaunchDesktopApp("org.example.Hidden"); err != nil {
			t.Fatal(err)
		}
		if err := LaunchDesktopApp("org.example.Missing"); err == nil {
			t.Fatal("missing app succeeded")
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
	script := filepath.Join(dir, "fixture")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf 'launched:%s\\n' \"$1\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	entry := "[Desktop Entry]\nType=Application\nName=Hidden\nExec=" + script + " expected\nNoDisplay=true\n"
	if err := os.WriteFile(filepath.Join(apps, "org.example.Hidden.desktop"), []byte(entry), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestLaunchDesktopApp$")
	cmd.Env = append(os.Environ(), "FROSTUI_LAUNCH_CHILD=1", "XDG_DATA_HOME="+dir, "XDG_DATA_DIRS="+dir, "PATH="+dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "launched:expected\n") {
		t.Fatalf("application was not launched with its argument: %s", out)
	}
}
