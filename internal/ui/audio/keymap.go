package audio

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/keys"
)

// audioKeyMap renders the bottom help bar, adapting to the active tab.
type audioKeyMap struct {
	m     Model
	width int
}

// relabel reuses an existing binding's keys under a different help
// description. Volume adjustment reuses the global Left/Right binding (h/l,
// arrow keys), which is normally labelled "scroll left/right" — here it
// means something else, so the help bar needs its own wording.
func relabel(b key.Binding, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(b.Keys()...), key.WithHelp(b.Help().Key, desc))
}

func (k audioKeyMap) ShortHelp() []key.Binding {
	g := keys.Keys.Global
	a := keys.Keys.Audio
	volDown, volUp := relabel(g.Left, "vol down"), relabel(g.Right, "vol up")

	switch k.m.active {
	case tabApps:
		return []key.Binding{g.Up, g.Down, volDown, volUp, a.Mute, g.CycleFocus, g.Help}
	default:
		return []key.Binding{g.Up, g.Down, volDown, volUp, a.Mute, a.Test, g.Help}
	}
}

func (k audioKeyMap) FullHelp() [][]key.Binding {
	g := keys.Keys.Global
	a := keys.Keys.Audio
	volDown, volUp := relabel(g.Left, "volume down"), relabel(g.Right, "volume up")

	var bindings []key.Binding
	switch k.m.active {
	case tabOutput:
		bindings = []key.Binding{g.Up, g.Down, volDown, volUp, a.Select, a.Mute, a.Test, a.Refresh, g.CycleFocus}
	case tabInput:
		bindings = []key.Binding{g.Up, g.Down, volDown, volUp, a.Select, a.Mute, relabel(a.Test, "test mic"), a.Refresh, g.CycleFocus}
	case tabApps:
		bindings = []key.Binding{g.Up, g.Down, volDown, volUp, a.Mute, a.Refresh, g.CycleFocus}
	}
	bindings = append(bindings, g.CommonBindings()...)
	return keys.ChunkByWidth(bindings, k.width)
}
