// Commandcenter opens frostui's command-center shell with a small mail
// sync for content: a top bar with the command palette (Ctrl K), a rail of
// pages with the running job, the console (Ctrl J) and a status bar. "Sync
// all" queues one job per folder; the jobs run one at a time and write
// their progress to the rail, the status bar and the console. Nothing
// talks to a server: a folder's sync waits a moment and makes up a count.
// A tray icon carries the status bar's text, a dot after a failed sync, and
// Open, "Sync all accounts" and Quit; with "Close to the tray" on in the
// settings, closing the window hides it while a panel shows the icon.
//
//	go run ./examples/commandcenter
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// appID names the example for the desktop (frostui.SetAppID).
const appID = "org.frostyard.FrostUI.CommandCenter"

func main() {
	if err := frostui.SetAppID(appID, "Command center"); err != nil {
		log.Printf("commandcenter: cannot set the app ID: %v", err)
	}
	mygo.App.SetName(appID)
	setIcons()

	p := &poster{}
	d := newDemo(p.post, simulatedFetch)
	d.quit = mygo.App.Quit
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title: "Command center", Width: 1280, Height: 800, MinWidth: 960, MinHeight: 600,
			TitleBarStyle: mygo.TitleBarHidden,
			Content:       ui.View(d.view),
		})
		p.attach(win)
		tray, err := frostui.NewTray(frostui.TrayOptions{
			ID: "commandcenter", ToolTip: "Command center", OpenLabel: "Open Command center",
			Icon: trayIcon(), Window: win, Items: d.trayItems(),
		})
		if err != nil {
			log.Printf("commandcenter: no tray icon: %v", err)
			return
		}
		d.tray = tray
		win.OnClose(func(e *mygo.CloseEvent) { tray.CloseToTray(e, d.closeToTray) })
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// setIcons gives the kit its icons. The logo is the one icon the example
// ships: on Linux the rest come from the desktop's icon theme, and
// elsewhere they draw as empty boxes.
func setIcons() {
	frostui.SetIcons(frostui.NewIcons(map[string][]byte{"commandcenter-symbolic": []byte(logoSVG)}))
}

// simulatedFetch stands in for a folder's sync: five steps of a fifth of a
// second, each reported to the job's progress and the console. The Spam
// folder of the Work account always fails, so the example shows a failure.
func simulatedFetch(ctx context.Context, p *frostui.Progress, account, folder string) (int, error) {
	for i := 1; i <= 5; i++ {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
		p.Set(float64(i)/5, fmt.Sprintf("%s: %d of 5", folder, i))
		p.Log(frostui.LevelOutput, "%s/%s: fetched block %d of 5", account, folder, i)
		if account == "Work" && folder == "Spam" && i == 3 {
			return 0, errors.New("the server closed the connection")
		}
	}
	return len(account) + 3*len(folder), nil
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

// trayIcon draws the example's mark for the tray: the six-armed star of
// logoSVG in the brand's blue, 32 pixels square, smoothed by 4 by 4
// samples a pixel. A tray icon is a PNG in color (frostui.TrayOptions).
func trayIcon() []byte {
	const size, samples = 32, 4
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			in := 0
			for sy := range samples {
				for sx := range samples {
					px := float64(x) + (float64(sx)+0.5)/samples - size/2
					py := float64(y) + (float64(sy)+0.5)/samples - size/2
					if onStar(px, py) {
						in++
					}
				}
			}
			if in > 0 {
				img.SetNRGBA(x, y, color.NRGBA{R: 0x47, G: 0xb8, B: 0xef, A: uint8(255 * in / (samples * samples))})
			}
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		panic(err) // an NRGBA image in memory always encodes
	}
	return b.Bytes()
}

// onStar reports whether a point, from the icon's center, is on one of
// the star's three arms: bars 28 long and 4 wide at 90, 30 and 150 degrees.
func onStar(x, y float64) bool {
	for _, deg := range []float64{90, 30, 150} {
		a := deg * math.Pi / 180
		along := x*math.Cos(a) + y*math.Sin(a)
		across := -x*math.Sin(a) + y*math.Cos(a)
		if math.Abs(along) <= 14 && math.Abs(across) <= 2 {
			return true
		}
	}
	return false
}

// logoSVG is the example's mark, a six-armed star, drawn in the text color.
const logoSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16"><path fill="#2e3436" d="M7.05 15 L8.95 15 L8.95 1 L7.05 1Z M1.46 10.68 L2.41 12.32 L14.54 5.32 L13.59 3.68Z M2.41 3.68 L1.46 5.32 L13.59 12.32 L14.54 10.68Z"/></svg>`
