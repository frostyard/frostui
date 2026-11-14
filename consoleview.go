package frostui

import (
	"github.com/egoist/mygo/ui"
)

// ConsoleView is the console pane: every command the app ran and its
// output, as it ran, with filters by level and text, follow, copy and
// clear.
type ConsoleView struct {
	Open   bool
	Query  string
	Levels LevelSet
	// Subtitle follows the pane's title; empty shows "every command the
	// app runs, as it runs".
	Subtitle string
	// Title names the pane, as "Activity"; empty is "Console".
	Title string
	// Offered are the levels the pane offers to filter by, every level
	// when 0. A level not offered has no toggle and its lines do not
	// show.
	Offered LevelSet
	// EmptyNote is what the empty pane says under "Nothing yet"; empty
	// says "Commands and their output show here as they run."
	EmptyNote string
	list      ui.ListState
	init      bool
}

// title is the pane's title.
func (v *ConsoleView) title() string {
	if v.Title != "" {
		return v.Title
	}
	return "Console"
}

// toggle shows the pane while hidden and hides it while shown.
func (v *ConsoleView) toggle() { v.Open = !v.Open }

// command is the command that shows or hides the pane, titled by what it
// does now and the pane's title.
func (v *ConsoleView) command() Command {
	verb := "Show "
	if v.Open {
		verb = "Hide "
	}
	return consoleCommand(verb+v.title(), v.toggle)
}

// showing is the levels whose lines show: those turned on and offered.
func (v *ConsoleView) showing() LevelSet {
	if v.Offered != 0 {
		return v.Levels & v.Offered
	}
	return v.Levels
}

// View draws the pane for con.
func (v *ConsoleView) View(c *ui.Context, con *Console) {
	if !v.init {
		v.init = true
		if v.Levels == 0 {
			v.Levels = DefaultLevels
		}
		v.list.FollowEnd = true
	}
	t := c.Theme()
	showing := v.showing()
	lines := con.Lines(&showing, v.Query)
	title := v.title()
	copied, clearLabel, closeLabel := "Console copied", "Clear the console", "Close the console"
	if v.Title != "" {
		copied, clearLabel, closeLabel = v.Title+" copied", "Clear "+v.Title, "Close "+v.Title
	}
	ui.Column(c).Fill().Background(ColorsOf(t).Code).AlignItems(ui.Stretch).Role(ui.RoleRegion).Label(title).Children(func() {
		ui.Row(c).Gap(SpaceS).Padding(SpaceS, SpaceL).AlignItems(ui.Center).Children(func() {
			Icon(c, "utilities-terminal").TextColor(t.TextMuted)
			ui.Text(c, title).Bold().FontSize(13)
			sub := v.Subtitle
			if sub == "" {
				sub = "every command the app runs, as it runs"
			}
			Muted(c, sub).SingleLine()
			ui.Spacer(c)
			ui.ToggleGroup(c, func() {
				for l := Level(0); l < levelCount; l++ {
					if v.Offered != 0 && !v.Offered.Has(l) {
						continue
					}
					on := v.Levels.Has(l)
					if ui.Toggle(c.Key(int(l)), &on, LevelNames[l]).Changed() {
						v.Levels.Toggle(l)
					}
				}
			}).Label("Levels")
			ui.SearchField(c, &v.Query).Label("Filter lines").Placeholder("Filter lines").Width(200)
			if b := IconButton(c, "go-bottom", "Follow the newest line").Selected(v.list.FollowEnd); b.Clicked() {
				v.list.FollowEnd = !v.list.FollowEnd
				if v.list.FollowEnd {
					v.list.ScrollToEnd()
				}
			}
			if IconButton(c, "edit-copy", "Copy the lines shown").Clicked() {
				c.WriteClipboard(Text(lines))
				c.Announce(copied)
			}
			if IconButton(c, "edit-clear-all", clearLabel).Clicked() {
				con.Clear()
			}
			if IconButton(c, "window-close", closeLabel).Clicked() {
				v.Open = false
			}
		})
		ui.Box(c).Height(1).Background(t.Border)
		if len(lines) == 0 {
			note := v.EmptyNote
			if note == "" {
				note = "Commands and their output show here as they run."
			}
			Empty(c, "utilities-terminal", "Nothing yet", note).Grow(1)
			return
		}
		v.list.Key = func(i int) any { return lines[i].Seq }
		v.list.Label = func(i int) string { return lines[i].Text }
		ui.List(c, &v.list, len(lines), func(i int) {
			l := lines[i]
			color := t.Text
			switch l.Level {
			case LevelDebug, LevelOutput:
				color = t.Text.Alpha(0.82)
			case LevelCommand, LevelInfo:
				color = t.Accent
			case LevelDryRun, LevelWarn:
				color = t.Warning
			case LevelError:
				color = t.Danger
			}
			ui.Row(c).Gap(SpaceM).Padding(1, SpaceL).AlignItems(ui.Start).Children(func() {
				ui.Text(c, l.At.Format("15:04:05")).Font("monospace").FontSize(12).TextColor(t.TextMuted).SingleLine()
				ui.Text(c, LevelNames[l.Level]).Font("monospace").FontSize(12).TextColor(color).Width(64).SingleLine()
				text := l.Text
				if l.Level == LevelCommand {
					text = "$ " + text
				}
				ui.Text(c, text).Font("monospace").FontSize(12).TextColor(color).Wrap().Grow(1).Selectable()
			})
		}).Grow(1).Padding(SpaceXS, 0)
	})
}
