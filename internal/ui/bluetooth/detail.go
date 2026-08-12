package bluetooth

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/bluetooth"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/util"
)

// deviceDetailState is the "device info" screen for whichever device (paired
// or not) the user asked about. It overlays the main list. Content can
// exceed a short terminal, so it's shown in a viewport.
type deviceDetailState struct {
	open    bool
	address string
	title   string

	loading bool
	err     error
	data    bluetooth.DeviceDetail

	pending  bool
	feedback string

	viewport viewport.Model
}

func (m *Model) openDetail(address, title string) tea.Cmd {
	w, h := m.contentBudget()
	vp := viewport.New()
	vp.SetWidth(w)
	vp.SetHeight(h)

	m.detail = deviceDetailState{
		open:     true,
		address:  address,
		title:    title,
		loading:  true,
		viewport: vp,
	}
	m.refreshDetailContent()
	return fetchDetailCmd(address)
}

// contentBudget returns the exact space available inside the bordered page
// box, below the tab bar and above the status bar. View() renders the box
// at this same size, so the detail viewport fills it exactly instead of
// leaving it short. Since the status bar's height depends on live state
// (e.g. whether help is expanded), callers must call resizeDetail whenever
// that state changes.
func (m Model) contentBudget() (w, h int) {
	frameW := windowStyle().GetHorizontalFrameSize()
	frameH := windowStyle().GetVerticalFrameSize()
	w = m.width - frameW

	statusH := strings.Count(m.renderStatusBar(), "\n") + 1
	tabBarH := lipgloss.Height(renderTabBar(w))

	h = m.height - tabBarH - frameH - statusH
	if h < 1 {
		h = 1
	}
	return w, h
}

func (m *Model) resizeDetail() {
	if !m.detail.open {
		return
	}
	w, h := m.contentBudget()
	m.detail.viewport.SetWidth(w)
	m.detail.viewport.SetHeight(h)
	m.refreshDetailContent()
}

func (m *Model) handleDetail(msg detailMsg) {
	if !m.detail.open {
		return
	}
	m.detail.loading = false
	m.detail.data = msg.detail
	m.detail.err = msg.err
	m.refreshDetailContent()
}

// closeDetailIfPending closes the info screen once the connect/disconnect
// it triggered has completed, but only if it's still the one waiting.
func (m *Model) closeDetailIfPending() {
	if m.detail.open && m.detail.pending {
		m.detail = deviceDetailState{}
	}
}

func (m *Model) updateDetail(msg tea.Msg) tea.Cmd {
	if wheel, ok := msg.(tea.MouseWheelMsg); ok {
		util.HandleMouseWheel(wheel, &m.detail.viewport)
		return nil
	}
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	g := keys.Keys.Global
	b := keys.Keys.Bluetooth

	switch {
	case key.Matches(keyMsg, g.Escape), key.Matches(keyMsg, g.CycleFocus):
		m.detail = deviceDetailState{}
	case key.Matches(keyMsg, g.Help):
		m.help.ShowAll = !m.help.ShowAll
		m.resizeDetail()
	case key.Matches(keyMsg, b.Select):
		m.detail.pending = true
		m.refreshDetailContent()
		switch {
		case m.detail.data.Connected:
			return disconnectCmd(m.detail.address)
		case m.detail.data.Paired:
			return connectCmd(m.detail.address)
		default:
			return pairCmd(m.detail.address)
		}
	case key.Matches(keyMsg, b.ToggleTrust):
		m.detail.pending = true
		m.refreshDetailContent()
		return trustCmd(m.detail.address, !m.detail.data.Trusted)
	case key.Matches(keyMsg, g.Up):
		util.ScrollLines(&m.detail.viewport, -1)
	case key.Matches(keyMsg, g.Down):
		util.ScrollLines(&m.detail.viewport, 1)
	case key.Matches(keyMsg, g.ScrollUp):
		util.ScrollViewport(&m.detail.viewport, -1)
	case key.Matches(keyMsg, g.ScrollDown):
		util.ScrollViewport(&m.detail.viewport, 1)
	}
	return nil
}

func (m Model) renderDetail() string {
	return m.detail.viewport.View()
}

func (m *Model) refreshDetailContent() {
	m.detail.viewport.SetContent(m.detailContent())
	util.RefreshScrollbar(&m.detail.viewport)
}

func (m Model) detailContent() string {
	d := m.detail
	if d.loading {
		return ilovetui.S.Faint.Render("Loading device info…")
	}

	title := lipgloss.NewStyle().Bold(true).Render(d.title)
	if d.pending {
		title += mutedStyle().Render("  working…")
	}
	rows := []string{title, subtleStyle().Render("Address:      ") + d.address}

	if d.err != nil {
		rows = append(rows, "", errorStyle().Render(d.err.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	rows = append(rows,
		subtleStyle().Render("Paired:       ")+yesNo(d.data.Paired),
		subtleStyle().Render("Trusted:      ")+yesNo(d.data.Trusted),
		subtleStyle().Render("Connected:    ")+yesNo(d.data.Connected),
		subtleStyle().Render("Blocked:      ")+yesNo(d.data.Blocked),
	)
	if d.data.Icon != "" {
		rows = append(rows, subtleStyle().Render("Type:         ")+d.data.Icon)
	}
	if d.data.Battery >= 0 {
		rows = append(rows, subtleStyle().Render("Battery:      ")+fmt.Sprintf("%d%%", d.data.Battery))
	}
	if d.data.Modalias != "" {
		rows = append(rows, subtleStyle().Render("Modalias:     ")+d.data.Modalias)
	}
	if len(d.data.UUIDs) > 0 {
		rows = append(rows, "", subtleStyle().Render("Services:"))
		for _, u := range d.data.UUIDs {
			rows = append(rows, "  "+u)
		}
	}

	if d.feedback != "" {
		rows = append(rows, "", errorStyle().Render(d.feedback))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func yesNo(b bool) string {
	if b {
		return successStyle().Render("yes")
	}
	return subtleStyle().Render("no")
}
