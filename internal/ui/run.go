package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/config"
)

func Run(cfg *config.Config) error {
	theme.Set(cfg.Theme)
	m, err := New(cfg)
	if err != nil {
		return err
	}
	defer m.Close()
	_, err = tea.NewProgram(m).Run()
	return err
}
