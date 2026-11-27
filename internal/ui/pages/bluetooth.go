package pages

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/confirm"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/bluetooth"
)

type btData struct {
	err     error
	ctl     bluetooth.ControllerInfo
	devices []bluetooth.Device
}

type btMsg btData

type scanDoneMsg struct{}

type forgetBtMsg struct{ d bluetooth.Device }

type Bluetooth struct {
	pending
	list     List
	d        btData
	busy     bool
	load     bool
	scanning bool
}

func NewBluetooth() *Bluetooth { return &Bluetooth{} }

func (p *Bluetooth) Available() bool { return bluetooth.Installed() }
func (p *Bluetooth) Name() string    { return "bluetooth" }
func (p *Bluetooth) Label() string   { return "Bluetooth" }
func (p *Bluetooth) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Bluetooth) Leave()          {}
func (p *Bluetooth) Close()          {}

func (p *Bluetooth) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		var d btData
		if d.err = bluetooth.CheckAvailable(); d.err != nil {
			return btMsg(d)
		}
		if d.ctl, d.err = bluetooth.ShowController(); d.err != nil {
			return btMsg(d)
		}
		if d.ctl.Powered {
			d.devices, d.err = bluetooth.ListDevices()
		}
		return btMsg(d)
	}
}

func (p *Bluetooth) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "bluetooth")
	switch m := msg.(type) {
	case forgetBtMsg:
		return p.do(act("bluetooth", "Forgot "+m.d.Name, func() error { return bluetooth.Remove(m.d.Address) }))
	case btMsg:
		p.busy, p.load = false, true
		p.d = btData(m)
		p.rebuild()
	case scanDoneMsg:
		p.scanning = false
		return p.Refresh(false)
	}
	return nil
}

func btIcon(d bluetooth.Device) string {
	g := map[string]string{
		"input-mouse":      "\U000f037d",
		"input-keyboard":   "\uf11c",
		"input-gaming":     "\uf11b",
		"audio-headset":    "\uf025",
		"audio-headphones": "\uf025",
		"audio-card":       "\U000f04c3",
		"phone":            "\uf10b",
		"computer":         "\uf108",
	}[d.Icon]
	if g == "" {
		g = "\U000f00af"
	}
	return icon(g, "")
}

func (p *Bluetooth) rebuild() {
	var paired, nearby []Item
	devs := append([]bluetooth.Device(nil), p.d.devices...)
	sort.Slice(devs, func(i, j int) bool { return strings.ToLower(devs[i].Name) < strings.ToLower(devs[j].Name) })
	for _, d := range devs {
		var note []string
		if d.Battery >= 0 {
			note = append(note, theme.Subtle().Render(itoa(d.Battery)+"%"))
		}
		if d.Connected {
			note = append(note, theme.Success().Render("connected"))
		}
		it := Item{Key: d.Address, Icon: btIcon(d), Label: d.Name, Note: strings.Join(note, " "), Data: d}
		if d.Paired {
			paired = append(paired, it)
		} else {
			it.Dim = true
			nearby = append(nearby, it)
		}
	}
	var items []Item
	if len(paired) > 0 {
		items = append(items, Item{Header: true, Label: "Paired"})
		items = append(items, paired...)
	}
	if len(nearby) > 0 {
		items = append(items, Item{Header: true, Label: "Nearby"})
		items = append(items, nearby...)
	}
	p.list.Set(items)
}

func (p *Bluetooth) dev() (bluetooth.Device, bool) {
	it, ok := p.list.Selected()
	if !ok {
		return bluetooth.Device{}, false
	}
	d, ok := it.Data.(bluetooth.Device)
	return d, ok
}

func (p *Bluetooth) connect() tea.Cmd {
	d, ok := p.dev()
	if !ok {
		return nil
	}
	start := notify.Info("Connecting to " + d.Name + "…")
	return p.do(tea.Batch(start, act("bluetooth", "Connected to "+d.Name, func() error {
		if !d.Paired {
			if err := bluetooth.Pair(d.Address); err != nil {
				return err
			}
			_ = bluetooth.SetTrusted(d.Address, true)
		}
		return bluetooth.Connect(d.Address)
	})))
}

func (p *Bluetooth) scan() tea.Cmd {
	if p.scanning {
		return nil
	}
	p.scanning = true
	return tea.Batch(notify.Info("Scanning…"), func() tea.Msg {
		_ = bluetooth.Scan(8)
		return scanDoneMsg{}
	})
}

func (p *Bluetooth) Commands() []commands.Command {
	sel := func(f func(bluetooth.Device) bool) func() bool {
		return func() bool { d, ok := p.dev(); return ok && f(d) }
	}
	cmds := []commands.Command{
		{ID: "bluetooth.connect", Title: "Connect (pair if needed)", Help: "connect", When: sel(func(d bluetooth.Device) bool { return !d.Connected }), Run: p.connect},
		{ID: "bluetooth.disconnect", Title: "Disconnect", Help: "disconnect", When: sel(func(d bluetooth.Device) bool { return d.Connected }), Run: func() tea.Cmd {
			d, _ := p.dev()
			return p.do(act("bluetooth", "Disconnected from "+d.Name, func() error { return bluetooth.Disconnect(d.Address) }))
		}},
		{ID: "bluetooth.trust", Title: "Toggle trust", When: sel(func(d bluetooth.Device) bool { return d.Paired }), Run: func() tea.Cmd {
			d, _ := p.dev()
			return p.do(act("bluetooth", "", func() error { return bluetooth.SetTrusted(d.Address, !d.Trusted) }))
		}},
		{ID: "bluetooth.forget", Title: "Forget the device", Help: "forget", When: sel(func(d bluetooth.Device) bool { return true }), Run: func() tea.Cmd {
			d, _ := p.dev()
			c := confirm.New("Forget", fmt.Sprintf("Forget %q?", d.Name), func() tea.Msg { return forgetBtMsg{d} })
			c.Danger = true
			return overlay.Open(c)
		}},
		{ID: "bluetooth.scan", Title: "Scan for devices", Help: "scan", When: func() bool { return p.d.ctl.Powered }, Run: p.scan},
		{ID: "bluetooth.power", Title: "Toggle the adapter", Help: "power", When: func() bool { return p.d.err == nil }, Run: func() tea.Cmd {
			on := !p.d.ctl.Powered
			return p.do(act("bluetooth", "Bluetooth "+onOff(on), func() error { return bluetooth.SetPower(on) }))
		}},
		{ID: "bluetooth.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Bluetooth", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func (p *Bluetooth) detailLines() []string {
	d, ok := p.dev()
	if !ok {
		return nil
	}
	lines := []string{kv("Name", d.Name), kv("Address", d.Address), kv("Paired", yesNo(d.Paired)), kv("Trusted", yesNo(d.Trusted)), kv("Connected", yesNo(d.Connected))}
	if d.Battery >= 0 {
		lines = append(lines, kv("Battery", level(d.Battery, 16)))
	}
	if d.Blocked {
		lines = append(lines, kv("Blocked", "yes"))
	}
	if d.Icon != "" {
		lines = append(lines, kv("Type", d.Icon))
	}
	return lines
}

func (p *Bluetooth) View(w, h int) string {
	f := Frame{Title: icon("\U000f00af ", "") + "Bluetooth", DetailTitle: "Details", Empty: "no device: scan with s"}
	switch {
	case p.d.err != nil:
		f.Empty = p.d.err.Error()
	case !p.load:
		f.Empty = "loading"
	case !p.d.ctl.Powered:
		f.Empty = "bluetooth is off: w to power on"
		f.Right = "off"
	default:
		f.Right = "on"
		if p.scanning {
			f.Right = "scanning"
		}
		f.Detail = p.detailLines()
	}
	if p.on {
		f.Right = "working…"
	}
	return p.list.Render(w, h, f)
}

func (p *Bluetooth) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) {
		if d, ok := p.dev(); ok && !d.Connected {
			return p.connect()
		}
	}
	return nil
}
