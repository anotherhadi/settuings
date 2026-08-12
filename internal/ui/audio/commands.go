package audio

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/audio"
	notificationsUI "github.com/anotherhadi/settuings/internal/ui/components/notifications"
)

const tickInterval = 5 * time.Second

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func checkCmd() tea.Cmd {
	return func() tea.Msg { return checkMsg{err: audio.CheckAvailable()} }
}

func fetchOutputCmd() tea.Cmd {
	return func() tea.Msg {
		devices, err := audio.ListSinks()
		return outputListMsg{devices: devices, err: err}
	}
}

func fetchInputCmd() tea.Cmd {
	return func() tea.Msg {
		devices, err := audio.ListSources()
		return inputListMsg{devices: devices, err: err}
	}
}

func fetchAppsCmd() tea.Cmd {
	return func() tea.Msg {
		streams, err := audio.ListStreams()
		return appsListMsg{streams: streams, err: err}
	}
}

func outputSetVolumeCmd(id string, pct int) tea.Cmd {
	return func() tea.Msg { return outputActionMsg{err: audio.SetVolume(id, pct)} }
}

func outputToggleMuteCmd(id string) tea.Cmd {
	return func() tea.Msg { return outputActionMsg{err: audio.ToggleMute(id)} }
}

func outputSetDefaultCmd(id string) tea.Cmd {
	return func() tea.Msg { return outputActionMsg{err: audio.SetDefault(id)} }
}

func outputTestToneCmd(sinkID string) tea.Cmd {
	return func() tea.Msg { return outputActionMsg{err: audio.PlayTestTone(sinkID)} }
}

func inputSetVolumeCmd(id string, pct int) tea.Cmd {
	return func() tea.Msg { return inputActionMsg{err: audio.SetVolume(id, pct)} }
}

func inputToggleMuteCmd(id string) tea.Cmd {
	return func() tea.Msg { return inputActionMsg{err: audio.ToggleMute(id)} }
}

func inputSetDefaultCmd(id string) tea.Cmd {
	return func() tea.Msg { return inputActionMsg{err: audio.SetDefault(id)} }
}

func appsSetVolumeCmd(id string, pct int) tea.Cmd {
	return func() tea.Msg { return appsActionMsg{err: audio.SetVolume(id, pct)} }
}

func appsToggleMuteCmd(id string) tea.Cmd {
	return func() tea.Msg { return appsActionMsg{err: audio.ToggleMute(id)} }
}

func notifyCmd(title, body string) tea.Cmd {
	return func() tea.Msg {
		return notificationsUI.NotificationMsg{Title: title, Body: body, Kind: notificationsUI.KindInfo}
	}
}

func startLevelMonitorCmd(sourceID string) tea.Cmd {
	return func() tea.Msg {
		mon, err := audio.StartLevelMonitor(sourceID)
		return levelStartedMsg{monitor: mon, err: err}
	}
}

func waitLevelCmd(mon *audio.LevelMonitor) tea.Cmd {
	return func() tea.Msg {
		peak, err := mon.Next()
		return levelMsg{peak: peak, err: err}
	}
}
