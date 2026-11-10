package frostui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// toastTester shows t once, on the first frame.
func toastTester(toast ActionToast) *ui.Tester {
	shown := false
	return ui.NewTester(func(c *ui.Context) {
		if !shown {
			shown = true
			ShowActionToast(c, toast)
		}
		TopToasts(c, "Close")
	}, 800, 400)
}

func TestActionToastRunsItsAction(t *testing.T) {
	acted, expired := 0, 0
	tt := toastTester(ActionToast{Title: "Deleted", Action: "Undo", OnAction: func() { acted++ }, OnExpire: func() { expired++ }})
	if !tt.HasText("Deleted") || !tt.HasText("Undo") {
		t.Fatalf("%q", tt.Texts())
	}
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if acted != 1 || expired != 0 || tt.HasText("Deleted") {
		t.Errorf("acted %d, expired %d, texts %q", acted, expired, tt.Texts())
	}
}

func TestActionToastReportsItsTimeEnding(t *testing.T) {
	acted, expired := 0, 0
	tt := toastTester(ActionToast{Title: "Deleted", Action: "Undo", OnAction: func() { acted++ }, OnExpire: func() { expired++ }})
	tt.Advance(7 * time.Second)
	if expired != 0 || !tt.HasText("Undo") {
		t.Fatalf("before its 8 seconds: expired %d, texts %q", expired, tt.Texts())
	}
	tt.Advance(2 * time.Second)
	if acted != 0 || expired != 1 || tt.HasText("Undo") {
		t.Errorf("after: acted %d, expired %d, texts %q", acted, expired, tt.Texts())
	}
}

func TestActionToastClosedReportsExpired(t *testing.T) {
	expired := 0
	tt := toastTester(ActionToast{Title: "Deleted", Action: "Undo", OnExpire: func() { expired++ }})
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if expired != 1 {
		t.Errorf("closed: expired %d", expired)
	}
}

func TestActionToastCountsDown(t *testing.T) {
	tt := toastTester(ActionToast{Title: "Sending", Action: "Undo", Timeout: 5 * time.Second, Countdown: true})
	timer := func(when, name string) {
		t.Helper()
		if _, ok := tt.AccessByRole(ui.RoleTimer, name); !ok {
			t.Errorf("%s: no timer named %q in %q", when, name, tt.Texts())
		}
	}
	timer("at first", "5 seconds left")
	tt.Advance(1500 * time.Millisecond)
	timer("after 1.5 s", "4 seconds left")
	tt.Advance(3 * time.Second)
	timer("after 4.5 s", "1 second left")
}

func TestActionToastEachCallShowsItsOwn(t *testing.T) {
	expired := 0
	shown := false
	tt := ui.NewTester(func(c *ui.Context) {
		if !shown {
			shown = true
			for range 2 {
				ShowActionToast(c, ActionToast{Title: "Sending", Action: "Undo", Timeout: time.Second, OnExpire: func() { expired++ }})
			}
		}
		TopToasts(c, "Close")
	}, 800, 400)
	tt.Advance(2 * time.Second)
	if expired != 2 {
		t.Errorf("two toasts of one title: %d expired, want both", expired)
	}
}

// TestPlainToastsKeepTheirLook leaves a toast not made by ShowActionToast
// without an action button in TopToasts, as before.
func TestPlainToastsKeepTheirLook(t *testing.T) {
	shown := false
	tt := ui.NewTester(func(c *ui.Context) {
		if !shown {
			shown = true
			c.ToastAction("Saved", "Open", func() {})
		}
		TopToasts(c, "Close")
	}, 800, 400)
	if !tt.HasText("Saved") || tt.HasText("Open") {
		t.Errorf("%q", tt.Texts())
	}
}
