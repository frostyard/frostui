package frostui

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/richtext"
)

// paletteOverEditor draws a shell around build with a table whose Editing
// is the handle bound to the editor, or nil without one.
func paletteOverEditor(s *Shell, h *ui.Handle, build func(c *ui.Context) ui.Element) *ui.Tester {
	table := NewCommandTable()
	if h != nil {
		table.Editing = func(c *ui.Context) bool { return c.FocusWithin(*h) }
	}
	return ui.NewTester(func(c *ui.Context) {
		s.View(c, Frame{Table: table, Content: func() {
			e := build(c).Height(200).AutoFocus()
			if h != nil {
				e.Bind(h)
			}
		}})
	}, 1000, 700)
}

func TestCtrlKYieldsToTheRichTextEditor(t *testing.T) {
	for _, named := range []bool{true, false} {
		var s Shell
		var h ui.Handle
		handle := &h
		if !named {
			handle = nil
		}
		linked := 0
		doc := richtext.NewDocument(richtext.NewText("Hello"))
		tt := paletteOverEditor(&s, handle, func(c *ui.Context) ui.Element {
			return richtext.Editor(c, doc, richtext.EditorOptions{OnLinkCommand: func() { linked++ }})
		})
		tt.Frame()
		tt.Key(ui.Cmd, ui.KeyK)
		if s.Palette.Open || linked != 1 {
			t.Errorf("named %v: Ctrl K: palette %v, link command ran %d times", named, s.Palette.Open, linked)
		}
		tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
		if !s.Palette.Open {
			t.Errorf("named %v: Ctrl+Shift+P did not open the palette", named)
		}
	}
}

func TestCtrlKYieldsToTheCodeEditor(t *testing.T) {
	for _, named := range []bool{true, false} {
		var s Shell
		var h ui.Handle
		handle := &h
		if !named {
			handle = nil
		}
		st := &editor.State{Buffer: editor.NewBuffer("x"), Bindings: []editor.Binding{
			{Mods: ui.Cmd, Key: ui.KeyK, Command: editor.DuplicateSelection},
		}}
		tt := paletteOverEditor(&s, handle, func(c *ui.Context) ui.Element { return editor.View(c, st) })
		tt.Frame()
		tt.Key(ui.Cmd, ui.KeyK)
		tt.Frame()
		if s.Palette.Open || st.Buffer.String() == "x" {
			t.Errorf("named %v: Ctrl K: palette %v, buffer %q", named, s.Palette.Open, st.Buffer.String())
		}
		tt.Key(ui.Cmd|ui.Shift, ui.KeyP)
		if !s.Palette.Open {
			t.Errorf("named %v: Ctrl+Shift+P did not open the palette", named)
		}
	}
}
