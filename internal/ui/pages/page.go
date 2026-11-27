package pages

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/theme"
)

type Page interface {
	Available() bool
	Name() string
	Label() string
	Init() tea.Cmd
	Refresh(manual bool) tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	Commands() []commands.Command
	View(w, h int) string
	Mouse(msg tea.MouseMsg) tea.Cmd
	Leave()
	Close()
}

type DoneMsg struct {
	Page string
	OK   string
	Err  error
}

func act(page, ok string, fn func() error) tea.Cmd {
	return func() tea.Msg { return DoneMsg{Page: page, OK: ok, Err: fn()} }
}

func icon(nerd, plain string) string { return theme.Icon(nerd, plain) }
