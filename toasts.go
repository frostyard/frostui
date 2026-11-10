package frostui

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

// ToastWidth is the widest a TopToasts toast grows, in DIPs.
const ToastWidth = 640

// TopToasts gives a window's toasts their look: at the top of the window,
// centered, each with a close button named closeLabel (translated by the
// app). MyGo's own toasts sit at the bottom, where they cover a wizard's
// Back and Next, and have no close button, so a toast that stays until
// closed would stay for good; resting the pointer on a toast also stops
// its time. A toast of Type "error" leads with a warning icon, and one
// made by ShowActionToast has its action button and its countdown. Plain
// text only (the Title). Call it once per frame, as ui.ToastViewportBase.
func TopToasts(c *ui.Context, closeLabel string) { toastsAt(c, closeLabel, SpaceM) }

// toastsAt is TopToasts with the first toast top DIPs from the top of the
// window.
func toastsAt(c *ui.Context, closeLabel string, top float32) {
	t := c.Theme()
	fill, text := t.Inverse, t.InverseText
	if fill.A == 0 {
		fill, text = t.Text, t.Background
	}
	ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Justify(ui.Start).AlignItems(ui.Center).Padding(top, SpaceXL, SpaceM, SpaceXL).Gap(SpaceS)
		for _, toast := range toasts {
			p := ui.ToastBase(c, toast)
			p.Root.Row().Gap(SpaceM).Padding(8, 8, 8, 16).Radius(10).MaxWidth(ToastWidth).AlignItems(ui.Center).
				Background(fill).TextColor(text).Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, 0.25))
			p.Root.Children(func() {
				if toast.Type == "error" {
					Icon(c, "dialog-warning").TextColor(t.Warning)
				}
				ui.Text(c, toast.Title).FontSize(13).Grow(1).Shrink(1)
				if action, ok := toast.Data.(actionToastLook); ok {
					if action.countdown {
						countdown(c, p.Left(), text)
					}
					if toast.Action != "" {
						p.ActionButton().Padding(6, 10).Radius(6).TextColor(t.Accent.Mix(text, 0.35)).Children(func() {
							ui.Text(c, toast.Action).FontSize(13).Bold().SingleLine()
						})
					}
				}
				p.CloseButton().Label(closeLabel).Tooltip(closeLabel).Padding(6).Radius(6).Children(func() {
					Icon(c, "window-close").TextColor(text)
				})
			})
		}
	})
}

// actionToastTime is how long an action toast shows when it does not say.
const actionToastTime = 8 * time.Second

// ActionToast is a toast with an action, as Undo after a delete, and an
// optional countdown, as Undo Send's seconds. Exactly one of OnAction and
// OnExpire runs.
type ActionToast struct {
	// Title says what happened.
	Title string
	// Action labels the button.
	Action string
	// Type is the toast's kind: "error" leads with a warning icon.
	Type string
	// Timeout is how long the toast shows: 8 seconds when 0, until it is
	// closed when negative.
	Timeout time.Duration
	// Countdown shows the seconds left before the button.
	Countdown bool
	// OnAction runs on the UI thread when the user chooses the action.
	OnAction func()
	// OnExpire runs on the UI thread when the toast closes without the
	// action: its time ended, or the user or Context.CloseToast closed it.
	OnExpire func()
}

// actionToastLook is the Data of a toast ShowActionToast made, which tells
// TopToasts to draw its action.
type actionToastLook struct{ countdown bool }

// ShowActionToast shows t and returns its ID, for Context.CloseToast. Each
// call shows a toast of its own, so one never replaces another and loses
// its OnExpire. Draw the window's toasts with TopToasts.
func ShowActionToast(c *ui.Context, t ActionToast) string {
	timeout := t.Timeout
	if timeout == 0 {
		timeout = actionToastTime
	}
	acted := false
	return c.AddToast(ui.Toast{
		Title: t.Title, Type: t.Type, Timeout: timeout, Action: t.Action,
		Data: actionToastLook{countdown: t.Countdown && timeout > 0},
		OnAction: func() {
			acted = true
			if t.OnAction != nil {
				t.OnAction()
			}
		},
		OnClose: func() {
			if !acted && t.OnExpire != nil {
				t.OnExpire()
			}
		},
	})
}

// countdown draws the whole seconds left, rounded up, and asks for the
// frame that changes them.
func countdown(c *ui.Context, left time.Duration, color ui.Color) {
	if left <= 0 {
		return
	}
	seconds := int((left + time.Second - 1) / time.Second)
	label := fmt.Sprintf("%d seconds left", seconds)
	if seconds == 1 {
		label = "1 second left"
	}
	ui.Text(c, fmt.Sprintf("%ds", seconds)).FontSize(13).TextColor(color.Alpha(0.75)).SingleLine().Role(ui.RoleTimer).Label(label)
	c.After(left - time.Duration(seconds-1)*time.Second)
}
