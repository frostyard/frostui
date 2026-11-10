package frostui

import (
	"context"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

// hasAll fails t for each text tt does not show.
func hasAll(t *testing.T, tt *ui.Tester, texts ...string) {
	t.Helper()
	for _, s := range texts {
		if !tt.HasText(s) {
			t.Errorf("missing %q in %q", s, tt.Texts())
		}
	}
}

func TestCardsHeadersAndStates(t *testing.T) {
	acted, closed, figure := 0, 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Gap(SpaceS).AlignItems(ui.Stretch).Children(func() {
			PageHeader(c, "Updates", "3 ready", func() { ui.Button(c, "Check now") })
			Card(c).Children(func() {
				CardHeader(c, "System image", "Staged with bootc", func() { Pill(c, "staged", Accent) })
				Loading(c, "Checking for updates")
				Problem(c, "brew outdated failed: exit 1")
				Empty(c, "object-select", "Up to date", "Nothing waits.")
			})
			a, x := Banner(c, Warning, "changes-prevent", "A tap is not trusted.", "Trust tap", true)
			if a {
				acted++
			}
			if x {
				closed++
			}
			Figures(c, []Figure{
				{Label: "Updates", Value: "12", Note: "across 5 sources", Tone: Accent, OnClick: func() { figure++ }},
				{Label: "Free space", Value: "288 GB"},
			})
		})
	}, 1000, 900)
	hasAll(t, tt, "Updates", "3 ready", "Check now", "System image", "Staged with bootc", "staged",
		"Checking for updates", "brew outdated failed: exit 1", "Up to date", "Nothing waits.",
		"A tap is not trusted.", "12", "across 5 sources", "288 GB")
	if err := tt.Click("Trust tap"); err != nil || acted != 1 {
		t.Errorf("banner action: %v, acted %d", err, acted)
	}
	if err := tt.Click("Dismiss"); err != nil || closed != 1 {
		t.Errorf("banner close: %v, closed %d", err, closed)
	}
	if err := tt.Click("Updates: 12"); err != nil || figure != 1 {
		t.Errorf("figure click: %v, clicks %d", err, figure)
	}
}

func TestChartsAndGauges(t *testing.T) {
	clicks := 0
	tt := ui.NewTesterWith(func(c *ui.Context) {
		ui.Column(c).Fill().Gap(SpaceS).AlignItems(ui.Stretch).Children(func() {
			LiveChart(c, "Processor and memory", 120,
				Series{Name: "CPU", Values: []float64{5, 40, 120, -3}, Tone: Accent},
				Series{Name: "Memory", Values: []float64{50}, Tone: Warning}) // too short: skipped
			Legend(c, Accent, "CPU 12%")
			Gauge(c, 288, 512, "288 GB", "free", Success, 160)
			Gauge(c, 1, 0, "?", "no total", Neutral, 80)
			SegmentBar(c, "Disk", 100, []Segment{{Label: "Home", Value: 60, Tone: Accent}, {Label: "Empty", Value: 0}}, func(s Segment) string {
				return s.Label + " part"
			})
			SegmentBar(c, "Nothing", 0, nil, nil)
			if ToggleTile(c, "application-x-addon", "Docker", "On · 28.5", true, false) {
				clicks++
			}
			if ToggleTile(c, "application-x-addon", "Incus", "Working…", false, true) {
				clicks += 10
			}
		})
	}, 900, 1000, ui.TesterOptions{Dark: true})
	hasAll(t, tt, "CPU 12%", "288 GB", "free", "Home part", "Docker", "Incus")
	for _, label := range []string{"Processor and memory", "free: 288 GB", "Disk"} {
		if _, ok := tt.Find(label); !ok {
			t.Errorf("no element named %q", label)
		}
	}
	if err := tt.Click("Docker"); err != nil || clicks != 1 {
		t.Errorf("tile click: %v, clicks %d", err, clicks)
	}
	if err := tt.Click("Incus"); err != nil || clicks != 1 {
		t.Errorf("a busy tile took a click: %v, clicks %d", err, clicks)
	}
}

func TestPipelineShowsEveryStepState(t *testing.T) {
	stages := []Stage{
		{Title: "System image", Icon: "security-high", Note: "1m 03s", Steps: []Step{{Title: "Snow Linux", State: StepDone}}},
		{Title: "Flatpak", Icon: "package-x-generic", Steps: []Step{
			{Title: "Firefox", State: StepRunning, Note: "9s"}, {Title: "Bazaar", State: StepQueued}, {Title: "Inkscape", State: StepReady},
		}},
		{Title: "Homebrew", Icon: "utilities-terminal", Steps: []Step{
			{Title: "gh", State: StepFailed}, {Title: "bluefin-cli", State: StepHeld}, {Title: "go", State: StepSkipped},
		}},
		{Title: "Features", Icon: "application-x-addon"},
		{Title: "Firmware", Icon: "drive-harddisk", Empty: "Waits for you"},
	}
	tt := ui.NewTester(func(c *ui.Context) { Pipeline(c, stages) }, 1300, 500)
	hasAll(t, tt, "System image", "1m 03s", "Firefox", "9s", "Bazaar", "Inkscape", "gh", "bluefin-cli", "go",
		"Nothing to do", "Waits for you")
	for _, label := range []string{"Firefox, running", "Bazaar, queued", "Inkscape, ready", "gh, failed", "bluefin-cli, held", "go, skipped", "Snow Linux, done"} {
		if _, ok := tt.Find(label); !ok {
			t.Errorf("no step named %q", label)
		}
	}
}

func TestJobCardShowsTheBatch(t *testing.T) {
	q := NewJobs(func(func()) {}, nil)
	tt := ui.NewTester(func(c *ui.Context) { JobCard(c, q) }, 400, 200)
	if len(tt.Texts()) != 0 {
		t.Errorf("an idle queue drew %q", tt.Texts())
	}
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })
	q.Enqueue(&Job{Title: "Updating Firefox", Run: func(_ context.Context, _ *Progress) error { <-block; return nil }},
		&Job{Title: "Updating gh"})
	tt.Frame()
	hasAll(t, tt, "Updating Firefox", "Job 1 of 2")
	q.Current().Status = "Downloading 3 of 5"
	tt.Frame()
	hasAll(t, tt, "Downloading 3 of 5")
}

func TestDialogsShowShortcutsAndAbout(t *testing.T) {
	var info Info
	tt := ui.NewTester(func(c *ui.Context) { info.View(c) }, 900, 700)
	if len(tt.Texts()) != 0 {
		t.Errorf("a closed dialog drew %q", tt.Texts())
	}
	info.Show("Keyboard shortcuts", ShortcutsBody([]ShortcutEntry{
		{Group: "Navigation", Title: "Go to Updates", Keys: []string{"Alt", "2"}},
		{Group: "General", Title: "Quit", Keys: []string{"Ctrl", "Q"}},
	}))
	tt.Frame()
	hasAll(t, tt, "Keyboard shortcuts", "Navigation", "Go to Updates", "Alt", "2", "General", "Quit")
	if err := tt.Click("Close"); err != nil || info.Open {
		t.Errorf("close: %v, open %v", err, info.Open)
	}
	info.Show("About Example", AboutBody(About{
		Name: "Example", Version: "v0.12.0", Summary: "A frostui command center.", Logo: "computer",
		Website: "https://example.org/site", Issues: "https://example.org/issues", License: "GNU GPL 3.0",
		Copyright: "© Frostyard", Developers: []string{"Brian", "Contributors"},
	}))
	tt.Frame()
	hasAll(t, tt, "Example", "v0.12.0", "A frostui command center.", "Developers", "Brian", "Contributors",
		"GNU GPL 3.0", "© Frostyard", "Website", "Report an issue")
}

func TestSetAppIDNamesTheProgram(t *testing.T) {
	err := SetAppID("org.frostyard.Test", "Test")
	if err != nil && strings.Contains(err.Error(), "libglib") {
		t.Skipf("no GLib here: %v", err)
	}
	if err != nil {
		t.Fatalf("SetAppID: %v", err)
	}
	// Only the first call counts, as GLib allows.
	if err := SetAppID("", ""); err != nil {
		t.Errorf("a second call = %v, want the first call's nil", err)
	}
}

func TestPipelineTakesTheCallersName(t *testing.T) {
	stages := []Stage{{Title: "Inbox", Icon: "mail-unread", Steps: []Step{{Title: "Fetch", State: StepRunning}}}}
	tt := ui.NewTester(func(c *ui.Context) {
		Pipeline(c.Key("default"), stages)
		Pipeline(c.Key("named"), stages).Label("Sync run")
	}, 900, 400)
	for _, name := range []string{"Pipeline", "Sync run"} {
		if _, ok := tt.AccessByRole(ui.RoleList, name); !ok {
			t.Errorf("no list named %q", name)
		}
	}
}
