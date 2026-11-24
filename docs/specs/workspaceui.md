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

### The parts of a list and a pane

The owner accepted these as the gallery draws them (2026-11-24,
[plan](../plans/workspaceui.md), Phase 2).

| Part | What it draws |
| --- | --- |
| `NavLabel` | the content of a row of NSGo's sidebar: an icon 16 in the accent, the name at 13 on 1 line, and `Count` at the row's end at 12 in figures of one width. A count of 0 is not shown; an `Unread` count is semibold in the text's color, any other is muted |
| `SummaryRow` | padding 8, gap 8: a gutter 8 wide whose dot (8, the accent) shows when `Unread`; NSGo's avatar at 32; then the title at 13 semibold with `When` at 12 muted at the row's end, the subject at 13 (semibold when `Unread`) with `Marks` at its end, and `Preview` at 12 muted on up to `PreviewLines` lines of 16 |
| `SoftRows` | the style of a list's rows: corners 8, the theme's hover fill, and `frost-selected` when chosen, with no change of the text's color |
| `Banner` | role status; padding 8 by 12, corners 8, a 1-point line: the tone's `Fill` and `Line`, an icon 16 in the tone's color (muted for `Neutral`), the sentence, and a small secondary button when `Action` is set |
| `EmptyState` | centred in its pane: an icon 32 in `faint`, the title in NSGo's title style, the hint as a muted caption |
| `DialogPanel` | the ground of a panel that floats over the window (a dialog, an alert, the command palette, a popover, a toast): `frost-panel`, opaque |
| `ConnectionItem` | a dot 8 in the tone's color and 1 line of muted caption |
| `SearchField` | NSGo's control height, corners 8, the field's ground and line, the accent's line while focused: a search icon 14, the platform's text field at 13, and at its end the keys of `Keys` while empty, or a "Clear search" button while it holds text |

### The control map

A need takes NSGo's control for it, used as NSGo's docs and gallery show;
the kit adds a part only where NSGo's `ui` package has none. The gallery
draws each row of this table in both themes.

| Need | The control | In the gallery |
| --- | --- | --- |
| The window and its navigation | `Shell` (NSGo's `ui.SidebarShell`), `ui.Sidebar` with `NavLabel` as an item's content | the standard window |
| A choice among a few views | `ui.Tabs`, segmented; with an icon alone and a `ui.Tooltip` where there is no room for words | buttons; the window's modules and filter |
| A toolbar | `ui.Toolbar`, `ui.ToolbarGroup`, `ui.ToolbarSeparator`, `ui.IconButton` (its label is its tooltip), `ui.ToggleButton` for a panel's toggle | buttons; the window |
| The one main action of a region | `ui.Button`, primary; every other is secondary, outline or ghost | buttons; the window's Compose |
| The palette's field | `ui.SidebarButton`, its field variant, with its shortcut | the window |
| A list and its selection | `ui.List` with a multiple `ui.Selection` and `SoftRows`; `SummaryRow` as a row's content | rows; the window |
| A person's picture | `ui.Avatar` | rows |
| A count or a state | `ui.Badge` | rows |
| Loading | `ui.Skeleton` | rows |
| A confirmation | `ui.AlertDialog` with `DialogPanel`: Cancel secondary and first, a destroying action in the danger variant | dialogs |
| The command palette | `ui.CommandDialog` with `DialogPanel`'s color | search, trees and popovers |
| Detail beside the content | `ui.Popover` and `ui.HoverCard` (a sender's card), the popover with `DialogPanel`'s color | search, trees and popovers |
| What was done, with a way back | `ui.Toaster` and `ui.Toasts` with `DialogPanel`; a toast's action is Undo | search, trees and popovers |
| A tree of folders or notes | `ui.Tree` | search, trees and popovers |
| A disclosure | `ui.Collapsible` | search, trees and popovers |
| Paging | `ui.Pagination` | search, trees and popovers |
| A settings view | `ui.Settings`: categories, groups of rows, its own search | settings |
| A form in a dialog | `ui.Dialog` with `DialogPanel`; Cancel secondary, the confirming button primary | dialogs |
| A menu | `ui.Menu` with icons, shortcuts, `ui.MenuSeparator` and the danger variant | dialogs |
| Settings rows | `ui.FormSection` and `ui.FormRow` | fields |
| An on/off setting | `ui.Switch`; `ui.Checkbox` for an item of a list and a task's done mark | fields |
| A choice from a list | `ui.Select` | fields |
| A date and a time | `ui.DatePicker`, `ui.TimeField` | fields |
| Recipients | `ui.TagInput` | fields |
| A notice about a pane's content | `Banner` (NSGo has none) | dialogs; the window |
| A pane with nothing in it | `EmptyState` (NSGo has none) | rows; the window |

| A search field | `SearchField` (NSGo has none) | search, trees and popovers; the window |

The owner accepted every row as the gallery draws it (2026-11-24), in 3
rounds. The search field's clear button was added after he accepted the
field: it is in the gallery's pictures, and he has not looked at it.

What NSGo draws that the theme cannot reach: a menu's ground is NSGo's
translucent one, since `ui.Menu` takes no style; and the side of
`ui.Settings` is a neutral gray, not `inset`.

### The standard for an app

What an app on the kit keeps to, so that the apps look and behave alike
(the owner's choices of 2026-11-24; frostyard/frostmail's plan 0016 orders
the work in each app).

1. **Theme.** `Apply` at the root and the views in a `Frame`. A choice of
   System, Light or Dark in the app's settings, kept on the device.
2. **No literal looks in app code.** No color, text size, corner or padding
   number outside the kit: an app takes a kit part, an NSGo control or a
   token, and a test in the app holds it.
3. **Buttons.** One filled button for each region of a window. Toolbar
   actions are icon buttons, whose labels are their tooltips. In a dialog,
   Cancel is secondary and comes before the confirming button; a
   destroying action is the danger variant.
4. **Dialogs.** A confirmation is `ui.AlertDialog`, and a form is
   `ui.Dialog`, both on `DialogPanel`. Settings is `ui.Settings`.
5. **Notices.** A notice takes its real tone. An action that can be undone
   answers with a toast whose action is Undo. What a pane has to say about
   its content is a `Banner`.
6. **States.** Every list and pane has an empty state (`EmptyState`), a
   loading state (`ui.Skeleton`) and an error state.
7. **Keys.** mod+K the palette (mod+Shift+P inside an editor), mod+, the
   settings, mod+N the app's new item, mod+F the search of what is shown.
8. **Window.** The size, the place and the panes' widths are kept, held so
   that the main pane never falls under its least width.
9. **Names.** Labels and test ids are the tests' handles: a change of the
   look keeps them, or changes the test in the same change.
10. **Live tests.** One suite on NSGo's `nsgotest`, on fixture data, that
    runs headless in the app's verify gate and in a real Linux window, and
    names a picture at each main view; the repository keeps the pictures,
    and a change of the look changes them in the same pull request.
11. **The right control.** The control map above. No view stands in for a
    control, and no control does another's job.

Not decided here: where the parts go in a window and how one moves between
them. That is each app's.

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

- `NavLabel`'s count ends where its row ends, for a short name and for one
  longer than the row, and a count of 0 is not shown
  (`TestNavLabelKeepsItsCountAtTheRowsEnd`,
  `TestNavLabelShowsNoCountOfZero`).
- `SummaryRow` keeps `When` and its marks whole and inside a row narrower
  than its text, takes the box's width, and limits the preview to
  `PreviewLines` (`TestSummaryRowKeepsItsTextInside`); its dot is the
  accent when unread and unpainted when read, in both themes, and a row
  with `PreviewLines` 0 has no preview
  (`TestSummaryRowMarksUnreadAndFollowsPreviewLines`).
- A chosen row's fill is `frost-selected` and its text's color is not set
  (`TestSoftRowsFillAChosenRowSoftly`).
- A `Banner` has its tone's fill and line, its button calls `OnAction`,
  and it has no button without `Action`
  (`TestBannerShowsItsToneAndAtMostOneButton`).
- `DialogPanel` is the theme's panel with no transparency, in both themes
  (`TestDialogPanelIsOpaque`).
- `SearchField` shows its keys while empty and a clear button while it
  holds text, reports what is typed and submitted, clears on its button,
  and takes its box's width (`TestSearchFieldShowsItsKeysOrAClearButton`).
- In the gallery, the filter is found by `role:tab`, the toolbar by
  `role:toolbar` and the confirmation by `role:alertdialog`, so a control
  of another kind fails its tour (`workspaceui/gallery/gallery_test.go`).

No test holds the other measures of the tables yet (the type sizes, the
paddings, the Shell's other widths). No headless test shows a pixel or a
loaded font: the themes, the bundled Inter and the compact shell's ground
were looked at through FrostMail's picture tours, in both appearances, in
a GTK window on Xvfb's X11, whose own sans-serif face is not Inter
(2026-11-24; the pictures are not kept in this repository).

## References

- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- Context: [design/workspaceui.md](../design/workspaceui.md)
- The colors the plan applies: [the Frostyard themes](frostyard-theme.md)
