package keys

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/config"
)

type PowerKeyMap struct {
	Select  key.Binding
	Refresh key.Binding
}

func newPowerKeyMap(cfg config.PowerKeys) PowerKeyMap {
	return PowerKeyMap{
		Select:  binding(cfg.Select, "apply profile"),
		Refresh: binding(cfg.Refresh, "refresh"),
	}
}

func (p PowerKeyMap) Bindings() []key.Binding {
	return []key.Binding{p.Select, p.Refresh}
}
