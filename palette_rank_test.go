package frostui

import (
	"reflect"
	"testing"
)

// With the query "ab", Match scores: "a b" 15, "ab" 13, "x ab" 10,
// "axb" 10, "xab" 5, "zz" no match. TestMatchScoresForRanking pins them.

func TestMatchScoresForRanking(t *testing.T) {
	want := map[string]int{"a b": 15, "ab": 13, "x ab": 10, "axb": 10, "xab": 5}
	for text, score := range want {
		if s, ok := Match("ab", text); !ok || s != score {
			t.Errorf("Match(ab, %q) = %d, %v; want %d, true", text, s, ok, score)
		}
	}
	if _, ok := Match("ab", "zz"); ok {
		t.Error("Match(ab, zz) matched")
	}
}

func TestFilterScoresHintsAndKeepsTies(t *testing.T) {
	cmds := []Command{
		{ID: "hint-only", Title: "zz", Hint: "ab"},    // 13-4 = 9
		{ID: "tie-a", Title: "x ab", Hint: "xab"},     // 10; hint 5-4 is lower
		{ID: "tie-b", Title: "axb", Hint: "zz"},       // 10
		{ID: "tie-c", Title: "x ab"},                  // 10
		{ID: "hint-wins", Title: "xab", Hint: "a b"},  // 15-4 = 11 beats 5
		{ID: "title-wins", Title: "ab", Hint: "ab"},   // 13; hint 9 is lower
		{ID: "none", Title: "zz"},                     // no match
		{ID: "empty-hint", Title: "zz", Hint: ""},     // no match
		{ID: "late-tie", Title: "zz", Hint: "x ab z"}, // hint 10-4 = 6
	}
	// "x ab z": a after space 6, b 4, len 6/16 = 0: 10, minus 4 = 6.
	got := ids(Filter(cmds, "ab"))
	want := []string{"title-wins", "hint-wins", "tie-a", "tie-b", "tie-c", "hint-only", "late-tie"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter(ab) = %v, want %v", got, want)
	}

	none := Filter(cmds, "qq")
	if none == nil || len(none) != 0 {
		t.Errorf("Filter with no match = %#v, want a non-nil empty slice", none)
	}
	if empty := Filter(nil, "ab"); empty == nil || len(empty) != 0 {
		t.Errorf("Filter(nil) = %#v, want a non-nil empty slice", empty)
	}
}

func TestRankKeepsTiesAndCutsAfterSorting(t *testing.T) {
	type item struct{ name string }
	items := []item{{"xab"}, {"x ab"}, {"zz"}, {"ab"}, {"axb"}, {"a b"}}
	text := func(it item) string { return it.name }
	names := func(its []item) []string {
		out := []string{}
		for _, it := range its {
			out = append(out, it.name)
		}
		return out
	}
	all := []string{"a b", "ab", "x ab", "axb", "xab"}
	for _, tc := range []struct {
		limit int
		want  []string
	}{
		{0, all},
		{-1, all},
		{10, all},
		{5, all},
		{2, all[:2]},
		{3, all[:3]},
		{4, all[:4]},
		{1, all[:1]},
	} {
		if got := names(Rank(items, "ab", tc.limit, text)); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Rank(ab, limit %d) = %q, want %q", tc.limit, got, tc.want)
		}
	}
	for _, limit := range []int{0, 2} {
		if got := Rank(items, "qq", limit, text); got == nil || len(got) != 0 {
			t.Errorf("Rank with no match, limit %d = %#v, want a non-nil empty slice", limit, got)
		}
	}
	if got := Rank([]item(nil), "ab", 0, text); got == nil || len(got) != 0 {
		t.Errorf("Rank(nil) = %#v, want a non-nil empty slice", got)
	}
}

func ids(cmds []Command) []string {
	out := []string{}
	for _, c := range cmds {
		out = append(out, c.ID)
	}
	return out
}
