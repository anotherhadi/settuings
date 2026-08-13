package keys

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/config"
)

type InputKeyMap struct {
	Select key.Binding
	Reset  key.Binding
	Filter key.Binding
}

func newInputKeyMap(cfg config.InputKeys) InputKeyMap {
	return InputKeyMap{
		Select: binding(cfg.Select, "apply layout"),
		Reset:  binding(cfg.Reset, "reset to config"),
		Filter: binding(cfg.Filter, "filter"),
	}
}

func (i InputKeyMap) Bindings() []key.Binding {
	return []key.Binding{i.Select, i.Reset, i.Filter}
}
