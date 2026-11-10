package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui"
)

// fetchFunc syncs one folder of an account and returns how many messages
// came: the program waits, the tests block or fail on purpose. It runs on
// the job's goroutine.
type fetchFunc func(ctx context.Context, p *frostui.Progress, account, folder string) (int, error)

// account is one mail account of the example.
type account struct {
	Name, Server string
	Folders      []string
	Messages     int
	Synced       time.Time
}

// demo is the example's state. It belongs to the UI thread, as frostui's
// models do; the jobs' goroutines only call fetch and Progress.
type demo struct {
	shell    frostui.Shell
	dialogs  frostui.Dialogs
	console  *frostui.Console
	jobs     *frostui.Jobs
	accounts []*account
	// lastRun is the last batch of jobs, kept after the queue went idle,
	// so the pipeline shows how it ended.
	lastRun []*frostui.Job

	openConsoleOnError, stopOnFailure bool

	fetch fetchFunc
	quit  func()
}

// pages are the rail's pages, in order, with the Alt+number that opens
// each.
var pages = []frostui.NavItem{
	{ID: "overview", Title: "Overview", Icon: "view-grid", Shortcut: "Alt+1"},
	{ID: "accounts", Title: "Accounts", Icon: "network-server", Shortcut: "Alt+2"},
	{ID: "settings", Title: "Settings", Icon: "preferences-system", Shortcut: "Alt+3"},
}

var pageKeys = []ui.Key{ui.Key1, ui.Key2, ui.Key3}

// newDemo returns the example's state. post hands work to the UI thread
// (Window.Update in the program, a frostuitest.Loop in the tests).
func newDemo(post func(func()), fetch fetchFunc) *demo {
	con := frostui.NewConsole(4000)
	con.SetNotify(frostui.NewWaker(post, 60*time.Millisecond).Wake)
	d := &demo{console: con, jobs: frostui.NewJobs(post, con), fetch: fetch, openConsoleOnError: true}
	d.shell.Brand = frostui.Brand{Name: "Command center", Tagline: "A frostui example", Logo: ui.MustParseSVG([]byte(logoSVG))}
	d.shell.Selected = "overview"
	d.shell.Dialogs = &d.dialogs
	d.shell.Console.Subtitle = "every job the example runs, as it runs"
	d.jobs.OnIdle = func(batch []*frostui.Job) {
		d.lastRun = batch
		for _, j := range batch {
			if j.State == frostui.JobFailed && d.openConsoleOnError {
				d.shell.Console.Open = true
			}
		}
	}
	d.accounts = []*account{
		{Name: "Personal", Server: "imaps://mail.example.org", Folders: []string{"Inbox", "Sent", "Archive"}},
		{Name: "Work", Server: "imaps://imap.example.com", Folders: []string{"Inbox", "Spam"}},
		{Name: "Lists", Server: "imaps://lists.example.net", Folders: []string{"Inbox"}},
	}
	con.Add(frostui.LevelInfo, "app", "Command center started")
	return d
}

// sync queues one job per folder of each account, after the jobs that
// already wait.
func (d *demo) sync(accts ...*account) {
	for _, a := range accts {
		for _, folder := range a.Folders {
			if d.jobs.Pending(a.Name, folder) {
				continue
			}
			var got int
			d.jobs.Enqueue(&frostui.Job{
				Title: fmt.Sprintf("Syncing %s · %s", a.Name, folder), Source: a.Name, Key: folder,
				Run: func(ctx context.Context, p *frostui.Progress) error {
					p.Log(frostui.LevelCommand, "fetch %s/%s", a.Server, folder)
					n, err := d.fetch(ctx, p, a.Name, folder)
					got = n
					return err
				},
				Done: func(err error) {
					switch {
					case err == nil:
						a.Messages += got
						a.Synced = time.Now()
					case d.stopOnFailure && !errors.Is(err, context.Canceled):
						d.jobs.CancelAll()
					}
				},
			})
		}
	}
}

// job is the newest job of an account's folder: queued, running, or in
// the last run.
func (d *demo) job(a *account, folder string) *frostui.Job {
	if j := d.jobs.Find(a.Name, folder); j != nil {
		return j
	}
	for i := len(d.lastRun) - 1; i >= 0; i-- {
		if j := d.lastRun[i]; j.Source == a.Name && j.Key == folder {
			return j
		}
	}
	return nil
}

func (d *demo) view(c *ui.Context) {
	for i, key := range pageKeys {
		id := pages[i].ID
		c.OnWindowShortcut(ui.Alt, key, func() { d.shell.Selected = id })
	}
	c.OnWindowShortcut(ui.Cmd, ui.KeyR, func() { d.sync(d.accounts...) })
	frostui.TopToasts(c, "Close")
	d.shell.View(c, frostui.Frame{
		Nav:        []frostui.NavSection{{Title: "Mail", Items: d.nav()}},
		Actions:    func() { d.actions(c) },
		Menu:       d.menu,
		RailFooter: func() { frostui.JobCard(c, d.jobs) },
		Content: func() {
			switch d.shell.Selected {
			case "accounts":
				d.accountsPage(c)
			case "settings":
				d.settingsPage(c)
			default:
				d.overviewPage(c)
			}
		},
		Commands:    d.commands(),
		Status:      d.status(),
		StatusEnd:   []frostui.StatusItem{{Icon: "help-about", Text: "frostui example", Progress: -1}},
		ConsoleData: d.console,
	})
}

func (d *demo) nav() []frostui.NavItem {
	items := append([]frostui.NavItem(nil), pages...)
	never := 0
	for _, a := range d.accounts {
		if a.Synced.IsZero() {
			never++
		}
	}
	if never > 0 {
		items[1].Badge, items[1].BadgeTone = fmt.Sprint(never), frostui.Accent
	}
	return items
}

func (d *demo) actions(c *ui.Context) {
	if frostui.IconButton(c, "view-refresh", "Sync all (Ctrl R)").Disabled(d.jobs.Busy()).Clicked() {
		d.sync(d.accounts...)
	}
}

func (d *demo) menu(m *ui.Menu) {
	if m.Item("Command palette").Shortcut(ui.Cmd, ui.KeyK).Chosen() {
		d.shell.Palette.Show()
	}
	if m.Item("Keyboard shortcuts").Chosen() {
		d.showShortcuts()
	}
	if m.Item("About Command center").Chosen() {
		d.showAbout()
	}
	if d.quit != nil {
		m.Separator()
		if m.Item("Quit").Shortcut(ui.Cmd, ui.KeyQ).Chosen() {
			d.quit()
		}
	}
}

func (d *demo) commands() []frostui.Command {
	busy := d.jobs.Busy()
	cmds := []frostui.Command{
		{ID: "sync", Title: "Sync all accounts", Group: "Actions", Icon: "view-refresh", Keys: []string{"Ctrl", "R"},
			Disabled: busy, Reason: "a sync runs", Run: func() { d.sync(d.accounts...) }},
		{ID: "stop", Title: "Stop the sync", Group: "Actions", Icon: "process-stop", Hint: "cancel",
			Disabled: !d.jobs.CanStop(), Reason: "nothing runs", Run: d.jobs.CancelAll},
		{ID: "clear", Title: "Clear the console", Group: "Actions", Icon: "edit-clear-all", Run: d.askClear},
	}
	for i, p := range pages {
		id := p.ID
		cmds = append(cmds, frostui.Command{ID: "go-" + id, Title: "Go to " + p.Title, Group: "Go to", Icon: p.Icon,
			Keys: []string{"Alt", fmt.Sprint(i + 1)}, Run: func() { d.shell.Selected = id }})
	}
	return append(cmds,
		frostui.Command{ID: "keys", Title: "Keyboard shortcuts", Group: "Help", Icon: "input-keyboard", Run: d.showShortcuts},
		frostui.Command{ID: "about", Title: "About Command center", Group: "Help", Icon: "help-about", Run: d.showAbout},
	)
}

func (d *demo) askClear() {
	d.dialogs.Ask(frostui.Confirm{
		Title: "Clear the console?", Body: "Every line goes; the jobs keep their results.",
		Action: "Clear", Destructive: true, OnConfirm: d.console.Clear,
	})
}

func (d *demo) showShortcuts() {
	entries := []frostui.ShortcutEntry{
		{Group: "General", Title: "Command palette", Keys: []string{"Ctrl", "K"}},
		{Group: "General", Title: "Console", Keys: []string{"Ctrl", "J"}},
		{Group: "General", Title: "Sync all", Keys: []string{"Ctrl", "R"}},
	}
	for i, p := range pages {
		entries = append(entries, frostui.ShortcutEntry{Group: "Go to", Title: p.Title, Keys: []string{"Alt", fmt.Sprint(i + 1)}})
	}
	d.shell.Info.Show("Keyboard shortcuts", frostui.ShortcutsBody(entries))
}

func (d *demo) showAbout() {
	d.shell.Info.Show("About Command center", frostui.AboutBody(frostui.About{
		Name: "Command center", Version: "example", Summary: "The command-center shell of frostui, with a pretend mail sync.",
		Logo: "commandcenter", Website: "https://github.com/frostyard/frostui", License: "MIT",
	}))
}

func (d *demo) status() []frostui.StatusItem {
	if cur := d.jobs.Current(); cur != nil {
		done, total := d.jobs.Batch()
		return []frostui.StatusItem{{Icon: "network-transmit-receive", Text: fmt.Sprintf("Syncing %d of %d", done+1, total),
			Tone: frostui.Accent, Tooltip: cur.Title, Progress: d.jobs.BatchFraction()}}
	}
	text := "Idle"
	if s := frostui.RunSummary(d.lastRun); s != "" {
		text = "Last sync: " + s
	}
	return []frostui.StatusItem{{Icon: "network-transmit-receive", Text: text, Progress: -1}}
}

func (d *demo) overviewPage(c *ui.Context) {
	messages, synced := 0, 0
	for _, a := range d.accounts {
		messages += a.Messages
		if !a.Synced.IsZero() {
			synced++
		}
	}
	warnings, errs := d.console.Counts()
	frostui.PageHeader(c, "Overview", fmt.Sprintf("%d accounts · %d messages", len(d.accounts), messages), func() {
		if frostui.LabeledButton(c, "view-refresh", "Sync all", true).Disabled(d.jobs.Busy()).Clicked() {
			d.sync(d.accounts...)
		}
	})
	failed := 0
	for _, j := range d.lastRun {
		if j.State == frostui.JobFailed {
			failed++
		}
	}
	if failed > 0 && !d.jobs.Busy() {
		if acted, _ := frostui.Banner(c, frostui.Warning, "dialog-warning",
			fmt.Sprintf("%d folders failed to sync. The console has the error.", failed), "Open the console", false); acted {
			d.shell.Console.Open = true
		}
	}
	frostui.Figures(c, []frostui.Figure{
		{Label: "Accounts", Value: fmt.Sprint(len(d.accounts)), Note: fmt.Sprintf("%d synced", synced), Tone: frostui.Success,
			OnClick: func() { d.shell.Selected = "accounts" }},
		{Label: "Messages", Value: fmt.Sprint(messages)},
		{Label: "Jobs run", Value: fmt.Sprint(len(d.jobs.History()))},
		{Label: "Problems", Value: fmt.Sprint(warnings + errs), Note: "in the console", Tone: frostui.Warning,
			OnClick: func() { d.shell.Console.Open = true }},
	})
	frostui.Card(c).Children(func() {
		sub := "Not run yet"
		if cur := d.jobs.Current(); cur != nil {
			done, total := d.jobs.Batch()
			sub = fmt.Sprintf("Job %d of %d", done+1, total)
		} else if s := frostui.RunSummary(d.lastRun); s != "" {
			sub = s
		}
		frostui.CardHeader(c, "Sync run", sub, nil)
		var stages []frostui.Stage
		for _, a := range d.accounts {
			st := frostui.Stage{Title: a.Name, Icon: "mail-unread"}
			for _, folder := range a.Folders {
				step := frostui.Step{Title: folder}
				if j := d.job(a, folder); j != nil {
					step.State = frostui.StateOf(j)
					if j.State.Ended() {
						step.Note = frostui.Duration(j.Ended.Sub(j.Started))
					}
				}
				st.Steps = append(st.Steps, step)
			}
			stages = append(stages, st)
		}
		frostui.Pipeline(c, stages).Label("Sync run")
	})
}

func (d *demo) accountsPage(c *ui.Context) {
	frostui.PageHeader(c, "Accounts", "Each folder syncs as a job of its own", nil)
	frostui.Card(c).Children(func() {
		for i, a := range d.accounts {
			if i > 0 {
				frostui.RowDivider(c.Key(fmt.Sprint("div", i)))
			}
			d.accountRow(c.Key(a.Name), a)
		}
	})
}

func (d *demo) accountRow(c *ui.Context, a *account) {
	pending, failed := false, ""
	for _, folder := range a.Folders {
		if j := d.job(a, folder); j != nil {
			pending = pending || !j.State.Ended()
			if j.State == frostui.JobFailed {
				failed = j.Err.Error()
			}
		}
	}
	frostui.CommandRow(c, frostui.CommandItem{
		Icon: "network-server", Title: a.Name, Command: a.Server,
		Status: func() {
			s := frostui.RowStatus{Label: "Sync " + a.Name}
			switch {
			case pending:
				s.State, s.Text = frostui.StepRunning, "Syncing"
			case failed != "":
				s.State, s.Text, s.Tooltip = frostui.StepFailed, "Failed", failed
			case !a.Synced.IsZero():
				s.State, s.Text, s.Tooltip = frostui.StepDone, "Synced", "at "+a.Synced.Format("15:04")
			default:
				s.State, s.Text = frostui.StepReady, "Never synced"
			}
			frostui.JobStatus(c, s)
		},
		Action: func() {
			if ui.Button(c, "Sync").Label("Sync " + a.Name).Disabled(pending).Clicked() {
				d.sync(a)
			}
		},
	})
}

func (d *demo) settingsPage(c *ui.Context) {
	frostui.PageHeader(c, "Settings", "Kept for this session only", nil)
	frostui.Card(c).Children(func() {
		frostui.SwitchRow(c.Key("console"), "Open the console on a failure", "", &d.openConsoleOnError)
		frostui.RowDivider(c)
		frostui.SwitchRow(c.Key("stop"), "Stop at the first failure", "Drop the folders that wait when one fails", &d.stopOnFailure)
	})
}
