package frostui

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/frostyard/frostui/frostuitest"
)

// noteSource answers at once from names, ranked, and counts the queries
// it was asked in asked.
func noteSource(names []string, asked *[]string, ran *string) PaletteSource {
	return PaletteSource{ID: "notes", Title: "Notes", Placeholder: "Find a note", Query: func(q *PaletteQuery) {
		*asked = append(*asked, q.Text)
		var items []PaletteItem
		for _, n := range Rank(names, q.Text, 0, func(s string) string { return s }) {
			items = append(items, PaletteItem{ID: n, Title: n, Detail: "Vault", Run: func() { *ran = n }})
		}
		q.Reply(items)
	}}
}

func paletteTester(p *Palette, cmds []Command) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		Title(c, "Page")
		p.View(c, cmds)
	}, 1000, 700)
}

func TestPaletteAsksSourcesLazilyOncePerQuery(t *testing.T) {
	var asked []string
	ran := ""
	p := &Palette{Sources: []PaletteSource{noteSource([]string{"Garden plan", "Reading list"}, &asked, &ran)}}
	tt := paletteTester(p, []Command{{ID: "c", Title: "Close tab", Group: "Tabs"}})
	tt.Frame()
	if len(asked) != 0 {
		t.Fatalf("a closed palette asked its source: %q", asked)
	}
	p.Show()
	tt.Frame()
	tt.Frame()
	if !reflect.DeepEqual(asked, []string{""}) {
		t.Fatalf("opening asked %q, want one empty query", asked)
	}
	for _, want := range []string{"Close tab", "Tabs", "Notes", "Garden plan", "Reading list", "Vault"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	tt.Type("gar")
	tt.Frame()
	if got := asked[len(asked)-1]; got != "gar" {
		t.Errorf("last query = %q in %q", got, asked)
	}
	n := len(asked)
	tt.Frame()
	tt.Frame()
	if len(asked) != n {
		t.Errorf("frames with the same query asked again: %q", asked)
	}
	if !tt.HasText("Garden plan") || tt.HasText("Reading list") {
		t.Errorf("results for gar: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if ran != "Garden plan" || p.Open {
		t.Errorf("Enter ran %q, open %v", ran, p.Open)
	}
}

func TestPaletteOpensOnOneSource(t *testing.T) {
	var asked, askedOther []string
	ran := ""
	notes := noteSource([]string{"Garden plan"}, &asked, &ran)
	other := PaletteSource{ID: "tags", Title: "Tags", Query: func(q *PaletteQuery) {
		askedOther = append(askedOther, q.Text)
		q.Reply([]PaletteItem{{ID: "t", Title: "#garden"}})
	}}
	p := &Palette{Sources: []PaletteSource{notes, other}}
	tt := paletteTester(p, []Command{{ID: "c", Title: "Close tab", Group: "Tabs"}})
	p.ShowSource("notes")
	tt.Frame()
	if p.ShownSource() != "notes" {
		t.Errorf("ShownSource = %q", p.ShownSource())
	}
	if !tt.HasText("Garden plan") || tt.HasText("Close tab") || tt.HasText("#garden") {
		t.Errorf("on notes alone: %q", tt.Texts())
	}
	if len(askedOther) != 0 {
		t.Errorf("a source not shown was asked: %q", askedOther)
	}
	a, ok := tt.AccessByRole(ui.RoleTextField, "Notes")
	if !ok || a.Placeholder != "Find a note" {
		t.Errorf("the field = %+v, %v; want the source's name and placeholder", a, ok)
	}
	tt.Key(0, ui.KeyEscape)
	p.ShowSource(PaletteCommands)
	tt.Frame()
	if !tt.HasText("Close tab") || tt.HasText("Garden plan") {
		t.Errorf("on the commands alone: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	p.Show()
	tt.Frame()
	for _, want := range []string{"Close tab", "Garden plan", "#garden"} {
		if !tt.HasText(want) {
			t.Errorf("on every source, missing %q in %q", want, tt.Texts())
		}
	}
}

func TestPaletteTakesAsyncResultsAndDropsStaleOnes(t *testing.T) {
	loop := frostuitest.NewLoop()
	var queries []*PaletteQuery
	p := &Palette{Sources: []PaletteSource{{ID: "mail", Title: "Messages", Query: func(q *PaletteQuery) {
		queries = append(queries, q)
	}}}}
	tt := paletteTester(p, nil)
	p.Show()
	tt.Frame()
	if len(queries) != 1 {
		t.Fatalf("asked %d times", len(queries))
	}
	if !tt.HasText("Searching…") {
		t.Errorf("no sign of the source working: %q", tt.Texts())
	}
	first := queries[0]
	tt.Type("hello")
	tt.Frame()
	second := queries[len(queries)-1]
	if second.Text != "hello" {
		t.Fatalf("second query = %q", second.Text)
	}
	select {
	case <-first.Context().Done():
	default:
		t.Error("the stale query's context is not canceled")
	}
	if second.Context().Err() != nil {
		t.Error("the current query's context is canceled")
	}
	// A worker of each query posts its results, the old one last.
	var tookOld, tookNew bool
	go loop.Post(func() { tookNew = second.Reply([]PaletteItem{{ID: "new", Title: "New result"}}) })
	loop.RunUntil(t, func() bool { return tookNew })
	go loop.Post(func() { tookOld = first.Reply([]PaletteItem{{ID: "old", Title: "Old result"}}) })
	done := false
	go loop.Post(func() { done = true })
	loop.RunUntil(t, func() bool { return done })
	tt.Frame()
	if tookOld || !first.Stale() {
		t.Error("a stale query's reply was taken")
	}
	if !tt.HasText("New result") || tt.HasText("Old result") {
		t.Errorf("results: %q", tt.Texts())
	}
	if tt.HasText("Searching…") {
		t.Error("still searching after the reply")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if second.Reply([]PaletteItem{{ID: "late", Title: "Late"}}) {
		t.Error("a reply after the palette closed was taken")
	}
	if second.Context().Err() == nil {
		t.Error("closing the palette left the query's context alive")
	}
}

func TestPaletteChoiceFollowsItsItemAsResultsCome(t *testing.T) {
	var slow, fast *PaletteQuery
	p := &Palette{Sources: []PaletteSource{
		{ID: "slow", Title: "Slow", Query: func(q *PaletteQuery) { slow = q }},
		{ID: "fast", Title: "Fast", Query: func(q *PaletteQuery) { fast = q }},
	}}
	tt := paletteTester(p, nil)
	p.Show()
	tt.Frame()
	fast.Reply([]PaletteItem{{ID: "a", Title: "Fast A"}, {ID: "b", Title: "Fast B"}})
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	if p.Selected != 1 {
		t.Fatalf("Selected = %d", p.Selected)
	}
	slow.Reply([]PaletteItem{{ID: "s", Title: "Slow S"}})
	tt.Frame()
	if p.Selected != 2 {
		t.Errorf("Selected = %d, want 2: the choice stays on Fast B as a row comes above it", p.Selected)
	}
}

func TestPaletteLimitsASourceOnlyBesideOthers(t *testing.T) {
	var names []string
	for i := range 20 {
		names = append(names, fmt.Sprintf("Note %02d", i))
	}
	var asked []string
	ran := ""
	src := noteSource(names, &asked, &ran)
	src.Limit = 3
	p := &Palette{Sources: []PaletteSource{src}}
	tt := paletteTester(p, nil)
	p.Show()
	tt.Frame()
	if !tt.HasText("Note 02") || tt.HasText("Note 03") {
		t.Errorf("beside others, Limit 3: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	p.ShowSource("notes")
	tt.Frame()
	if !tt.HasText("Note 05") {
		t.Errorf("alone, the source shows more: %q", tt.Texts())
	}
}

func TestPaletteRunWithGetsTheModifiers(t *testing.T) {
	var got []ui.Modifiers
	src := PaletteSource{ID: "notes", Title: "Notes", Query: func(q *PaletteQuery) {
		q.Reply([]PaletteItem{{ID: "a", Title: "Note", RunWith: func(m ui.Modifiers) { got = append(got, m) }}})
	}}
	p := &Palette{Sources: []PaletteSource{src}}
	tt := paletteTester(p, nil)
	for _, mods := range []ui.Modifiers{0, ui.Cmd, ui.Cmd | ui.Alt, ui.Shift} {
		p.ShowSource("notes")
		tt.Frame()
		tt.HoldModifiers(mods)
		tt.Key(mods, ui.KeyEnter)
		tt.HoldModifiers(0)
		tt.Frame()
		if p.Open {
			t.Fatalf("mods %v left the palette open", mods)
		}
	}
	if want := []ui.Modifiers{0, ui.Cmd, ui.Cmd | ui.Alt, ui.Shift}; !reflect.DeepEqual(got, want) {
		t.Errorf("RunWith got %v, want %v", got, want)
	}
}

func TestPaletteDisabledItemDoesNotRun(t *testing.T) {
	ran := false
	src := PaletteSource{ID: "n", Title: "Notes", Query: func(q *PaletteQuery) {
		q.Reply([]PaletteItem{{ID: "a", Title: "Locked", Disabled: true, Reason: "Read only", Run: func() { ran = true }}})
	}}
	p := &Palette{Sources: []PaletteSource{src}}
	tt := paletteTester(p, nil)
	p.Show()
	tt.Frame()
	if !tt.HasText("Read only") {
		t.Errorf("no reason: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if ran || !p.Open {
		t.Errorf("ran %v, open %v", ran, p.Open)
	}
}

func TestPaletteSaysWhenNothingMatches(t *testing.T) {
	p := &Palette{Sources: []PaletteSource{{ID: "n", Title: "Notes", Query: func(q *PaletteQuery) { q.Reply(nil) }}}}
	tt := paletteTester(p, nil)
	p.Show()
	tt.Frame()
	if !tt.HasText("Nothing matches") {
		t.Errorf("%q", tt.Texts())
	}
}

func TestRankOrdersAndLimits(t *testing.T) {
	names := []string{"Stage the update", "Update everything", "Nothing here", "Updates"}
	id := func(s string) string { return s }
	got := Rank(names, "upd", 0, id)
	if len(got) != 3 || !strings.HasPrefix(got[0], "Update") || got[2] != "Stage the update" {
		t.Errorf("Rank = %q", got)
	}
	if got := Rank(names, "upd", 1, id); len(got) != 1 {
		t.Errorf("Rank limit = %q", got)
	}
	if got := Rank(names, " ", 2, id); !reflect.DeepEqual(got, names[:2]) {
		t.Errorf("Rank of an empty query = %q", got)
	}
}
