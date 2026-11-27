package pages

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/notify"
)

var KeyLabel func(id string) string

func keyOf(id, fallback string) string {
	if KeyLabel != nil {
		if k := KeyLabel(id); k != "" {
			return k
		}
	}
	return fallback
}

func copyText(text, what string) tea.Cmd {
	return tea.Batch(tea.SetClipboard(text), notify.Success("Copied "+what))
}
