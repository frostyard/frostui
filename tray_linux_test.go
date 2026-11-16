package frostui

import (
	"errors"
	"testing"
)

// TestTrayHostedAsksForTheWatcher injects the D-Bus check: TrayHosted asks
// for org.kde.StatusNotifierWatcher and reports its owner, and false when
// the bus cannot be asked.
func TestTrayHostedAsksForTheWatcher(t *testing.T) {
	saved := trayNameHasOwner
	t.Cleanup(func() { trayNameHasOwner = saved })

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
