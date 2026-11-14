package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestShellConsoleCommandShowsAndHides pins the command center's console
// command: its title follows the console's state and the pane's title, it
// keeps the console command's ID and group, and its Run flips the console.
func TestShellConsoleCommandShowsAndHides(t *testing.T) {
	var s Shell
	cmd := s.ConsoleCommand()
	if cmd.ID != ConsoleCommandID || cmd.Title != "Show Console" || cmd.Group != "View" || cmd.Icon != "utilities-terminal" {
		t.Fatalf("closed: %+v", cmd)
	}
	cmd.Run()
	if !s.Console.Open {
		t.Fatal("Run did not open the console")
	}
	cmd = s.ConsoleCommand()
	if cmd.Title != "Hide Console" {
		t.Errorf("open: title %q", cmd.Title)
	}
	cmd.Run()
	if s.Console.Open {
		t.Error("Run did not close the console")
	}
	s.Console.Title = "Activity"
	if got := s.ConsoleCommand().Title; got != "Show Activity" {
		t.Errorf("titled, closed: %q", got)
	}
	s.Console.Open = true
	if got := s.ConsoleCommand().Title; got != "Hide Activity" {
		t.Errorf("titled, open: %q", got)
	}
}

// shellConsoleTester draws a Shell with a console, and a main menu when
// menu is set.
func shellConsoleTester(s *Shell, menu bool) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		f := Frame{ConsoleData: NewConsole(16), Content: func() {}}
		if menu {
			f.Menu = func(m *ui.Menu) { m.Item("About") }
		}
		s.View(c, f)
	}, 1280, 800)
}

// TestShellConsoleControlsFlipTheConsole pins that the top bar's console
// button, the status bar's console item and Ctrl J each open and close
// the console of the command center.
func TestShellConsoleControlsFlipTheConsole(t *testing.T) {
	for _, name := range []string{"Console (Ctrl J)", "Show or hide the console (Ctrl J)"} {
		var s Shell
		tt := shellConsoleTester(&s, false)
		if a, ok := tt.AccessByRole(ui.RoleButton, name); !ok {
			t.Fatalf("no button %q: %+v", name, a)
		}
		for _, want := range []bool{true, false, true} {
			if err := tt.Click(name); err != nil {
				t.Fatalf("%q: %v", name, err)
			}
			if s.Console.Open != want {
				t.Fatalf("%q: console open %v, want %v", name, s.Console.Open, want)
			}
		}
	}
	var s Shell
	tt := shellConsoleTester(&s, false)
	for _, want := range []bool{true, false} {
		tt.Key(ui.Cmd, ui.KeyJ)
		if s.Console.Open != want {
			t.Fatalf("Ctrl J: console open %v, want %v", s.Console.Open, want)
		}
	}
}

// TestShellMainMenuShowsOnlyWithAMenu pins the command center's main
// menu button: it shows, named and described "Main menu", only with
// Frame.Menu, and opens the menu.
func TestShellMainMenuShowsOnlyWithAMenu(t *testing.T) {
	var s Shell
	tt := shellConsoleTester(&s, false)
	if _, ok := tt.Find("Main menu"); ok {
		t.Error("a main menu shows with no Frame.Menu")
	}
	var m Shell
	tt = shellConsoleTester(&m, true)
	a, ok := tt.Access("Main menu")
	if !ok || a.Description != "Main menu" {
		t.Fatalf("main menu: %v, %+v", ok, a)
	}
	if err := tt.Click("Main menu"); err != nil {
		t.Fatal(err)
	}
	if got := tt.Menu(); len(got) != 1 || got[0] != "About" {
		t.Errorf("the main menu shows %v", got)
	}
}
