package frostui

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// settingRow is a setting's line: its title, an optional line under it,
// and the control at its end, as SwitchRow lays them out.
func settingRow(c *ui.Context, title, subtitle string, control func()) {
	ui.Row(c).Gap(SpaceM).AlignItems(ui.Center).Children(func() {
		if subtitle == "" {
			Body(c, title).Grow(1).MinWidth(0).Wrap()
		} else {
			ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
				Strong(c, title).Wrap()
				Muted(c, subtitle).Wrap()
			})
		}
		control()
	})
}

// SelectRow is a setting with a drop-down of options at its end, labelled
// with the title, laid out as SwitchRow. It returns the drop-down, for
// Disabled, Tooltip and Changed. Key it with c.Key.
func SelectRow(c *ui.Context, title, subtitle string, selected *string, options []string) ui.Element {
	var e ui.Element
	settingRow(c, title, subtitle, func() { e = ui.Select(c, selected, options).Label(title).Width(200).Shrink(0) })
	return e
}

// TextRow is a setting with a text field at its end, labelled with the
// title, laid out as SwitchRow. It returns the field, for Placeholder,
// Disabled and Changed. Key it with c.Key.
func TextRow(c *ui.Context, title, subtitle string, value *string) ui.Element {
	var e ui.Element
	settingRow(c, title, subtitle, func() { e = ui.TextInput(c, value).Label(title).Width(240).Shrink(0) })
	return e
}

// NumberRow is a setting with a number field from lo to hi at its end,
// labelled with the title, laid out as SwitchRow. It returns the field,
// for Disabled and Changed. Key it with c.Key.
func NumberRow(c *ui.Context, title, subtitle string, value *float64, lo, hi, step float64) ui.Element {
	var e ui.Element
	settingRow(c, title, subtitle, func() { e = ui.NumberInput(c, value, lo, hi, step).Label(title).Width(120).Shrink(0) })
	return e
}

// Setting is one row of a settings page: a title, an optional line under
// it, and one control at its end, which the first field set of Switch,
// Choice, Text, Number and Control says. With none it shows Value.
type Setting struct {
	// Title names the setting; Subtitle explains it under the title.
	Title, Subtitle string
	// Keywords are matched by the search with the title and the
	// subtitle, but not shown: other words for the setting.
	Keywords string
	// Switch draws a switch of the value.
	Switch *bool
	// Choice draws a drop-down of Options.
	Choice  *string
	Options []string
	// Text draws a text field, with Placeholder when it is empty.
	Text        *string
	Placeholder string
	// Number draws a number field from Min to Max, in Step's steps.
	Number         *float64
	Min, Max, Step float64
	// Control draws a control of the app's, as a button.
	Control func(c *ui.Context)
	// Value is a read-only value, as a version or a key.
	Value string
	// Disabled grays out the control.
	Disabled bool
	// OnChange runs as the user changes the control's value; for a text
	// or number field, as they type.
	OnChange func()
}

// SettingsGroup is a titled card of settings on a page.
type SettingsGroup struct {
	// Title heads the group; Note explains it under the title.
	Title, Note string
	// Settings are the card's rows.
	Settings []Setting
	// Values are read-only lines under the card, drawn with KeyValues.
	Values []KeyValue
}

// SettingsPage is one section of a settings layout: an entry of the list
// at the left, and what shows at the right as it is chosen.
type SettingsPage struct {
	// ID identifies the page, for Settings.Selected.
	ID string
	// Title names the page, in the list and over its content; Icon is a
	// freedesktop icon name.
	Title, Icon string
	// Keywords are matched by the search: a match on them or the title
	// shows the whole page.
	Keywords string
	// Groups are the page's cards of settings.
	Groups []SettingsGroup
	// Body, when set, draws what is not a row after the groups, as a
	// list of accounts. The search does not read it.
	Body func(c *ui.Context)
}

// Settings is a settings layout: a list of sections at the left and the
// chosen section's rows at the right, for a settings window (View) or a
// dialog (Dialog). Keep one in the app's state; it belongs to the UI
// thread and to one window.
type Settings struct {
	// Selected is the ID of the page shown; empty or unknown, the first.
	Selected string
	// Search shows a field over the list, which finds settings on every
	// page; Query is its text.
	Search bool
	Query  string
	// Open says whether Dialog shows.
	Open bool
}

// Show opens the dialog on the page with the ID, or as it was for "".
func (s *Settings) Show(id string) {
	s.Open, s.Query = true, ""
	if id != "" {
		s.Selected = id
	}
}

// matchesSetting reports whether every word of query is in one of texts,
// ignoring case. Words, rather than Match's scattered letters: the
// subtitles are sentences, in which nearly any letters come in order.
func matchesSetting(query string, texts ...string) bool {
	all := strings.ToLower(strings.Join(texts, " "))
	for _, w := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(all, w) {
			return false
		}
	}
	return true
}

// found is what the search shows of a page: all of it when the page's
// title or keywords match, else its groups holding settings that match.
func (p SettingsPage) found(query string) (whole bool, groups []SettingsGroup) {
	if matchesSetting(query, p.Title, p.Keywords) {
		return true, p.Groups
	}
	for _, g := range p.Groups {
		if matchesSetting(query, g.Title, g.Note) {
			groups = append(groups, g)
			continue
		}
		hit := SettingsGroup{Title: g.Title, Note: g.Note}
		for _, st := range g.Settings {
			if matchesSetting(query, st.Title, st.Subtitle, st.Keywords, st.Value) {
				hit.Settings = append(hit.Settings, st)
			}
		}
		for _, kv := range g.Values {
			if matchesSetting(query, kv.Key, kv.Value) {
				hit.Values = append(hit.Values, kv)
			}
		}
		if len(hit.Settings) > 0 || len(hit.Values) > 0 {
			groups = append(groups, hit)
		}
	}
	return false, groups
}

// page returns the page chosen, nil without pages.
func (s *Settings) page(pages []SettingsPage) *SettingsPage {
	for i := range pages {
		if pages[i].ID == s.Selected {
			return &pages[i]
		}
	}
	if len(pages) == 0 {
		return nil
	}
	s.Selected = pages[0].ID
	return &pages[0]
}

// View draws the layout, filling its parent, and returns its row.
func (s *Settings) View(c *ui.Context, pages []SettingsPage) ui.Element {
	t := c.Theme()
	page := s.page(pages)
	return ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Width(208).Shrink(0).AlignItems(ui.Stretch).Background(Chrome(t)).Children(func() {
			if s.Search {
				ui.Box(c).Padding(SpaceS, SpaceS, 0, SpaceS).Children(func() {
					ui.SearchField(c, &s.Query).Label("Search settings").Placeholder("Search settings")
				})
			}
			selected := s.Selected
			ui.Sidebar(c, &selected, func() {
				for _, p := range pages {
					if ui.SidebarItem(c.Key(p.ID), p.ID, IconSVG(p.Icon), p.Title).Clicked() {
						s.Query = ""
					}
				}
			}).Grow(1).Padding(SpaceS, SpaceS).Background(Chrome(t)).Label("Settings sections")
			s.Selected = selected
		})
		ui.Box(c).Width(1).Background(t.Border)
		// Keyed by what it shows, so a new page starts at its top.
		key := s.Selected
		searching := strings.TrimSpace(s.Query) != ""
		if searching {
			key = "\x00search"
		}
		ui.Scroll(c.Key(key)).Grow(1).Padding(SpaceXL, 28).Gap(18).Children(func() {
			ui.Column(c).Gap(18).AlignItems(ui.Stretch).Children(func() {
				switch {
				case searching:
					s.results(c, pages)
				case page != nil:
					Title(c, page.Title).Role(ui.RoleHeading).Level(1)
					s.groups(c, page.ID, page.Groups)
					if page.Body != nil {
						page.Body(c)
					}
				}
			})
		})
	})
}

// results draws what the search found, page by page.
func (s *Settings) results(c *ui.Context, pages []SettingsPage) {
	shown := false
	for _, p := range pages {
		whole, groups := p.found(s.Query)
		if !whole && len(groups) == 0 {
			continue
		}
		shown = true
		ui.Column(c.Key(p.ID)).Gap(18).AlignItems(ui.Stretch).Children(func() {
			Title(c, p.Title).Role(ui.RoleHeading).Level(1)
			s.groups(c, p.ID, groups)
			if whole && p.Body != nil {
				p.Body(c)
			}
		})
	}
	if !shown {
		Empty(c, "edit-find", "No setting matches", "")
	}
}

// groups draws a page's cards.
func (s *Settings) groups(c *ui.Context, page string, groups []SettingsGroup) {
	for gi, g := range groups {
		ui.Column(c.Key([2]any{page, gi})).Gap(SpaceS).AlignItems(ui.Stretch).Children(func() {
			if g.Title != "" {
				Heading(c, g.Title).Role(ui.RoleHeading).Level(2)
			}
			if g.Note != "" {
				Muted(c, g.Note).Wrap()
			}
			if len(g.Settings) > 0 {
				Card(c).Children(func() {
					for i, st := range g.Settings {
						if i > 0 {
							RowDivider(c.Key([2]any{"divider", i}))
						}
						setting(c.Key(i), st)
					}
				})
			}
			if len(g.Values) > 0 {
				KeyValues(c, g.Values)
			}
		})
	}
}

// setting draws one row.
func setting(c *ui.Context, st Setting) {
	var control ui.Element
	switch {
	case st.Switch != nil:
		control = SwitchRow(c, st.Title, st.Subtitle, st.Switch)
	case st.Choice != nil:
		control = SelectRow(c, st.Title, st.Subtitle, st.Choice, st.Options)
	case st.Text != nil:
		control = TextRow(c, st.Title, st.Subtitle, st.Text).Placeholder(st.Placeholder)
	case st.Number != nil:
		control = NumberRow(c, st.Title, st.Subtitle, st.Number, st.Min, st.Max, st.Step)
	case st.Control != nil:
		settingRow(c, st.Title, st.Subtitle, func() { st.Control(c) })
		return
	default:
		settingRow(c, st.Title, st.Subtitle, func() {
			ui.Text(c, st.Value).FontSize(13).TextColor(c.Theme().TextMuted).Selectable()
		})
		return
	}
	if st.Disabled {
		control.Disabled(true)
	}
	if st.OnChange != nil && control.Changed() {
		st.OnChange()
	}
}

// Dialog draws the layout in a dialog while Open is set, under a title
// and a close button.
func (s *Settings) Dialog(c *ui.Context, title string, pages []SettingsPage) {
	if !s.Open {
		return
	}
	ui.Modal(c, &s.Open, func() {
		ui.Column(c).Width(760).Height(540).AlignItems(ui.Stretch).Role(ui.RoleDialog).Label(title).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Padding(0, 0, SpaceM, 0).Children(func() {
				ui.Text(c, title).FontSize(18).Bold().Grow(1)
				if IconButton(c, "window-close", "Close").Clicked() {
					s.Open = false
				}
			})
			RowDivider(c)
			ui.Column(c).Grow(1).AlignItems(ui.Stretch).Children(func() { s.View(c, pages) })
		})
	})
}
