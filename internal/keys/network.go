package keys

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/config"
)

type NetworkKeyMap struct {
	Select       key.Binding
	Refresh      key.Binding
	Forget       key.Binding
	ToggleAuto   key.Binding
	RevealSecret key.Binding
	Known        key.Binding
	Info         key.Binding
	Disconnect   key.Binding
	ToggleRadio  key.Binding
}

func newNetworkKeyMap(cfg config.NetworkKeys) NetworkKeyMap {
	return NetworkKeyMap{
		Select:       binding(cfg.Select, "select"),
		Refresh:      binding(cfg.Refresh, "refresh"),
		Forget:       binding(cfg.Forget, "forget"),
		ToggleAuto:   binding(cfg.ToggleAuto, "toggle autoconnect"),
		RevealSecret: binding(cfg.RevealSecret, "show/hide password"),
		Known:        binding(cfg.Known, "known networks"),
		Info:         binding(cfg.Info, "network info"),
		Disconnect:   binding(cfg.Disconnect, "disconnect"),
		ToggleRadio:  binding(cfg.ToggleRadio, "toggle wifi"),
	}
}

func (n NetworkKeyMap) Bindings() []key.Binding {
	return []key.Binding{
		n.Select, n.Refresh, n.Known, n.Info, n.Disconnect, n.ToggleRadio,
		n.ToggleAuto, n.RevealSecret, n.Forget,
	}
}
