package frostui

import (
	_ "embed"
	"log/slog"
	"sync"

	"github.com/egoist/mygo/ui"
)

// FontFamily is the family the kit's themes set their text in: Inter,
// embedded in the kit (fonts/, under the SIL Open Font License 1.1 in
// fonts/OFL.txt) and registered under this name on the first Adapt.
const FontFamily = "Inter"

var (
	//go:embed fonts/InterVariable.ttf
	interUpright []byte
	//go:embed fonts/InterVariable-Italic.ttf
	interItalic []byte

	registerFonts sync.Once
	// interErr is why the text system did not take the kit's font, nil
	// when it did.
	interErr error
)

// registerInter adds the embedded Inter to the text system once. It runs
// from Adapt, on the UI thread, never at package init: a program that
// never opens a window must not load a text engine. A text system that
// cannot add fonts draws the kit's text in its default family.
func registerInter() {
	registerFonts.Do(func() {
		for _, data := range [][]byte{interUpright, interItalic} {
			if err := ui.RegisterFont(data, FontFamily); err != nil {
				interErr = err
				slog.Warn("frostui: the kit's font is not available; text falls back to the system's", "family", FontFamily, "err", err)
				return
			}
		}
	})
}
