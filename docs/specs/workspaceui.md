# Spec: The workspaceui module

This contract governs `github.com/frostyard/frostui/workspaceui`, the kit
that frostyard's NSGo apps build their windows from
([ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)): the
module's boundary, its versions, and the measures its parts draw with
today. Apps consume it through their go.mod and the exported parts;
[the design](../design/workspaceui.md) shows where it fits. The look that
ADR-0014 decides is not part of this contract until
[the plan](../plans/workspaceui.md) builds it.

## Interface

### The module

| Item | Value |
| --- | --- |
| Module path | `github.com/frostyard/frostui/workspaceui` |
| Package | `workspaceui` |
| Direct requirement | `github.com/nsgo-dev/nsgo` v0.3.0, and no other |
| Tags | `workspaceui/vMAJOR.MINOR.PATCH` |
| Targets | linux, windows and darwin on amd64 and arm64, and js/wasm, with `CGO_ENABLED=0` |

### `Shell`

| Measure | Value |
| --- | --- |
| Navigation width | 246; a saved width is held between 180 and 420 |
| Inspector width | 300; between 240 and 420 |
| Auxiliary panel width | 510; between 340 and 900 |
| Compact layout | a window narrower than 720: one pane, with a Navigation button |
| Docked inspector | a window of 1000 or wider; narrower, the inspector takes the content's place |
| Both side panels | a window of 1550 or wider; narrower, opening one closes the other |
| `OnWidthsChange` | called 200 ms after the last resize, and once more when the shell goes away with a save pending |
| Title bar | on Linux, in a desktop window, the system's (`desktop.TitleBarDefault`) |

Accessible names the shell sets: "Navigation", "Resize navigation",
"Resize right panel", and "Resize " with the auxiliary panel's title.

### The other parts

| Part | Measures |
| --- | --- |
| `Page` | a scrolling column at most 920 wide, centred, padding 20, gap 14 |
| `Row`, `Stack` | gap 8; `Row` wraps and centres on the cross axis |
| `Card` | padding 16, gap 8, corner 12, a 1-point border |
| `Section` | a disclosure with its count as a badge; its content is built only while open |
| `BranchLabel`, `ChildLabel` | text 13; a child below the first level indents `8 + 20 × Depth` beside a guide 32 high |
| `Notice` | 380 wide (the window's width when narrower), 56 under the window's top; its icon and tone by `Kind` |
| `FileExplorer` | compact under 620; the tree between 110 and 440, the preview at least 170 |
| `Conversation` | content 810 wide |

## Rules

- Every file of the module, tests included, imports only the standard
  library and NSGo's packages (`TestKitImportsOnlyNSGo`).
- go.mod requires NSGo directly and nothing else
  (`TestModuleRequiresOnlyNSGo`).
- The module builds and vets for every target above (`make cross`).
- A part takes every color from the current theme (`ui.ColorsOf`): the
  module holds no literal color.
- A saved navigation width outside its range is drawn at the range's end
  (`TestShellKeepsItsNavigationWidth`).
- A compact shell shows its navigation only after its button is pressed,
  and closes it when a destination is chosen
  (`TestCompactShellClosesNavigationAfterAChoice`).
- While the module is v0, a minor version may change its API. A change of
  a measure in this spec is made with the code and the test that hold it.

No test holds the other measures of the tables yet; the plan's Phase 2
adds them with the parts it builds.

## References

- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- Context: [design/workspaceui.md](../design/workspaceui.md)
- The colors the plan applies: [the Frostyard themes](frostyard-theme.md)
