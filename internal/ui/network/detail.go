package network

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/network"
	"github.com/anotherhadi/settuings/internal/util"
)

// deviceDetailState is the "network info" screen: full addressing/hardware
// details for whichever connected device (wifi or ethernet) the user asked
// about, plus a way to disconnect it. It overlays whichever tab opened it.
// Content can exceed a short terminal, so it's shown in a viewport.
type deviceDetailState struct {
	open           bool
	kind           tab // tabWifi or tabEthernet: decides which disconnect action to run
	device         string
	title          string
	security       string // wifi only
	signal         int    // wifi only, <=0 means n/a
	connectionName string // wifi only: profile name, for secret lookup

	loading bool
	err     error
	data    network.DeviceDetail

	secretLoading  bool
	secretValue    string
	secretRevealed bool
	secretErr      error

	pending  bool
	feedback string

	viewport viewport.Model
}

func (m *Model) openDeviceDetail(kind tab, device, title, security string, signal int) tea.Cmd {
	w, h := m.contentBudget()
	vp := viewport.New()
	vp.SetWidth(w)
	vp.SetHeight(h)

	m.deviceDetail = deviceDetailState{
		open:           true,
		kind:           kind,
		device:         device,
		title:          title,
		security:       security,
		signal:         signal,
		connectionName: m.wifi.connectionName,
		loading:        true,
		secretLoading:  kind == tabWifi,
		viewport:       vp,
	}
	m.refreshDeviceDetailContent()

	cmds := []tea.Cmd{fetchDeviceDetailCmd(device)}
	if kind == tabWifi && m.wifi.connectionName != "" {
		cmds = append(cmds, fetchSecretCmd(m.wifi.connectionName))
	}
	return tea.Batch(cmds...)
}

// contentBudget returns the exact space available inside the bordered tab
// box, below the tab bar and above the status bar. View() renders the box
// at this same size, so anything sized against contentBudget (e.g. a detail
// viewport) fills the box exactly instead of leaving it short. Since the
// status bar's height depends on live state (e.g. whether help is
// expanded), callers must re-size whenever that state changes — see
// resizeDeviceDetail/resizeKnownDetail.
func (m Model) contentBudget() (w, h int) {
	frameW := windowStyle().GetHorizontalFrameSize()
	frameH := windowStyle().GetVerticalFrameSize()
	w = m.width - frameW

	statusH := strings.Count(m.renderStatusBar(), "\n") + 1
	tabBarH := lipgloss.Height(m.renderTabBar(w))

	h = m.height - tabBarH - frameH - statusH
	if h < 1 {
		h = 1
	}
	return w, h
}

// resizeDetails re-sizes every open detail viewport to match the current
// content budget. Call it whenever that budget could have changed: on
// window resize, or when the status bar's height changes (e.g. toggling
// help).
func (m *Model) resizeDetails() {
	m.resizeDeviceDetail()
	m.resizeKnownDetail()
}

func (m *Model) resizeDeviceDetail() {
	if !m.deviceDetail.open {
		return
	}
	w, h := m.contentBudget()
	m.deviceDetail.viewport.SetWidth(w)
	m.deviceDetail.viewport.SetHeight(h)
	m.refreshDeviceDetailContent()
}

func (m *Model) handleDeviceDetail(msg deviceDetailMsg) {
	if !m.deviceDetail.open {
		return
	}
	m.deviceDetail.loading = false
	m.deviceDetail.data = msg.detail
	m.deviceDetail.err = msg.err
	m.refreshDeviceDetailContent()
}

// handleDeviceDetailSecret feeds a knownSecretMsg to the info screen if it's
// the one that asked for it. The known-networks tab checks the same message
// for its own detail view; both can be true at once only in theory (they're
// on different tabs), so this is safe to call unconditionally.
func (m *Model) handleDeviceDetailSecret(msg knownSecretMsg) {
	if !m.deviceDetail.open || m.deviceDetail.kind != tabWifi || m.deviceDetail.connectionName != msg.name {
		return
	}
	m.deviceDetail.secretLoading = false
	m.deviceDetail.secretValue = msg.secret
	m.deviceDetail.secretErr = msg.err
	m.refreshDeviceDetailContent()
}

// closeDeviceDetailIfPending closes the info screen once the disconnect it
// triggered has completed, but only if it's still the one waiting on kind.
func (m *Model) closeDeviceDetailIfPending(kind tab) {
	if m.deviceDetail.open && m.deviceDetail.kind == kind && m.deviceDetail.pending {
		m.deviceDetail = deviceDetailState{}
	}
}

func (m *Model) updateDeviceDetail(msg tea.Msg) tea.Cmd {
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		util.HandleMouseWheel(wheel, &m.deviceDetail.viewport)
		return nil
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	n := keys.Keys.Network

	switch {
	case key.Matches(keyMsg, g.Escape), key.Matches(keyMsg, g.CycleFocus):
		m.deviceDetail = deviceDetailState{}
	case key.Matches(keyMsg, g.Help):
		m.help.ShowAll = !m.help.ShowAll
		m.resizeDetails()
	case key.Matches(keyMsg, n.Disconnect):
		m.deviceDetail.pending = true
		m.refreshDeviceDetailContent()
		switch m.deviceDetail.kind {
		case tabWifi:
			return disconnectWifiCmd()
		case tabEthernet:
			return ethernetDownCmd(m.deviceDetail.device)
		}
	case key.Matches(keyMsg, n.RevealSecret):
		if m.deviceDetail.kind == tabWifi {
			m.deviceDetail.secretRevealed = !m.deviceDetail.secretRevealed
			m.refreshDeviceDetailContent()
		}
	case key.Matches(keyMsg, g.Up):
		util.ScrollLines(&m.deviceDetail.viewport, -1)
	case key.Matches(keyMsg, g.Down):
		util.ScrollLines(&m.deviceDetail.viewport, 1)
	case key.Matches(keyMsg, g.ScrollUp):
		util.ScrollViewport(&m.deviceDetail.viewport, -1)
	case key.Matches(keyMsg, g.ScrollDown):
		util.ScrollViewport(&m.deviceDetail.viewport, 1)
	}
	return nil
}

func (m Model) renderDeviceDetail() string {
	return m.deviceDetail.viewport.View()
}

func (m *Model) refreshDeviceDetailContent() {
	m.deviceDetail.viewport.SetContent(m.deviceDetailContent())
	util.RefreshScrollbar(&m.deviceDetail.viewport)
}

func (m Model) deviceDetailContent() string {
	d := m.deviceDetail
	if d.loading {
		return ilovetui.S.Faint.Render("Loading network info…")
	}

	title := lipgloss.NewStyle().Bold(true).Render(d.title)
	if d.pending {
		title += mutedStyle().Render("  disconnecting…")
	}
	rows := []string{title}
	if d.kind == tabWifi {
		rows = append(rows, subtleStyle().Render("Security:     ")+securityLabel(d.security))
		if d.signal > 0 {
			rows = append(rows, subtleStyle().Render("Signal:       ")+fmt.Sprintf("%d%%", d.signal))
		}
		rows = append(rows, subtleStyle().Render("Password:     ")+m.renderDeviceSecret())
	}
	rows = append(rows, subtleStyle().Render("Interface:    ")+d.device)

	if d.err != nil {
		rows = append(rows, "", errorStyle().Render(d.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	rows = append(rows,
		subtleStyle().Render("IPv4 Address: ")+valueOrDash(d.data.IPv4),
		subtleStyle().Render("Gateway:      ")+valueOrDash(d.data.Gateway4),
		subtleStyle().Render("DNS:          ")+valueOrDash(strings.Join(d.data.DNS4, ", ")),
		subtleStyle().Render("MAC Address:  ")+valueOrDash(d.data.HWAddr),
		subtleStyle().Render("MTU:          ")+valueOrDash(d.data.MTU),
	)

	if len(d.data.IPv6) > 0 {
		rows = append(rows, "", subtleStyle().Render("IPv6 Address: ")+d.data.IPv6[0])
		for _, addr := range d.data.IPv6[1:] {
			rows = append(rows, "              "+addr)
		}
	}
	if d.data.Gateway6 != "" {
		rows = append(rows, subtleStyle().Render("IPv6 Gateway: ")+d.data.Gateway6)
	}
	if len(d.data.DNS6) > 0 {
		rows = append(rows, subtleStyle().Render("IPv6 DNS:     ")+strings.Join(d.data.DNS6, ", "))
	}

	if d.feedback != "" {
		rows = append(rows, "", errorStyle().Render(d.feedback))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderDeviceSecret() string {
	d := m.deviceDetail
	if d.secretLoading {
		return mutedStyle().Render("loading…")
	}
	if d.secretErr != nil {
		return errorStyle().Render(d.secretErr.Error())
	}
	if d.secretValue == "" {
		return subtleStyle().Render("(none)")
	}
	if !d.secretRevealed {
		return subtleStyle().Render(strings.Repeat("•", len(d.secretValue)))
	}
	return d.secretValue
}

func valueOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
