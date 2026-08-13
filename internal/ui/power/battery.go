package power

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/power"
)

func (m *Model) handleBattery(msg batteryMsg) {
	m.batteryLoading = false
	m.batteryErr = msg.err
	if msg.err != nil {
		return
	}
	m.batteries = msg.batteries
	m.ac = msg.ac
}

func (m Model) renderBattery(width int) string {
	title := lipgloss.NewStyle().Bold(true).Render("Battery")

	switch {
	case m.batteryLoading:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", ilovetui.S.Faint.Render("Reading battery status…"))
	case m.batteryErr != nil:
		return lipgloss.JoinVertical(lipgloss.Left, title, "", errorStyle().Render(m.batteryErr.Error()))
	}

	lines := []string{title, ""}
	if len(m.batteries) == 0 {
		lines = append(lines, subtleStyle().Render("No battery detected — this looks like a desktop."))
	}
	for i, b := range m.batteries {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, renderBatteryInfo(b, width)...)
	}

	lines = append(lines, "", subtleStyle().Render("Power source:  ")+acLabel(m.ac))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderBatteryInfo(b power.Battery, width int) []string {
	rows := []string{
		subtleStyle().Render("Name:          ") + b.Name,
		subtleStyle().Render("Status:        ") + statusLabel(b.Status),
		subtleStyle().Render("Charge:        ") + renderBatteryBar(b.Capacity, width),
	}
	if remaining := timeRemainingLabel(b); remaining != "" {
		rows = append(rows, subtleStyle().Render("Time left:     ")+remaining)
	}
	if b.Health >= 0 {
		rows = append(rows, subtleStyle().Render("Health:        ")+fmt.Sprintf("%d%%", b.Health))
	}
	if b.CycleCount >= 0 {
		rows = append(rows, subtleStyle().Render("Cycle count:   ")+fmt.Sprintf("%d", b.CycleCount))
	}
	if b.Technology != "" {
		rows = append(rows, subtleStyle().Render("Technology:    ")+b.Technology)
	}
	if model := strings.TrimSpace(b.Manufacturer + " " + b.Model); model != "" {
		rows = append(rows, subtleStyle().Render("Model:         ")+model)
	}
	return rows
}

func timeRemainingLabel(b power.Battery) string {
	switch {
	case b.TimeToEmpty != "":
		return b.TimeToEmpty + " until empty"
	case b.TimeToFull != "":
		return b.TimeToFull + " until full"
	}
	return ""
}

func statusLabel(status string) string {
	switch status {
	case "Charging":
		return successStyle().Render("⚡ Charging")
	case "Full":
		return successStyle().Render(status)
	case "Discharging":
		return status
	default:
		return mutedStyle().Render(status)
	}
}

func acLabel(ac power.ACStatus) string {
	switch {
	case !ac.Present:
		return mutedStyle().Render("unknown")
	case ac.Online:
		return successStyle().Render("connected")
	default:
		return mutedStyle().Render("disconnected")
	}
}

// renderBatteryBar draws a static block-bar gauge for the charge level,
// colored by how low it is, same spirit as the audio page's volume bar.
func renderBatteryBar(pct, width int) string {
	if pct < 0 {
		return subtleStyle().Render("unknown")
	}

	barWidth := width / 3
	switch {
	case barWidth < 10:
		barWidth = 10
	case barWidth > 30:
		barWidth = 30
	}

	filled := pct * barWidth / 100
	switch {
	case filled > barWidth:
		filled = barWidth
	case filled < 0:
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	color := ilovetui.S.Success
	switch {
	case pct <= 15:
		color = ilovetui.S.Error
	case pct <= 35:
		color = ilovetui.S.Warning
	}
	gauge := lipgloss.NewStyle().Foreground(color).Render("[" + bar + "]")
	return fmt.Sprintf("%s %3d%%", gauge, pct)
}
