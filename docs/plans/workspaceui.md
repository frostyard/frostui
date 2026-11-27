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

## Phase 1 — The themes and the type (medium, done 2026-11-24)

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
  file names "Inter Variable".
- **The apps (2026-11-24).** The maintainer's tag is `workspaceui/v0.2.0`,
  on the merge commit of frostyard/frostui#72. FrostMail
  (frostyard/frostmail#395), Frostnotes (frostyard/frostnotes#278) and
  PocketGo (frostyard/pocketgo#168) call `Apply`, sit in a `Frame` and
  show the choice of appearance, and each made its kept pictures again in
  both themes: they show Frostyard's colors and Inter. In Frostnotes and
  in PocketGo the bundled file was told from the host's own Inter: with
  the host's hidden the pictures did not change, and in Frostnotes with
  the registration also removed the text was Noto Sans.
- **Left of "Done when": one title row.** No picture shows it, since a
  capture holds the window's content and not the system's title bar; the
  pictures show the navigation's head without the app's name on Linux.
  The phase closes when the owner sees it on a desktop, or accepts closing
  without it. Not looked at, and not part of "Done when": the web build's
  font and Windows.
- **The owner's picture (2026-11-24).** The owner's screenshot of
  PocketGo's window on his desktop, in Frostyard, shows the system's title
  bar with the app's name and, under it, the navigation's head with its
  toggle and no name: the name once. It shows PocketGo only; FrostMail's
  and Frostnotes' windows were not looked at on a desktop, and the same
  `Shell` draws their heads.
- **Closed (2026-11-24).** The owner accepted closing the phase on that
  picture. The missing piece of "Done when": a desktop picture with one
  title row for FrostMail and for Frostnotes.
- **Seen in the apps' pictures, for Phase 2:** a dialog's and a palette's
  panel let the view under them show through and take `background`, not
  `raised`; in a wide window the content's ground is 1 step of 255 off
  `background` (`#f4f9fc` for `#f5f9fc` on Linux); a settings dialog's
  sidebar is a neutral gray, not `inset`.

## Phase 2 — The parts and the gallery (large, done 2026-11-24)

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
- **Round 3 (2026-11-24).** The gallery draws the rest of the control
  map: a tree, a disclosure, paging, a popover, a hover card, toasts, the
  command palette and NSGo's settings view, and a prototype of a search
  field with its icon and key hint. The palette, the popover and the
  toasts had NSGo's translucent ground, as dialogs had, and take
  `DialogPanel`. 30 pictures in both themes (the round's first record
  said 26, which missed the 2 new pages). The owner looked at the round's
  pictures and accepted it, the search field included (2026-11-24).
- **Not part of this phase,** by the bounds the owner set: draft windows
  for a calendar, people, tasks, compose and settings. They are layout,
  and each app's own phase draws them from these controls.
- **Round 4 (2026-11-24).** `SearchField` is in the kit with its test, and
  gains a button that clears it, which is in the gallery's kept pictures;
  the owner has looked at every one of them. The standard's rules for an app are in
  [the spec](../specs/workspaceui.md#the-standard-for-an-app). The gallery's
  narrow window folds its toolbar into one row: its header had overlapped.
- **In a browser (2026-11-24).** `make gallery-pictures-web`, in a headless
  Chromium: the 10 tours pass and write 30 pictures, each view in both
  themes. The agent looked at 1, the standard window in Frostyard: the
  themes' colors, the text in Inter, and the app's name at the navigation's
  head, as no system title bar shows it there. 1 fault seen: the list's
  scroll bar is drawn light in the dark theme. The pictures are not kept
  (6.9 MB), and were not given to the owner.
- **Done (2026-11-24).** The gallery shows every part in both themes on
  Linux and in a browser, by its tours on the 2 targets. The owner accepted
  the look in 3 rounds, from the Linux pictures and the app run on his
  desktop.

## Phase 2.5 — What the first app asked for (small)

FrostMail's main window, built on the kit (frostyard/frostmail's plan
0016, Phases 3 and 4), asked the kit for 4 things it did not have.

- **Spacing.** `SpaceXS` to `SpaceXL` and `Indent`: the standard's rule 2
  keeps a padding written as a number out of an app, and the kit had no
  step to take in its place. FrostMail's style test counts 13 such
  paddings.
- **Flags.** `ColorsOf(cx).Flag(n)`: 7 colors in each theme, the values
  FrostMail's other desktop client draws its flags in. Without them a
  flag drew in a tone, with yellow as orange and purple as the accent
  (frostyard/frostmail#401).
- **A page.** `ColorsOf(cx).Page()`, white in both themes, and
  `PageFrame`, which sets a page that is not the app's own on it inside
  a hairline: so that a part of a message's page not yet painted is not
  the dark window showing through.
- **A row's text size.** `SummaryRow.Scale`, for an app's own text size
  setting (frostyard/frostmail#400).
- **Built (2026-11-24).** The 4 are in the kit with their tests, and the
  gallery has a page that draws them, "Flags, a page, sizes": 32
  pictures, 2 of them new. The flags' values are the agent's choice,
  taken from that client, and are drawn for the owner to judge.
- **Done when:** the owner accepts the page's pictures, the kit is
  tagged, and FrostMail's style test lists no padding.

## Phase 2.6 — What the apps asked for next (small)

FrostMail and Frostnotes, on `workspaceui/v0.4.0`, each held something
the kit should (frostyard/frostmail's plan 0017, Phase 1).

- **A flag's row.** `NavLabel.Tint`: FrostMail's navigation drew every
  flag's row in the accent (frostyard/frostmail#401).
- **An empty pane's actions.** `EmptyState.Actions`: Frostnotes' pane
  with no vault drew its own buttons under the kit's part.
- **Headings.** `Heading(level)`: Frostnotes' style test had 1 exception,
  a heading's size written as a number.
- **The window.** `KeepWindow`: FrostMail and Frostnotes each kept the
  window's size and place with the same code.
- **The tree's width and the right panel's state.**
  `FileExplorer.OnTreeWidthChange`, and `Inspector.Open` with
  `OnOpenChange`: the kit held both and told no app, so neither was kept
  between runs.
- **Built (2026-11-24).** The 6 are in the kit with their tests. The
  gallery draws the 3 that have a look: the flags' rows and the headings
  on "Flags, a page, sizes", and an empty pane with 2 buttons on "Search,
  trees and popovers". 10 of its 32 pictures changed: those 2 pages, and
  the 3 overlays taken over the second, in both themes.
- **Done when:** the owner accepts the changed pictures, the kit is
  tagged, and the apps take the parts in place of their own.

## Phase 3 — NSGo's gaps (small, alongside)

- An issue upstream for each gap the kit works around: edge resize and
  window controls for an undecorated GTK window; themable corners and type
  scale in `ui`; a `Style` on `IconButton`, `Badge`, `SidebarItem` and
  `Menu`; row hover, and sticky headers with selection, in `ui.List`; the
  floating panels' translucent ground where no blur is drawn (Linux); a
  search field; a scroll bar that follows the theme on the web; pictures
  from every app a test process starts on the web target. The owner files
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
