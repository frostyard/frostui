package frostui

import (
	"sync"
	"testing"
	"time"
)

// fakeAfter holds what a Waker scheduled, so a test runs it without a clock.
type fakeAfter struct {
	mu    sync.Mutex
	delay []time.Duration
	fns   []func()
}

func (f *fakeAfter) after(d time.Duration, fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delay = append(f.delay, d)
	f.fns = append(f.fns, fn)
}

func TestWakerCoalescesWakes(t *testing.T) {
	posts := 0
	var f fakeAfter
	w := NewWaker(func(fn func()) { fn(); posts++ }, 60*time.Millisecond)
	w.after = f.after
	var wg sync.WaitGroup
	for range 500 {
		wg.Add(1)
		go func() { defer wg.Done(); w.Wake() }()
	}
	wg.Wait()
	if len(f.fns) != 1 || f.delay[0] != 60*time.Millisecond {
		t.Fatalf("500 wakes scheduled %d posts after %v, want 1 after 60ms", len(f.fns), f.delay)
	}
	if posts != 0 {
		t.Fatalf("%d posts before the interval ended", posts)
	}
	f.fns[0]()
	if posts != 1 {
		t.Fatalf("the interval ended with %d posts", posts)
	}
	w.Wake()
	if len(f.fns) != 2 {
		t.Fatalf("a wake after the post scheduled %d in all, want 2", len(f.fns))
	}
}

// TestWakerPostsThroughATimer runs the real timer, waiting on the post.
func TestWakerPostsThroughATimer(t *testing.T) {
	posted := make(chan struct{})
	w := NewWaker(func(func()) { close(posted) }, time.Millisecond)
	w.Wake()
	select {
	case <-posted:
	case <-time.After(5 * time.Second):
		t.Fatal("the waker never posted")
	}
}
