//go:build !linux

package frostui

import "errors"

// setAppID is SetAppID off Linux: MyGo does not draw on GTK there, so no
// GLib program name needs setting. It loads no library.
func setAppID(id, _ string) error {
	if id == "" {
		return errors.New("frostui: empty app ID")
	}
	return nil
}
