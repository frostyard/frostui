# Spec: The workspaceui module

This contract governs `github.com/frostyard/frostui/workspaceui`, the kit
that frostyard's NSGo apps build their windows from
([ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)): the
module's boundary, its versions, its themes, and the measures its parts
draw with today. Apps consume it through their go.mod and the exported
parts; [the design](../design/workspaceui.md) shows where it fits. The rest
of the look that ADR-0014 decides is not part of this contract until
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

### The themes

`Theme()` returns one NSGo theme with two palettes: Frostyard for the dark
appearance and Frostyard light for the light one. Every color is a value of
[the themes' spec](frostyard-theme.md); this table says which.

| NSGo role | The themes' spec |
| --- | --- |
| `background`, `surface`, `border`, `text`, `textMuted` | `Background`, `Surface`, `Border`, `Text`, `TextMuted` |
| `surfaceHover`, `surfacePressed` | `Text` at 6% and at 12%, as an overlay: over `Surface` they are the spec's `SurfaceHover` and `SurfacePressed`, within 1 step of 255 |
| `accent`, `accentHover`, `accentPressed`, `accentText` | `Accent`, `AccentHover`, `AccentPressed`, `AccentText` |
| `focus`, `accentFill`, `info` | `Accent` |
| `success`, `warning`, `danger`, `down` | `Success`, `Warning`, `Danger`, `Danger` |
| `selection`, `scrollbar` | `Selection`, `Scrollbar` |
| `inset` (the shell's ground) | `Chrome` |
| `raised` (cards, menus, dialogs) | `Panel` |
| `field`, `fieldBorder` | `Code`, `Line(Neutral)` |
| `line` | `Border` |
| `faint` | `Label` |
| `light`, `onLight` (a tooltip) | `Text`, `Background` |
| `track`, `knob` | `Text` at 25%; `Text` in Frostyard, white in Frostyard light |
| `code` (the color of code's text) | `Text` |

shadcn/ui's tokens, which a class can name, take the same colors (`card`
and `popover` are `Panel`, `primary` is `Accent`, `muted` is `Surface`,
`sidebar` is `Chrome`); the `chart-` colors stay NSGo's.

`ColorsOf(cx)` reads the colors NSGo has no role for, in the appearance
shown:

| Method | Palette name | The themes' spec |
| --- | --- | --- |
| `Chrome()`, `Panel()` | `frost-chrome`, `frost-panel` | `Chrome`, `Panel` |
| `Selected()` | `frost-selected` | `SelectedFill` |
| `SelectedText()` | `frost-selected-text` | `SelectedText` |
| `Label()` | `frost-label` | `Label` |
| `Fill(tone)` | `frost-fill-neutral`, `-accent`, `-success`, `-warning`, `-danger` | `Fill(tone)` |
| `Line(tone)` | `frost-line-` and the same 5 | `Line(tone)` |

| Item | Value |
| --- | --- |
| `Theme().Radius` | NSGo's default (8): the kit keeps NSGo's corners (ADR-0014) |
| `FontFamily`, `Theme().Sans` | `"Inter Variable", Inter, sans-serif`: NSGo finds a registered font by the family its file names, and the variable file names "Inter Variable" |
| `Apply(cx)` | sets `Theme()` on `cx` and registers Inter 4.001's upright variable file (`fonts/InterVariable.ttf`, embedded); the theme is set even when the font is refused |
| `Frame(children...)` | a view that fills the window, paints `background`, and sets the text's color and `FontFamily`, which its texts inherit |
| `AppearanceChoice` | a segmented `ui.Tabs` of System, Light and Dark; a press calls `SetAppearance`, then `OnChange` |

Not carried over from the MyGo kit: the high-contrast values of the
themes' spec, the gradient colors, the italic file, and the text styles
(the kit uses NSGo's `ui.Type` styles).

### `Shell`

| Measure | Value |
| --- | --- |
| Navigation width | 246; a saved width is held between 180 and 420 |
| Inspector width | 300; between 240 and 420 |
| Auxiliary panel width | 510; between 340 and 900 |
| Compact layout | a window narrower than 720: one pane, with a Navigation button, on a ground of `background` |
| Docked inspector | a window of 1000 or wider; narrower, the inspector takes the content's place |
| Both side panels | a window of 1550 or wider; narrower, opening one closes the other |
| `OnWidthsChange` | called 200 ms after the last resize, and once more when the shell goes away with a save pending |
| Title bar | on Linux, in a desktop window, the system's (`desktop.TitleBarDefault`) |
| `Brand` | at the head of the navigation, except where the system's title bar names the app (Linux, a desktop window) |

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
- A part takes every color from the current theme (`ui.ColorsOf`,
  `ColorsOf`): the module's only literal colors are the two palettes of
  `theme.go`.
- The palettes hold the values of the themes' spec
  (`TestThemeHoldsTheSpecsColors`).
- No role that NSGo's default theme names is left at NSGo's color, the
  `chart-` colors aside, so a role NSGo adds later fails the test until it
  has a Frostyard color (`TestThemeSetsEveryRoleNSGoNames`).
- In both themes these pairs hold 4.5:1 (WCAG 2), 69 pairs each: on each of
  `background`, `inset`, `raised`, `surface` and `field`, the colors
  `text`, `textMuted`, `faint`, `accent`, `success`, `warning`, `danger`
  and the selected text, and the text on a hovered, a pressed and a chosen
  row; the accent's text on its 3 faces; a tooltip's text; and each tone on
  its fill over the panel (`TestThemeContrast`). The lowest is Frostyard
  light's accent on `surface`, 4.53:1.
- `ColorsOf` and NSGo's controls follow the appearance shown, and a text in
  a `Frame` has the kit's family (`TestColorsOfFollowsTheAppearance`,
  `TestAppearanceChoiceShowsAndReportsTheChoice`).
- A saved navigation width outside its range is drawn at the range's end
  (`TestShellKeepsItsNavigationWidth`).
- A compact shell paints the theme's ground and text color in both
  appearances (`TestCompactShellPaintsItsGround`). Before
  `workspaceui/v0.2.0` it painted none, and a dark appearance showed its
  light text on the window's white.
- A compact shell shows its navigation only after its button is pressed,
  and closes it when a destination is chosen
  (`TestCompactShellClosesNavigationAfterAChoice`).
- While the module is v0, a minor version may change its API. A change of
  a measure in this spec is made with the code and the test that hold it.

No test holds the other measures of the tables yet; the plan's Phase 2
adds them with the parts it builds. No headless test shows a pixel or a
loaded font: the themes, the bundled Inter and the compact shell's ground
were looked at through FrostMail's picture tours, in both appearances, in
a GTK window on Xvfb's X11, whose own sans-serif face is not Inter
(2026-11-24; the pictures are not kept in this repository).

## References

- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- Context: [design/workspaceui.md](../design/workspaceui.md)
- The colors the plan applies: [the Frostyard themes](frostyard-theme.md)
