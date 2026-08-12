package network

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/network"
)

type vpnState struct {
	loading  bool
	err      error
	conns    []network.VPNConnection
	cursor   int
	pending  bool
	feedback string
}

func newVPNState() vpnState {
	return vpnState{loading: true}
}

func (m *Model) handleVPNList(msg vpnListMsg) {
	m.vpn.loading = false
	m.vpn.err = msg.err
	if msg.err != nil {
		return
	}
	m.vpn.conns = msg.conns
	if m.vpn.cursor >= len(m.vpn.conns) {
		m.vpn.cursor = max(0, len(m.vpn.conns)-1)
	}
}

func (m *Model) handleVPNAction(msg vpnActionMsg) tea.Cmd {
	m.vpn.pending = false
	if msg.err != nil {
		m.vpn.feedback = msg.err.Error()
		return nil
	}
	m.vpn.feedback = ""
	return fetchVPNCmd()
}

func (m *Model) updateVPN(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	n := keys.Keys.Network

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.vpn.cursor > 0 {
			m.vpn.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.vpn.cursor < len(m.vpn.conns)-1 {
			m.vpn.cursor++
		}
	case key.Matches(keyMsg, n.Refresh):
		return fetchVPNCmd()
	case key.Matches(keyMsg, n.Select):
		return m.selectVPN()
	}
	return nil
}

func (m *Model) selectVPN() tea.Cmd {
	if m.vpn.cursor < 0 || m.vpn.cursor >= len(m.vpn.conns) {
		return nil
	}
	conn := m.vpn.conns[m.vpn.cursor]
	m.vpn.pending = true
	if conn.Active {
		return vpnDownCmd(conn.Name)
	}
	return vpnUpCmd(conn.Name)
}

func (m Model) renderVPN(height int) string {
	if m.vpn.loading {
		return ilovetui.S.Faint.Render("Loading VPN connections…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("VPN CONNECTIONS")
	if m.vpn.pending {
		header += mutedStyle().Render("  working…")
	}

	if m.vpn.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.vpn.err.Error()))
	}
	if len(m.vpn.conns) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No VPN connections configured."))
	}

	available := height - 3 // header + blank line + a trailing line for the paginator/feedback
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.vpn.cursor, len(m.vpn.conns), available)

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, m.renderVPNRow(m.vpn.conns[i], i == m.vpn.cursor))
	}
	list := strings.Join(rows, "\n")

	footer := renderPaginatorDots(p)
	if m.vpn.feedback != "" {
		footer = errorStyle().Render(m.vpn.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, footer)
}

func (m Model) renderVPNRow(c network.VPNConnection, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if c.Active {
		check = successStyle().Render(icons.I.Check)
	}
	return fmt.Sprintf("%s%s%s  %s", cursor, check, textStyle.Render(c.Name), subtleStyle().Render(c.Type))
}
