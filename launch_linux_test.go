package frostui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLaunchDesktopAppStartsGtkLaunch(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "args")
	// Only shell builtins: PATH holds nothing but this script.
	script := "#!/bin/sh\necho \"$@\" > " + out + "\n"
	if err := os.WriteFile(filepath.Join(dir, "gtk-launch"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if err := LaunchDesktopApp("org.example.App"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		data, err := os.ReadFile(out)
		if err == nil && strings.HasSuffix(string(data), "\n") {
			if got := strings.TrimSpace(string(data)); got != "org.example.App" {
				t.Errorf("gtk-launch got %q", got)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("gtk-launch did not run")
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Setenv("PATH", t.TempDir())
	if err := LaunchDesktopApp("org.example.App"); err == nil || !errors.Is(err, os.ErrNotExist) && !strings.Contains(err.Error(), "not found") {
		t.Errorf("without gtk-launch: %v", err)
	}
}
