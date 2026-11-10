//go:build !linux

package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestIconsOffLinuxAreTheEmbeddedOnes(t *testing.T) {
	t.Setenv("FROSTUI_ICON_THEME", "Adwaita")
	if got := DesktopIconTheme(); got != "" {
		t.Errorf("DesktopIconTheme = %q, want none off Linux", got)
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"><circle cx="8" cy="8" r="8"/></svg>`)
	ic := NewIcons(map[string][]byte{"go-next-symbolic": svg})
	if ic.Theme != "" || ic.Get("go-next") == nil || ic.Get("no-such-icon") != nil {
		t.Errorf("theme %q: want the embedded go-next and no theme", ic.Theme)
	}
	if ic.Get("edit-copy") == nil || ic.Get("edit-copy-symbolic") == nil {
		t.Error("edit-copy, which the app does not embed, does not come from the kit's set")
	}
	SetIcons(ic)
	t.Cleanup(func() { SetIcons(nil) })
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
