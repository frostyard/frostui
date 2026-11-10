# 0005 — Build everywhere, with the freedesktop parts behind build tags

- **Status:** Accepted; its icon fallback superseded by [0007](0007-embed-a-lucide-icon-set-as-the-last-fallback.md)
- **Date:** 2026-11-10

## Context

Frostmail ships on Windows as well as Linux. MyGo draws on GTK 3 on Linux
(amd64 and arm64), on Cocoa on macOS, on Win32 on Windows, and in a browser
under js/wasm, all without CGO. Four parts of the kit assume a freedesktop
Linux desktop:

- `SetAppID` loads GLib with purego and calls `g_set_prgname`.
  `purego.Dlopen` does not exist on Windows, and purego does not compile
  for js/wasm.
- `DesktopIconTheme` runs `gsettings` and reads KDE's `kdeglobals`.
- `NewIcons` searches the freedesktop icon directories
  (`~/.local/share/icons`, `$XDG_DATA_DIRS/icons`, `/usr/share/icons`).
- `LaunchDesktopApp` runs `gtk-launch` with a desktop file ID.

## Decision

- **Targets.** The module builds and vets with `CGO_ENABLED=0` for
  linux/amd64, linux/arm64, windows/amd64, windows/arm64, darwin/amd64,
  darwin/arm64 and js/wasm. js/wasm is a compile check for reuse on the
  web; it is not a supported runtime.
- **Split by file, not by API.** Each of the four exported functions stays
  in an untagged file with its one doc comment and calls a lower-case
  function. `*_linux.go` files hold the freedesktop code; `*_other.go`
  files (`//go:build !linux`) hold the fallbacks. A file that builds only
  on some platforms exports nothing, so the API is the same everywhere.
- **The fallbacks off Linux:**
  - `SetAppID` loads no library and only rejects an empty ID. The system
    names the app from its bundle, its executable or its page. Only the
    first call counts on every system.
  - `DesktopIconTheme` returns "", and `NewIcons` searches no directory:
    it finds the app's embedded icons alone, and an icon the app does not
    embed draws as an empty box of the text size, so layouts do not move.
  - `LaunchDesktopApp` starts nothing and returns an error that wraps
    `errors.ErrUnsupported`.
- **`linux`, not MyGo's GTK condition.** The freedesktop files build on
  every Linux architecture, also where MyGo has no GTK backend; a GLib
  program name there is harmless.
- **One look.** `Adapt`, which keeps the accent in windows without the
  focus as GNOME and KDE do, applies on every platform.
- **Guards.** `TestKitCrossCompiles` builds and vets the module for every
  target, `TestPlatformFilesExportNothing` keeps the API the same,
  `make lint` lints as linux, windows, darwin and js/wasm, `make
  test-wasm` runs the tests as js/wasm under Node, and CI's cross-build
  job runs one target per runner.

## Consequences

- Frostmail can import the kit on Windows and macOS. An app there embeds
  every icon it draws, the kit's own included, or shows empty boxes.
- The off-Linux fallbacks run in tests under js/wasm only; Windows and
  macOS are compiled, vetted and linted, not run, in CI.
- A new freedesktop call goes in a `_linux.go` file with a fallback in an
  `_other.go` file, or the cross-build fails.

## Alternatives considered

- **Linux only:** Frostmail could not use the kit on Windows.
- **Windows and macOS equivalents (AppUserModelID, the macOS bundle
  name, `ShellExecute`):** each needs platform code MyGo already owns or
  an identity the kit does not have; the apps can add them when they need
  them.
- **An icon set in the kit:** it sets a license and a look for every app;
  the apps embed what they draw instead.

## References

- Shapes: [design: frostui](../design/frostui.md), `AGENTS.md`
- Builds on: [ADR-0004](0004-extract-the-ui-kit-into-its-own-module.md)
