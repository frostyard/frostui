//go:build !linux

package frostui

import "testing"

func TestTrayIsHostedOffLinux(t *testing.T) {
	if !TrayHosted() {
		t.Error("TrayHosted = false off Linux, want true")
	}
}
