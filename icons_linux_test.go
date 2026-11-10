package frostui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// writeIcon writes an SVG icon at root/rel.
func writeIcon(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><rect width="16" height="16"/></svg>`
	if err := os.WriteFile(path, []byte(svg), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestThemeIconsFindFallbackAndEmbedded(t *testing.T) {
	data := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_DIRS", data+":"+data) // the second copy is deduped
	t.Setenv("FROSTUI_ICON_THEME", "Frost")
	writeIcon(t, data, "icons/Frost/symbolic/actions/edit-copy-symbolic.svg") // the theme's own
	writeIcon(t, data, "icons/breeze/actions/16/go-next.svg")                 // a Breeze layout
	writeIcon(t, data, "icons/hicolor/scalable/apps/app.svg")                 // the last fallback

	ic := NewIcons(map[string][]byte{"brand-symbolic": []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><circle cx="8" cy="8" r="8"/></svg>`)})
	if ic.Theme != "Frost" {
		t.Errorf("Theme = %q, want the FROSTUI_ICON_THEME override", ic.Theme)
	}
	for _, name := range []string{"edit-copy", "edit-copy-symbolic", "go-next", "app", "brand", "brand-symbolic"} {
		if ic.Get(name) == nil {
			t.Errorf("Get(%q) = nil", name)
		}
	}
	if ic.Get("no-such-icon") != nil || ic.Get("") != nil {
		t.Error("a missing icon is not nil")
	}
	if _, cached := ic.cache["no-such-icon"]; !cached {
		t.Error("a missing icon is not cached")
	}
	var none *Icons
	if none.Get("edit-copy") != nil {
		t.Error("a nil Icons returned an icon")
	}

	SetIcons(ic)
	t.Cleanup(func() { SetIcons(nil) })
	if IconSVG("go-next") == nil {
		t.Error("IconSVG does not read the kit's icons")
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Children(func() {
			Icon(c, "go-next").Label("next")
			Icon(c, "no-such-icon").Label("placeholder")
		})
	}, 200, 100)
	icon, ok1 := tt.Find("next")
	hole, ok2 := tt.Find("placeholder")
	if !ok1 || !ok2 || icon.W != hole.W || icon.H != hole.H || icon.W == 0 {
		t.Errorf("icon %+v and missing icon %+v differ in size: layouts would move", icon, hole)
	}
}

func TestThemeIconsWinOverTheKitSet(t *testing.T) {
	data := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_DIRS", data)
	t.Setenv("FROSTUI_ICON_THEME", "Frost")
	writeIcon(t, data, "icons/Frost/symbolic/actions/edit-copy-symbolic.svg") // 16 wide

	ic := NewIcons(nil)
	// The search also reads /usr/share/icons, which may have other themes:
	// keep it to the temporary directory.
	ic.roots = []string{filepath.Join(data, "icons")}
	theme := ic.Get("edit-copy")
	if theme == nil {
		t.Fatal("the theme's edit-copy is missing")
	}
	if w, _ := theme.Size(); w != 16 {
		t.Errorf("edit-copy is %v wide: the kit's set won over the theme", w)
	}
	kit := ic.Get("edit-find")
	if kit == nil {
		t.Fatal("edit-find, which the theme lacks, does not come from the kit's set")
	}
	if w, _ := kit.Size(); w != 24 {
		t.Errorf("edit-find is %v wide, want the kit's 24", w)
	}
}

func TestDesktopIconThemeFromKDE(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("FROSTUI_ICON_THEME", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "KDE")
	if got := DesktopIconTheme(); got != "breeze" {
		t.Errorf("KDE without kdeglobals = %q, want breeze", got)
	}
	cfg := filepath.Join(home, ".config")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	globals := "[General]\nTheme=wrong\n[Icons]\nTheme=breeze-dark\n"
	if err := os.WriteFile(filepath.Join(cfg, "kdeglobals"), []byte(globals), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := DesktopIconTheme(); got != "breeze-dark" {
		t.Errorf("KDE theme = %q, want breeze-dark", got)
	}
	if got := strings.Join(dedupe([]string{"a", "b", "a", "c", "b"}), ","); got != "a,b,c" {
		t.Errorf("dedupe = %q", got)
	}
}
