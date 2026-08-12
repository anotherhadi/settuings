package network

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/network"
)

type wifiState struct {
	loading        bool
	err            error
	ip             string
	device         string // wifi interface name, e.g. "wlo1"
	connectionName string // active connection profile name, e.g. "Djoon 1"
	radioOn        bool
	current        *network.WifiNetwork // the AP we're associated with, if any
	others         []network.WifiNetwork
	cursor         int
	pending        bool // a connect/disconnect/rescan/radio-toggle is in flight
	feedback       string

	connecting  bool
	connectSSID string
	password    textinput.Model
}

func newWifiState() wifiState {
	ti := textinput.New()
	ti.Prompt = "Password: "
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	s := ti.Styles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(ilovetui.S.Primary)
	ti.SetStyles(s)
	return wifiState{password: ti, loading: true}
}

func (m Model) isKnownSSID(ssid string) bool {
	for _, k := range m.known.networks {
		if k.Name == ssid {
			return true
		}
	}
	return false
}

func (m *Model) handleWifiList(msg wifiListMsg) {
	m.wifi.loading = false
	m.wifi.err = msg.err
	m.wifi.radioOn = msg.radioOn
	if msg.err != nil {
		return
	}
	m.wifi.ip = msg.ip
	m.wifi.device = msg.device
	m.wifi.connectionName = msg.connectionName
	m.wifi.current = nil
	m.wifi.others = m.wifi.others[:0]
	for _, n := range msg.networks {
		n := n
		if n.Active {
			m.wifi.current = &n
			continue
		}
		m.wifi.others = append(m.wifi.others, n)
	}
	if m.wifi.cursor >= len(m.wifi.others) {
		m.wifi.cursor = max(0, len(m.wifi.others)-1)
	}
}

func (m *Model) handleWifiAction(msg wifiActionMsg) tea.Cmd {
	m.wifi.pending = false
	if msg.err != nil {
		m.wifi.feedback = msg.err.Error()
		m.deviceDetail.pending = false
		m.deviceDetail.feedback = msg.err.Error()
		m.refreshDeviceDetailContent()
		return nil
	}
	m.wifi.feedback = ""
	m.closeDeviceDetailIfPending(tabWifi)
	return tea.Batch(fetchWifiCmd(), fetchKnownCmd())
}

func (m *Model) handleRadioAction(msg radioActionMsg) tea.Cmd {
	m.wifi.pending = false
	if msg.err != nil {
		m.wifi.feedback = msg.err.Error()
		return nil
	}
	m.wifi.feedback = ""
	return fetchWifiCmd()
}

func (m *Model) updateWifi(msg tea.Msg) tea.Cmd {
	if m.wifi.connecting {
		return m.updateWifiPasswordPrompt(msg)
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	n := keys.Keys.Network

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.wifi.cursor > 0 {
			m.wifi.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.wifi.cursor < len(m.wifi.others)-1 {
			m.wifi.cursor++
		}
	case key.Matches(keyMsg, n.Refresh):
		m.wifi.pending = true
		return tea.Batch(rescanWifiCmd(), fetchWifiCmd())
	case key.Matches(keyMsg, n.Select):
		return m.selectWifi()
	case key.Matches(keyMsg, n.ToggleRadio):
		m.wifi.pending = true
		return toggleWifiRadioCmd(!m.wifi.radioOn)
	case key.Matches(keyMsg, n.Disconnect):
		if m.wifi.current == nil {
			m.wifi.feedback = "Not connected to a network."
			return nil
		}
		m.wifi.pending = true
		return disconnectWifiCmd()
	case key.Matches(keyMsg, n.Info):
		if m.wifi.current == nil {
			m.wifi.feedback = "Not connected to a network."
			return nil
		}
		cur := m.wifi.current
		return m.openDeviceDetail(tabWifi, m.wifi.device, cur.SSID, cur.Security, cur.Signal)
	}
	return nil
}

func (m *Model) selectWifi() tea.Cmd {
	if len(m.wifi.others) == 0 {
		return nil
	}
	target := m.wifi.others[m.wifi.cursor]
	if target.Secured() && !m.isKnownSSID(target.SSID) {
		m.wifi.connecting = true
		m.wifi.connectSSID = target.SSID
		m.wifi.password.SetValue("")
		m.wifi.password.Focus()
		return nil
	}
	m.wifi.pending = true
	return connectWifiCmd(target.SSID, "")
}

func (m *Model) updateWifiPasswordPrompt(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	switch {
	case key.Matches(keyMsg, g.Escape):
		m.wifi.connecting = false
		m.wifi.password.Blur()
		return nil
	case keyMsg.String() == "enter":
		ssid := m.wifi.connectSSID
		password := m.wifi.password.Value()
		m.wifi.connecting = false
		m.wifi.password.Blur()
		m.wifi.pending = true
		return connectWifiCmd(ssid, password)
	default:
		var cmd tea.Cmd
		m.wifi.password, cmd = m.wifi.password.Update(keyMsg)
		return cmd
	}
}

func (m Model) renderWifi(width, height int) string {
	if m.wifi.connecting {
		return m.renderWifiPasswordPrompt()
	}

	if m.wifi.loading {
		return ilovetui.S.Faint.Render("Loading Wi-Fi networks…")
	}

	othersHeader := lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("OTHER NETWORKS")
	if m.wifi.pending {
		othersHeader += mutedStyle().Render("  refreshing…")
	}

	fixed := lipgloss.JoinVertical(lipgloss.Left,
		m.renderWifiRadioStatus(),
		"",
		lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render("CURRENT NETWORK"),
		"",
		m.renderCurrentWifiCard(width),
		"",
		othersHeader,
		"",
	)
	fixedH := lipgloss.Height(fixed)

	if m.wifi.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, fixed, errorStyle().Render(m.wifi.err.Error()))
	}
	if len(m.wifi.others) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, fixed, ilovetui.S.Faint.Render("No other networks in range."))
	}

	available := height - fixedH - 1 // reserve a trailing line for the paginator/feedback
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.wifi.cursor, len(m.wifi.others), available)

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, m.renderWifiRow(m.wifi.others[i], i == m.wifi.cursor))
	}
	list := strings.Join(rows, "\n")

	footer := renderPaginatorDots(p)
	if m.wifi.feedback != "" {
		footer = errorStyle().Render(m.wifi.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, fixed, list, footer)
}

func (m Model) renderWifiRadioStatus() string {
	label := "Wi-Fi: "
	if m.wifi.radioOn {
		return label + successStyle().Bold(true).Render("On")
	}
	return label + errorStyle().Bold(true).Render("Off")
}

func (m Model) renderCurrentWifiCard(width int) string {
	if m.wifi.current == nil {
		return ilovetui.S.Faint.Render("Not connected")
	}
	cur := m.wifi.current
	check := lipgloss.NewStyle().Foreground(ilovetui.S.Success).Render(icons.I.Check)
	title := lipgloss.NewStyle().Bold(true).Render(cur.SSID)
	line1 := check + title
	line2 := lipgloss.NewStyle().Foreground(ilovetui.S.Subtle).Render(
		fmt.Sprintf("%s · %d%%", securityLabel(cur.Security), cur.Signal))
	if m.wifi.ip != "" {
		line2 += lipgloss.NewStyle().Foreground(ilovetui.S.Subtle).Render(" · " + m.wifi.ip)
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, line1, line2))
}

func (m Model) renderWifiRow(n network.WifiNetwork, selected bool) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	subtleStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Subtle)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	lock := ""
	if n.Secured() {
		lock = " " + icons.I.Lock
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	return fmt.Sprintf("%s%s%s %s", cursor, textStyle.Render(n.SSID), lock, subtleStyle.Render(fmt.Sprintf("%d%%", n.Signal)))
}

func (m Model) renderWifiPasswordPrompt() string {
	title := lipgloss.NewStyle().Bold(true).Render("Connect to " + m.wifi.connectSSID)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", m.wifi.password.View())
}

func securityLabel(security string) string {
	if security == "" || security == "--" {
		return "Open"
	}
	return security
}
