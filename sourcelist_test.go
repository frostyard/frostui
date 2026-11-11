package frostui

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

// mailTree is a small tree of mailboxes by name.
var mailTree = map[string][]string{
	"Projects": {"Alpha", "Beta"},
	"Alpha":    {"Drafts"},
}

func mailSections() []SourceSection[string] {
	return []SourceSection[string]{
		{ID: "fav", Title: "Favorites", Items: []string{"Inbox", "Flagged"}},
		{ID: "work", Title: "Work", Items: []string{"Inbox", "Projects", "Spam"}},
	}
}

func mailList() *SourceList[string] {
	return &SourceList[string]{
		Label:    "Mailboxes",
		Children: func(item string) []string { return mailTree[item] },
		Row: func(item string) SourceRow {
			r := SourceRow{Title: item, Icon: "folder"}
			switch item {
			case "Inbox":
				r.Count, r.CountTone = "12", Accent
			case "Spam":
				r.Alert = "The server refused the login"
			case "Flagged":
				r.Busy, r.Tag = true, "VIP"
			}
			return r
		},
	}
}

func listTester(l *SourceList[string], sections []SourceSection[string]) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		ui.Column(c).Width(280).Height(600).AlignItems(ui.Stretch).Children(func() {
			l.View(c, sections).Grow(1)
		})
	}, 700, 640)
}

func TestSourceListShowsSectionsAndRows(t *testing.T) {
	tt := listTester(mailList(), mailSections())
	for _, want := range []string{"Favorites", "Work", "Inbox", "12", "Flagged", "VIP", "Projects", "Spam"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q in %q", want, tt.Texts())
		}
	}
	if tt.HasText("Alpha") {
		t.Error("the child of a closed item shows")
	}
	for _, name := range []string{"Inbox, 12", "Spam, The server refused the login", "Flagged, VIP, busy"} {
		if _, ok := tt.AccessByRole(ui.RoleTreeItem, name); !ok {
			t.Errorf("no tree item named %q", name)
		}
	}
}

func TestSourceListSectionHeaderHidesItsItems(t *testing.T) {
	l := mailList()
	tt := listTester(l, mailSections())
	if err := tt.Click("Work"); err != nil {
		t.Fatal(err)
	}
	if tt.HasText("Projects") || l.SectionOpen("work") {
		t.Errorf("Work did not close: %q", tt.Texts())
	}
	if !tt.HasText("Flagged") {
		t.Error("closing Work closed Favorites")
	}
	if err := tt.Click("Work"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Projects") || !l.SectionOpen("work") {
		t.Errorf("Work did not open: %q", tt.Texts())
	}
	l.SetSectionOpen("fav", false)
	tt.Frame()
	if tt.HasText("Flagged") {
		t.Error("SetSectionOpen did not hide the section")
	}
}

func TestSourceListChoosesNothingAtFirst(t *testing.T) {
	l := mailList()
	fired := 0
	l.OnSelect = func(string) { fired++ }
	// Calls before the first frame must not make the first row a choice.
	l.SetSectionOpen("files", true)
	l.SetOpen("files", "Projects", true)
	tt := listTester(l, []SourceSection[string]{{ID: "files", Items: []string{"Projects", "Spam"}}})
	tt.Frame()
	if item, _, ok := l.Selected(); ok || fired != 0 {
		t.Errorf("Selected = %q, %v; OnSelect ran %d times", item, ok, fired)
	}
}

func TestSourceListF2WhileRenamingKeepsTheDraft(t *testing.T) {
	l := mailList()
	l.Rename = func(string, string) {}
	tt := listTester(l, mailSections())
	l.StartRename("work", "Spam")
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Junk")
	tt.Key(0, ui.KeyF2)
	tt.Frame()
	if l.draft != "Junk" {
		t.Errorf("draft = %q after F2 in the field", l.draft)
	}
}

func TestSourceListWithoutATitleHasNoHeader(t *testing.T) {
	l := mailList()
	tt := listTester(l, []SourceSection[string]{{ID: "files", Items: []string{"Projects", "Spam"}}})
	if !tt.HasText("Projects") || !tt.HasText("Spam") {
		t.Errorf("%q", tt.Texts())
	}
	a, ok := tt.AccessByRole(ui.RoleTreeItem, "Projects")
	if !ok || a.Level != 1 {
		t.Errorf("Projects = %+v, %v; want level 1", a, ok)
	}
}

func TestSourceListChoosesByClickAndKeys(t *testing.T) {
	l := mailList()
	var chosen []string
	l.OnSelect = func(item string) { chosen = append(chosen, item) }
	tt := listTester(l, mailSections())
	if err := tt.Click("Flagged"); err != nil {
		t.Fatal(err)
	}
	if item, section, ok := l.Selected(); !ok || item != "Flagged" || section != "fav" {
		t.Fatalf("Selected = %q, %q, %v", item, section, ok)
	}
	// Down passes the Work header, which is never chosen.
	tt.Key(0, ui.KeyDown)
	if item, section, ok := l.Selected(); !ok || item != "Inbox" || section != "work" {
		t.Fatalf("after Down: %q, %q, %v", item, section, ok)
	}
	tt.Key(0, ui.KeyDown)
	if item, _, _ := l.Selected(); item != "Projects" {
		t.Fatalf("after Down twice: %q", item)
	}
	tt.Key(0, ui.KeyRight)
	if !l.IsOpen("work", "Projects") || !tt.HasText("Alpha") {
		t.Errorf("Right did not open Projects: %q", tt.Texts())
	}
	tt.Key(0, ui.KeyRight)
	if item, _, _ := l.Selected(); item != "Alpha" {
		t.Errorf("Right on an open item chose %q, want its first child", item)
	}
	tt.Key(0, ui.KeyLeft)
	if item, _, _ := l.Selected(); item != "Projects" {
		t.Errorf("Left on a leaf chose %q, want its parent", item)
	}
	tt.Key(0, ui.KeyLeft)
	if l.IsOpen("work", "Projects") || tt.HasText("Alpha") {
		t.Error("Left did not close Projects")
	}
	want := []string{"Flagged", "Inbox", "Projects", "Alpha", "Projects"}
	if !reflect.DeepEqual(chosen, want) {
		t.Errorf("OnSelect saw %q, want %q", chosen, want)
	}
}

func TestSourceListKeepsTheChoiceThroughAClosedSection(t *testing.T) {
	l := mailList()
	fired := 0
	tt := listTester(l, mailSections())
	if err := tt.Click("Projects"); err != nil {
		t.Fatal(err)
	}
	l.OnSelect = func(string) { fired++ }
	if err := tt.Click("Work"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if item, _, ok := l.Selected(); !ok || item != "Projects" {
		t.Errorf("closing the section lost the choice: %q, %v", item, ok)
	}
	if err := tt.Click("Work"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a, _ := tt.AccessByRole(ui.RoleTreeItem, "Projects"); !a.Selected {
		t.Errorf("the choice does not show again: %+v", a)
	}
	if fired != 0 {
		t.Errorf("OnSelect ran %d times: the user chose nothing", fired)
	}
}

func TestSourceListForgetsAChoiceThatIsGone(t *testing.T) {
	l := mailList()
	sections := mailSections()
	tt := listTester(l, sections)
	if err := tt.Click("Spam"); err != nil {
		t.Fatal(err)
	}
	sections[1].Items = []string{"Inbox", "Projects"}
	tt.Frame()
	tt.Frame()
	if item, _, ok := l.Selected(); ok {
		t.Errorf("a removed item is still chosen: %q", item)
	}
}

func TestSourceListTellsAnItemInTwoSectionsApart(t *testing.T) {
	l := mailList()
	tt := listTester(l, mailSections())
	l.Select("work", "Inbox")
	tt.Frame()
	if _, section, ok := l.Selected(); !ok || section != "work" {
		t.Errorf("section = %q, %v", section, ok)
	}
	l.Select("", "Inbox")
	tt.Frame()
	if _, section, _ := l.Selected(); section != "fav" {
		t.Errorf("Select in any section chose %q, want the first that holds it", section)
	}
}

func TestSourceListSelectOpensTheWayToAnItem(t *testing.T) {
	l := mailList()
	fired := 0
	l.OnSelect = func(string) { fired++ }
	tt := listTester(l, mailSections())
	l.SetSectionOpen("work", false)
	tt.Frame()
	l.Select("work", "Drafts")
	tt.Frame()
	if !l.SectionOpen("work") || !l.IsOpen("work", "Projects") || !l.IsOpen("work", "Alpha") {
		t.Errorf("the way to Drafts is not open: %v", l.OpenItems("work"))
	}
	if !tt.HasText("Drafts") {
		t.Errorf("Drafts does not show: %q", tt.Texts())
	}
	if item, _, ok := l.Selected(); !ok || item != "Drafts" {
		t.Errorf("Selected = %q, %v", item, ok)
	}
	if a, ok := tt.AccessByRole(ui.RoleTreeItem, "Drafts"); !ok || !a.Selected {
		t.Errorf("Drafts = %+v, %v, want the row chosen", a, ok)
	}
	if fired != 0 {
		t.Errorf("Select ran OnSelect %d times", fired)
	}
	l.ClearSelection()
	tt.Frame()
	if _, _, ok := l.Selected(); ok {
		t.Error("ClearSelection left a choice")
	}
	l.Select("work", "Missing")
	tt.Frame()
	if _, _, ok := l.Selected(); ok {
		t.Error("selecting an item no section holds chose one")
	}
}

func TestSourceListCollapseAndExpandAll(t *testing.T) {
	l := mailList()
	tt := listTester(l, mailSections())
	l.ExpandAll()
	tt.Frame()
	if !tt.HasText("Drafts") {
		t.Errorf("ExpandAll: %q", tt.Texts())
	}
	got := l.OpenItems("work")
	slices.Sort(got)
	if !reflect.DeepEqual(got, []string{"Alpha", "Projects"}) {
		t.Errorf("OpenItems = %q", got)
	}
	l.CollapseAll()
	tt.Frame()
	if tt.HasText("Alpha") || !tt.HasText("Projects") {
		t.Errorf("CollapseAll: %q", tt.Texts())
	}
}

func TestSourceListActivates(t *testing.T) {
	l := mailList()
	var got []string
	l.OnActivate = func(item string, mods ui.Modifiers) { got = append(got, fmt.Sprint(item, "/", mods)) }
	tt := listTester(l, mailSections())
	if err := tt.Click("Spam"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Key(ui.Cmd, ui.KeyEnter)
	want := []string{fmt.Sprint("Spam/", ui.Modifiers(0)), fmt.Sprint("Spam/", ui.Cmd)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("OnActivate saw %q, want %q", got, want)
	}
}

func TestSourceListBuildsContextMenus(t *testing.T) {
	l := mailList()
	got := ""
	l.Menu = func(m *ui.Menu, item, section string) {
		if m.Item("Empty " + item).Chosen() {
			got = section + "/" + item
		}
	}
	tt := listTester(l, mailSections())
	if err := tt.RightClick("Spam"); err != nil {
		t.Fatal(err)
	}
	if m := tt.Menu(); len(m) != 1 || m[0] != "Empty Spam" {
		t.Fatalf("menu = %q", m)
	}
	if err := tt.ChooseMenuItem("Empty Spam"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got != "work/Spam" {
		t.Errorf("menu acted on %q", got)
	}
}

// dragTo drags from the center of the element named from to the one named
// to, and lets go.
func dragTo(t *testing.T, tt *ui.Tester, from, to string, hold func()) {
	t.Helper()
	a, ok := tt.Find(from)
	b, ok2 := tt.Find(to)
	if !ok || !ok2 {
		t.Fatalf("cannot find %q or %q in %q", from, to, tt.Texts())
	}
	ax, ay, bx, by := a.X+a.W/2, a.Y+a.H/2, b.X+b.W/2, b.Y+b.H/2
	tt.Press(ax, ay)
	tt.Move(ax+8, ay+8)
	tt.Move(bx, by)
	tt.Frame()
	if hold != nil {
		hold()
	}
	tt.Release(bx, by)
	tt.Frame()
}

type message struct{ subject string }

func TestSourceListTakesDrops(t *testing.T) {
	l := mailList()
	type dropped struct {
		box  string
		subj string
		mods ui.Modifiers
	}
	var got []dropped
	DropOn(l, func(item string, _ message) bool { return item != "Flagged" }, func(item string, m message, mods ui.Modifiers) {
		got = append(got, dropped{item, m.subject, mods})
	})
	sections := mailSections()
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(280).AlignItems(ui.Stretch).Children(func() { l.View(c, sections).Grow(1) })
			ui.Box(c).Size(120, 40).Drag(message{"Hello"}).Children(func() { ui.Text(c, "A message") })
		})
	}, 700, 640)
	tint := func(name string) any {
		r, _ := tt.Find(name)
		return tt.Image().RGBAAt(int(r.X+r.W)-2, int(r.Y+r.H/2))
	}
	before := tint("Spam")
	var during any
	dragTo(t, tt, "A message", "Spam", func() { during = tint("Spam") })
	if len(got) != 1 || got[0].box != "Spam" || got[0].subj != "Hello" {
		t.Fatalf("drops = %+v", got)
	}
	if during == before {
		t.Error("the row does not show that it takes the value")
	}
	got = nil
	before = tint("Flagged")
	dragTo(t, tt, "A message", "Flagged", func() { during = tint("Flagged") })
	if len(got) != 0 {
		t.Errorf("a row that refuses the value took it: %+v", got)
	}
	if during != before {
		t.Error("a row that refuses the value shows it would take it")
	}
}

func TestSourceListDragsItsItems(t *testing.T) {
	l := mailList()
	l.DragItems = func(items []string) any { return items }
	var moved []string
	var into string
	DropOn(l, func(item string, _ []string) bool { return mailTree[item] != nil },
		func(item string, v []string, _ ui.Modifiers) { into, moved = item, v })
	tt := listTester(l, mailSections())
	dragTo(t, tt, "Spam", "Projects", nil)
	if into != "Projects" || !reflect.DeepEqual(moved, []string{"Spam"}) {
		t.Errorf("into %q moved %q", into, moved)
	}
}

func TestSourceListRenamesInPlace(t *testing.T) {
	l := mailList()
	var renamed [][2]string
	l.Rename = func(item, name string) { renamed = append(renamed, [2]string{item, name}) }
	l.CanRename = func(item string) bool { return item != "Inbox" }
	tt := listTester(l, mailSections())
	if err := tt.Click("Spam"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyF2)
	if !l.Renaming() {
		t.Fatal("F2 did not start renaming")
	}
	if !tt.Focused("Name") {
		t.Fatal("the field did not take the focus")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Junk")
	tt.Key(0, ui.KeyEnter)
	if l.Renaming() {
		t.Fatal("Enter did not end renaming")
	}
	if want := [][2]string{{"Spam", "Junk"}}; !reflect.DeepEqual(renamed, want) {
		t.Errorf("renamed = %v, want %v", renamed, want)
	}
	if a, ok := tt.AccessByRole(ui.RoleTree, "Mailboxes"); !ok || !a.Focused && !tt.Focused("Mailboxes") {
		t.Errorf("the focus did not go back to the tree: %+v, %v", a, ok)
	}

	renamed = nil
	tt.Key(0, ui.KeyF2)
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Other")
	tt.Key(0, ui.KeyEscape)
	if l.Renaming() || renamed != nil {
		t.Errorf("Escape: renaming %v, renamed %v", l.Renaming(), renamed)
	}

	l.StartRename("work", "Inbox")
	if l.Renaming() {
		t.Error("an item CanRename refuses started renaming")
	}
	l.StartRename("work", "Projects")
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if renamed != nil {
		t.Errorf("keeping the name renamed it: %v", renamed)
	}
}

func TestSourceListRenamesAnItemFoundInAnySection(t *testing.T) {
	l := mailList()
	var renamed []string
	l.Rename = func(item, name string) { renamed = append(renamed, item+"="+name) }
	tt := listTester(l, mailSections())
	l.StartRename("", "Spam")
	tt.Frame()
	if !tt.Focused("Name") {
		t.Fatalf("no field took the focus: %q", tt.Texts())
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Junk")
	tt.Key(0, ui.KeyEnter)
	if !reflect.DeepEqual(renamed, []string{"Spam=Junk"}) {
		t.Errorf("renamed = %q", renamed)
	}
	l.StartRename("", "Missing")
	tt.Frame()
	if l.Renaming() {
		t.Error("renaming an item no section holds stays on with no field")
	}
}

func TestSourceListRenameEndsWhenItsRowGoes(t *testing.T) {
	l := mailList()
	var renamed, activated []string
	l.Rename = func(item, name string) { renamed = append(renamed, item+"="+name) }
	l.OnActivate = func(item string, _ ui.Modifiers) { activated = append(activated, item) }
	sections := mailSections()
	tt := listTester(l, sections)
	l.StartRename("work", "Spam")
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Junk")
	sections[1].Items = []string{"Inbox", "Projects"}
	tt.Frame()
	tt.Frame()
	if l.Renaming() {
		t.Fatal("the rename stayed on after its row went")
	}
	if renamed != nil {
		t.Errorf("a rename whose row went called Rename: %q", renamed)
	}
	if err := tt.Click("Projects"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	if !reflect.DeepEqual(activated, []string{"Projects"}) {
		t.Errorf("Enter after the rename activated %q", activated)
	}
	tt.Key(0, ui.KeyF2)
	if !l.Renaming() {
		t.Error("F2 after the rename did not start another")
	}
}

func TestSourceListTrimsANewName(t *testing.T) {
	for _, tc := range []struct {
		typed string
		want  []string
	}{
		{"  Junk  ", []string{"Spam=Junk"}},
		{"   ", nil},
		{"  Spam ", nil},
	} {
		l := mailList()
		var renamed []string
		l.Rename = func(item, name string) { renamed = append(renamed, item+"="+name) }
		tt := listTester(l, mailSections())
		l.StartRename("work", "Spam")
		tt.Frame()
		tt.Key(ui.Cmd, ui.KeyA)
		tt.Type(tc.typed)
		tt.Key(0, ui.KeyEnter)
		if !reflect.DeepEqual(renamed, tc.want) {
			t.Errorf("typed %q: renamed %q, want %q", tc.typed, renamed, tc.want)
		}
	}
}

func TestSourceListRenameEndsWhenTheFocusLeaves(t *testing.T) {
	l := mailList()
	var renamed []string
	l.Rename = func(_, name string) { renamed = append(renamed, name) }
	other := ""
	sections := mailSections()
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		ui.Column(c).Width(280).Height(600).AlignItems(ui.Stretch).Children(func() {
			ui.TextInput(c, &other).Label("Other")
			l.View(c, sections).Grow(1)
		})
	}, 700, 640)
	l.StartRename("work", "Spam")
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Junk")
	if err := tt.Click("Other"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if l.Renaming() || !reflect.DeepEqual(renamed, []string{"Junk"}) {
		t.Errorf("renaming %v, renamed %v", l.Renaming(), renamed)
	}
}

func TestSourceListIsATreeForAssistiveTechnology(t *testing.T) {
	l := mailList()
	tt := listTester(l, mailSections())
	if _, ok := tt.AccessByRole(ui.RoleTree, "Mailboxes"); !ok {
		t.Fatal("no tree named Mailboxes")
	}
	head, ok := tt.AccessByRole(ui.RoleTreeItem, "Work")
	if !ok || head.Level != 1 || !head.Expandable || !head.Expanded {
		t.Errorf("Work = %+v, %v; want level 1, open", head, ok)
	}
	item, ok := tt.AccessByRole(ui.RoleTreeItem, "Projects")
	if !ok || item.Level != 2 || !item.Expandable || item.Expanded {
		t.Errorf("Projects = %+v, %v; want level 2, closed", item, ok)
	}
	l.SetOpen("work", "Projects", true)
	tt.Frame()
	if item, _ = tt.AccessByRole(ui.RoleTreeItem, "Projects"); !item.Expanded {
		t.Errorf("Projects open = %+v", item)
	}
	if leaf, ok := tt.AccessByRole(ui.RoleTreeItem, "Beta"); !ok || leaf.Level != 3 || leaf.Expandable {
		t.Errorf("Beta = %+v, %v; want level 3, a leaf", leaf, ok)
	}
	if err := tt.Click("Spam"); err != nil {
		t.Fatal(err)
	}
	if spam, _ := tt.AccessByRole(ui.RoleTreeItem, "Spam, The server refused the login"); !spam.Selected {
		t.Errorf("Spam = %+v, want selected", spam)
	}
}

func TestSourceListSpokenReplacesTheStateRead(t *testing.T) {
	l := mailList()
	row := l.Row
	l.Row = func(item string) SourceRow {
		r := row(item)
		if item == "Inbox" {
			r.Spoken = "12 unread"
		}
		return r
	}
	tt := listTester(l, mailSections())
	if _, ok := tt.AccessByRole(ui.RoleTreeItem, "Inbox, 12 unread"); !ok {
		t.Errorf("no tree item reads the spoken count")
	}
}

func TestSourceListBuildsOnlyTheRowsInView(t *testing.T) {
	items := make([]int, 10000)
	for i := range items {
		items[i] = i
	}
	rows := 0
	l := &SourceList[int]{
		Label:    "Files",
		Children: func(int) []int { return nil },
		Row: func(i int) SourceRow {
			rows++
			return SourceRow{Title: fmt.Sprintf("Note %d", i)}
		},
	}
	sections := []SourceSection[int]{{ID: "v", Title: "Vault", Items: items}}
	tt := ui.NewTester(func(c *ui.Context) {
		Adapt(c)
		ui.Column(c).Width(280).Height(400).AlignItems(ui.Stretch).Children(func() { l.View(c, sections).Grow(1) })
	}, 700, 420)
	tt.Frame()
	rows = 0
	tt.Frame()
	if rows == 0 || rows > 500 {
		t.Errorf("a frame asked for %d rows of 10000", rows)
	}
	l.Select("v", 9000)
	tt.Frame()
	if !tt.HasText("Note 9000") {
		t.Errorf("Select did not scroll to a far row: %q", tt.Texts())
	}
}

func TestSourceListClickIsIndependentOfKeyboardSelection(t *testing.T) {
	l := mailList()
	var clicked []string
	l.OnClick = func(item string, _ ui.Modifiers) { clicked = append(clicked, item) }
	tt := listTester(l, mailSections())
	if err := tt.Click("Inbox"); err != nil {
		t.Fatal(err)
	}
	tt.Move(1000, 1000)
	tt.Key(0, ui.KeyDown)
	if !slices.Equal(clicked, []string{"Inbox"}) {
		t.Fatalf("clicks %v", clicked)
	}
}

func TestSourceListClickIncludesEnter(t *testing.T) {
	l := mailList()
	var clicked []string
	l.OnActivate = nil
	l.OnClick = func(item string, _ ui.Modifiers) { clicked = append(clicked, item) }
	tt := listTester(l, mailSections())
	if err := tt.Click("Inbox"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEnter)
	if !slices.Equal(clicked, []string{"Inbox", "Inbox"}) {
		t.Fatalf("click and Enter: %v", clicked)
	}
}

func TestSourceListMiddleClick(t *testing.T) {
	l := mailList()
	var clicked string
	l.OnMiddleClick = func(item string, _ ui.Modifiers) { clicked = item }
	tt := listTester(l, mailSections())
	r, ok := tt.Find("Inbox")
	if !ok {
		t.Fatal("no Inbox")
	}
	tt.PressButton(2, r.X+r.W/2, r.Y+r.H/2)
	tt.ReleaseButton(2, r.X+r.W/2, r.Y+r.H/2)
	if clicked != "Inbox" {
		t.Fatalf("middle click %q", clicked)
	}
}

func TestSourceListOpenGeneration(t *testing.T) {
	l := mailList()
	tt := listTester(l, mailSections())
	before := l.OpenGeneration()
	tt.Frame()
	if l.OpenGeneration() != before {
		t.Fatal("idle frame changed open state")
	}
	l.SetOpen("account", "inbox", true)
	if l.OpenGeneration() == before {
		t.Fatal("opening did not change generation")
	}
}
