# 0008 — Open behind allow-lists, and let editors keep their keys

- **Status:** Accepted
- **Date:** 2026-11-10

## Context

Frostyard's desktop apps show content that comes from outside them: links in
messages and notes, and attachments. Handing such a target to the system's
handler can start a program, install software, or follow a link to another
place. The apps share one kit, so the rule for what may be opened is written
once.

The same apps bind window shortcuts that their text editors also use: Ctrl K
opens the command palette, and in an editor it is Insert Link. MyGo gives an
editor's own key handler (`Element.OnTextKey`, `Element.HandleInput`) a key
only when no shortcut claims it, and a window shortcut claims it.
`Context.FocusTakesText` tells that an element takes text, but not whether it
is an editor or a one-line field.

## Decision

**The opener's policy.** `Opener` refuses anything not allowed, with an error
wrapping `ErrNotAllowed`:

- Without `Schemes`, it opens http, https and mailto URLs. The `file` scheme
  is refused even when listed, so files go through the file rules alone.
- It opens no file without `Roots`.
- It follows symbolic links before it checks a path, in the file and in each
  root, so a link under a root that leads out of it is refused.
- It refuses non-regular files, executable files, and a deny list of types
  that run, install or redirect. The deny list is a backstop that cannot be
  complete: an app sets `Extensions`, an allow-list, for files from outside
  it, as attachments.
- It accepts a window between the check and the open: a file swapped in that
  window is opened as swapped. The roots are the app's own directories.

**The key-scope rule.** Each `KeyBinding` has a `KeyScope`. A key that
`OutsideEditors` or `OutsideText` makes step aside is not registered with the
window while it would step aside, so the editor's own handler receives it.
`OutsideEditors` is the zero value, `OutsideText` also steps aside in a
one-line field, and `Everywhere` always works. `CommandTable.Editing` names
the editors; left nil, any element that takes text counts as one, so a key
never goes past a field. An element that binds a key itself comes first in
every scope.

## Consequences

- An app that opens attachments has one more field to set, and is safer for
  it. An app that leaves `Extensions` empty relies on the deny list.
- A new dangerous file type needs a line in the deny list, and its test.
- A swap inside a root between the check and the open is not caught.
- Filesystems that mark every file executable refuse every file.
- An app with editors writes one function to name them. Until it does, a key
  of the zero scope also steps aside in its one-line fields.

## Alternatives considered

- **A deny list only:** it cannot name every type that runs on three
  systems, so it stays a backstop to `Extensions`.
- **Opening a file descriptor to close the check-then-open window:** the
  system's handlers take a path or a URL, not a descriptor.
- **Binding every key and having the editor take it back:** MyGo gives the key
  to the shortcut before the editor's handler, and has no way back.
- **`FocusTakesText` alone as the rule:** it would take Ctrl K from one-line
  fields in apps that want it to open the palette there.

## References

- Shapes: [design: frostui](../design/frostui.md#opener), [command table](../design/frostui.md#command-table)
- Builds on: [ADR-0004](0004-extract-the-ui-kit-into-its-own-module.md)
