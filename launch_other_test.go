//go:build !linux

package frostui

import (
	"errors"
	"testing"
)

func TestLaunchDesktopAppIsUnsupportedOffLinux(t *testing.T) {
	if err := LaunchDesktopApp("org.example.App"); !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("LaunchDesktopApp = %v, want an error that wraps errors.ErrUnsupported", err)
	}
}
