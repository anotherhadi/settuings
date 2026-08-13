package input

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/keys"
)

// inputKeyMap renders the bottom help bar, adapting to which section
// currently has focus.
type inputKeyMap struct {
	m     Model
	width int
}

// relabel reuses an existing binding's keys under a different help
// description, same trick the audio page uses for its volume keys.
func relabel(b key.Binding, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(b.Keys()...), key.WithHelp(b.Help().Key, desc))
}

func (k inputKeyMap) ShortHelp() []key.Binding {
	g := keys.Keys.Global
	in := keys.Keys.Input
	if k.m.checkErr != nil {
		return []key.Binding{g.Help}
	}
	switch k.m.focus {
	case focusMouse:
		speedDown, speedUp := relabel(g.Left, "speed down"), relabel(g.Right, "speed up")
		return []key.Binding{speedDown, speedUp, in.Reset, g.CycleFocus, g.Help}
	default:
		return []key.Binding{g.Up, g.Down, in.Select, in.Reset, g.CycleFocus, g.Help}
	}
}

func (k inputKeyMap) FullHelp() [][]key.Binding {
	g := keys.Keys.Global
	in := keys.Keys.Input

	var bindings []key.Binding
	switch {
	case k.m.checkErr != nil:
		bindings = nil
	case k.m.focus == focusMouse:
		speedDown, speedUp := relabel(g.Left, "speed down"), relabel(g.Right, "speed up")
		bindings = []key.Binding{speedDown, speedUp, in.Reset, g.CycleFocus}
	default:
		bindings = []key.Binding{g.Up, g.Down, in.Select, in.Reset, g.CycleFocus}
	}
	bindings = append(bindings, g.CommonBindings()...)
	return keys.ChunkByWidth(bindings, k.width)
}
