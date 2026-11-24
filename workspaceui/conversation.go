package workspaceui

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui/chat"
)

// Conversation uses the gallery's scroll-following transcript, jump control and glass composer.
// A caller can substitute a read-only footer while retaining the same transcript.
type Conversation struct {
	State      *chat.ConversationState
	Transcript []core.Node
	Footer     core.Node
	// FooterMode selects an application-owned decision or read-only footer.
	// The empty mode uses Footer. Reactive updates within a mode keep its editor.
	FooterMode func() string
	Alternate  func(*core.Scope, string) core.Node
}

func (v Conversation) Render(cx *core.Scope) core.Node {
	footer := v.Footer
	if v.FooterMode != nil && v.Alternate != nil {
		mode := core.NewMemo(cx, v.FooterMode)
		footer = core.Dynamic(func(cx *core.Scope) core.Node {
			if key := mode.Get(); key != "" {
				return v.Alternate(cx, key)
			}
			return v.Footer
		})
	}
	return chat.FloatingConversation{State: v.State, ContentWidth: 810, Children: v.Transcript, Composer: []core.Node{footer}, JumpLabel: "Jump to latest"}
}

// Message renders selectable, searchable Markdown in the gallery's incoming or outgoing layout.
func Message(outgoing bool, body core.Node, extras ...core.Node) core.Node {
	if outgoing {
		body = chat.Bubble{Children: []core.Node{body}}
	}
	return chat.Message{Outgoing: outgoing, Children: append([]core.Node{body}, extras...)}
}

// ConversationEntries mounts each application-supplied item once per key.
// Reveal names a mounted item to scroll to after layout. The caller owns
// paging, ordering and the meaning of each entry.
type ConversationEntries[T any] struct {
	Items      func() []T
	Key        func(T) string
	Content    func(*core.Scope, func() T) core.Node
	Reveal     func() string
	OnRevealed func(string)
}

func (v ConversationEntries[T]) Render(*core.Scope) core.Node {
	return core.Each(v.Items, v.Key, func(cx *core.Scope, item func() T, _ func() int) core.Node {
		child := v.Content(cx, item)
		if v.Reveal == nil {
			return child
		}
		key := v.Key(core.Untracked(cx, item))
		ref := &core.Ref{}
		reveal := func() {
			if cx.Disposed() || core.Untracked(cx, v.Reveal) != key {
				return
			}
			if _, ok := ref.Bounds(); !ok {
				return
			}
			ref.Reveal(core.RevealOptions{Vertical: core.ScrollAlignCenter})
			if v.OnRevealed != nil {
				v.OnRevealed(key)
			}
		}
		cx.Effect(func() {
			if v.Reveal() == key {
				cx.Post(reveal)
			}
		})
		return core.View{Ref: ref, OnLayout: func(core.Bounds) {
			if core.Untracked(cx, v.Reveal) == key {
				cx.Post(reveal)
			}
		}, Children: []core.Node{child}}
	})
}
