package power

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/power"
)

const tickInterval = 5 * time.Second

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func checkCmd() tea.Cmd {
	return func() tea.Msg { return checkMsg{err: power.CheckAvailable()} }
}

func fetchProfilesCmd() tea.Cmd {
	return func() tea.Msg {
		profiles, err := power.ListProfiles()
		return profilesMsg{profiles: profiles, err: err}
	}
}

func setProfileCmd(name string) tea.Cmd {
	return func() tea.Msg { return setProfileMsg{err: power.SetActive(name)} }
}

func fetchBatteryCmd() tea.Cmd {
	return func() tea.Msg {
		batteries, err := power.ReadBatteries()
		if err != nil {
			return batteryMsg{err: err}
		}
		ac, err := power.ReadAC()
		return batteryMsg{batteries: batteries, ac: ac, err: err}
	}
}
