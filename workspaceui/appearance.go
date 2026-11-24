package workspaceui

import (
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/ui"
)

// AppearanceChoice is the setting that shows an app light, dark or as the
// system does: a segmented control of the three. A press shows the choice
// at once (Scope.SetAppearance) and then calls OnChange, where the app
// saves it. At its start an app sets a saved choice itself, with
// SetAppearance, after Apply.
type AppearanceChoice struct {
	// Value is the appearance chosen; nil is the system's.
	Value    func() core.Appearance
	OnChange func(core.Appearance)
}

func (a AppearanceChoice) Render(cx *core.Scope) core.Node {
	return ui.Tabs[core.Appearance]{
		Variant: ui.TabsSegmented, Label: "Appearance",
		Options: []ui.Option[core.Appearance]{
			{Value: core.AppearanceSystem, Label: "System"},
			{Value: core.AppearanceLight, Label: "Light"},
			{Value: core.AppearanceDark, Label: "Dark"},
		},
		Value: func() core.Appearance {
			if a.Value == nil {
				return core.AppearanceSystem
			}
			return a.Value()
		},
		OnChange: func(chosen core.Appearance) {
			cx.SetAppearance(chosen)
			if a.OnChange != nil {
				a.OnChange(chosen)
			}
		},
	}
}
