package style

import (
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

type Styles struct {
	PanelEditing lipgloss.Style
}

var S *Styles

func Init() {
	S = &Styles{
		PanelEditing: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ilovetui.S.Base0E),
	}
}
