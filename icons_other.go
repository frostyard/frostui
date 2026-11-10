//go:build !linux

package frostui

// iconSearch is where NewIcons looks off Linux: nowhere. Windows, macOS and
// a browser have no freedesktop icon theme, so only the app's embedded
// icons are found, and a missing one draws as an empty box.
func iconSearch() (theme string, roots, themes []string) { return "", nil, nil }

func desktopIconTheme() string { return "" }
