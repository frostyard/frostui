//go:build !linux

package frostui

// trayHosted reports true: macOS and Windows always show the icon, and
// on js/wasm NewTray fails before a close can ask.
func trayHosted() bool { return true }
