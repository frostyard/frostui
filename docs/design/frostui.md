# frostui: frostyard's desktop UI kit

Living document. Rationale: [ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md), [ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md).
Contracts: [spec: API stability](../specs/api-stability.md).
Built in: [roadmap](../plans/roadmap.md).

## Overview

`github.com/frostyard/frostui` is a thin layer over MyGo's native UI package
(`github.com/egoist/mygo/ui`) that gives frostyard's desktop apps one look
and one set of behaviors. It came out of ChairLift's `internal/frostui`
([ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md)); ChairLift,
Frostmail and Frostnotes build their windows from it. It builds without CGO
for Linux, Windows, macOS and js/wasm
([ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)).

```
app state ──► Shell.View(c, Frame{…}) ──► top bar · rail · page · console · status bar
                    │                          ▲
app pages ──► Card, Figures, CommandRow, …     │ Window.Update (post)
                                               │
app changes ──► Jobs.Enqueue(Job) ──► Run on a goroutine ──► Done on the UI thread
                         └──► Console (start, output, end) ──► ConsoleView
```

![The command-center example mid-sync](../images/commandcenter.png)

The picture is `examples/commandcenter` drawn by its snapshot test
(`make snapshot`).

## Design

### The command-center shell

`Shell` (`shell.go`) is the window: a top bar (brand, a field that opens the
command palette, the app's buttons, a main menu), a left rail of pages with
counts and a footer, the page area (scrolled, padded, at most 1320 DIPs
wide), an optional console split under it, and a status bar of clickable
items. The app keeps one `Shell` in its state and calls
`Shell.View(c, Frame{…})` from its root view every frame; the `Frame` says
what goes where: `Nav` (sections of `NavItem`, with a `Badge` and a
`Shortcut` for the tooltip), `Actions`, `Menu`, `RailFooter` (the running
job's `JobCard`, a host card), `Content` (the selected page), `Commands`,
`Status` and `StatusEnd`, and `ConsoleData`. `Shell.Selected` is the rail's
page; the app reads it to pick `Content` and writes it to navigate.

The shell binds Ctrl K (palette) and, with a console, Ctrl J (console), and
draws the shell's `Dialogs`, its `Info` dialog and the palette over the
page. An app binds its own shortcuts with `c.OnWindowShortcut` and lists
them with `ShortcutsBody` in `Info`.

**The pattern** an app follows: a rail of the places it has (pages,
mailboxes, notebooks), the palette for every command, including "Go to"
each page, the console as an activity log of what the app runs, the job
queue for every change, and the status bar for the state of the current
run. `examples/commandcenter` is the reference: a pretend mail sync with an
overview, an accounts page of `CommandRow`s and a settings page of
`SwitchRow`s.

### Command palette

`palette.go`: `Command{ID, Title, Group, Icon, Keys, Hint, Disabled, Reason,
Run}`. `Match` scores an ordered subsequence, ignoring case: word starts,
runs and a match at the start count more, and shorter texts win among
equals. `Filter` ranks by title and hint (a hint match counts 4 less) and,
with an empty query, keeps the commands in order, grouped by first-seen
group. Up, Down, Enter and Escape work in the field. Opening the palette
(`Show`) and changing the query select the first command and scroll the
list back to it, so the command Enter runs is the one shown highlighted,
never one scrolled off by an earlier opening. A chosen command runs
after the palette closes, before the page is built (`RunPending`), so a
command that changes the page or asks a question shows in the same frame.

### Job queue

`jobs.go`: `Jobs` runs `Job`s one at a time, in order, so two changes never
race. `Run` works on a goroutine with a cancelable context and a `Progress`
(`Set(fraction, status)`, coalesced so that at most one post waits;
`Log(level, …)` to the console under the job's `Source`); `Done` runs on the
UI thread with `Run`'s error. The queue counts batches (`Batch`: "Job 3 of
10"; a batch is every job queued since the queue was last idle), lists the
batch's jobs in queue order (`BatchJobs`), finds the job of an item by
`Source` and `Key`, and calls `OnIdle` with the batch when the last job
ends. `Find` returns the newest job of the pair that waits or runs, in
`Enqueue` order (the newest queued one, else the running one), else the
newest that ended in the current batch; `Pending` reports whether any job
of the pair waits or runs, so a newer job that `CancelAll` dropped does not
hide an `Unstoppable` one of the same pair that still runs. It recovers a
panic in `Run` as the job's error and logs each job's start ("» title"),
success ("✓ title"), cancel and failure to the console. `CancelAll` drops
the queued jobs (their `Done` gets `context.Canceled`) and cancels the
running one, except an `Unstoppable` job, which finishes and reports its
own result while the queue stays busy (`Stopping`, `CanStop`). A job that
finished despite a cancel keeps its own result. `Duration` formats a job's
time ("1m 03s") and `RunSummary` how a batch ended ("9 done · 1 failed").

### Console

`console.go`, `consoleview.go`: `Console` is a thread-safe ring buffer of
lines with levels (debug, output, info, command, dry run, warn, error),
counts of warnings and errors since the last `Clear`, and a notify function
called after each `Add` from the adding goroutine. `Lines(levels, query)`
filters by level and by text or source; `Text` renders lines for the
clipboard. `Console.Writer` adapts it as a `log` output, with an optional
`Classify` that picks the level of each line: a write of several lines is
split first, so its lines can differ in level and count as warnings or
errors one by one. The writer keeps no partial line between writes (a
write without a final newline is a line of its own) and does not strip
the standard logger's timestamp. `ConsoleView` is the pane: filters by level
(`DefaultLevels` hides debug) and text, follow, copy, clear and close; its
`Subtitle` follows the title ("every command the app runs, as it runs"
when empty).

### Dialogs

`dialogs.go`: `Dialogs.Ask(Confirm{…})` queues confirmations and shows them
one at a time (a destructive action in the danger color, a `Note` with a
password icon for administrator authentication). `Info` shows content
dialogs; `ShortcutsBody` and `AboutBody` build the usual two. `CodeDialog`
shows text exactly as it is (a report before it is sent) in a scrollable
monospaced box with Cancel and an action, and holds no state of its own.

### Building blocks, rows and visualizations

- `components.go`, `text.go`: `Card`, `CardHeader`, `PageHeader`, `Section`
  and `SectionHeader` (`SectionHead`: icon, title, note, aside, `Quiet`),
  `Figures`, `Pill`, `IconPill`, `ColorPill`, `DotPill`, `SquareAvatar` and
  `Initials`, `IconTile`, `IconButton`, `LabeledButton`, `Banner`, `Empty`,
  `Loading`, `Problem`, `KeyValues` (with `KeyValuesStyle`), `PillTabs`,
  `SoftSelection`, `RowDivider`, `Kbd`/`Shortcut`, and the text styles
  `Title`, `Heading`, `Strong`, `Body`, `Muted`, `Caption`, `Mono`, `Toned`.
- `rows.go`: `JobStatus(c, RowStatus{…})` draws a row's state at its end (a
  spinner while queued or running, a check, an error, or a note; `Compact`
  for dense rows), `VersionChange` ("1.2 → 1.3"), `CommandRow` (icon tile,
  title with an optional "admin" pill, the command line in monospace, a
  132 DIP status cell and a 72 DIP action cell), `StepTile`, `SwitchRow`
  (returns its switch), `DisclosureCard` (opens in the frame of the click)
  and `TileGrid`.
- `wizard.go`, `toasts.go`: `Hero`, `StepDots` (a progress for screen
  readers, "2 / 5") and `TopToasts` (at the top of the window, each with a
  close button named by the app).
- `viz.go`, `pipeline.go`: `LiveChart` (0–100 series, smoothed), `Legend`,
  `Gauge`, `SegmentBar`, `ToggleTile`, `Pipeline` (stages of steps with
  states; its list is named "Pipeline", and an app chains `Label` to name
  its run), `JobCard` (the rail's running job with a Stop button) and
  `StateOf` (a job's state as a step's).

### Theming

The look follows the desktop, not a palette of its own: MyGo reads the
accent, light or dark mode, contrast and text size from the system, and the
kit adds metrics (`SpaceXS`…`SpaceXL`, `RadiusCard`, `RadiusChip`,
`RailWidth`, `TopBarHeight`, `ToastWidth`, `HeroSize`), tones (`Neutral`,
`Accent`, `Success`, `Warning`, `Danger`) mapped onto the theme by
`ToneColor` and `ToneFill`, and two derived surfaces (`Chrome` for the bars
and rail, `Panel` for cards). `Adapt` keeps the accent in windows without
the focus, as GNOME and KDE do (MyGo fades it to gray, as macOS does), on
every platform. MyGo resets the theme at each build pass, so `Adapt` runs
first in every frame; `Shell.View` calls it.

### Icons

`icons.go`: `Icons` finds icons by freedesktop name, lazily, and caches them
(a missing one is cached as nil). The app's embedded icons, by name, win;
`Get` prefers the `-symbolic` form. On Linux, `NewIcons` then searches the
desktop's theme (`DesktopIconTheme`: `FROSTUI_ICON_THEME`, else KDE's
`kdeglobals`, else `gsettings`, else Adwaita), then Adwaita, Breeze and
hicolor, under `~/.local/share/icons`, `~/.icons`, `$XDG_DATA_DIRS/icons`
and `/usr/share/icons`, in the layouts of Adwaita, Breeze and hicolor. Off
Linux it searches nothing. `SetIcons` makes an `Icons` the kit's; `Icon`
draws a missing icon as an empty box of the text size, so layouts do not
move.

The kit itself draws these names, which an app off Linux embeds to show
them: `changes-prevent`, `content-loading`, `dialog-error`,
`dialog-password`, `dialog-warning`, `edit-clear-all`, `edit-copy`,
`edit-find`, `go-bottom`, `go-next`, `media-skip-forward`, `object-select`,
`open-menu`, `pan-down`, `pan-up`, `process-stop`,
`software-update-available`, `system-search`, `utilities-terminal`,
`window-close`.

### Platforms

Four functions have a Linux part and a fallback
([ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)).
Each exported function sits in an untagged file and calls a lower-case
function from `*_linux.go` or `*_other.go` (`//go:build !linux`):

| Function | Linux | Elsewhere |
| --- | --- | --- |
| `SetAppID` (`appid*.go`) | `g_set_prgname` and `g_set_application_name` through purego, so the Wayland app ID and X11 `WM_CLASS` match the `.desktop` file | loads nothing; rejects an empty ID |
| `DesktopIconTheme` (`icons*.go`) | as above | "" |
| `NewIcons` (`icons*.go`) | embedded icons, then the theme directories | embedded icons only |
| `LaunchDesktopApp` (`launch*.go`) | `gtk-launch <id>`, not waited for | an error wrapping `errors.ErrUnsupported` |

Only the first `SetAppID` call counts, on every system.

### Rules

- **Imports.** The kit imports the standard library, MyGo, purego and its
  own packages only, and `frostuitest` the standard library only
  (`internal/repocheck.TestKitImportsOnlyMyGoAndPurego`, over every file,
  tests included). go.mod requires MyGo and purego directly and nothing
  else (`TestModuleRequiresOnlyMyGoAndPurego`).
- **No app code.** A helper that is about one app stays in that app; one
  that any app could use (a pill, a row, a status, a dialog) belongs here.
- **Threads.** Views and every model belong to the UI thread, except
  `Console.Add`/`Addf`/`Writer` and `Progress.Set`/`Log`, which any
  goroutine may call. Apps hand work back with MyGo's `Window.Update`. It
  rebuilds the view on every call, so wake the window from a burst of
  console lines through a coalescing waker (the example's `coalesce`).
- **State outside the view.** `ListState`, `ScrollState`, dialog and
  palette state and parsed SVGs live in structs (`Shell` holds the
  palette, console view, dialogs and info dialog), never in a view
  function. A `ListState` belongs to one window: two Testers drawing one
  `Shell` panic ("a ListState shown by two Lists").
- **Lists need a height.** A `ui.List` or `ui.Table` inside an unsized
  column gets almost none; size it or its column. Rows that hold controls
  must not set `ListState.Label` (the Tester then targets the row, not the
  control).
- **Clickable custom rows are buttons**: `ui.ButtonBase(c)`, for the
  keyboard and screen readers. Key a kit container from the caller with
  `c.Key` (`DisclosureCard(c.Key(id), …)`).
- **Platforms.** A freedesktop call goes in a `_linux.go` file with a
  fallback in an `_other.go` file; a platform file exports nothing
  (`TestPlatformFilesExportNothing`).

## Operational notes

- **Testing.** Every part has headless tests with `ui.NewTester`, which
  renders on the CPU and needs Pango and cairo but no display and no GTK.
  `frostui_test.go` covers palette ranking and running, the console, the
  job queue's order, serialization, cancellation and panic recovery,
  dialogs and the shell; `components_test.go` and `widgets_test.go` the
  pieces; `*_linux_test.go` the freedesktop parts (`LaunchDesktopApp`
  against a fake `gtk-launch` on `PATH`); `*_other_test.go` the fallbacks,
  which run under `make test-wasm`. `frostuitest.Loop` stands in for the UI
  thread: workers `Post` from any goroutine and the test runs what they
  posted with `Flush`, `RunUntil(t, cond)` or `Settle(quiet)`.
- **Pictures.** `Tester.Image()` gives a frame as an image.
  `FROSTUI_SNAPSHOT_DIR=<dir>` makes the example's snapshot test write
  `commandcenter.png` and `commandcenter-dark.png` (`make snapshot` writes
  them to `build/snapshots`). To prove a refactor changes nothing on screen,
  render before and after and compare the PNGs: spinners and wall-clock
  times differ between runs, nothing else.
- **MyGo.** After a MyGo bump, read its source in `GOMODCACHE` for the
  theme reset per frame, `Window.Update` semantics and the Tester before
  merging. MyGo loads GTK only in `App.Run`; never load a UI library in an
  `init` function or a package-level variable, or every package that
  imports the kit becomes untestable.

## Moved from ChairLift

The module is chairlift's `internal/frostui` at 09e7184. The extraction
changed no exported identifier or signature: the API extractor run over
chairlift's copy gives the same 350 lines as `api.txt`. ChairLift consumes
it by:

1. `go get github.com/frostyard/frostui@<version>`, then replacing
   `github.com/frostyard/chairlift/internal/frostui` with
   `github.com/frostyard/frostui` (and `…/frostuitest`) in every import.
2. Deleting `internal/frostui` and, in `internal/installcheck`:
   - deleting `TestFrostUIImportsNoChairLift` and its fixtures, which walk
     the deleted directory; this repository holds the guard now;
   - renaming `github.com/frostyard/chairlift/internal/frostui` to
     `github.com/frostyard/frostui` in `firstrun_test.go`, where
     `TestFirstRunGateImportsNoGTK` lists it as a UI package the first-run
     gate must not reach and the window must reach; with the old path the
     guard no longer sees the kit;
   - dropping the `internal/frostui/` exemption of
     `TestEveryCommandIsTraced` (`cmdlog_test.go`), which then matches
     nothing.
3. Chaining `.Label("Update run")` on `frostui.Pipeline` to keep its old
   accessible name.
4. Updating the docs that describe `internal/frostui` (`AGENTS.md`,
   `README.md`, `docs/design/frostui.md`, ADR-0017's references) to point
   at this module.

What changed in the move: the four freedesktop functions moved their
platform code into `_linux.go` and `_other.go` files; the comments lost
ChairLift's examples (Homebrew taps, firmware, bootc deployments); and
`Pipeline`'s list is named "Pipeline" instead of "Update run".

## API changes

None since the module began.

## References

- Rationale: [ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md), [ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)
- Contracts: [spec: API stability](../specs/api-stability.md)
- Built in: [roadmap — Phase 1](../plans/roadmap.md#phase-1--extract-the-kit-done)
