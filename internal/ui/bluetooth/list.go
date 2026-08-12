package bluetooth

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/bluetooth"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
)

// deviceListState is the main screen: paired devices (fixed profiles)
// followed by other devices seen during a scan but not yet paired. Both
// live in one cursor-navigable, height-budgeted, paginated list.
type deviceListState struct {
	loading  bool
	err      error
	paired   []bluetooth.Device
	others   []bluetooth.Device
	cursor   int
	pending  bool
	scanning bool
	feedback string

	confirmRemove bool
}

func newDeviceListState() deviceListState {
	return deviceListState{loading: true}
}

func (s deviceListState) total() int { return len(s.paired) + len(s.others) }

func (s deviceListState) selected() *bluetooth.Device {
	if s.cursor < 0 || s.cursor >= s.total() {
		return nil
	}
	if s.cursor < len(s.paired) {
		return &s.paired[s.cursor]
	}
	return &s.others[s.cursor-len(s.paired)]
}

func (m *Model) handleDeviceList(msg deviceListMsg) {
	m.list.loading = false
	m.list.err = msg.err
	if msg.err != nil {
		return
	}

	paired := make([]bluetooth.Device, 0, len(msg.devices))
	others := make([]bluetooth.Device, 0, len(msg.devices))
	for _, d := range msg.devices {
		if d.Paired {
			paired = append(paired, d)
		} else {
			others = append(others, d)
		}
	}
	sort.Slice(paired, func(i, j int) bool {
		if paired[i].Connected != paired[j].Connected {
			return paired[i].Connected
		}
		return strings.ToLower(paired[i].Name) < strings.ToLower(paired[j].Name)
	})
	sort.Slice(others, func(i, j int) bool {
		return strings.ToLower(others[i].Name) < strings.ToLower(others[j].Name)
	})
	m.list.paired = paired
	m.list.others = others
	if m.list.cursor >= m.list.total() {
		m.list.cursor = max(0, m.list.total()-1)
	}
}

func (m *Model) handleDeviceAction(msg deviceActionMsg) tea.Cmd {
	m.list.pending = false
	if msg.err != nil {
		m.list.feedback = msg.err.Error()
		m.detail.pending = false
		m.detail.feedback = msg.err.Error()
		m.refreshDetailContent()
		return nil
	}
	m.list.feedback = ""
	m.closeDetailIfPending()
	return fetchDevicesCmd()
}

func (m *Model) handlePowerAction(msg powerActionMsg) tea.Cmd {
	m.list.pending = false
	if msg.err != nil {
		m.list.feedback = msg.err.Error()
		return nil
	}
	m.list.feedback = ""
	return tea.Batch(fetchControllerCmd(), fetchDevicesCmd())
}

func (m *Model) handleScanDone(msg scanDoneMsg) tea.Cmd {
	m.list.scanning = false
	if msg.err != nil {
		m.list.feedback = msg.err.Error()
		return nil
	}
	m.list.feedback = ""
	return fetchDevicesCmd()
}

func (m *Model) updateList(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	if m.list.confirmRemove {
		return m.updateConfirmRemove(keyMsg)
	}

	g := keys.Keys.Global
	b := keys.Keys.Bluetooth

	switch {
	case key.Matches(keyMsg, g.Up):
		if m.list.cursor > 0 {
			m.list.cursor--
		}
	case key.Matches(keyMsg, g.Down):
		if m.list.cursor < m.list.total()-1 {
			m.list.cursor++
		}
	case key.Matches(keyMsg, b.Refresh):
		m.list.pending = true
		m.list.scanning = true
		return scanCmd()
	case key.Matches(keyMsg, b.Select):
		return m.selectDevice()
	case key.Matches(keyMsg, b.Info):
		return m.openInfo()
	case key.Matches(keyMsg, b.ToggleTrust):
		return m.toggleTrust()
	case key.Matches(keyMsg, b.Forget):
		if m.list.selected() != nil {
			m.list.confirmRemove = true
		}
	case key.Matches(keyMsg, b.TogglePower):
		m.list.pending = true
		return togglePowerCmd(!m.controller.Powered)
	}
	return nil
}

func (m *Model) updateConfirmRemove(keyMsg tea.KeyPressMsg) tea.Cmd {
	switch keyMsg.String() {
	case "y", "enter":
		sel := m.list.selected()
		m.list.confirmRemove = false
		if sel == nil {
			return nil
		}
		m.list.pending = true
		return removeCmd(sel.Address)
	default:
		m.list.confirmRemove = false
	}
	return nil
}

func (m *Model) selectDevice() tea.Cmd {
	sel := m.list.selected()
	if sel == nil {
		return nil
	}
	m.list.pending = true
	if sel.Paired {
		if sel.Connected {
			return disconnectCmd(sel.Address)
		}
		return connectCmd(sel.Address)
	}
	return pairCmd(sel.Address)
}

func (m *Model) toggleTrust() tea.Cmd {
	sel := m.list.selected()
	if sel == nil {
		return nil
	}
	m.list.pending = true
	return trustCmd(sel.Address, !sel.Trusted)
}

func (m *Model) openInfo() tea.Cmd {
	sel := m.list.selected()
	if sel == nil {
		return nil
	}
	return m.openDetail(sel.Address, sel.Name)
}

func (m Model) renderList(width, height int) string {
	if m.list.loading {
		return ilovetui.S.Faint.Render("Loading Bluetooth devices…")
	}
	if m.list.err != nil {
		return errorStyle().Render(m.list.err.Error())
	}
	if m.list.total() == 0 {
		return ilovetui.S.Faint.Render("No devices found. Press r to scan for nearby devices.")
	}

	// Reserve a line for each section header that could appear, so the
	// paginated device rows never overflow the available height even on
	// the one page where a section boundary falls.
	headerReserve := 0
	if len(m.list.paired) > 0 {
		headerReserve++
	}
	if len(m.list.others) > 0 {
		headerReserve++
	}

	available := height - headerReserve - 1 // trailing line for paginator/feedback
	if available < 1 {
		available = 1
	}
	p, start, end := paginate(m.list.cursor, m.list.total(), available)

	var lines []string
	for i := start; i < end; i++ {
		if i == 0 && len(m.list.paired) > 0 {
			lines = append(lines, sectionHeader("PAIRED DEVICES"))
		}
		if i == len(m.list.paired) && len(m.list.others) > 0 {
			lines = append(lines, sectionHeader("OTHER DEVICES"))
		}
		if i < len(m.list.paired) {
			lines = append(lines, m.renderDeviceRow(m.list.paired[i], i == m.list.cursor, width))
		} else {
			lines = append(lines, m.renderDeviceRow(m.list.others[i-len(m.list.paired)], i == m.list.cursor, width))
		}
	}
	list := strings.Join(lines, "\n")

	footer := renderPaginatorDots(p)
	switch {
	case m.list.confirmRemove:
		name := ""
		if sel := m.list.selected(); sel != nil {
			name = sel.Name
		}
		footer = errorStyle().Render(fmt.Sprintf("Forget %q? press y to confirm, any other key to cancel", name))
	case m.list.feedback != "":
		footer = errorStyle().Render(m.list.feedback)
	}

	return lipgloss.JoinVertical(lipgloss.Left, list, footer)
}

func sectionHeader(text string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(ilovetui.S.Subtle).Render(text)
}

func (m Model) renderDeviceRow(d bluetooth.Device, selected bool, width int) string {
	textStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Text)
	if selected {
		textStyle = textStyle.Bold(true).Foreground(ilovetui.S.Primary)
	}
	cursor := "  "
	if selected {
		cursor = "> "
	}
	check := " "
	if d.Connected {
		check = successStyle().Render(icons.I.Check)
	}

	var tags []string
	if d.Trusted {
		tags = append(tags, "trusted")
	}
	if d.Battery >= 0 {
		tags = append(tags, fmt.Sprintf("%d%%", d.Battery))
	}
	tag := strings.Join(tags, " · ")

	row := fmt.Sprintf("%s%s%s", cursor, check, textStyle.Render(d.Name))
	if tag == "" {
		return row
	}
	pad := width - lipgloss.Width(row) - lipgloss.Width(tag)
	if pad < 1 {
		pad = 1
	}
	return row + strings.Repeat(" ", pad) + subtleStyle().Render(tag)
}
