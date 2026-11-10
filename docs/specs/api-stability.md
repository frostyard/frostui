# Spec: API stability and versions

This contract governs what an app that imports `github.com/frostyard/frostui`
can rely on between versions: the exported API of the packages `frostui` and
`frostuitest`, and the version tag that a change to it requires. Its
consumers are the apps (ChairLift, Frostmail, Frostnotes), the maintainer who
tags releases, and agents that change the kit.

## Interface

The exported API is the list in `internal/repocheck/testdata/api.txt`: one
line per exported declaration of the non-test files of both packages, sorted,
in these forms (parameter names are left out):

| Line | Declares |
| --- | --- |
| `func frostui.Card(*ui.Context) ui.Element` | a function |
| `method (*Jobs) frostui.Enqueue(...*Job)` | a method of an exported type |
| `type frostui.Shell struct` | a struct type |
| `field frostui.Shell.Console ConsoleView` | an exported field |
| `embed frostui.T U` | an exported embedded field |
| `type frostui.Tone int` | another type, `=` for an alias |
| `type frostui.I interface`, `imethod frostui.I.M(int) error` | an interface and its methods |
| `const frostui.SpaceL = 16`, `const frostui.Accent Tone` | an untyped constant with its value, a typed one with its type |
| `var frostui.LevelNames [levelCount]string` | a variable with its type |

Versions are git tags `vMAJOR.MINOR.PATCH`. svu derives the next one from the
Conventional Commits since the last tag (`.svu.yml`: prefix `v`, `v0: true`,
`always: true`), and `make bump` creates and pushes it.

## Rules

- `TestExportedAPIIsUnchanged` (`internal/repocheck/api_test.go`) MUST pass:
  the list it computes from the source equals `api.txt`.
- A change that only adds lines to `api.txt` is a minor change: a `feat:`
  commit.
- A change that removes or changes a line breaks callers. It MUST be a
  `feat!:` commit or carry a `BREAKING CHANGE:` footer, it MUST bump the
  major version, or the minor version while the module is v0, and it MUST
  add an entry under "API changes" in [design/frostui.md](../design/frostui.md#api-changes)
  that names the old and the new form and what callers change.
- After a deliberate change, `api.txt` is rewritten with
  `go test ./internal/repocheck -run TestExportedAPIIsUnchanged -update`, in
  the same commit as the code.
- A file that builds only on some platforms (a `//go:build` line or a
  `_GOOS`/`_GOARCH` file name) MUST NOT declare an exported identifier, so
  the list is the same on every platform (`TestPlatformFilesExportNothing`).
- A change of behavior with the same signature (a default string, a size, a
  color) is not an API change; it is a `fix:` or `feat:` and is named in the
  commit message.

## Derived artifacts

| Artifact | Derivation |
| --- | --- |
| `internal/repocheck/testdata/api.txt` | `go test ./internal/repocheck -run TestExportedAPIIsUnchanged -update` |
| The release tag | `svu next` over the commits since the last tag (`make bump`) |

## References

- Rationale: [ADR-0004](../adr/0004-extract-the-ui-kit-into-its-own-module.md), [ADR-0005](../adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)
- Context: [design/frostui.md](../design/frostui.md)
