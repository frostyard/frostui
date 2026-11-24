# Plan: The NSGo kit and its look

This plan orders the work on `workspaceui`
([design](../design/workspaceui.md)), the kit for frostyard's NSGo apps:
its move here, the Frostyard look, and the parts the three apps place.
It is the kit's share of frostyard/frostmail's plan 0016, which also holds
each app's phases and, until Phase 2 here turns it into
[the kit's spec](../specs/workspaceui.md), the draft of the standard (the
window, the rules for every app, the parts). The
[roadmap](roadmap.md) stays the MyGo kit's.

## Phase 0 — The module moves here (small, done 2026-11-24)

- `workspaceui/` holds the 6 files of frostyard/pocketgo's
  `desktop/workspaceui/v0.1.0-rc.1` under the new module path, with the
  import guard and the shell's tests
  ([ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md),
  [spec](../specs/workspaceui.md)).
- Every gate runs in both modules, and the module has its tag series.
- The maintainer tags `workspaceui/v0.1.0`. FrostMail, Frostnotes and
  PocketGo then take the new import path, and PocketGo removes its copy.
- **Done when:** the three apps build and pass their gates on
  `workspaceui/v0.1.0`, with no change in what they draw.
- **Done (2026-11-24).** The maintainer tagged `workspaceui/v0.1.0` on
  the merge commit of frostyard/frostui#70. Each app requires that
  version, and each change merged with every check of its pull request
  passing: frostyard/frostmail#387, frostyard/frostnotes#275 and
  frostyard/pocketgo#167, which also removed PocketGo's copy. For what
  PocketGo draws, its pull request records a run of both real targets on
  one host: 12 of 13 pictures byte-identical to `main`'s on each target,
  and the 13th differing between runs of the same tree. No picture
  compares what FrostMail or Frostnotes drew before and after: their
  suites keep none yet. For those two the phase closes on the module's 6
  source files being byte-identical to the ones they ran before, which the
  owner accepts by merging this record.

## Phase 1 — The themes and the type (medium, the kit's part built)

- The Frostyard and Frostyard light themes as an NSGo theme, from
  [the themes' spec](../specs/frostyard-theme.md): the color roles, and
  named roles for the colors NSGo has no role for (chrome, panel, the
  selected fill, the label, the tones' fills and lines).
- Inter for the native builds and for the web build.
- One call that an app makes at its root, and an appearance choice
  (System, Light, Dark) an app can show in its settings.
- The contrast table of the themes' spec, measured for the NSGo roles.
- `Shell` draws no brand where the system's title bar shows the app's name.
- **Done when:** a picture of each app's main window in each theme shows
  Frostyard's colors and Inter with one title row, and the contrast test
  passes.
- **The kit's part (2026-11-24).** `Theme`, `Apply`, `Frame`, `ColorsOf`
  and `AppearanceChoice` are built, with the contrast test (69 pairs in
  each theme, each at 4.5:1 or over), and `Shell` leaves the app's name to
  the system's title bar ([spec](../specs/workspaceui.md#the-themes)).
  The compact shell paints its ground, which Frostnotes' pictures showed
  missing in a dark narrow window (frostyard/frostnotes#277).
  FrostMail's picture tours, run against this code in a GTK window on
  Xvfb's X11 in both appearances, drew Frostyard's colors in the shell, the
  list, a menu and a dialog, and their text in the bundled Inter, where the
  same tours without `Apply` drew Xvfb's own sans-serif face. The first
  trial named the family "Inter" and drew that fallback face: the variable
  file names "Inter Variable". Left: a tag of the module, the three apps
  calling `Apply` and keeping their pictures, and a look at Inter in the
  web build and on Windows.

## Phase 2 — The parts and the gallery (large, in progress)

- The parts the standard names, each with a layout test: the toolbar's
  segments and buttons, the navigation row with its count, the summary row
  with its avatar and marks, the search field with its key hint, the
  banner, the empty and loading states, the field and option rows, the
  form dialog and the confirmation.
- A gallery app that shows every part in both themes, built for Linux and
  the web, with screenshot comparisons.
- The standard moves from frostmail's plan 0016 into the kit's spec, each
  measure with the test that holds it.
- **Done when:** the gallery shows every part in both themes on Linux and
  in a browser, and the owner accepts its pictures.
- **Round 1 (2026-11-24).** The gallery, with the standard window and 5
  pages of controls, 20 pictures in both themes from a GTK window on
  Xvfb's X11. The owner looked at the pictures and ran the app, and
  accepted the look as drawn: the soft fill of a chosen row, the toolbar
  on NSGo's inset strip, NSGo's tinted danger button, an opaque panel
  under a dialog, initials on a neutral avatar, the summary row of 3
  lines, modules as tabs of an icon alone, and NSGo's corners and sizes.
  He set the phase's bounds: it decides colors, controls and the inside of
  a part; where parts go in a window, and how one moves between them, is
  each app's later phase.
- **Round 2 (2026-11-24).** The accepted parts are in the kit with layout
  tests (`NavLabel`, `SummaryRow`, `SoftRows`, `Banner`, `EmptyState`,
  `DialogPanel`, `ConnectionItem`), the gallery draws them from the kit
  with its pictures unchanged, the control map is in
  [the spec](../specs/workspaceui.md#the-control-map), and the gallery is
  a module of the repository's gates.
- **Left:** the controls of the map the gallery does not draw yet (a
  tree, a popover and a hover card, a toast, paging, a disclosure, NSGo's
  settings, the command palette, a search field's icon and key hint);
  draft windows for a calendar, people, tasks, compose and settings;
  pictures from the web target; and the rest of the standard (frostmail's
  plan 0016: the rules for every app) in the spec.

## Phase 3 — NSGo's gaps (small, alongside)

- An issue upstream for each gap the kit works around: edge resize and
  window controls for an undecorated GTK window; themable corners and type
  scale in `ui`; a `Style` on `IconButton`, `Badge` and `SidebarItem`; row
  hover, and sticky headers with selection, in `ui.List`. The owner files
  them, or approves the text an agent drafts.
- **Done when:** each gap has an issue whose link is in the kit's spec,
  beside the workaround to remove when it lands.

## Later / ideas

- The high-contrast values of the themes' spec, when NSGo reports the
  desktop's high contrast to an app.
- Inter's italic file, if a view needs italics that the platform does not
  slant well.
- A pinned list of the module's exported API, as the root's `api.txt`,
  when the parts of Phase 2 have settled.
- A selected fill in the accent while a list holds the focus.

## Open questions

- **The web build under Lodge:** whether the kit's themes or Lodge's own
  tokens decide the colors when the two differ. By Phase 1.

## References

- Implements: [design/workspaceui.md](../design/workspaceui.md),
  [specs/workspaceui.md](../specs/workspaceui.md),
  [specs/frostyard-theme.md](../specs/frostyard-theme.md)
- Rationale: [ADR-0014](../adr/0014-hold-the-nsgo-kit-as-a-second-module.md)
- frostyard/frostmail: plan 0016 (a shared look for the NSGo apps) and
  ADR-0062
