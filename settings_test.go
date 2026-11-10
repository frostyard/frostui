package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// settingsApp is a settings layout over some values.
type settingsApp struct {
	s       Settings
	sync    bool
	theme   string
	name    string
	size    float64
	changes []string
	clicked int
}

func (a *settingsApp) pages() []SettingsPage {
	change := func(what string) func() { return func() { a.changes = append(a.changes, what) } }
	return []SettingsPage{
		{ID: "general", Title: "General", Icon: "preferences-system", Groups: []SettingsGroup{
			{Title: "Mail", Note: "How mail arrives.", Settings: []Setting{
				{Title: "Sync in the background", Subtitle: "Fetch new mail while closed", Switch: &a.sync, OnChange: change("sync")},
				{Title: "Theme", Keywords: "dark light appearance", Choice: &a.theme, Options: []string{"Light", "Dark"}, OnChange: change("theme")},
				{Title: "Display name", Text: &a.name, Placeholder: "Your name", OnChange: change("name")},
				{Title: "Font size", Number: &a.size, Min: 8, Max: 32, Step: 1},
				{Title: "Version", Value: "1.2.3"},
			}},
		}},
		{ID: "accounts", Title: "Accounts", Icon: "network-server", Keywords: "imap smtp", Groups: []SettingsGroup{
			{Title: "Server", Settings: []Setting{
				{Title: "Reset", Control: func(c *ui.Context) {
					if ui.Button(c, "Reset now").Clicked() {
						a.clicked++
					}
				}},
				{Title: "Locked", Switch: &a.sync, Disabled: true},
			}, Values: []KeyValue{{Key: "Host", Value: "mail.example.org"}}},
		}, Body: func(c *ui.Context) { Body(c, "A list of accounts") }},
	}
}

func settingsTester(a *settingsApp) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		a.s.View(c, a.pages())
	}, 900, 620)
}

func TestSettingsShowsTheFirstPageAndChangesPage(t *testing.T) {
	a := &settingsApp{theme: "Light", size: 12}
	tt := settingsTester(a)
	if a.s.Selected != "general" {
		t.Errorf("Selected = %q", a.s.Selected)
	}
	for _, want := range []string{"General", "Accounts", "Mail", "How mail arrives.", "Sync in the background", "Fetch new mail while closed", "1.2.3"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("A list of accounts") {
		t.Error("another page's body shows")
	}
	if err := tt.Click("Accounts"); err != nil {
		t.Fatal(err)
	}
	if a.s.Selected != "accounts" || !tt.HasText("A list of accounts") || !tt.HasText("mail.example.org") {
		t.Errorf("Accounts: %q", tt.Texts())
	}
	if err := tt.Click("Reset now"); err != nil {
		t.Fatal(err)
	}
	if a.clicked != 1 {
		t.Error("the app's control did not run")
	}
}

func TestSettingsControlsChangeTheirValues(t *testing.T) {
	a := &settingsApp{theme: "Light", size: 12}
	tt := settingsTester(a)
	clickRole := func(role ui.Role, name string) {
		t.Helper()
		a, ok := tt.AccessByRole(role, name)
		if !ok {
			t.Fatalf("no element of role %v named %q", role, name)
		}
		tt.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	}
	clickRole(ui.RoleSwitch, "Sync in the background")
	if !a.sync {
		t.Error("the switch did not change")
	}
	clickRole(ui.RoleTextField, "Display name")
	tt.Type("Ada")
	tt.Frame()
	if a.name != "Ada" {
		t.Errorf("name = %q", a.name)
	}
	if err := tt.Click("Light"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Dark"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.theme != "Dark" {
		t.Errorf("theme = %q", a.theme)
	}
	var sync, theme, name int
	for _, c := range a.changes {
		switch c {
		case "sync":
			sync++
		case "theme":
			theme++
		case "name":
			name++
		}
	}
	if sync != 1 || theme != 1 || name == 0 {
		t.Errorf("OnChange ran: %q", a.changes)
	}
}

func TestSettingsDisabledControlDoesNotChange(t *testing.T) {
	a := &settingsApp{s: Settings{Selected: "accounts"}}
	tt := settingsTester(a)
	sw, ok := tt.AccessByRole(ui.RoleSwitch, "Locked")
	if !ok || !sw.Disabled {
		t.Errorf("Locked = %+v, %v", sw, ok)
	}
}

func TestSettingsSearchFindsAcrossPages(t *testing.T) {
	a := &settingsApp{theme: "Light", s: Settings{Search: true}}
	tt := settingsTester(a)
	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	tt.Type("dark")
	tt.Frame()
	if !tt.HasText("Theme") || tt.HasText("Sync in the background") || tt.HasText("Reset now") {
		t.Errorf("a keyword search: %q", tt.Texts())
	}
	a.s.Query = "imap"
	tt.Frame()
	if !tt.HasText("A list of accounts") || !tt.HasText("Reset now") {
		t.Errorf("a page's keywords show the whole page: %q", tt.Texts())
	}
	a.s.Query = "example.org"
	tt.Frame()
	if !tt.HasText("mail.example.org") || tt.HasText("Reset now") {
		t.Errorf("a value's search: %q", tt.Texts())
	}
	a.s.Query = "sync mail"
	tt.Frame()
	if !tt.HasText("Sync in the background") {
		t.Errorf("words in any order across title and subtitle: %q", tt.Texts())
	}
	a.s.Query = "zzz"
	tt.Frame()
	if !tt.HasText("No setting matches") {
		t.Errorf("no match: %q", tt.Texts())
	}
	if err := tt.Click("Accounts"); err != nil {
		t.Fatal(err)
	}
	if a.s.Query != "" || !tt.HasText("A list of accounts") {
		t.Errorf("choosing a page keeps the search: %q", a.s.Query)
	}
}

func TestSettingsDialogOpensOnAPageAndCloses(t *testing.T) {
	a := &settingsApp{theme: "Light"}
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		Title(c, "Page")
		a.s.Dialog(c, "Preferences", a.pages())
	}, 1000, 700)
	if tt.HasText("Preferences") {
		t.Fatal("a closed dialog shows")
	}
	a.s.Show("accounts")
	tt.Frame()
	if _, ok := tt.AccessByRole(ui.RoleDialog, "Preferences"); !ok {
		t.Error("no dialog named Preferences")
	}
	if _, ok := tt.AccessByRole(ui.RoleTree, "Settings sections"); !ok {
		t.Error("no tree of the sections")
	}
	if _, ok := tt.AccessByRole(ui.RoleHeading, "Accounts"); !ok {
		t.Error("no heading of the page")
	}
	if !tt.HasText("A list of accounts") {
		t.Errorf("not on the page asked for: %q", tt.Texts())
	}
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	if a.s.Open || tt.HasText("Preferences") {
		t.Error("Close did not close it")
	}
	a.s.Show("")
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if a.s.Open {
		t.Error("Escape did not close it")
	}
}

func TestSettingRowsLabelTheirControls(t *testing.T) {
	choice, text, number := "a", "x", 2.0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Gap(SpaceS).Children(func() {
			SelectRow(c.Key(1), "Pick", "One of these", &choice, []string{"a", "b"})
			TextRow(c.Key(2), "Name", "", &text)
			NumberRow(c.Key(3), "Count", "", &number, 0, 9, 1)
		})
	}, 700, 300)
	for _, want := range []string{"Pick", "One of these", "Name", "Count"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if _, ok := tt.AccessByRole(ui.RoleTextField, "Name"); !ok {
		t.Error("the text field is not named by its title")
	}
}

func TestSettingsSearchKeepsAFieldsStateAsTheQueryChanges(t *testing.T) {
	var alpha, bravo, charlie string
	pages := []SettingsPage{{ID: "p", Title: "Page", Groups: []SettingsGroup{{Title: "G", Settings: []Setting{
		{Title: "Alpha field", Text: &alpha},
		{Title: "Bravo field", Text: &bravo},
		{Title: "Charlie field", Text: &charlie},
	}}}}}
	s := Settings{Search: true, Query: "field"}
	tt := ui.NewTester(func(c *ui.Context) { Adapt(c); s.View(c, pages) }, 900, 620)
	tt.Frame()
	f, ok := tt.AccessByRole(ui.RoleTextField, "Bravo field")
	if !ok {
		t.Fatal("no Bravo field")
	}
	tt.ClickAt(f.Bounds.X+f.Bounds.W/2, f.Bounds.Y+f.Bounds.H/2)
	if !tt.Focused("Bravo field") {
		t.Fatal("the field did not take the focus")
	}
	// Bravo moves from the second result to the first.
	s.Query = "bravo"
	tt.Frame()
	if !tt.Focused("Bravo field") {
		t.Error("the focus did not stay with the setting as the query changed")
	}
	tt.Type("x")
	tt.Frame()
	if bravo != "x" || alpha != "" || charlie != "" {
		t.Errorf("typing went to alpha %q, bravo %q, charlie %q", alpha, bravo, charlie)
	}
}

func TestSettingsTellsSameTitlesApart(t *testing.T) {
	var a, b bool
	pages := []SettingsPage{{ID: "p", Title: "Page", Groups: []SettingsGroup{
		{Settings: []Setting{{Title: "Same", Switch: &a}, {Title: "Same", Switch: &b}}},
		{Settings: []Setting{{Title: "Same", Switch: &a}}},
	}}}
	var s Settings
	// A Tester panics on two keys alike under one parent.
	tt := ui.NewTester(func(c *ui.Context) { Adapt(c); s.View(c, pages) }, 900, 620)
	tt.Frame()
	if !tt.HasText("Same") {
		t.Errorf("%q", tt.Texts())
	}
}
