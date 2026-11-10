package frostui

import (
	"sync"
	"time"
)

// Waker coalesces wake-ups from many goroutines into at most one post per
// interval. MyGo's Window.Update rebuilds the view on every call, so a
// stream such as a console's output must not call it once per line: give
// the console Wake (Console.SetNotify), and Window.Update as post.
type Waker struct {
	post  func(func())
	every time.Duration
	after func(time.Duration, func())

	mu      sync.Mutex
	pending bool
}

// NewWaker returns a Waker that posts an empty function through post at
// most once every interval.
func NewWaker(post func(func()), every time.Duration) *Waker {
	return &Waker{post: post, every: every, after: func(d time.Duration, fn func()) { time.AfterFunc(d, fn) }}
}

// Wake asks for a rebuild, one interval from now unless one is already
// due. It is safe from any goroutine.
func (w *Waker) Wake() {
	w.mu.Lock()
	if w.pending {
		w.mu.Unlock()
		return
	}
	w.pending = true
	w.mu.Unlock()
	w.after(w.every, func() {
		w.mu.Lock()
		w.pending = false
		w.mu.Unlock()
		w.post(func() {})
	})
}
