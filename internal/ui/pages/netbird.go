package pages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/netbird"
)

type nbData struct {
	err   error
	state netbird.State
	nets  []netbird.Network
}

type nbMsg nbData

type NetBird struct {
	pending
	list List
	d    nbData
	busy bool
	load bool
}

func NewNetBird() *NetBird { return &NetBird{} }

func (p *NetBird) Available() bool { return netbird.Installed() }
func (p *NetBird) Name() string    { return "netbird" }
func (p *NetBird) Label() string   { return "NetBird" }
func (p *NetBird) Init() tea.Cmd   { return p.Refresh(true) }
func (p *NetBird) Leave()          {}
func (p *NetBird) Close()          {}

func (p *NetBird) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		var d nbData
		if d.state, d.err = netbird.Read(); d.err == nil && d.state.Up() {
			d.nets, _ = netbird.Networks()
		}
		return nbMsg(d)
	}
}

func (p *NetBird) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "netbird")
	m, ok := msg.(nbMsg)
	if !ok {
		return nil
	}
	p.busy, p.load = false, true
	p.d = nbData(m)
	p.rebuild()
	return nil
}

func (p *NetBird) rebuild() {
	var items []Item
	if p.d.err == nil && p.d.state.Up() {
		var exits, others []Item
		for _, n := range p.d.nets {
			note := ""
			if n.Selected {
				note = theme.Success().Render(map[bool]string{true: "active", false: "on"}[n.ExitNode()])
			}
			it := Item{Key: "net:" + n.ID, Label: n.ID, Note: note, Data: n}
			if n.ExitNode() {
				it.Icon = icon("\U000f059f", "")
				exits = append(exits, it)
			} else {
				it.Icon = icon("\U000f0317", "")
				if !n.Selected {
					it.Dim = true
				}
				others = append(others, it)
			}
		}
		hdr := Item{Header: true, Label: "Exit nodes"}
		for _, n := range p.d.nets {
			if n.Selected && n.ExitNode() {
				hdr.Key, hdr.Badge = "hdr:exit", badge(true)
			}
		}
		items = append(items, hdr)
		if len(exits) == 0 {
			items = append(items, Item{Header: true, Dim: true, Label: "no exit node available"})
		}
		items = append(items, exits...)
		if len(others) > 0 {
			items = append(items, Item{Header: true, Label: "Networks"})
			items = append(items, others...)
		}
		if len(p.d.state.Peers) > 0 {
			items = append(items, Item{Header: true, Label: "Peers"})
		}
		for _, peer := range p.d.state.Peers {
			note := theme.Subtle().Render(strings.ToLower(peer.Status))
			if peer.Connected() {
				note = theme.Success().Render(latency(peer.Latency))
			}
			items = append(items, Item{Key: "peer:" + peer.FQDN, Icon: icon("\uf233", ""), Label: strings.TrimSuffix(peer.FQDN, ".nb"), Note: note, Dim: !peer.Connected(), Data: peer})
		}
	}
	p.list.Set(items)
}

func latency(d time.Duration) string {
	switch {
	case d <= 0:
		return "connected"
	case d < time.Millisecond:
		return "<1ms"
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func (p *NetBird) network() (netbird.Network, bool) {
	it, ok := p.list.Selected()
	if !ok {
		return netbird.Network{}, false
	}
	n, ok := it.Data.(netbird.Network)
	return n, ok
}

func (p *NetBird) toggle() tea.Cmd {
	n, ok := p.network()
	if !ok {
		return nil
	}
	nets := append([]netbird.Network(nil), p.d.nets...)
	ok1 := "Using " + n.ID + " as exit node"
	switch {
	case n.Selected && n.ExitNode():
		ok1 = "Stopped using " + n.ID
	case n.Selected:
		ok1 = n.ID + " off"
	case !n.ExitNode():
		ok1 = n.ID + " on"
	}
	return p.do(act("netbird", ok1, func() error { return netbird.Toggle(n.ID, nets) }))
}

func (p *NetBird) loginHelp() (string, bool) {
	switch {
	case errors.Is(p.d.err, netbird.ErrNoDaemon):
		return netbird.StartCommand, true
	case p.d.err == nil && p.d.state.NeedsLogin():
		return netbird.LoginCommand, true
	}
	return "", false
}

func (p *NetBird) Commands() []commands.Command {
	ready := func() bool { return p.load && p.d.err == nil }
	cmds := []commands.Command{
		{ID: "netbird.select", Title: "Use as exit node, or toggle the network", Help: "select", When: func() bool { _, ok := p.network(); return ok }, Run: p.toggle},
		{ID: "netbird.connect", Title: "Connect to NetBird", Help: "connect", When: func() bool { return ready() && p.d.state.Daemon == netbird.Idle }, Run: func() tea.Cmd {
			return p.do(tea.Batch(notify.Info("Connecting to NetBird…"), act("netbird", "Connected to NetBird", netbird.Up)))
		}},
		{ID: "netbird.disconnect", Title: "Disconnect from NetBird", Help: "disconnect", When: func() bool { return ready() && p.d.state.Up() }, Run: func() tea.Cmd {
			return p.do(act("netbird", "Disconnected from NetBird", netbird.Down))
		}},
		{ID: "netbird.copy", Title: "Copy the command", Help: "copy", When: func() bool { _, ok := p.loginHelp(); return ok }, Run: func() tea.Cmd {
			c, _ := p.loginHelp()
			return copyText(c, "the command")
		}},
		{ID: "netbird.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("NetBird", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func bytesText(n int64) string {
	f, unit := float64(n), "B"
	for _, u := range []string{"KiB", "MiB", "GiB"} {
		if f < 1024 {
			break
		}
		f, unit = f/1024, u
	}
	if unit == "B" {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.1f %s", f, unit)
}

func (p *NetBird) detailLines() []string {
	it, ok := p.list.Selected()
	if !ok {
		return nil
	}
	switch v := it.Data.(type) {
	case netbird.Network:
		kind, state := "network", theme.Subtle().Render("off")
		if v.ExitNode() {
			kind = "exit node"
			state = theme.Subtle().Render("not used")
		}
		if v.Selected {
			state = theme.Success().Render("active")
		}
		lines := []string{kv("ID", v.ID), kv("Type", kind), kv("State", state)}
		if len(v.Ranges) > 0 {
			lines = append(lines, kv("Routes", strings.Join(v.Ranges, ", ")))
		}
		if len(v.Domains) > 0 {
			lines = append(lines, kv("Domains", strings.Join(v.Domains, ", ")))
		}
		if v.ExitNode() {
			lines = append(lines, blank(), theme.Subtle().Render("All your traffic goes through this peer while it is active."))
		}
		return lines
	case netbird.Peer:
		lines := []string{kv("Name", v.FQDN), kv("IP", v.IP), kv("Status", v.Status)}
		if v.Type != "" && v.Type != "-" {
			lines = append(lines, kv("Connection", v.Type))
		}
		if v.Connected() && v.Latency > 0 {
			lines = append(lines, kv("Latency", latency(v.Latency)))
		}
		if !v.Handshake.IsZero() {
			lines = append(lines, kv("Handshake", ago(v.Handshake)))
		}
		if v.Received+v.Sent > 0 {
			lines = append(lines, kv("Received", bytesText(v.Received)), kv("Sent", bytesText(v.Sent)))
		}
		return lines
	}
	return nil
}

func ago(t time.Time) string {
	d := time.Since(t).Round(time.Second)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

func (p *NetBird) message() []string {
	code := theme.Primary().Style().Bold(true)
	switch {
	case errors.Is(p.d.err, netbird.ErrNoDaemon):
		return []string{
			theme.Warning().Render("The NetBird daemon is not running."), "",
			"Start it with:", "", "  " + code.Render(netbird.StartCommand), "",
			theme.Subtle().Render("Then press r to refresh."),
		}
	case p.d.err != nil:
		return []string{theme.Danger().Render(p.d.err.Error())}
	}
	head := "You are not logged in to NetBird."
	if p.d.state.Daemon == netbird.SessionExpired {
		head = "Your NetBird session has expired."
	}
	return []string{
		theme.Warning().Render(head), "",
		"Log in from a terminal:", "", "  " + code.Render(netbird.LoginCommand), "",
		"It opens your browser, or prints a link. With a setup key:", "",
		"  " + code.Render("netbird up --setup-key <KEY>"), "",
		theme.Subtle().Render("Then press r to refresh."),
	}
}

func (p *NetBird) View(w, h int) string {
	f := Frame{Title: icon("\U000f059f ", "") + "NetBird", DetailTitle: "Details", Empty: "loading"}
	_, help := p.loginHelp()
	switch {
	case !p.load:
	case help || p.d.err != nil:
		iw, ih := max(w-2, 0), max(h-2, 0)
		var lines []string
		for _, l := range p.message() {
			for _, r := range strings.Split(ansi.Wrap(l, max(iw-2, 1), ""), "\n") {
				lines = append(lines, " "+r)
			}
		}
		return draw.Frame(draw.Fit(strings.Join(lines, "\n"), iw, ih), w, h, draw.FrameOpts{Title: f.Title, Active: true})
	case !p.d.state.Up():
		f.Empty = "disconnected: press c to connect"
		f.Right = "disconnected"
	default:
		f.Right = strings.SplitN(p.d.state.IP, "/", 2)[0]
		f.Detail = p.detailLines()
	}
	if p.on {
		f.Right = "working…"
	}
	return p.list.Render(w, h, f)
}

func (p *NetBird) clearExit() tea.Cmd {
	nets := append([]netbird.Network(nil), p.d.nets...)
	for _, n := range nets {
		if n.Selected && n.ExitNode() {
			return p.do(act("netbird", "Stopped using "+n.ID, func() error { return netbird.Toggle(n.ID, nets) }))
		}
	}
	return nil
}

func (p *NetBird) Mouse(msg tea.MouseMsg) tea.Cmd {
	if k, ok := p.list.BadgeClick(msg); ok && k == "hdr:exit" {
		return p.clearExit()
	}
	if p.list.Mouse(msg) {
		return p.toggle()
	}
	return nil
}
