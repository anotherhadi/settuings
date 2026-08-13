package app

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/icons"
	aboutUI "github.com/anotherhadi/settuings/internal/ui/about"
	audioUI "github.com/anotherhadi/settuings/internal/ui/audio"
	bluetoothUI "github.com/anotherhadi/settuings/internal/ui/bluetooth"
	networkUI "github.com/anotherhadi/settuings/internal/ui/network"
	powerUI "github.com/anotherhadi/settuings/internal/ui/power"
)

type page string

const (
	pageAbout     page = "About"
	pageNetwork   page = "Network"
	pageBluetooth page = "Bluetooth"
	pageAudio     page = "Audio"
	pagePower     page = "Power"
)

// pageEntry describes a page and all its integration hooks.
type pageEntry struct {
	id   page
	icon func() string

	// render returns the page's view content. nil = show "empty".
	render func(m *Model) string
	// update is called when this page is active. nil = no-op.
	update func(m *Model, msg tea.Msg) tea.Cmd
	// isEditing reports whether the page is in text-editing mode.
	isEditing func(m *Model) bool
	// resize propagates a new (w, h) to the page model.
	resize func(m *Model, w, h int)
	// hasUpdate reports whether the page has unseen updates.
	hasUpdate func(m *Model) bool
	// activate is called once when the user navigates onto this page.
	// nil = no-op.
	activate func(m *Model) tea.Cmd
	// deactivate is called synchronously once when the user navigates away
	// from this page (or quits the app). nil = no-op. Unlike activate, this
	// runs synchronously rather than as a tea.Cmd: pages that hold a
	// background OS process (e.g. Audio's live input-level meter) need it
	// torn down immediately, not racing an async command against exit.
	deactivate func(m *Model)
}

var pageRegistry = []pageEntry{
	{
		id:   pageAbout,
		icon: func() string { return icons.I.About },

		render: func(m *Model) string { return m.about.View().Content },
		update: func(m *Model, msg tea.Msg) tea.Cmd {
			updated, cmd := m.about.Update(msg)
			m.about = updated.(aboutUI.Model)
			return cmd
		},
		isEditing: func(m *Model) bool { return m.about.IsEditing() },
		resize:    func(m *Model, w, h int) { m.about.SetSize(w, h) },
	},
	{
		id:   pageNetwork,
		icon: func() string { return icons.I.Network },

		render: func(m *Model) string { return m.network.View().Content },
		update: func(m *Model, msg tea.Msg) tea.Cmd {
			updated, cmd := m.network.Update(msg)
			m.network = updated.(networkUI.Model)
			return cmd
		},
		isEditing: func(m *Model) bool { return m.network.IsEditing() },
		resize:    func(m *Model, w, h int) { m.network.SetSize(w, h) },
		activate:  func(m *Model) tea.Cmd { return m.network.Activate() },
	},
	{
		id:   pageBluetooth,
		icon: func() string { return icons.I.Bluetooth },

		render: func(m *Model) string { return m.bluetooth.View().Content },
		update: func(m *Model, msg tea.Msg) tea.Cmd {
			updated, cmd := m.bluetooth.Update(msg)
			m.bluetooth = updated.(bluetoothUI.Model)
			return cmd
		},
		isEditing: func(m *Model) bool { return m.bluetooth.IsEditing() },
		resize:    func(m *Model, w, h int) { m.bluetooth.SetSize(w, h) },
		activate:  func(m *Model) tea.Cmd { return m.bluetooth.Activate() },
	},
	{
		id:   pageAudio,
		icon: func() string { return icons.I.Audio },

		render: func(m *Model) string { return m.audio.View().Content },
		update: func(m *Model, msg tea.Msg) tea.Cmd {
			updated, cmd := m.audio.Update(msg)
			m.audio = updated.(audioUI.Model)
			return cmd
		},
		isEditing:  func(m *Model) bool { return m.audio.IsEditing() },
		resize:     func(m *Model, w, h int) { m.audio.SetSize(w, h) },
		activate:   func(m *Model) tea.Cmd { return m.audio.Activate() },
		deactivate: func(m *Model) { m.audio.Deactivate() },
	},
	{
		id:   pagePower,
		icon: func() string { return icons.I.Power },

		render: func(m *Model) string { return m.power.View().Content },
		update: func(m *Model, msg tea.Msg) tea.Cmd {
			updated, cmd := m.power.Update(msg)
			m.power = updated.(powerUI.Model)
			return cmd
		},
		isEditing: func(m *Model) bool { return m.power.IsEditing() },
		resize:    func(m *Model, w, h int) { m.power.SetSize(w, h) },
		activate:  func(m *Model) tea.Cmd { return m.power.Activate() },
	},
}

// PageNames returns the valid page identifiers, in sidebar order, usable
// with the --page flag.
func PageNames() []string {
	names := make([]string, len(pageRegistry))
	for i, e := range pageRegistry {
		names[i] = string(e.id)
	}
	return names
}

// lookupPage resolves a case-insensitive page name (as passed on the
// command line) to its registry id.
func lookupPage(name string) (page, bool) {
	for _, e := range pageRegistry {
		if strings.EqualFold(string(e.id), name) {
			return e.id, true
		}
	}
	return "", false
}

// visiblePages returns pageRegistry entries not listed in the user's
// tui.hidden_pages config, preserving registry order. A page hidden this
// way is still reachable via --page: hiding only affects the sidebar and
// its numbered shortcuts.
func visiblePages() []pageEntry {
	hiddenNames := config.Global.TUI.HiddenPages
	if len(hiddenNames) == 0 {
		return pageRegistry
	}

	hidden := make(map[page]bool, len(hiddenNames))
	for _, name := range hiddenNames {
		if p, ok := lookupPage(name); ok {
			hidden[p] = true
		}
	}
	if len(hidden) == 0 {
		return pageRegistry
	}

	visible := make([]pageEntry, 0, len(pageRegistry))
	for _, e := range pageRegistry {
		if !hidden[e.id] {
			visible = append(visible, e)
		}
	}
	return visible
}
