package pages

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/settuings/internal/about"
	"github.com/anotherhadi/settuings/internal/monitor"
	"github.com/anotherhadi/settuings/internal/netbird"
	"github.com/anotherhadi/settuings/internal/power"
)

func TestSectionSpacing(t *testing.T) {
	var l List
	l.Set([]Item{
		{Header: true, Label: "One"}, {Key: "a", Label: "a"}, {Key: "b", Label: "b"},
		{Header: true, Label: "Two"}, {Key: "c", Label: "c"},
	})
	var got []string
	for _, it := range l.items {
		got = append(got, it.Label)
	}
	if want := "One||a|b||Two||c"; strings.Join(got, "|") != want {
		t.Fatalf("items = %q, want %q", strings.Join(got, "|"), want)
	}
	if it, ok := l.Selected(); !ok || it.Key != "a" {
		t.Fatalf("selected = %+v", it)
	}
	l.Move(2)
	if it, _ := l.Selected(); it.Key != "c" {
		t.Fatalf("after moving past a header: %+v", it)
	}
	l.Move(-1)
	if it, _ := l.Selected(); it.Key != "b" {
		t.Fatalf("moving back: %+v", it)
	}
}

func TestPowerIgnoresASecondUse(t *testing.T) {
	p := NewPower()
	p.d.profiles = []power.Profile{{Name: "balanced", Active: true}, {Name: "performance"}}
	p.rebuild()
	if p.use("performance") == nil {
		t.Fatal("first use did nothing")
	}
	if p.use("performance") != nil {
		t.Fatal("second use ran again while the first is pending")
	}
	if !p.d.profiles[1].Active || p.d.profiles[0].Active {
		t.Fatalf("not applied locally: %+v", p.d.profiles)
	}
	p.Update(DoneMsg{Page: "power"})
	if p.use("balanced") == nil {
		t.Fatal("blocked after the first finished")
	}
}

func TestAboutWraps(t *testing.T) {
	logo := []string{strings.Repeat("#", 20), strings.Repeat("#", 20)}
	p := NewAbout()
	p.info = about.Info{Logo: logo, Title: "me@host", Fields: []about.Field{
		{Key: "Packages", Value: "2827 (nix-system), 2273 (nix-user), 40 (flatpak-system), 12 (snap)"},
		{Key: "Display (BOE0B7D)", Value: strings.Repeat("verylongword", 8)},
	}}
	for _, w := range []int{120, 70, 50, 30, 12} {
		for _, l := range p.lines(w, 30) {
			if got := ansi.StringWidth(l); got > w {
				t.Fatalf("w=%d: line is %d wide: %q", w, got, ansi.Strip(l))
			}
		}
	}
	joined := ansi.Strip(strings.Join(p.lines(40, 30), "\n"))
	if !strings.Contains(joined, "flatpak-") || !strings.Contains(joined, "(snap)") {
		t.Fatalf("text lost when wrapping:\n%s", joined)
	}
}

func TestKeepOrRevert(t *testing.T) {
	p := NewMonitors(5)
	p.prev = []monitor.Output{{Name: "eDP-1", Active: true}}
	apply := appliedMsg{prev: p.prev, desc: "Mode X"}

	p.askKeep(apply)
	gen := p.gen
	if p.left != revertAfter || p.keep == nil {
		t.Fatalf("not armed: left=%d keep=%v", p.left, p.keep)
	}
	if cmd := p.Update(keepTickMsg{gen}); cmd == nil || p.left != revertAfter-1 {
		t.Fatalf("tick did not count down: left=%d", p.left)
	}
	if !strings.Contains(p.keep.Question, "9s") {
		t.Fatalf("question = %q", p.keep.Question)
	}
	if cmd := p.Update(keepTickMsg{gen - 1}); cmd != nil || p.left != revertAfter-1 {
		t.Fatal("a stale tick counted")
	}
	p.Update(keepMsg{gen})
	if p.gen == gen {
		t.Fatal("keeping did not disarm")
	}
	if cmd := p.Update(keepTickMsg{gen}); cmd != nil {
		t.Fatal("a tick after keeping did something")
	}
	if cmd := p.Update(revertMsg{gen}); cmd != nil {
		t.Fatal("a revert after keeping did something")
	}

	p.askKeep(apply)
	gen = p.gen
	p.left = 1
	if cmd := p.Update(keepTickMsg{gen}); cmd == nil || p.gen == gen {
		t.Fatal("the deadline did not revert")
	}
	if cmd := p.Update(keepMsg{gen}); cmd != nil {
		t.Fatal("answering after the deadline did something")
	}

	p.askKeep(apply)
	gen = p.gen
	if cmd := p.Update(revertMsg{gen}); cmd == nil || p.gen == gen {
		t.Fatal("saying no did not revert")
	}
}

func TestLayoutPreviewFits(t *testing.T) {
	outs := []monitor.Output{
		{Name: "eDP-1", Active: true, Rect: monitor.Rect{X: 0, Y: 0, Width: 2560, Height: 1440}},
		{Name: "DP-9", Active: true, Rect: monitor.Rect{X: 2560, Y: -300, Width: 3440, Height: 1440}},
		{Name: "HDMI-A-1", Active: false},
	}
	for _, size := range [][2]int{{36, 8}, {20, 5}, {80, 12}, {16, 5}} {
		lines := layoutPreview(outs, "DP-9", size[0], size[1])
		if len(lines) != size[1] {
			t.Fatalf("%v: %d lines", size, len(lines))
		}
		for _, l := range lines {
			if w := ansi.StringWidth(l); w != size[0] {
				t.Fatalf("%v: line is %d wide", size, w)
			}
		}
	}
	if layoutPreview(outs, "DP-9", 10, 4) != nil || layoutPreview(nil, "", 36, 8) != nil {
		t.Fatal("too small or empty should draw nothing")
	}
	joined := ansi.Strip(strings.Join(layoutPreview(outs, "DP-9", 36, 8), "\n"))
	if !strings.Contains(joined, "DP-9") || !strings.Contains(joined, "eDP-1") || strings.Contains(joined, "HDMI") {
		t.Fatalf("preview:\n%s", joined)
	}
}

func enabled(p Page, id string) bool {
	for _, c := range p.Commands() {
		if c.ID == id {
			return c.Enabled()
		}
	}
	return false
}

func nbView(p *NetBird) string { return ansi.Strip(p.View(90, 24)) }

func TestNetBirdStates(t *testing.T) {
	nets := []netbird.Network{
		{ID: "fr", Ranges: []string{"0.0.0.0/0", "::/0"}},
		{ID: "ch", Ranges: []string{"0.0.0.0/0", "::/0"}, Selected: true},
		{ID: "lab", Ranges: []string{"10.1.0.0/16"}, Selected: true},
	}
	peers := []netbird.Peer{{FQDN: "a.nb", IP: "100.1.1.1", Status: "Connected", Latency: 5 * time.Millisecond}}

	set := func(d nbData) *NetBird {
		p := NewNetBird()
		p.load, p.d = true, d
		p.rebuild()
		return p
	}

	p := set(nbData{state: netbird.State{Daemon: netbird.NeedsLogin}})
	if v := nbView(p); !strings.Contains(v, "not logged in") || !strings.Contains(v, "netbird up") {
		t.Fatalf("login view:\n%s", v)
	}
	if !enabled(p, "netbird.copy") || enabled(p, "netbird.connect") || enabled(p, "netbird.disconnect") || enabled(p, "netbird.select") {
		t.Fatal("login state commands")
	}
	if v := nbView(set(nbData{state: netbird.State{Daemon: netbird.SessionExpired}})); !strings.Contains(v, "expired") || !strings.Contains(v, "netbird up") {
		t.Fatalf("expired view:\n%s", v)
	}

	p = set(nbData{err: netbird.ErrNoDaemon})
	if v := nbView(p); !strings.Contains(v, "daemon is not running") || !strings.Contains(v, "systemctl start netbird") {
		t.Fatalf("no daemon view:\n%s", v)
	}
	if !enabled(p, "netbird.copy") || enabled(p, "netbird.connect") {
		t.Fatal("no daemon commands")
	}

	p = set(nbData{err: errors.New("boom")})
	if v := nbView(p); !strings.Contains(v, "boom") || enabled(p, "netbird.copy") {
		t.Fatalf("other error:\n%s", v)
	}

	p = set(nbData{state: netbird.State{Daemon: netbird.Idle}})
	if v := nbView(p); !strings.Contains(v, "press c to connect") {
		t.Fatalf("idle view:\n%s", v)
	}
	if !enabled(p, "netbird.connect") || enabled(p, "netbird.disconnect") || enabled(p, "netbird.copy") {
		t.Fatal("idle commands")
	}

	p = set(nbData{state: netbird.State{Daemon: netbird.Connected, IP: "100.126.115.18/16", Peers: peers}, nets: nets})
	v := nbView(p)
	for _, want := range []string{"Exit nodes", "Networks", "Peers", "ch", "active", "lab", "a", "100.126.115.18"} {
		if !strings.Contains(v, want) {
			t.Fatalf("connected view misses %q:\n%s", want, v)
		}
	}
	if !enabled(p, "netbird.disconnect") || enabled(p, "netbird.connect") || !enabled(p, "netbird.select") {
		t.Fatal("connected commands")
	}
	if p.toggle() == nil {
		t.Fatal("toggle did nothing")
	}
	if p.toggle() != nil {
		t.Fatal("a second toggle ran while the first is pending")
	}
	p.Update(DoneMsg{Page: "netbird"})
	if p.toggle() == nil {
		t.Fatal("blocked after the first finished")
	}
	p.list.Move(100)
	if it, _ := p.list.Selected(); it.Key != "peer:a.nb" || enabled(p, "netbird.select") {
		t.Fatalf("a peer is not selectable as a network: %+v", it)
	}
	if len(p.detailLines()) == 0 {
		t.Fatal("no peer details")
	}
}

func TestCompactAndHiddenHint(t *testing.T) {
	var l List
	items := []Item{{Header: true, Label: "A"}}
	for i := 0; i < 6; i++ {
		items = append(items, Item{Key: "a" + itoa(i), Label: "a" + itoa(i)})
	}
	items = append(items, Item{Header: true, Label: "B"}, Item{Key: "b", Label: "b"})
	l.Set(items)
	if len(l.items) != 12 {
		t.Fatalf("spaced items = %d", len(l.items))
	}
	l.Move(4)
	sel, _ := l.Selected()

	small := ansi.Strip(l.Render(40, 8, Frame{Title: "T"}))
	if len(l.items) != 9 {
		t.Fatalf("a short list keeps its blank lines: %d items", len(l.items))
	}
	if it, _ := l.Selected(); it.Key != sel.Key {
		t.Fatalf("selection moved by the relayout: %q -> %q", sel.Key, it.Key)
	}
	if !strings.Contains(small, "↓") {
		t.Fatalf("no hint that rows are hidden below:\n%s", small)
	}
	l.Move(100)
	if v := ansi.Strip(l.Render(40, 8, Frame{Title: "T"})); !strings.Contains(v, "↑") || strings.Contains(v, "↓") {
		t.Fatalf("hint at the end of the list:\n%s", v)
	}

	tall := ansi.Strip(l.Render(40, 30, Frame{Title: "T"}))
	if len(l.items) != 12 || strings.Contains(tall, "↑") || strings.Contains(tall, "↓") {
		t.Fatalf("a tall list: %d items\n%s", len(l.items), tall)
	}
}

func TestPreviewKeepsTouchingDisplaysTouching(t *testing.T) {
	mk := func(name string, x, y, w, h int) monitor.Output {
		return monitor.Output{Name: name, Active: true, Rect: monitor.Rect{X: x, Y: y, Width: w, Height: h}}
	}
	// the laptop at several scales, the monitor always against its right edge
	for _, w := range []int{2560, 2133, 1920, 1707, 1463, 1280, 1024} {
		h := w * 1440 / 2560
		outs := []monitor.Output{mk("eDP-1", 0, 0, w, h), mk("DP-9", w, 0, 3440, 1440)}
		for _, size := range [][2]int{{36, 8}, {30, 8}, {60, 12}, {24, 6}} {
			lines := layoutPreview(outs, "DP-9", size[0], size[1])
			joined := ansi.Strip(strings.Join(lines, "\n"))
			if !strings.Contains(joined, "╮╭") {
				t.Fatalf("laptop %d wide, canvas %v: the displays do not touch:\n%s", w, size, joined)
			}
			if size[0] >= 30 && (!strings.Contains(joined, "eDP-1") || !strings.Contains(joined, "DP-9")) {
				t.Fatalf("laptop %d wide, canvas %v: a display is cut:\n%s", w, size, joined)
			}
		}
	}
	// stacked
	outs := []monitor.Output{mk("eDP-1", 0, 0, 2560, 1440), mk("DP-9", 0, 1440, 2560, 1440)}
	if j := ansi.Strip(strings.Join(layoutPreview(outs, "DP-9", 36, 10), "\n")); !strings.Contains(j, "╯\n│") && !strings.Contains(j, "╰") {
		t.Fatalf("stacked:\n%s", j)
	}
}

func TestConfigBoxShowsItsKeys(t *testing.T) {
	b := &configBox{outs: []monitor.Output{{Name: "eDP-1", Active: true, Scale: 1, Transform: "normal"}}}
	v := ansi.Strip(b.View(100, 20))
	for _, want := range []string{"y copy", "tab sway / kanshi", "esc close"} {
		if !strings.Contains(v, want) {
			t.Fatalf("the popup does not say %q:\n%s", want, v)
		}
	}
	KeyLabel = func(id string) string { return map[string]string{"config.copy": "ctrl+y"}[id] }
	defer func() { KeyLabel = nil }()
	if v := ansi.Strip(b.View(100, 20)); !strings.Contains(v, "ctrl+y copy") || !strings.Contains(v, "tab sway / kanshi") {
		t.Fatalf("remapped keys not shown:\n%s", v)
	}
}

func TestWorkingIndicator(t *testing.T) {
	p := NewPower()
	p.d.profiles = []power.Profile{{Name: "balanced", Active: true}, {Name: "performance"}}
	p.load = true
	p.rebuild()
	if strings.Contains(ansi.Strip(p.View(100, 20)), "working") {
		t.Fatal("working while nothing runs")
	}
	p.use("performance")
	if !strings.Contains(ansi.Strip(p.View(100, 20)), "working…") {
		t.Fatal("no sign that the action is running")
	}
	p.Update(DoneMsg{Page: "power"})
	if strings.Contains(ansi.Strip(p.View(100, 20)), "working") {
		t.Fatal("still working after the action finished")
	}
}
