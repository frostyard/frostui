package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestBannerCloseNamesItsButton(t *testing.T) {
	closed := false
	tt := ui.NewTester(func(c *ui.Context) {
		if _, x := BannerClose(c, Warning, "dialog-warning", "Deleted elsewhere", "", "Close"); x {
			closed = true
		}
		ui.Text(c, "after")
	}, 600, 200)
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if !closed {
		t.Error("the close button did not report")
	}
	tt2 := ui.NewTester(func(c *ui.Context) {
		BannerClose(c, Warning, "dialog-warning", "A notice", "", "")
		Banner(c, Warning, "dialog-warning", "Another", "", true)
	}, 600, 200)
	if _, ok := tt2.Find("Close"); ok {
		t.Error("an empty label drew a close button")
	}
	if _, ok := tt2.Find("Dismiss"); !ok {
		t.Error("Banner's button is no longer Dismiss")
	}
}
