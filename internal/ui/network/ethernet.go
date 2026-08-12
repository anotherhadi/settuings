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

type ethernetState struct {
	loading  bool
	err      error
	devices  []network.EthernetDevice
	cursor   int
	pending  bool
	feedback string
}

func newEthernetState() ethernetState {
	return ethernetState{loading: true}
}

func (m *Model) handleEthernetList(msg ethernetListMsg) {
	m.ethernet.loading = false
	m.ethernet.err = msg.err
	if msg.err != nil {
		return
	}
	m.ethernet.devices = msg.devices
	if m.ethernet.cursor >= len(m.ethernet.devices) {
		m.ethernet.cursor = max(0, len(m.ethernet.devices)-1)
	}
}

func (m *Model) handleEthernetAction(msg ethernetActionMsg) tea.Cmd {
	m.ethernet.pending = false
	if msg.err != nil {
		m.ethernet.feedback = msg.err.Error()
		m.deviceDetail.pending = false
		m.deviceDetail.feedback = msg.err.Error()
		m.refreshDeviceDetailContent()
		return nil
	}
	m.ethernet.feedback = ""
	m.closeDeviceDetailIfPending(tabEthernet)
	return fetchEthernetCmd()
}

func (m *Model) updateEthernet(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	n := keys.Keys.Network

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.ethernet.cursor > 0 {
			m.ethernet.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.ethernet.cursor < len(m.ethernet.devices)-1 {
			m.ethernet.cursor++
		}
	case key.Matches(keyMsg, n.Refresh):
		return fetchEthernetCmd()
	case key.Matches(keyMsg, n.Select):
		return m.selectEthernet()
	case key.Matches(keyMsg, n.Info):
		return m.infoEthernet()
	}
	return nil
}

func (m *Model) infoEthernet() tea.Cmd {
	if m.ethernet.cursor < 0 || m.ethernet.cursor >= len(m.ethernet.devices) {
		return nil
	}
	dev := m.ethernet.devices[m.ethernet.cursor]
	if !dev.Connected() {
		m.ethernet.feedback = "Not connected."
		return nil
	}
	return m.openDeviceDetail(tabEthernet, dev.Device, dev.Device, "", 0)
}

func (m *Model) selectEthernet() tea.Cmd {
	if m.ethernet.cursor < 0 || m.ethernet.cursor >= len(m.ethernet.devices) {
		return nil
	}
	dev := m.ethernet.devices[m.ethernet.cursor]
	m.ethernet.pending = true
	if dev.Connected() {
		return ethernetDownCmd(dev.Device)
	}
	return ethernetConnectCmd(dev.Device)
}

func (m Model) renderEthernet(height int) string {
	if m.ethernet.loading {
		return ilovetui.S.Faint.Render("Loading ethernet devices…")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("ETHERNET DEVICES")
	if m.ethernet.pending {
		header += mutedStyle().Render("  working…")
	}

	if m.ethernet.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", errorStyle().Render(m.ethernet.err.Error()))
	}
	if len(m.ethernet.devices) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, "", ilovetui.S.Faint.Render("No ethernet devices found."))
	}

	available := height - 3 // header + blank line + a trailing line for the paginator/feedback
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.ethernet.cursor, len(m.ethernet.devices), available)

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, m.renderEthernetRow(m.ethernet.devices[i], i == m.ethernet.cursor))
	}
	list := strings.Join(rows, "\n")

	footer := renderPaginatorDots(p)
	if m.ethernet.feedback != "" {
		footer = errorStyle().Render(m.ethernet.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", list, footer)
}

func (m Model) renderEthernetRow(d network.EthernetDevice, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if d.Connected() {
		check = successStyle().Render(icons.I.Check)
	}
	name := d.Device
	if d.Connection != "" {
		name = d.Connection
	}
	details := d.State
	if d.IP4 != "" {
		details += " · " + d.IP4
	}
	return fmt.Sprintf("%s%s%s %s  %s", cursor, check, textStyle.Render(name), subtleStyle().Render("("+d.Device+")"), subtleStyle().Render(details))
}
