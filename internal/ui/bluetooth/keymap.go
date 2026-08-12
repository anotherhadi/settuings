package bluetooth

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/keys"
)

// bluetoothKeyMap renders the bottom help bar, adapting to whether the
// confirm-forget prompt or the device info screen is open.
type bluetoothKeyMap struct {
	m     Model
	width int
}

func (k bluetoothKeyMap) ShortHelp() []key.Binding {
	g := keys.Keys.Global
	b := keys.Keys.Bluetooth

	switch {
	case k.m.list.confirmRemove:
		return []key.Binding{b.Forget, g.Escape}
	case k.m.detail.open:
		return []key.Binding{g.Up, g.Down, b.Select, b.ToggleTrust, g.Escape, g.Help}
	default:
		return []key.Binding{g.Up, g.Down, b.Select, b.Refresh, b.Info, g.Help}
	}
}

func (k bluetoothKeyMap) FullHelp() [][]key.Binding {
	g := keys.Keys.Global
	b := keys.Keys.Bluetooth

	var bindings []key.Binding
	switch {
	case k.m.list.confirmRemove:
		bindings = []key.Binding{b.Forget, g.Escape}
	case k.m.detail.open:
		bindings = []key.Binding{g.Up, g.Down, g.ScrollUp, g.ScrollDown, b.Select, b.ToggleTrust, g.Escape}
	default:
		bindings = []key.Binding{g.Up, g.Down, b.Select, b.Refresh, b.Info, b.ToggleTrust, b.Forget, b.TogglePower}
	}
	bindings = append(bindings, g.CommonBindings()...)
	return keys.ChunkByWidth(bindings, k.width)
}
