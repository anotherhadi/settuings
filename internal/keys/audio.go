package keys

import (
	"charm.land/bubbles/v2/key"
	"github.com/anotherhadi/settuings/internal/config"
)

type AudioKeyMap struct {
	Select  key.Binding
	Refresh key.Binding
	Mute    key.Binding
	Test    key.Binding
}

func newAudioKeyMap(cfg config.AudioKeys) AudioKeyMap {
	return AudioKeyMap{
		Select:  binding(cfg.Select, "set as default"),
		Refresh: binding(cfg.Refresh, "refresh"),
		Mute:    binding(cfg.Mute, "mute/unmute"),
		Test:    binding(cfg.Test, "test"),
	}
}

func (a AudioKeyMap) Bindings() []key.Binding {
	return []key.Binding{a.Select, a.Refresh, a.Mute, a.Test}
}
