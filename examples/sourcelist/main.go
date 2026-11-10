// Sourcelist opens a pretend notes app built from frostui's shared
// pieces: a source list of folders and notes (drag a note onto a folder,
// F2 renames, right-click for a menu), one command table behind the keys,
// the main menu, the palette and the shortcuts dialog (F1), a palette
// source that finds notes on a worker (Ctrl O), and a settings dialog
// (Ctrl ,). Ctrl K opens the palette, except in the note's editor, where
// it inserts a link; Ctrl+Shift+P opens it everywhere.
//
//	go run ./examples/sourcelist
package main

import (
	"log"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// appID names the example for the desktop (frostui.SetAppID).
const appID = "org.frostyard.FrostUI.SourceList"

func main() {
	if err := frostui.SetAppID(appID, "Source list"); err != nil {
		log.Printf("sourcelist: cannot set the app ID: %v", err)
	}
	mygo.App.SetName(appID)
	setIcons()

	p := &poster{}
	d := newDemo(p.post)
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title: "Source list", Width: 1000, Height: 680, MinWidth: 720, MinHeight: 480,
			Content: ui.View(d.view),
		})
		p.attach(win)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// setIcons gives the kit its icons: on Linux they come from the desktop's
// icon theme, and elsewhere they draw as empty boxes.
func setIcons() { frostui.SetIcons(frostui.NewIcons(nil)) }

// poster hands functions to the window's UI thread with Window.Update. It
// is safe from any goroutine and keeps what is posted before the window
// exists until attach.
type poster struct {
	mu      sync.Mutex
	win     *mygo.Window
	pending []func()
}

func (p *poster) post(fn func()) {
	p.mu.Lock()
	if p.win == nil {
		p.pending = append(p.pending, fn)
		p.mu.Unlock()
		return
	}
	w := p.win
	p.mu.Unlock()
	w.Update(fn)
}

func (p *poster) attach(w *mygo.Window) {
	p.mu.Lock()
	p.win = w
	pending := p.pending
	p.pending = nil
	p.mu.Unlock()
	for _, fn := range pending {
		w.Update(fn)
	}
}
