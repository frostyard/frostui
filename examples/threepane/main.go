// Threepane opens frostui's frame without the rail with a small mail
// reader for content: three panes under a top bar whose segments follow
// them, a console called Activity (Ctrl J), the connection in the status
// bar, and Undo after a delete. Nothing talks to a server: the connection
// is a timer, and the messages are made up.
//
//	go run ./examples/threepane
package main

import (
	"log"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// appID names the example for the desktop (frostui.SetAppID).
const appID = "org.frostyard.FrostUI.ThreePane"

func main() {
	if err := frostui.SetAppID(appID, "Three panes"); err != nil {
		log.Printf("threepane: cannot set the app ID: %v", err)
	}
	mygo.App.SetName(appID)
	// On Linux the icons come from the desktop's theme and elsewhere from
	// the kit's own set.
	frostui.SetIcons(frostui.NewIcons(nil))

	p := &poster{}
	d := newDemo(p.post)
	d.quit = mygo.App.Quit
	// The first connection takes a moment, to show the connecting screen.
	d.conn.State = frostui.Connecting
	time.AfterFunc(1500*time.Millisecond, func() { p.post(d.connected) })
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title: "Three panes", Width: 1280, Height: 800, MinWidth: 800, MinHeight: 500,
			TitleBarStyle: mygo.TitleBarHidden,
			Content:       ui.View(d.view),
		})
		p.attach(win)
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

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
