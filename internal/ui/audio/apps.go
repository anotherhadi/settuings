package audio

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/audio"
	"github.com/anotherhadi/settuings/internal/keys"
)

type appsState struct {
	loading  bool
	err      error
	streams  []audio.Stream
	cursor   int
	pending  bool
	feedback string
}

func newAppsState() appsState {
	return appsState{loading: true}
}

func (m *Model) handleAppsList(msg appsListMsg) {
	m.apps.loading = false
	m.apps.err = msg.err
	if msg.err != nil {
		return
	}
	m.apps.streams = msg.streams
	if m.apps.cursor >= len(m.apps.streams) {
		m.apps.cursor = max(0, len(m.apps.streams)-1)
	}
}

func (m *Model) handleAppsAction(msg appsActionMsg) tea.Cmd {
	m.apps.pending = false
	if msg.err != nil {
		m.apps.feedback = msg.err.Error()
		return nil
	}
	return fetchAppsCmd()
}

func (m Model) selectedStream() *audio.Stream {
	if m.apps.cursor < 0 || m.apps.cursor >= len(m.apps.streams) {
		return nil
	}
	return &m.apps.streams[m.apps.cursor]
}

func (m *Model) updateApps(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	a := keys.Keys.Audio

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.apps.cursor > 0 {
			m.apps.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.apps.cursor < len(m.apps.streams)-1 {
			m.apps.cursor++
		}
	case key.Matches(keyMsg, g.Left):
		return m.adjustStreamVolume(-volumeStep)
	case key.Matches(keyMsg, g.Right):
		return m.adjustStreamVolume(volumeStep)
	case key.Matches(keyMsg, a.Refresh):
		m.apps.pending = true
		return fetchAppsCmd()
	case key.Matches(keyMsg, a.Mute):
		s := m.selectedStream()
		if s == nil {
			return nil
		}
		m.apps.feedback = ""
		m.apps.pending = true
		return appsToggleMuteCmd(s.ID)
	}
	return nil
}

func (m *Model) adjustStreamVolume(delta int) tea.Cmd {
	s := m.selectedStream()
	if s == nil {
		return nil
	}
	pct := s.Volume + delta
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	s.Volume = pct
	m.apps.feedback = ""
	return appsSetVolumeCmd(s.ID, pct)
}

func sectionTitle(k audio.StreamKind) string {
	if k == audio.StreamInput {
		return "INPUT STREAMS (recording)"
	}
	return "OUTPUT STREAMS (playback)"
}

// windowAroundCursor picks a [start, end) slice of size items (or fewer, if
// total is smaller) that keeps cursor in view.
func windowAroundCursor(cursor, total, size int) (start, end int) {
	if size >= total {
		return 0, total
	}
	start = cursor - size/2
	if start < 0 {
		start = 0
	}
	if start+size > total {
		start = total - size
	}
	return start, start + size
}

func (m Model) renderApps(width, height int) string {
	if m.apps.loading {
		return ilovetui.S.Faint.Render("Loading application streams…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("APPLICATIONS")
	if m.apps.pending {
		header += mutedStyle().Render("  refreshing…")
	}

	if m.apps.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.apps.err.Error()))
	}
	if len(m.apps.streams) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No application is currently playing or recording audio."))
	}

	maxItems := (height - 3) / streamRowHeight
	if maxItems < 1 {
		maxItems = 1
	}
	start, end := windowAroundCursor(m.apps.cursor, len(m.apps.streams), maxItems)

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle)
	var b strings.Builder
	lastKind := audio.StreamKind(-1)
	for i := start; i < end; i++ {
		s := m.apps.streams[i]
		if s.Kind != lastKind {
			b.WriteString(sectionStyle.Render(sectionTitle(s.Kind)) + "\n\n")
			lastKind = s.Kind
		}
		b.WriteString(renderStreamRow(s, i == m.apps.cursor, width))
		b.WriteString("\n\n")
	}
	list := strings.TrimRight(b.String(), "\n")

	footer := ""
	if m.apps.feedback != "" {
		footer = subtleStyle().Render(m.apps.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, footer)
}
