# Documentation

Docs are split by the question they answer:

| Directory | Question | Contents |
| --- | --- | --- |
| [adr/](adr/) | **Why** did we choose this? | Architecture Decision Records — immutable once accepted; superseded, never edited |
| [design/](design/) | **How** does it fit together? | Living documents describing the current architecture |
| [specs/](specs/) | **What exactly** is the contract? | Precise, testable interface definitions |
| [plans/](plans/) | **When/in what order** do we build? | Roadmaps and phase plans; updated as work lands |

## Index

### Decisions (ADRs)

- [0001 — Record architecture decisions](adr/0001-record-architecture-decisions.md)
- [0002 — Agent-portable instruction surface](adr/0002-agent-portable-instruction-surface.md)
- [0003 — Verify gate, risk paths and corrections](adr/0003-verify-gate-risk-paths-and-corrections.md)
- [0004 — Extract the UI kit into its own module](adr/0004-extract-the-ui-kit-into-its-own-module.md)
- [0005 — Build everywhere, with the freedesktop parts behind build tags](adr/0005-build-everywhere-with-freedesktop-parts-behind-build-tags.md)
- [0006 — A frame without the rail, whose top-bar segments live in the panes' columns](adr/0006-a-frame-without-the-rail.md)
- [0007 — Embed a Lucide icon set as the last fallback](adr/0007-embed-a-lucide-icon-set-as-the-last-fallback.md)
- [0008 — Open behind allow-lists, and let editors keep their keys](adr/0008-open-behind-allow-lists-and-let-editors-keep-their-keys.md)
- [0009 — The top bar's end belongs to the window, and the palette field turns compact](adr/0009-the-top-bars-end-belongs-to-the-window.md)
- [0010 — Keep only a mailto link's addressees, subject and body](adr/0010-keep-only-a-mailto-links-addressees-subject-and-body.md)

### Org-wide decisions

- [Org-wide decisions (frostyard/core ADRs)](org-adrs.md)

### Guides

- [After you create a repository from this template](new-repository.md): the admin steps for frostyard/frostui

### Design

- [frostui: frostyard's desktop UI kit](design/frostui.md): start here
- [Agent-ready repository](design/agent-ready-repository.md)

### Specs

- [API stability and versions](specs/api-stability.md)

### Plans

- [Roadmap](plans/roadmap.md)

## Conventions

- **New docs start from their category's `TEMPLATE.md`** (in each directory).
- New decision → new ADR with the next number; if it reverses an old one, mark
  the old one `Superseded by NNNN` rather than editing it.
- Design docs are updated in place to always reflect reality.
- Specs change only alongside the code that implements them.
- Cross-links between categories are mandatory in both directions — see the
  documentation rules in [AGENTS.md](../AGENTS.md) (CLAUDE.md/GEMINI.md are
  symlinks to it, ADR-0002).
- Adding a doc means adding it to the index above.
