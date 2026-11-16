# 0012 — One tray component, with a close-to-tray check the app switches

- **Status:** Accepted
- **Date:** 2026-11-15

## Context

Frostmail and Frostnotes want an icon in the notification area or menu bar,
and ChairLift and the Pocket client can use one too. MyGo v0.9.1 gives
`mygo.NewTray` with a PNG icon, a menu, a tooltip, `OnClick` and
`OnRightClick`, but each app would still have to settle the same points:

- **Clicks differ by platform.** On Linux the icon is an AppIndicator
  (StatusNotifierItem and DBusMenu): the panel draws the menu, a click opens
  it, and `OnClick` is the panel's activation (GNOME's AppIndicator
  extension sends it for a double click). Windows sends a left click to
  `OnClick` and opens the menu on a right click. macOS opens the menu on a
  click while one is set.
- **A missing panel is silent.** GNOME without the AppIndicator extension has
  no `org.kde.StatusNotifierWatcher` on the session bus. libayatana then
  falls back to an XEmbed icon, which no Wayland session shows, and
  `mygo.NewTray` reports no error. A window that hides "to the tray" there is
  lost to the user.
- **A tray keeps nothing alive.** Close to tray is `PreventDefault` and `Hide`
  in the window's `OnClose`, and must step aside while the app quits.
- **No badge.** No platform puts a badge on a tray icon: a count or a dot is
  part of the image.
- **A hidden window costs memory:** it keeps GTK (and WebKit in Frostmail)
  loaded. frostyard's apps make close to tray an opt-in setting, off by
  default, so a close quits as it did before the tray. They start no hidden
  window at login: the tray shows while the app runs.

## Decision

The kit has **one tray component**, `Tray` (`tray.go`), made by
`NewTray(TrayOptions)` in `mygo.App.WhenReady`. It belongs to the UI thread.

- **One menu shape.** A disabled status line (`SetStatus`) and a separator
  when there is a status, Open (`TrayOptions.OpenLabel`, "Open" when empty),
  the app's `TrayItem`s (`SetItems`), a separator, and Quit (MyGo's
  `RoleQuit`, which runs `App.Quit` and takes the platform's label). The
  status line also follows the tooltip, since GNOME shows no tooltip.
- **The menu is always set,** on every platform, and `OnClick` shows and
  focuses `TrayOptions.Window`. So Linux and macOS open the menu on a click,
  the panel's activation and a Windows click open the window, and a Windows
  right click opens the menu. MyGo has no way to pop up a menu that is not
  set on the icon, so macOS keeps its own convention.
- **A dot, drawn in.** `SetDot` swaps the icon for a copy with a disc of the
  Frostyard dark theme's Danger in its top right corner, cut out of the icon
  by a clear ring. The copy is drawn once, in `NewTray`, which rejects an
  icon that is not a PNG. The dot is never the only cue: the status line
  says what it means.
- **The close check, switched by the app.** `CloseToTray(e, enabled)` runs in
  the app's own `OnClose`. It hides the window and prevents the close only
  when `enabled` (the app's setting), the event is for `TrayOptions.Window`,
  no quit runs (the tray follows `App.OnBeforeQuit` and `OnQuitCanceled`),
  the tray is not destroyed, and `TrayHosted` reports true. Else the close
  goes on as without a tray. The setting stays the app's, which offers it
  opt-in and off by default and passes it at each close; the kit does
  nothing at login.
- **The watcher check at each close.** `TrayHosted` asks the session bus
  `NameHasOwner("org.kde.StatusNotifierWatcher")` through GIO with purego
  (`tray_linux.go`, as `launch_linux.go` loads GIO), with a 0.5 second
  timeout, and reports false when the bus cannot be asked. It is asked at
  each close, not once, because the panel or its extension can go and come
  back while the app runs. Off Linux it reports true (`tray_other.go`).

## Consequences

- An app adds a tray in a few lines and gets the same menu, clicks, dot and
  close rule as the other apps; a fix lands once for all of them.
- The close rule fails safe: no panel, no setting, a quit, or a bus that does
  not answer, and the window closes as it did before the tray.
- A close on Linux with the setting on waits for 1 D-Bus round trip on the UI
  thread, at most 0.5 seconds when the bus does not answer.
- The menu shape is fixed: an app that needs submenus, check items or another
  order uses `mygo.NewTray` itself, or the kit grows the shape in a new
  additive change.
- On macOS a click opens the menu, not the window: Open is the menu's first
  item.
- The apps keep the parts that are theirs: the opt-in setting (off by
  default), `App.OnSecondInstance` showing a hidden window, and the Flatpak
  pieces (the AppIndicator libraries, `--talk-name=org.kde.StatusNotifierWatcher`,
  a `TMPDIR` the panel can read).
- `tray_linux.go` is native code and a risk path (L2).

## Alternatives considered

- **Each app calls `mygo.NewTray` itself:** 4 copies of the click rules, the
  watcher check and the close rule, which would drift; frostui exists to
  hold what every app needs once.
- **The kit owns the close setting:** where an app stores it and how it
  shows it are the app's, so the app passes its setting at each close.
- **Close to tray by default:** a hidden window keeps its toolkit in memory,
  and a close that does not quit surprises a user who never chose it.
- **A hidden window at login, so the tray shows before the app is opened:**
  the Background portal keeps 1 login command per app, so it changes each
  app's launcher, and a background daemon can already notify.
- **Check the watcher once, at `NewTray`:** misses a panel that goes or comes
  later, and a hidden window is then lost.
- **Leave the menu unset off Linux, so a click opens the window and a right
  click pops the menu up:** MyGo's `PopUpMenu` shows only the menu set on the
  icon, so it cannot be done without setting it.

## References

- Shapes: [design: tray](../design/frostui.md#tray),
  [design: platforms](../design/frostui.md#platforms),
  [spec: API stability](../specs/api-stability.md)
- Builds on: [ADR-0005](0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md),
  [ADR-0011](0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md)
