package frostui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// DiffView draws a read-only unified diff with old/new line numbers and signed,
// colored additions and removals. Keep one instance per window. SetPatch parses
// only when the text changes; View builds only visible rows and wraps long lines.
// The zero value is ready to use. File operations belong to the caller.
type DiffView struct {
	patch  string
	lines  []diffLine
	list   ui.ListState
	digits int
}

type diffLine struct {
	text      string
	kind      byte
	old, next int
}

var diffHunk = regexp.MustCompile(`^@@ -([0-9]+)(?:,([0-9]+))? \+([0-9]+)(?:,([0-9]+))? @@`)

// SetPatch replaces the unified diff and scrolls to its beginning. An identical
// patch preserves the current scroll position. Metadata and binary notices are
// displayed literally; unrecognized text never becomes executable content.
func (v *DiffView) SetPatch(patch string) {
	if v.patch == patch {
		return
	}
	v.patch = patch
	v.lines = parseDiff(patch)
	v.digits = 4
	for _, line := range v.lines {
		v.digits = max(v.digits, len(strconv.Itoa(max(line.old, line.next))))
	}
	v.list.ScrollTo(0, ui.Start)
}

func parseDiff(patch string) []diffLine {
	if patch == "" {
		return nil
	}
	var lines []diffLine
	old, next, leftOld, leftNew := 0, 0, 0, 0
	for _, text := range strings.Split(strings.TrimSuffix(patch, "\n"), "\n") {
		text = strings.TrimSuffix(text, "\r")
		line := diffLine{text: text}
		if match := diffHunk.FindStringSubmatch(text); match != nil {
			old, _ = strconv.Atoi(match[1])
			next, _ = strconv.Atoi(match[3])
			leftOld, leftNew = 1, 1
			if match[2] != "" {
				leftOld, _ = strconv.Atoi(match[2])
			}
			if match[4] != "" {
				leftNew, _ = strconv.Atoi(match[4])
			}
			line.kind = '@'
		} else if len(text) > 0 && (leftOld > 0 || leftNew > 0) {
			switch text[0] {
			case ' ':
				if leftOld > 0 && leftNew > 0 {
					line.old, line.next = old, next
					old++
					next++
					leftOld--
					leftNew--
				}
			case '-':
				if leftOld > 0 {
					line.kind, line.old = '-', old
					old++
					leftOld--
				}
			case '+':
				if leftNew > 0 {
					line.kind, line.next = '+', next
					next++
					leftNew--
				}
			case '\\': // A no-final-newline marker consumes neither side.
			default:
				leftOld, leftNew = 0, 0
			}
		}
		lines = append(lines, line)
	}
	return lines
}

// View draws the diff. Give the returned element a height or a share of a sized
// parent with Grow. Text can be selected and copied; the original +/- prefixes
// distinguish changes without relying on color.
func (v *DiffView) View(c *ui.Context) ui.Element {
	t := c.Theme()
	width := float32(max(v.digits, 4)) * TextSize(c, 8)
	v.list.Label = func(i int) string { return v.lines[i].text }
	return ui.List(c, &v.list, len(v.lines), func(i int) {
		line := v.lines[i]
		bg, fg := ColorsOf(t).Code, t.Text
		switch line.kind {
		case '+':
			bg = ToneFill(t, Success)
		case '-':
			bg = ToneFill(t, Danger)
		case '@':
			bg = ToneFill(t, Accent)
			fg = ToneColor(t, Accent)
		}
		ui.Row(c).Gap(SpaceS).Padding(2, SpaceS).AlignItems(ui.Start).Background(bg).Children(func() {
			old, next := "", ""
			if line.old > 0 {
				old = fmt.Sprint(line.old)
			}
			if line.next > 0 {
				next = fmt.Sprint(line.next)
			}
			Mono(c, old).Width(width).TextColor(t.TextMuted).AccessHidden(true)
			Mono(c, next).Width(width).TextColor(t.TextMuted).AccessHidden(true)
			Mono(c, line.text).TextColor(fg).Wrap().Grow(1).Basis(0).Selectable()
		})
	}).Label("Changes").Background(ColorsOf(t).Code)
}
