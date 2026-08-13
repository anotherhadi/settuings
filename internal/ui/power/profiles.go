package power

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/power"
)

func (m *Model) handleProfiles(msg profilesMsg) {
	m.profilesLoading = false
	m.profilesErr = msg.err
	if msg.err != nil {
		return
	}
	m.profiles = msg.profiles
	if m.cursor >= len(m.profiles) {
		m.cursor = max(0, len(m.profiles)-1)
	}
}

func (m *Model) handleSetProfile(msg setProfileMsg) tea.Cmd {
	m.pending = false
	if msg.err != nil {
		m.feedback = msg.err.Error()
		return nil
	}
	m.feedback = ""
	return fetchProfilesCmd()
}

func (m Model) selectedProfile() *power.Profile {
	if m.cursor < 0 || m.cursor >= len(m.profiles) {
		return nil
	}
	return &m.profiles[m.cursor]
}

func (m *Model) updateProfiles(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if m.checkErr != nil {
		if key.Matches(keyMsg, keys.Keys.Power.Refresh) {
			return checkCmd()
		}
		return nil
	}

	g := keys.Keys.Global
	p := keys.Keys.Power

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, p.Refresh):
		return tea.Batch(fetchProfilesCmd(), fetchBatteryCmd())
	case key.Matches(keyMsg, p.Select):
		return m.applySelected()
	}
	return nil
}

func (m *Model) applySelected() tea.Cmd {
	sel := m.selectedProfile()
	if sel == nil || sel.Active {
		return nil
	}
	m.pending = true
	return setProfileCmd(sel.Name)
}

func (m Model) renderProfiles() string {
	title := lipgloss.NewStyle().Bold(true).Render("Power Profile")

	switch {
	case !m.checked:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", ilovetui.S.Faint.Render("Checking powerprofilesctl…"))
	case m.checkErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", subtleStyle().Render(unavailableHint(m.checkErr)))
	case m.profilesLoading:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", ilovetui.S.Faint.Render("Loading profiles…"))
	case m.profilesErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", errorStyle().Render(m.profilesErr.Error()))
	}

	lines := []string{title, ""}
	for i, p := range m.profiles {
		lines = append(lines, m.renderProfileRow(p, i == m.cursor))
	}
	switch {
	case m.pending:
		lines = append(lines, "", mutedStyle().Render("applying…"))
	case m.feedback != "":
		lines = append(lines, "", errorStyle().Render(m.feedback))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) renderProfileRow(p power.Profile, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if p.Active {
		check = successStyle().Render(icons.I.Check)
	}
	return fmt.Sprintf("%s%s%s", cursor, check, textStyle.Render(displayName(p.Name)))
}

// displayName turns a powerprofilesctl identifier ("power-saver") into a
// readable label ("Power Saver").
func displayName(name string) string {
	words := strings.Split(strings.ReplaceAll(name, "-", " "), " ")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

func unavailableHint(err error) string {
	if errors.Is(err, power.ErrNotFound) {
		return "powerprofilesctl was not found in PATH.\nInstall power-profiles-daemon to select a profile."
	}
	return err.Error()
}
