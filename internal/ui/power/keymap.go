package power

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/keys"
)

// powerKeyMap renders the bottom help bar.
type powerKeyMap struct {
	m     Model
	width int
}

func (k powerKeyMap) ShortHelp() []key.Binding {
	g := keys.Keys.Global
	p := keys.Keys.Power
	if k.m.checkErr != nil {
		return []key.Binding{p.Refresh, g.Help}
	}
	return []key.Binding{g.Up, g.Down, p.Select, p.Refresh, g.Help}
}

func (k powerKeyMap) FullHelp() [][]key.Binding {
	g := keys.Keys.Global
	p := keys.Keys.Power

	var bindings []key.Binding
	if k.m.checkErr != nil {
		bindings = []key.Binding{p.Refresh}
	} else {
		bindings = []key.Binding{g.Up, g.Down, p.Select, p.Refresh}
	}
	bindings = append(bindings, g.CommonBindings()...)
	return keys.ChunkByWidth(bindings, k.width)
}
