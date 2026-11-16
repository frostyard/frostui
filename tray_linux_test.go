package frostui

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/frostyard/frostui/frostuitest"
)

// TestTrayHostedAsksForTheWatcher injects the D-Bus check: TrayHosted asks
// for org.kde.StatusNotifierWatcher and reports its owner, and false when
// the bus cannot be asked.
func TestTrayHostedAsksForTheWatcher(t *testing.T) {
	savedAsk, savedDeadline := trayNameHasOwner, trayDeadline
	t.Cleanup(func() { trayNameHasOwner, trayDeadline = savedAsk, savedDeadline })
	trayDeadline = func() <-chan time.Time { return nil } // never fires

	cases := []struct {
		name  string
		owned bool
		err   error
		want  bool
	}{
		{name: "a panel owns the watcher", owned: true, want: true},
		{name: "no panel", owned: false, want: false},
		{name: "no session bus", err: errors.New("no bus"), want: false},
		{name: "an error with an answer", owned: true, err: errors.New("timeout"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var asked []string
			trayNameHasOwner = func(name string) (bool, error) {
				asked = append(asked, name)
				return tc.owned, tc.err
			}
			if got := TrayHosted(); got != tc.want {
				t.Errorf("TrayHosted = %v, want %v", got, tc.want)
			}
			if len(asked) != 1 || asked[0] != "org.kde.StatusNotifierWatcher" {
				t.Errorf("asked for %q, want org.kde.StatusNotifierWatcher once", asked)
			}
		})
	}
}

// TestTrayHostedGivesUpAtTheDeadline stalls the D-Bus check, as a session
// bus that never finishes its connection: TrayHosted reports false when
// its deadline fires, the calls after it wait on the same query instead of
// starting more, and once the query ends the next call has its answer.
func TestTrayHostedGivesUpAtTheDeadline(t *testing.T) {
	savedAsk, savedDeadline := trayNameHasOwner, trayDeadline
	t.Cleanup(func() { trayNameHasOwner, trayDeadline = savedAsk, savedDeadline })

	release := make(chan struct{})
	var asks atomic.Int32
	trayNameHasOwner = func(string) (bool, error) {
		asks.Add(1)
		<-release
		return true, nil
	}
	fired := make(chan time.Time)
	close(fired) // every deadline has passed
	trayDeadline = func() <-chan time.Time { return fired }

	// A TrayHosted that ignores its deadline would block here for good:
	// frostuitest.Timeout turns that into a failure, not a hang.
	stalled := make(chan bool, 3)
	go func() {
		for range 3 {
			stalled <- TrayHosted()
		}
	}()
	for i := range 3 {
		select {
		case hosted := <-stalled:
			if hosted {
				t.Fatalf("call %d: TrayHosted = true while the query stalls, want false at the deadline", i+1)
			}
		case <-time.After(frostuitest.Timeout):
			close(release)
			t.Fatalf("call %d: TrayHosted still waits on a stalled query after its deadline fired", i+1)
		}
	}
	close(release)
	trayDeadline = func() <-chan time.Time { return nil } // never fires
	// The stalled query may still be in flight: the next call waits on it
	// and reads its answer, or starts a new one once it ended.
	if !TrayHosted() {
		t.Error("TrayHosted = false after the query answered true")
	}
	if n := asks.Load(); n < 1 || n > 2 {
		t.Errorf("the watcher was asked %d times, want 1 for the stalled calls and at most 1 more", n)
	}
}

// TestGIONameHasOwnerAnswersOrFails calls GIO for a name nobody owns: on a
// machine with a session bus it answers false, without one (as in CI) it
// fails with an error. It never panics, and never reports an owner.
func TestGIONameHasOwnerAnswersOrFails(t *testing.T) {
	owned, err := gioNameHasOwner("org.frostyard.FrostUI.Test.NobodyOwnsThis")
	if err == nil && owned {
		t.Error("gioNameHasOwner reports an owner for a name nobody owns")
	}
	if err != nil {
		t.Logf("no answer from the session bus: %v", err)
	}
}

func TestGIONameHasOwnerRejectsBadNames(t *testing.T) {
	for _, name := range []string{"", "a\x00b"} {
		if _, err := gioNameHasOwner(name); err == nil {
			t.Errorf("gioNameHasOwner(%q) succeeded, want an error", name)
		}
	}
}
