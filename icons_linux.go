package frostui

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// iconSearch is where NewIcons looks on Linux: the desktop's theme, then
// Adwaita, Breeze and hicolor, under the user's and XDG_DATA_DIRS' icon
// directories and /usr/share/icons.
func iconSearch() (theme string, roots, themes []string) {
	home, _ := os.UserHomeDir()
	theme = DesktopIconTheme()
	roots = []string{filepath.Join(home, ".local/share/icons"), filepath.Join(home, ".icons")}
	for _, dir := range filepath.SplitList(os.Getenv("XDG_DATA_DIRS")) {
		if dir != "" {
			roots = append(roots, filepath.Join(dir, "icons"))
		}
	}
	roots = append(roots, "/usr/share/icons")
	themes = []string{theme}
	for _, fallback := range []string{"Adwaita", "breeze", "hicolor"} {
		if !strings.EqualFold(fallback, theme) {
			themes = append(themes, fallback)
		}
	}
	return theme, dedupe(roots), themes
}

func desktopIconTheme() string {
	if env := os.Getenv("FROSTUI_ICON_THEME"); env != "" {
		return env
	}
	if strings.Contains(strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP")), "KDE") {
		home, _ := os.UserHomeDir()
		if name := kdeIconTheme(filepath.Join(home, ".config", "kdeglobals")); name != "" {
			return name
		}
		return "breeze"
	}
	if out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "icon-theme").Output(); err == nil {
		if name := strings.Trim(strings.TrimSpace(string(out)), "'"); name != "" {
			return name
		}
	}
	return "Adwaita"
}

func kdeIconTheme(path string) string {
	f, err := os.Open(path) //nolint:gosec // the user's own kdeglobals
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	section := ""
	for s := bufio.NewScanner(f); s.Scan(); {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "[") {
			section = line
		} else if section == "[Icons]" && strings.HasPrefix(line, "Theme=") {
			return strings.TrimPrefix(line, "Theme=")
		}
	}
	return ""
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
