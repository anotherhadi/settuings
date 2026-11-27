package pages

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/skip2/go-qrcode"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/theme"
)

func qrAuth(security string) string {
	switch s := strings.ToLower(security); {
	case s == "":
		return "nopass"
	case strings.Contains(s, "sae"), strings.Contains(s, "wpa3"):
		return "SAE"
	case strings.Contains(s, "wpa"):
		return "WPA"
	}
	return "WEP"
}

var wifiEscape = strings.NewReplacer(`\`, `\\`, `;`, `\;`, `,`, `\,`, `:`, `\:`, `"`, `\"`)

func wifiPayload(ssid, security, secret string) string {
	auth := qrAuth(security)
	if auth == "nopass" {
		return "WIFI:T:nopass;S:" + wifiEscape.Replace(ssid) + ";;"
	}
	return "WIFI:T:" + auth + ";S:" + wifiEscape.Replace(ssid) + ";P:" + wifiEscape.Replace(secret) + ";;"
}

func qrLines(content string) ([]string, error) {
	q, err := qrcode.New(content, qrcode.Low)
	if err != nil {
		return nil, err
	}
	bits := q.Bitmap()
	shade := func(dark bool) color.Color {
		if dark {
			return lipgloss.Color("#000000")
		}
		return lipgloss.Color("#ffffff")
	}
	var lines []string
	for y := 0; y < len(bits); y += 2 {
		var b strings.Builder
		for x := range bits[y] {
			top, bottom := bits[y][x], false
			if y+1 < len(bits) {
				bottom = bits[y+1][x]
			}
			b.WriteString(lipgloss.NewStyle().Foreground(shade(top)).Background(shade(bottom)).Render("▀"))
		}
		lines = append(lines, b.String())
	}
	return lines, nil
}

type qrBox struct {
	name  string
	lines []string
	err   error
}

func newQR(name, payload string) *qrBox {
	l, err := qrLines(payload)
	return &qrBox{name: draw.Clean(name), lines: l, err: err}
}

func (m *qrBox) Scope() commands.Scope {
	return commands.Scope{Name: "QR code", Commands: []commands.Command{
		{ID: "help.close", Help: "close", Run: func() tea.Cmd { return overlay.Close }},
	}}
}

func (m *qrBox) Update(tea.Msg) tea.Cmd { return nil }

func (m *qrBox) View(maxW, maxH int) string {
	title := "Share · " + m.name
	if m.err != nil {
		return draw.Frame(" "+m.err.Error(), min(48, maxW), 3, draw.FrameOpts{Title: title, Active: true})
	}
	cols, rows := lipgloss.Width(m.lines[0]), len(m.lines)
	if cols+2 > maxW || rows+3 > maxH {
		msg := "too small for the QR code"
		need := theme.Subtle().Render(" needs " + itoa(cols+2) + " × " + itoa(rows+3))
		w := min(max(len(msg)+4, 40), maxW)
		return draw.Frame(" "+msg+"\n"+need, w, 4, draw.FrameOpts{Title: title, Active: true})
	}
	caption := theme.Subtle().Render("scan to join")
	pad := strings.Repeat(" ", max((cols-lipgloss.Width(caption))/2, 0))
	body := strings.Join(m.lines, "\n") + "\n" + pad + caption + strings.Repeat(" ", max(cols-lipgloss.Width(pad+caption), 0))
	return draw.Frame(body, cols+2, rows+3, draw.FrameOpts{Title: title, Active: true})
}
