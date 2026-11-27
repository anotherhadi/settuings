package pages

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/confirm"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/prompt"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/network"
)

type netKind int

const (
	netWifi netKind = iota
	netEth
	netVPN
	netKnown
)

type netRow struct {
	kind  netKind
	wifi  network.WifiNetwork
	eth   network.EthernetDevice
	vpn   network.VPNConnection
	known network.KnownNetwork
}

type netData struct {
	err     error
	radio   bool
	hasWifi bool
	dev     network.Device
	wifi    []network.WifiNetwork
	eth     []network.EthernetDevice
	vpn     []network.VPNConnection
	known   []network.KnownNetwork
	detail  map[string]network.DeviceDetail
}

type netMsg netData

type needSecretMsg struct{ name string }

type forgetNetMsg struct{ name string }

type qrMsg struct {
	name, security, secret string
	err                    error
}

type Network struct {
	pending
	list List
	d    netData
	busy bool
	load bool
}

func NewNetwork() *Network { return &Network{} }

func (p *Network) Available() bool { return network.Installed() }
func (p *Network) Name() string    { return "network" }
func (p *Network) Label() string   { return "Network" }
func (p *Network) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Network) Leave()          {}
func (p *Network) Close()          {}

func (p *Network) Refresh(manual bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg { return netMsg(fetchNetwork(manual)) }
}

func fetchNetwork(rescan bool) netData {
	var d netData
	if d.err = network.CheckAvailable(); d.err != nil {
		return d
	}
	d.radio, _ = network.RadioEnabled()
	dev, err := network.WifiDevice()
	d.hasWifi = err == nil
	d.dev = dev
	if d.hasWifi && d.radio {
		if rescan {
			_ = network.Rescan()
		}
		d.wifi, _ = network.ListWifi()
	}
	d.eth, _ = network.ListEthernet()
	d.vpn, _ = network.ListVPN()
	d.known, _ = network.ListKnownWifi()
	d.detail = map[string]network.DeviceDetail{}
	if d.hasWifi && dev.Connected() {
		d.detail[dev.Name], _ = network.DeviceDetails(dev.Name)
	}
	for _, e := range d.eth {
		if e.Connected() {
			d.detail[e.Device], _ = network.DeviceDetails(e.Device)
		}
	}
	return d
}

func (p *Network) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "network")
	switch m := msg.(type) {
	case needSecretMsg:
		p.on = false
		return p.askPassword(m.name, func(pw string) tea.Cmd {
			return p.do(tea.Batch(notify.Info("Connecting to "+m.name+"…"), act("network", "Connected to "+m.name, func() error {
				return network.ConnectKnownWithSecret(m.name, pw)
			})))
		})
	case forgetNetMsg:
		return p.do(act("network", "Forgot "+m.name, func() error { return network.Forget(m.name) }))
	case netMsg:
		p.busy, p.load = false, true
		p.d = netData(m)
		p.rebuild()
	case qrMsg:
		switch {
		case m.err != nil:
			return notify.Err(m.err)
		case m.secret == "" && qrAuth(m.security) != "nopass":
			return notify.Warning(m.name + ": no password to share")
		}
		return overlay.Open(newQR(m.name, wifiPayload(m.name, m.security, m.secret)))
	}
	return nil
}

func (p *Network) qrcode() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name, sec := r.name(), r.known.Security
	if r.kind == netWifi {
		sec = r.wifi.Security
		for _, k := range p.d.known {
			if k.Name == name {
				sec = k.Security
			}
		}
	}
	return func() tea.Msg {
		if qrAuth(sec) == "nopass" {
			return qrMsg{name: name, security: sec}
		}
		s, err := network.Secret(name)
		return qrMsg{name: name, security: sec, secret: s, err: err}
	}
}

func (p *Network) isKnown(ssid string) bool {
	for _, k := range p.d.known {
		if k.Name == ssid {
			return true
		}
	}
	return false
}

func (p *Network) rebuild() {
	d := p.d
	var items []Item
	if d.hasWifi {
		items = append(items, Item{Header: true, Label: "Wi-Fi"})
		switch {
		case !d.radio:
			items = append(items, Item{Header: true, Dim: true, Label: "Wi-Fi is off"})
		case len(d.wifi) == 0:
			items = append(items, Item{Header: true, Dim: true, Label: "no network in range"})
		}
		for _, w := range d.wifi {
			note := theme.Subtle().Render(itoa(w.Signal) + "%")
			if w.Secured() {
				note = theme.Muted().Render(icon("\uf023", "*")) + " " + note
			}
			if w.Active {
				note = theme.Success().Render("connected")
			}
			items = append(items, Item{Key: "wifi:" + w.SSID, Icon: icon("\uf1eb", ""), Label: w.SSID, Note: note, Data: netRow{kind: netWifi, wifi: w}})
		}
	}
	if len(d.eth) > 0 {
		items = append(items, Item{Header: true, Label: "Ethernet"})
	}
	for _, e := range d.eth {
		note := theme.Subtle().Render(e.State)
		if e.Connected() {
			note = theme.Success().Render(e.IP4)
			if e.IP4 == "" {
				note = theme.Success().Render("connected")
			}
		}
		items = append(items, Item{Key: "eth:" + e.Device, Icon: icon("\U000f0200", ""), Label: e.Device, Note: note, Data: netRow{kind: netEth, eth: e}})
	}
	if len(d.vpn) > 0 {
		items = append(items, Item{Header: true, Label: "VPN"})
	}
	for _, v := range d.vpn {
		note := theme.Subtle().Render(v.Type)
		if v.Active {
			note = theme.Success().Render("active")
		}
		items = append(items, Item{Key: "vpn:" + v.Name, Icon: icon("\U000f0582", ""), Label: v.Name, Note: note, Data: netRow{kind: netVPN, vpn: v}})
	}
	if len(d.known) > 0 {
		items = append(items, Item{Header: true, Label: "Known networks"})
	}
	for _, k := range d.known {
		note := ""
		if k.AutoConnect {
			note = theme.Subtle().Render("auto")
		}
		if k.Active {
			note = theme.Success().Render("active")
		}
		items = append(items, Item{Key: "known:" + k.Name, Icon: icon("\uf1eb", ""), Label: k.Name, Note: note, Data: netRow{kind: netKnown, known: k}})
	}
	p.list.Set(items)
}

func (p *Network) row() (netRow, bool) {
	it, ok := p.list.Selected()
	if !ok {
		return netRow{}, false
	}
	r, ok := it.Data.(netRow)
	return r, ok
}

func (r netRow) name() string {
	switch r.kind {
	case netWifi:
		return r.wifi.SSID
	case netEth:
		return r.eth.Device
	case netVPN:
		return r.vpn.Name
	}
	return r.known.Name
}

func (r netRow) active() bool {
	switch r.kind {
	case netWifi:
		return r.wifi.Active
	case netEth:
		return r.eth.Connected()
	case netVPN:
		return r.vpn.Active
	}
	return r.known.Active
}

func (p *Network) askPassword(name string, submit func(string) tea.Cmd) tea.Cmd {
	pr := prompt.New("Password · "+name, "", submit)
	pr.Secret = true
	pr.Validate = func(s string) error {
		if s == "" {
			return network.ErrEmptyPassword
		}
		return nil
	}
	return overlay.Open(pr)
}

func (p *Network) up(name, done string) tea.Cmd {
	return func() tea.Msg {
		err := network.ConnectKnown(name)
		if errors.Is(err, network.ErrSecrets) {
			return needSecretMsg{name: name}
		}
		return DoneMsg{Page: "network", OK: done, Err: err}
	}
}

func (p *Network) connect() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name := r.name()
	start := notify.Info("Connecting to " + name + "…")
	done := "Connected to " + name
	switch r.kind {
	case netWifi:
		switch {
		case p.isKnown(name):
			return p.do(tea.Batch(start, p.up(name, done)))
		case r.wifi.Secured():
			return p.askPassword(name, func(pw string) tea.Cmd {
				return p.do(tea.Batch(start, act("network", done, func() error { return network.ConnectWifi(name, pw) })))
			})
		}
		return p.do(tea.Batch(start, act("network", done, func() error { return network.ConnectWifi(name, "") })))
	case netEth:
		return p.do(tea.Batch(start, act("network", done, func() error { return network.DeviceConnect(name) })))
	case netVPN:
		return p.do(tea.Batch(start, act("network", done, func() error { return network.VPNUp(name) })))
	}
	return p.do(tea.Batch(start, p.up(name, done)))
}

func (p *Network) disconnect() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name := r.name()
	done := "Disconnected from " + name
	switch r.kind {
	case netEth:
		return p.do(act("network", done, func() error { return network.EthernetDown(name) }))
	case netVPN:
		return p.do(act("network", done, func() error { return network.VPNDown(name) }))
	}
	return p.do(act("network", done, func() error { return network.DisconnectWifi(p.d.dev.Name) }))
}

func (p *Network) forget() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name := r.name()
	c := confirm.New("Forget", fmt.Sprintf("Forget %q and its saved password?", name),
		func() tea.Msg { return forgetNetMsg{name} })
	c.Danger = true
	return overlay.Open(c)
}

func (p *Network) autoconnect() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name := r.name()
	on := true
	for _, k := range p.d.known {
		if k.Name == name {
			on = !k.AutoConnect
		}
	}
	return p.do(act("network", "Auto-connect "+onOff(on)+" for "+name, func() error { return network.SetAutoConnect(name, on) }))
}

func (p *Network) secret() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	name := r.name()
	return func() tea.Msg {
		s, err := network.Secret(name)
		switch {
		case err != nil:
			return notify.Msg{Level: notify.ErrorLevel, Text: err.Error()}
		case s == "":
			return notify.Msg{Level: notify.InfoLevel, Text: name + ": no saved password"}
		}
		return notify.Msg{Level: notify.InfoLevel, Text: name + ": " + s, Duration: 15e9}
	}
}

func (p *Network) Commands() []commands.Command {
	sel := func(f func(netRow) bool) func() bool {
		return func() bool { r, ok := p.row(); return ok && f(r) }
	}
	saved := func(r netRow) bool { return r.kind == netKnown || r.kind == netWifi && p.isKnown(r.wifi.SSID) }
	cmds := []commands.Command{
		{ID: "network.connect", Title: "Connect", Help: "connect", When: sel(func(r netRow) bool { return !r.active() && !(r.kind == netEth && r.eth.State == "unavailable") }), Run: p.connect},
		{ID: "network.disconnect", Title: "Disconnect", Help: "disconnect", When: sel(netRow.active), Run: p.disconnect},
		{ID: "network.forget", Title: "Forget the saved network", Help: "forget", When: sel(saved), Run: p.forget},
		{ID: "network.autoconnect", Title: "Toggle auto-connect", When: sel(saved), Run: p.autoconnect},
		{ID: "network.secret", Title: "Show the saved password", When: sel(saved), Run: p.secret},
		{ID: "network.qrcode", Title: "Share the network with a QR code", Help: "qr", When: sel(saved), Run: p.qrcode},
		{ID: "network.radio", Title: "Toggle the Wi-Fi radio", Help: "radio", When: func() bool { return p.d.hasWifi }, Run: func() tea.Cmd {
			on := !p.d.radio
			return p.do(act("network", "Wi-Fi "+onOff(on), func() error { return network.SetWifiRadio(on) }))
		}},
		{ID: "network.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Network", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func (p *Network) ipLines(dev string) []string {
	d, ok := p.d.detail[dev]
	if !ok {
		return nil
	}
	lines := []string{blank()}
	add := func(k, v string) {
		if v != "" {
			lines = append(lines, kv(k, v))
		}
	}
	add("IPv4", d.IPv4)
	add("Gateway", d.Gateway4)
	add("DNS", strings.Join(d.DNS4, ", "))
	for _, a := range d.IPv6 {
		add("IPv6", a)
	}
	add("MAC", d.HWAddr)
	add("MTU", d.MTU)
	return lines
}

func (p *Network) detailLines() []string {
	r, ok := p.row()
	if !ok {
		return nil
	}
	state := func(on bool, yes, no string) string {
		if on {
			return theme.Success().Render(yes)
		}
		return theme.Subtle().Render(no)
	}
	switch r.kind {
	case netWifi:
		w := r.wifi
		sec := w.Security
		if !w.Secured() {
			sec = "open"
		}
		lines := []string{
			kv("SSID", w.SSID),
			kv("State", state(w.Active, "connected", map[bool]string{true: "saved", false: "new"}[p.isKnown(w.SSID)])),
			kv("Signal", level(w.Signal, 16)),
			kv("Security", sec),
			kv("BSSID", w.BSSID),
		}
		if w.Active {
			lines = append(lines, p.ipLines(p.d.dev.Name)...)
		}
		return lines
	case netEth:
		e := r.eth
		return append([]string{kv("Device", e.Device), kv("State", state(e.Connected(), e.State, e.State)), kv("Profile", e.Connection)}, p.ipLines(e.Device)...)
	case netVPN:
		v := r.vpn
		return []string{kv("Name", v.Name), kv("Type", v.Type), kv("State", state(v.Active, "active", "inactive")), kv("Device", v.Device)}
	}
	k := r.known
	return []string{kv("Name", k.Name), kv("Security", k.Security), kv("Auto-connect", yesNo(k.AutoConnect)), kv("State", state(k.Active, "active", "inactive"))}
}

func (p *Network) View(w, h int) string {
	f := Frame{Title: icon("\U000f05a9 ", "") + "Network", Right: "", DetailTitle: "Details", Empty: "no network device"}
	if p.d.hasWifi {
		f.Right = "Wi-Fi " + onOff(p.d.radio)
	}
	switch {
	case p.d.err != nil:
		f.Empty = p.d.err.Error()
	case !p.load:
		f.Empty = "loading"
	default:
		f.Detail = p.detailLines()
	}
	if p.on {
		f.Right = "working…"
	}
	return p.list.Render(w, h, f)
}

func (p *Network) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) {
		if r, ok := p.row(); ok && !r.active() {
			return p.connect()
		}
	}
	return nil
}
