package workspaceui

import (
	"encoding/json"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
)

// windowKey is where KeepWindow keeps the window in the app's storage.
const windowKey = "window"

// keptWindow is the window as the device keeps it.
type keptWindow struct {
	Width, Height core.Points
	X, Y          core.Points
	// Placed says X and Y are a place the platform reported.
	Placed bool
}

// KeepWindow shows the app's window as the device kept it, and keeps it
// as it changes (the standard for an app, rule 8): its size, and its place
// where the platform reports one. The size kept is the window's own where
// the platform reports it, and the content's elsewhere. It is kept in the app's storage under
// "window". A window under minWidth by minHeight is not kept, nor a
// maximized one's size, which leaves the size it had before. A browser's
// window is the browser's: there KeepWindow does nothing. Call it once,
// from the root.
func KeepWindow(cx *core.Scope, minWidth, minHeight core.Points) {
	if !desktop.IsAvailable(cx) {
		return
	}
	if text, ok := cx.Storage().GetItem(windowKey); ok {
		var k keptWindow
		if json.Unmarshal([]byte(text), &k) == nil && k.Width >= minWidth && k.Height >= minHeight {
			desktop.SetWindowFrame(cx, core.WindowFrame{Width: k.Width, Height: k.Height, X: k.X, Y: k.Y, Move: k.Placed})
		}
	}
	var pending func() bool
	first := true
	cx.Effect(func() {
		size, frame, large := cx.Env().Window(), desktop.Frame(cx), desktop.Maximized(cx)
		if first {
			// The size the window opens at is not a change of it.
			first = false
			return
		}
		if large || size.Width < minWidth || size.Height < minHeight {
			return
		}
		k := keptWindow{Width: size.Width, Height: size.Height, X: frame.X, Y: frame.Y, Placed: frame.Move}
		// Where the platform says how large the window is, that is kept:
		// SetWindowFrame takes the window's size, and a window with a
		// frame of its own is larger than its content. Keeping the
		// content's size there made each start's window smaller.
		if frame.Width > 0 && frame.Height > 0 {
			k.Width, k.Height = frame.Width, frame.Height
		}
		if pending != nil {
			pending()
		}
		// A drag of the window's edge reports many sizes: keep the last.
		pending = cx.AfterFunc(300*time.Millisecond, func() {
			if text, err := json.Marshal(k); err == nil {
				_ = cx.Storage().SetItem(windowKey, string(text))
			}
		})
	})
}
