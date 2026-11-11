# Plan: The Frostyard look

This plan builds [ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md):
the kit draws the frostyard brand in a dark and a light theme, and
ChairLift, Frostmail and Frostnotes adopt it. Phases 1 and 2 change the
kit and end in a minor release; phases 3 to 5 move each app onto it and can
run in parallel once that release exists. It follows the
[roadmap](roadmap.md).

The values below come from the Frostyard Desktop design system, which
measured every text pair in both themes. Phase 1 copies them into a spec
with the code; until then this table is the source.

| Role | Frostyard | Frostyard light | Draws |
| --- | --- | --- | --- |
| background | `#06111d` | `#f5f9fc` | the window, lists, the reader |
| chrome | `#081b2d` | `#e8f0f6` | top bar, rail, pane bars, status bar, console, source list |
| panel | `#0d263b` | `#ffffff` | cards, dialogs, the palette, toasts |
| panel gradient (160°, to 62%) | `#0f3048` → `#091c2d` | `#ffffff` → `#eef5fa` | cards and figures |
| surface | `#0d2a41` | `#dfeaf2` | `ui.Button` faces, avatars |
| border | `rgba(162,222,255,0.18)` | `rgba(6,17,29,0.14)` | every hairline |
| text | `#edf8ff` | `#06111d` | body text |
| text muted | `#b6d0df` | `#3f5a6b` | secondary text, previews |
| accent | `#47b8ef` | `#0b6ea6` | links, list icons, the focus ring, the selected row's rule |
| on accent | `#02131f` | `#02131f` | text on the primary gradient |
| selected fill | sky at 8% | accent at 10% | the selected rail item, source-list row, chosen list row |
| selected text | `#aee9ff` | `#064a72` | text and icon of the selected row |
| hover text | `#aee9ff` | `#064a72` | hovered buttons, rows and links (faces do not change) |
| success | `#7edafa` | `#076377` | tone |
| warning | `#e0b15c` | `#8a5300` | tone |
| danger | `#e07a72` | `#b0322b` | tone |
| tone fills (success, warning, danger; accent) | `#0e3247`, `#33291433`, `#3a181e`; sky at 8% | `#ddf1f6`, `#f8ecd6`, `#f9e2df`; accent at 8% | banners, icon tiles |
| tone lines (accent and success, warning, danger; neutral) | `#47b8ef4d`, `#e0b15c55`, `#e07a7255`; `#a6dffa38` | each tone at 40%; ink at 22% | pill and badge outlines |
| label | `#82b6d0` | `#3b6680` | uppercase section titles |
| tag text | `#b9e9fb` | `#24485e` | neutral pill text |
| code | `#061826` | `#ffffff` | fields, the console |
| line strong | `#2c6380` | `#2c6380` | the toast's outline |

## Phase 1 — Themes, palette and text (medium)

- `FrostyardTheme(dark bool) *ui.Theme` returns the two themes: the colors
  above in MyGo's fields, `InactiveAccent` equal to `Accent`, `Radius` 4,
  `Font` "Inter". Values the design system lacks (`Selection`,
  `Scrollbar`, `Inverse` for tooltips) are chosen here, measured, and added
  to the design system.
- The kit palette holds the roles `ui.Theme` has no field for. `Chrome(t)`
  and `Panel(t)` keep their signatures and return the palette's values; new
  accessors return the rest (see the open question on how they find it).
  `ToneColor`, `ToneFill` and a new `ToneLine` read the palette.
- `Adapt` takes Frostyard or Frostyard light from the theme MyGo picked
  (`Dark`), applies `WithPreferences` with the accent cleared, so high
  contrast and text size still apply, and sets it. Apps with their own
  appearance setting (Frostnotes' light, dark or system; ChairLift's
  first-run preview) call `FrostyardTheme` and apply the preferences the
  same way.
- Inter is embedded under `fonts/` with `OFL.txt`, listed in `NOTICE`, and
  registered with `ui.RegisterFont` on the first `Adapt`, on the UI thread,
  never at package init: a program that never opens a window (ChairLift's
  `--send-pending-reports`) must not load a text engine.
- `Title`, `Heading`, `Strong`, `Body`, `Muted`, `Caption`, `Mono` and
  `Toned` size their text as multiples of the theme's `FontSize`
  (`Theme.Rem`), so the desktop's text size reaches them. A new `Label`
  draws an uppercase monospace label with letter spacing, in the label
  color.
- `docs/specs/frostyard-theme.md` lists every value per theme, high
  contrast included, and links ADR-0011 and the design doc. A test computes
  each pair it lists: text on background, chrome, panel, surface and
  selected fill; muted text on the same; tone text on its fill; the focus
  ring on background and chrome. It fails below 4.5:1 for text and 3:1 for
  the ring and meaningful marks (closes frostyard/frostui#16).
- Rewrite [design: theming](../design/frostui.md#theming); add an entry
  under [API changes](../design/frostui.md#api-changes) and the new exports
  to `internal/repocheck/testdata/api.txt`.
- **Done when:** `make snapshot` draws `examples/commandcenter` and
  `examples/threepane` in both themes in Inter, the contrast test passes in
  both themes with and without high contrast, and `make ci` passes.

## Phase 2 — Components in the brand (medium)

- Metrics: `RadiusCard` 0, `RadiusChip` 2, toasts and the palette 0,
  banners and fields 2, `Kbd` 2. Remove the toast's shadow; give it a
  line-strong hairline on `panel`.
- `Card` and `Figures` fill with the panel gradient over a hairline.
- `Pill`, `IconPill`, `DotPill` and `ColorPill` draw a 1px outline in the
  tone's line around a `Label`-style text in the tone's color; neutral pills
  take the tag text color.
- `IconButton` and `LabeledButton` draw a secondary button as a ghost (no
  face, a hairline, hover text) and the primary on the brand gradient with
  on-accent text.
- [Shell](../design/frostui.md#the-command-center-shell) and
  [PaneShell](../design/frostui.md#the-frame-without-the-rail): the brand
  tile on the gradient at radius 4; the palette field on the code color at
  radius 2; rail section titles as `Label`; the selected rail item on the
  selected fill with a 1px accent rule on its left edge and selected text;
  a badge on the selected item stays readable (closes
  frostyard/frostui#15).
- [Source list](../design/frostui.md#source-list): headers as `Label`; the
  selected row as the rail's; counts as outlined badges.
- [Command palette](../design/frostui.md#command-palette): headings as
  `Label`; the chosen row on the selected fill with the accent rule.
- [Console](../design/frostui.md#console) on the code color.
  `SoftSelection` uses the selected fill. Dialogs, rows, the wizard's step
  dots, `Gauge`, `SegmentBar`, `ConnectionStatus` and `Settings` take the
  new radii and colors.
- Copy the new snapshots to `docs/images/`; update the design doc's
  sections above, `README.md` and `AGENTS.md` where they describe the look.
- **Done when:** the three examples' snapshots match the design system's
  component previews in both themes in a side-by-side review, and `make ci`
  passes. The maintainer then tags the next minor version.

## Phase 3 — ChairLift (small)

- Require the new frostui; the first-run appearance preview uses
  `FrostyardTheme`; its card's shadow goes.
- Move its literal radii (5) and font sizes (26) to kit metrics and styles.
- **Done when:** ChairLift's `make ci` passes and its snapshot pages
  (`CHAIRLIFT_SNAPSHOT_DIR`) show Overview, Updates and the first run in both
  themes.

## Phase 4 — Frostmail (large)

- Require the new frostui. Move the 148 literal font sizes (most in
  `internal/desktop/pim`: calendar, tasks, people and the event window) and
  27 literal radii to kit styles and metrics.
- `uikit.SummaryRow` takes the selected fill and outlined counts, and the
  row fixes in frostyard/frostmail#267, #270 and #271 land with it, since
  they change the same rows and sidebar.
- The reader frames the forced-white HTML page in both themes (see the open
  question).
- **Done when:** Frostmail's mail, calendar, people and tasks views render
  in both themes in its snapshot test and its `make ci` passes.

## Phase 5 — Frostnotes (small)

- Require the new frostui. Its light, dark or system setting picks
  `FrostyardTheme`; the editor's code blocks and fences use the code color,
  headings Inter at the kit's sizes; move its 13 literal font sizes and 5
  radii.
- **Done when:** a Tester test draws the main window over a vault in both
  themes and checks the background and font it draws with, the window is
  reviewed by hand in both, and its `make ci` passes.

## Later / ideas

- Refresh the apps' README screenshots and store listings (Flathub,
  Homebrew) once phases 3 to 5 ship.
- A high-contrast pass beyond MyGo's mix, if the measured high-contrast
  values in the spec fall short of 7:1 for text.

## Open questions

- **How the palette is found from a `*ui.Theme`:** kit functions take only
  the theme. Either the kit keeps the palette of each theme it built and
  looks it up by the theme's background color, or a kit theme type wraps
  `ui.Theme` (an API change). Decide in phase 1; record it in the design
  doc.
- **Inter's files:** whether `ui.RegisterFont` takes Inter's variable font
  on every platform engine (Pango on Linux, DirectWrite, Core Text, wasm),
  or the kit ships static weights (400, 570, 650, 700, 720 and 750 are
  used), and
  whether italics ship. Decide in phase 1 by drawing each weight on each
  platform.
- **Badge size for counts:** the brand's 9 DIP label text is small for
  unread counts in a sidebar; counts may take the `pill` size in the same
  outline. Decide at phase 2's review.
- **HTML mail in Frostyard:** Frostmail draws a message's HTML on a forced
  white page. In Frostyard it sits inside the dark reader: a framed page, or
  a dark rendering of simple messages. Frostmail decides by phase 4.
- **An org-wide record:** the brand lives in frostyard/core. Whether core
  records that frostyard's desktop apps draw it, as an org ADR the apps
  list, is core's call before phase 3.
- **Where the values live after phase 1:** the kit's spec, core's
  `frostyard-design` skill, or both with a check that they agree.

## References

- Implements: [ADR-0011](../adr/0011-draw-the-frostyard-brand-in-a-dark-and-a-light-theme.md),
  [design: theming](../design/frostui.md#theming),
  [design: the command-center shell](../design/frostui.md#the-command-center-shell),
  [design: the frame without the rail](../design/frostui.md#the-frame-without-the-rail),
  [design: source list](../design/frostui.md#source-list),
  [design: command palette](../design/frostui.md#command-palette),
  [design: console](../design/frostui.md#console)
- Builds on: [ADR-0007](../adr/0007-embed-a-lucide-icon-set-as-the-last-fallback.md),
  [roadmap](roadmap.md)
- Fixes: frostyard/frostui#15, frostyard/frostui#16
