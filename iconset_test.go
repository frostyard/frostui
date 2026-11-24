package frostui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestKitIconSetParses reads every mapped icon as the kit would.
func TestKitIconSetParses(t *testing.T) {
	for name, file := range kitIconNames {
		data := kitIcon(name)
		if data == nil {
			t.Errorf("%s: no file icons/%s.svg", name, file)
			continue
		}
		if _, err := ui.ParseSVG(data); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		if strings.HasSuffix(name, "-symbolic") {
			t.Errorf("%s: map the name without -symbolic", name)
		}
	}
	if kitIcon("edit-copy-symbolic") == nil {
		t.Error("the -symbolic form of a mapped name finds nothing")
	}
	if kitIcon("no-such-icon") != nil {
		t.Error("an unmapped name finds an icon")
	}
}

func TestCPUEmbeddedFallback(t *testing.T) {
	// No app icons or theme search roots, even on a desktop that has a CPU icon.
	ic := &Icons{cache: map[string]*ui.SVG{}}
	for _, name := range []string{"cpu", "cpu-symbolic"} {
		t.Run(name, func(t *testing.T) {
			svg := ic.Get(name)
			if svg == nil {
				t.Fatalf("Get(%q) returned no embedded CPU glyph", name)
			}
			// Sample the rendered pin, package outline, and inner die, plus
			// gaps between the pins and inside the die. Use 4x to avoid edges.
			img := svg.Image(96, 96, nil)
			for _, p := range []struct {
				x, y int
				ink  bool
			}{
				{10, 48, true}, {16, 36, true}, {32, 48, true},
				{10, 36, false}, {48, 48, false},
			} {
				if got := img.RGBAAt(p.x, p.y).A > 127; got != p.ink {
					t.Errorf("Get(%q): CPU glyph ink at (%d, %d) = %t, want %t", name, p.x, p.y, got, p.ink)
				}
			}
		})
	}
}

// TestKitIconSetHoldsOnlyMappedIcons keeps the binary to what is named: an
// embedded file nothing maps is size for nothing.
func TestKitIconSetHoldsOnlyMappedIcons(t *testing.T) {
	entries, err := kitIconFiles.ReadDir("icons")
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, file := range kitIconNames {
		used[file+".svg"] = true
	}
	for _, e := range entries {
		if !used[e.Name()] {
			t.Errorf("icons/%s is embedded and nothing maps it", e.Name())
		}
	}
}

// iconCall finds the icon names the kit's source passes to its drawing
// functions, and to the Icon field of a struct.
var iconCall = regexp.MustCompile(`(?:\b(?:Icon|IconButton|IconTile|IconPill|LabeledButton|Empty|toggleButton)\(c[^,"]*, "([a-z][a-z0-9-]*)"|\bIcon: "([a-z][a-z0-9-]*)")`)

// TestKitDrawsOnlyMappedIcons fails when the kit starts to draw a name its
// own set lacks, which off Linux would be an empty box.
func TestKitDrawsOnlyMappedIcons(t *testing.T) {
	if !iconCall.MatchString(`toggleButton(c, "sidebar-show", "Sidebar", open)`) {
		t.Error("the pattern does not see toggleButton's icon")
	}
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(path) //nolint:gosec // the kit's own source
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range iconCall.FindAllStringSubmatch(string(data), -1) {
			name := m[1] + m[2]
			found++
			if _, ok := kitIconNames[name]; !ok {
				t.Errorf("%s draws %q, which the kit's icon set does not map", path, name)
			}
		}
	}
	if found < 15 {
		t.Fatalf("found %d icon names in the kit's source, too few: the pattern missed them", found)
	}
}

// TestKitIconSetKeepsItsLicense keeps the upstream notice with the icons.
func TestKitIconSetKeepsItsLicense(t *testing.T) {
	license, err := os.ReadFile(filepath.Join("icons", "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ISC License", "Lucide Icons and Contributors", "Cole Bemis"} {
		if !strings.Contains(string(license), want) {
			t.Errorf("icons/LICENSE lacks %q", want)
		}
	}
	notice, err := os.ReadFile("NOTICE")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notice), "lucide.dev") || !strings.Contains(string(notice), "icons/LICENSE") {
		t.Errorf("NOTICE does not attribute the icons: %s", notice)
	}
}
