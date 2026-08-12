package network

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/keys"
)

// networkKeyMap renders the bottom help bar, adapting to the active tab and
// whether a sub-view (password prompt, device info, known-network detail)
// is open.
type networkKeyMap struct {
	m     Model
	width int
}

func (k networkKeyMap) ShortHelp() []key.Binding {
	g := keys.Keys.Global
	n := keys.Keys.Network

	switch {
	case k.m.wifi.connecting:
		return []key.Binding{n.Select, g.Escape, g.Help}
	case k.m.deviceDetail.open && k.m.deviceDetail.kind == tabWifi:
		return []key.Binding{g.Up, g.Down, n.RevealSecret, n.Disconnect, g.Escape, g.Help}
	case k.m.deviceDetail.open:
		return []key.Binding{g.Up, g.Down, n.Disconnect, g.Escape, g.Help}
	case k.m.active == tabKnown && k.m.known.detail:
		return []key.Binding{n.RevealSecret, n.ToggleAuto, n.Forget, g.Escape, g.Help}
	case k.m.active == tabKnown:
		return []key.Binding{g.Up, g.Down, n.Select, n.Known, g.Help}
	case k.m.active == tabWifi:
		return []key.Binding{g.Up, g.Down, n.Select, n.Disconnect, n.ToggleRadio, g.Help}
	default:
		return []key.Binding{g.Up, g.Down, n.Select, g.CycleFocus, n.Known, g.Help}
	}
}

func (k networkKeyMap) FullHelp() [][]key.Binding {
	g := keys.Keys.Global
	n := keys.Keys.Network

	var bindings []key.Binding
	switch {
	case k.m.wifi.connecting:
		bindings = []key.Binding{n.Select, g.Escape}
	case k.m.deviceDetail.open && k.m.deviceDetail.kind == tabWifi:
		bindings = []key.Binding{g.Up, g.Down, g.ScrollUp, g.ScrollDown, n.RevealSecret, n.Disconnect, g.Escape}
	case k.m.deviceDetail.open:
		bindings = []key.Binding{g.Up, g.Down, g.ScrollUp, g.ScrollDown, n.Disconnect, g.Escape}
	case k.m.active == tabKnown && k.m.known.detail:
		bindings = []key.Binding{n.RevealSecret, n.ToggleAuto, n.Forget, g.Escape}
	case k.m.active == tabKnown:
		bindings = []key.Binding{g.Up, g.Down, n.Select, n.Refresh, n.Known}
	case k.m.active == tabWifi:
		bindings = []key.Binding{g.Up, g.Down, n.Select, n.Refresh, n.Info, n.Disconnect, n.ToggleRadio, g.CycleFocus, n.Known}
	case k.m.active == tabEthernet:
		bindings = []key.Binding{g.Up, g.Down, n.Select, n.Refresh, n.Info, g.CycleFocus, n.Known}
	default:
		bindings = []key.Binding{g.Up, g.Down, n.Select, n.Refresh, g.CycleFocus, n.Known}
	}
	bindings = append(bindings, g.CommonBindings()...)
	return keys.ChunkByWidth(bindings, k.width)
}
