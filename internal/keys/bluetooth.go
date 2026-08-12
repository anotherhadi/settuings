package keys

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/config"
)

type BluetoothKeyMap struct {
	Select      key.Binding
	Refresh     key.Binding
	Forget      key.Binding
	ToggleTrust key.Binding
	Info        key.Binding
	TogglePower key.Binding
}

func newBluetoothKeyMap(cfg config.BluetoothKeys) BluetoothKeyMap {
	return BluetoothKeyMap{
		Select:      binding(cfg.Select, "connect/pair"),
		Refresh:     binding(cfg.Refresh, "scan"),
		Forget:      binding(cfg.Forget, "forget"),
		ToggleTrust: binding(cfg.ToggleTrust, "toggle trust"),
		Info:        binding(cfg.Info, "device info"),
		TogglePower: binding(cfg.TogglePower, "toggle bluetooth"),
	}
}

func (b BluetoothKeyMap) Bindings() []key.Binding {
	return []key.Binding{
		b.Select, b.Refresh, b.Info, b.ToggleTrust, b.Forget, b.TogglePower,
	}
}
