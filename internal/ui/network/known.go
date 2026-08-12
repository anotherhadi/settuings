package network

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/network"
	"github.com/anotherhadi/settuings/internal/util"
)

type knownState struct {
	loading  bool
	err      error
	networks []network.KnownNetwork
	cursor   int
	pending  bool
	feedback string

	detail         bool
	secretLoading  bool
	secretRevealed bool
	secretValue    string
	secretErr      error

	confirmForget bool

	// Content can exceed a short terminal, so the detail screen is shown in
	// a viewport.
	viewport viewport.Model
}

func newKnownState() knownState {
	return knownState{loading: true}
}

func (m *Model) handleKnownList(msg knownListMsg) {
	m.known.loading = false
	m.known.err = msg.err
	if msg.err != nil {
		return
	}
	m.known.networks = msg.networks
	if m.known.cursor >= len(m.known.networks) {
		m.known.cursor = max(0, len(m.known.networks)-1)
	}
}

func (m *Model) handleKnownSecret(msg knownSecretMsg) {
	if m.known.selected() == nil || m.known.selected().Name != msg.name {
		return // stale response for a network we've since navigated away from
	}
	m.known.secretLoading = false
	m.known.secretValue = msg.secret
	m.known.secretErr = msg.err
	m.refreshKnownDetailContent()
}

func (m *Model) handleKnownAction(msg knownActionMsg) tea.Cmd {
	m.known.pending = false
	if msg.err != nil {
		m.known.feedback = msg.err.Error()
		m.refreshKnownDetailContent()
		return nil
	}
	m.known.feedback = ""
	m.known.detail = false
	return fetchKnownCmd()
}

func (s knownState) selected() *network.KnownNetwork {
	if s.cursor < 0 || s.cursor >= len(s.networks) {
		return nil
	}
	return &s.networks[s.cursor]
}

func (m *Model) updateKnown(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	if m.known.detail {
		return m.updateKnownDetail(keyMsg)
	}

	g := keys.Keys.Global
	n := keys.Keys.Network
	switch {
	case key.Matches(keyMsg, g.Up):
		if m.known.cursor > 0 {
			m.known.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.known.cursor < len(m.known.networks)-1 {
			m.known.cursor++
		}
	case key.Matches(keyMsg, n.Refresh):
		return fetchKnownCmd()
	case key.Matches(keyMsg, n.Select):
		if sel := m.known.selected(); sel != nil {
			m.known.detail = true
			m.known.secretRevealed = false
			m.known.secretLoading = true
			m.known.secretValue = ""
			m.known.secretErr = nil
			m.openKnownDetail()
			return fetchSecretCmd(sel.Name)
		}
	}
	return nil
}

func (m *Model) updateKnownDetail(keyMsg tea.KeyPressMsg) tea.Cmd {
	g := keys.Keys.Global
	n := keys.Keys.Network

	if m.known.confirmForget {
		m.known.confirmForget = false
		defer m.refreshKnownDetailContent()
		if keyMsg.String() == "y" || keyMsg.String() == "enter" {
			if sel := m.known.selected(); sel != nil {
				m.known.pending = true
				return forgetCmd(sel.Name)
			}
		}
		return nil
	}

	switch {
	case key.Matches(keyMsg, g.Escape):
		m.known.detail = false
	case key.Matches(keyMsg, n.RevealSecret):
		m.known.secretRevealed = !m.known.secretRevealed
		m.refreshKnownDetailContent()
	case key.Matches(keyMsg, n.ToggleAuto):
		if sel := m.known.selected(); sel != nil {
			m.known.pending = true
			return toggleAutoConnectCmd(sel.Name, !sel.AutoConnect)
		}
	case key.Matches(keyMsg, n.Forget):
		m.known.confirmForget = true
		m.refreshKnownDetailContent()
	case key.Matches(keyMsg, g.Up):
		util.ScrollLines(&m.known.viewport, -1)
	case key.Matches(keyMsg, g.Down):
		util.ScrollLines(&m.known.viewport, 1)
	case key.Matches(keyMsg, g.ScrollUp):
		util.ScrollViewport(&m.known.viewport, -1)
	case key.Matches(keyMsg, g.ScrollDown):
		util.ScrollViewport(&m.known.viewport, 1)
	}
	return nil
}

// openKnownDetail sizes and populates the detail viewport for the currently
// selected network. Call it whenever the detail screen is opened.
func (m *Model) openKnownDetail() {
	w, h := m.contentBudget()
	vp := viewport.New()
	vp.SetWidth(w)
	vp.SetHeight(h)
	m.known.viewport = vp
	m.refreshKnownDetailContent()
}

// resizeKnownDetail keeps the detail viewport sized to the available space
// whenever the terminal is resized.
func (m *Model) resizeKnownDetail() {
	if !m.known.detail {
		return
	}
	w, h := m.contentBudget()
	m.known.viewport.SetWidth(w)
	m.known.viewport.SetHeight(h)
	m.refreshKnownDetailContent()
}

func (m *Model) refreshKnownDetailContent() {
	m.known.viewport.SetContent(m.knownDetailContent())
	util.RefreshScrollbar(&m.known.viewport)
}

func (m Model) renderKnown(height int) string {
	if m.known.detail {
		return m.renderKnownDetail()
	}
	if m.known.loading {
		return ilovetui.S.Faint.Render("Loading known networks…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("KNOWN NETWORKS")

	if m.known.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.known.err.Error()))
	}
	if len(m.known.networks) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No saved Wi-Fi networks."))
	}

	available := height - 3 // header + blank line + a trailing line for the paginator
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.known.cursor, len(m.known.networks), available)

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, m.renderKnownRow(m.known.networks[i], i == m.known.cursor))
	}
	list := strings.Join(rows, "\n")

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, renderPaginatorDots(p))
}

func (m Model) renderKnownRow(k network.KnownNetwork, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if k.Active {
		check = successStyle().Render(icons.I.Check)
	}
	auto := "manual"
	if k.AutoConnect {
		auto = "auto-join"
	}
	return fmt.Sprintf("%s%s%s  %s", cursor, check, textStyle.Render(k.Name), subtleStyle().Render(auto))
}

func (m Model) renderKnownDetail() string {
	return m.known.viewport.View()
}

func (m Model) knownDetailContent() string {
	sel := m.known.selected()
	if sel == nil {
		return ilovetui.S.Faint.Render("No network selected.")
	}

	rows := []string{
		lipgloss.NewStyle().Bold(true).Render(sel.Name),
		"",
		subtleStyle().Render("Security:    ") + securityLabel(sel.Security),
		subtleStyle().Render("Auto-join:   ") + fmt.Sprintf("%v", sel.AutoConnect),
		subtleStyle().Render("Active:      ") + fmt.Sprintf("%v", sel.Active),
		subtleStyle().Render("Password:    ") + m.renderKnownSecret(),
	}

	if m.known.feedback != "" {
		rows = append(rows, "", errorStyle().Render(m.known.feedback))
	}

	if m.known.confirmForget {
		rows = append(rows, "", errorStyle().Render(fmt.Sprintf("Forget %q? press y to confirm, any other key to cancel", sel.Name)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderKnownSecret() string {
	if m.known.secretLoading {
		return mutedStyle().Render("loading…")
	}
	if m.known.secretErr != nil {
		return errorStyle().Render(m.known.secretErr.Error())
	}
	if m.known.secretValue == "" {
		return subtleStyle().Render("(none)")
	}
	if !m.known.secretRevealed {
		return subtleStyle().Render(strings.Repeat("•", len(m.known.secretValue)))
	}
	return m.known.secretValue
}
