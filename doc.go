// Package frostui is frostyard's desktop UI kit on MyGo
// (github.com/egoist/mygo/ui): the command-center shell (top bar, left
// rail, status bar, console, command palette), a frame without the rail
// for apps made of panes (PaneShell, Panes), the cards, figures, pills,
// banners and dialogs that every page is built from, a serial job queue
// whose output streams to the console, a command table behind the keys,
// menus and palette, a source list, a settings layout, an allow-listed
// opener, and desktop integration (the desktop's icon theme with an
// embedded set behind it, its accent kept in inactive windows, and the
// Wayland app ID).
//
// It imports only the standard library, MyGo and purego, so any frostyard
// desktop app can take it, and it builds with CGO_ENABLED=0 for Linux,
// Windows, macOS and js/wasm. The freedesktop parts (SetAppID,
// DesktopIconTheme, the icon theme of NewIcons, LaunchDesktopApp) work on
// Linux and fall back elsewhere, as each one's comment says. The
// repository's docs/design/frostui.md describes the parts and the rules.
//
// # Threads
//
// Views run on the UI thread. Every model here (Console, Jobs, Palette,
// Dialogs, CommandTable, SourceList, Settings) is changed on the UI
// thread, except Console.Add, Waker.Wake and an Opener, which any
// goroutine may use. Workers hand results back with the post function the
// app gives NewJobs (MyGo's Window.Update), as a palette source's
// PaletteQuery.Reply; in tests, a frostuitest.Loop stands in for it.
package frostui
