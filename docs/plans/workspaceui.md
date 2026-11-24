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

## Phase 1 — The themes and the type (medium)

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

## Phase 2 — The parts and the gallery (large)

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

## Phase 3 — NSGo's gaps (small, alongside)

- An issue upstream for each gap the kit works around: edge resize and
  window controls for an undecorated GTK window; themable corners and type
  scale in `ui`; a `Style` on `IconButton`, `Badge` and `SidebarItem`; row
  hover, and sticky headers with selection, in `ui.List`. The owner files
  them, or approves the text an agent drafts.
- **Done when:** each gap has an issue whose link is in the kit's spec,
  beside the workaround to remove when it lands.

## Later / ideas

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
