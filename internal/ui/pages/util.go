package pages

import (
	"strconv"

	tea "charm.land/bubbletea/v2"
)

func itoa(n int) string { return strconv.Itoa(n) }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

type pending struct{ on bool }

func (g *pending) do(cmd tea.Cmd) tea.Cmd {
	if g.on {
		return nil
	}
	g.on = true
	return cmd
}

func (g *pending) clear(msg tea.Msg, page string) {
	if d, ok := msg.(DoneMsg); ok && d.Page == page {
		g.on = false
	}
}
