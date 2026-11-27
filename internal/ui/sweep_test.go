package ui

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
)

func newMonkey(t *testing.T) *monkey {
	t.Helper()
	m := newModel(t)
	k := &monkey{t: t, m: m, rng: rand.New(rand.NewSource(1)), queue: make(chan tea.Msg, 1<<16)}
	k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
	k.loadPages()
	return k
}

func fakes() bool { return os.Getenv("SETTUINGS_FAKES") != "" }

// sizes are the edges of the layouts by default (32x10 is the least the app
// draws, the details sit beside the list from 90 columns, a short list loses
// its blank lines under 16 rows...). SETTUINGS_SWEEP=dense or full tries more.
func sizes() (ws, hs []int) {
	ws = []int{1, 2, 3, 5, 8, 13, 20, 31, 32, 33, 40, 45, 60, 80, 89, 90, 91, 100, 120, 150}
	hs = []int{1, 2, 3, 5, 8, 9, 10, 11, 14, 18, 21, 22, 23, 30, 40, 55}
	switch os.Getenv("SETTUINGS_SWEEP") {
	case "dense":
		ws, hs = nil, nil
		for w := 1; w <= 150; w++ {
			if w <= 45 || w%7 == 0 {
				ws = append(ws, w)
			}
		}
		for h := 1; h <= 55; h++ {
			if h <= 30 || h%5 == 0 {
				hs = append(hs, h)
			}
		}
	case "full":
		ws, hs = nil, nil
		for w := 1; w <= 200; w++ {
			ws = append(ws, w)
		}
		for h := 1; h <= 70; h++ {
			hs = append(hs, h)
		}
	}
	return ws, hs
}

// every page, every size
func TestSweepSizes(t *testing.T) {
	optIn(t)
	k := newMonkey(t)
	ws, hs := sizes()
	for i, p := range k.m.pages {
		k.m.cur = i
		for _, w := range ws {
			for _, h := range hs {
				k.feed(tea.WindowSizeMsg{Width: w, Height: h})
				k.check(w*1000 + h)
			}
		}
		t.Logf("%s: %d sizes", p.Name(), len(ws)*len(hs))
	}
}

// the popups of the app, at every size
func TestSweepPopups(t *testing.T) {
	optIn(t)
	k := newMonkey(t)
	ws, hs := sizes()
	for _, open := range []string{"?", ":"} {
		for i := range k.m.pages {
			k.m.cur = i
			k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
			k.feed(tea.KeyPressMsg(tea.Key{Code: rune(open[0]), Text: open}))
			if k.m.overlays.Len() == 0 {
				k.settle(10 * time.Millisecond)
			}
			if k.m.overlays.Len() != 1 {
				t.Fatalf("%q on %s opened %d popups", open, k.m.pages[i].Name(), k.m.overlays.Len())
			}
			for _, w := range ws {
				for _, h := range hs {
					k.feed(tea.WindowSizeMsg{Width: w, Height: h})
					k.check(w*1000 + h)
				}
			}
			k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
			k.feed(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
			k.settle(15 * time.Millisecond)
			if k.m.overlays.Len() != 0 {
				t.Fatalf("esc did not close the %q popup on %s", open, k.m.pages[i].Name())
			}
		}
	}
}

// the popup of every command that opens one, at every size
func TestSweepEveryPopup(t *testing.T) {
	optIn(t)
	k := newMonkey(t)
	ws, hs := sizes()
	rows := envInt("SETTUINGS_ROWS", 14)
	opened := map[string]bool{}
	for i, p := range k.m.pages {
		k.m.cur = i
		k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
		k.feed(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
		for row := 0; row < rows; row++ {
			for _, c := range p.Commands() {
				key := p.Name() + " " + c.ID
				if !c.Enabled() || opened[key] || strings.HasPrefix(c.ID, "list.") {
					continue
				}
				k.exec(c.Run())
				k.settle(25 * time.Millisecond)
				if k.m.overlays.Len() == 0 {
					continue
				}
				opened[key] = true
				for _, w := range ws {
					for _, h := range hs {
						k.feed(tea.WindowSizeMsg{Width: w, Height: h})
						k.check(w*1000 + h)
					}
				}
				k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
				k.closePopups()
			}
			k.feed(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
		}
	}
	t.Logf("%d popups swept at %d sizes each", len(opened), len(ws)*len(hs))
	if fakes() && os.Getenv("FAKE_EMPTY") == "" && os.Getenv("FAKE_NB") == "" && os.Getenv("FAKE_SWAY") == "" {
		for _, want := range []string{"network network.connect", "network network.forget", "network network.qrcode", "bluetooth bluetooth.forget",
			"monitors monitors.mode", "monitors monitors.rotate", "monitors monitors.place", "monitors monitors.config",
			"inputs inputs.layouts"} {
			if !opened[want] {
				t.Errorf("%s never opened a popup", want)
			}
		}
	}
}

func (k *monkey) closePopups() {
	for i := 0; i < 6 && k.m.overlays.Len() > 0; i++ {
		k.feed(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
		k.settle(3 * time.Millisecond)
	}
	if k.m.overlays.Len() > 0 {
		k.t.Fatalf("%d popups stay open after esc", k.m.overlays.Len())
	}
}

// every command of every page, on every row, then every popup it opens at several sizes
func TestEveryCommand(t *testing.T) {
	optIn(t)
	k := newMonkey(t)
	rows := envInt("SETTUINGS_ROWS", 14)
	shown := [][2]int{{130, 40}, {80, 24}, {60, 18}, {45, 14}, {34, 10}}
	ran := map[string]int{}
	for i, p := range k.m.pages {
		k.m.cur = i
		k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
		k.feed(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
		for row := 0; row < rows; row++ {
			for _, c := range p.Commands() {
				if !c.Enabled() || strings.HasPrefix(c.ID, "list.") {
					continue
				}
				ran[p.Name()+" "+c.ID]++
				k.exec(c.Run())
				k.settle(25 * time.Millisecond)
				for _, s := range shown {
					k.feed(tea.WindowSizeMsg{Width: s[0], Height: s[1]})
					k.check(row)
				}
				k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
				k.closePopups()
			}
			k.feed(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
		}
	}
	for _, c := range k.m.global() {
		if c.ID == "app.quit" || !c.Enabled() {
			continue
		}
		ran["app "+c.ID]++
		k.exec(c.Run())
		k.settle(25 * time.Millisecond)
		for _, s := range shown {
			k.feed(tea.WindowSizeMsg{Width: s[0], Height: s[1]})
			k.check(0)
		}
		k.feed(tea.WindowSizeMsg{Width: 130, Height: 40})
		k.closePopups()
	}
	if fakes() && os.Getenv("FAKE_EMPTY") == "" && os.Getenv("FAKE_NB") == "" {
		for _, want := range []string{
			"network network.connect", "network network.disconnect", "network network.forget", "network network.qrcode", "network network.secret",
			"bluetooth bluetooth.connect", "bluetooth bluetooth.forget", "bluetooth bluetooth.trust",
			"audio audio.mute", "audio audio.test", "audio audio.default", "power power.set",
			"inputs inputs.faster", "inputs inputs.layouts",
			"monitors monitors.mode", "monitors monitors.rotate", "monitors monitors.place", "monitors monitors.config", "monitors monitors.power",
			"netbird netbird.select", "netbird netbird.disconnect",
		} {
			if ran[want] == 0 {
				t.Errorf("%s never ran: the page shows no row it applies to", want)
			}
		}
	}
	checkWrites(t, os.Getenv("FAKE_LOG"))
}

// with the fake tools, the data has to be on screen, and nothing hostile with it
func TestRichData(t *testing.T) {
	if !fakes() || os.Getenv("FAKE_EMPTY") != "" || os.Getenv("FAKE_NB") != "" || os.Getenv("FAKE_SWAY") != "" || os.Getenv("FAKE_FAIL") != "" {
		t.Skip("needs the default fake tools")
	}
	k := newMonkey(t)
	k.feed(tea.WindowSizeMsg{Width: 140, Height: 110})
	want := map[string][]string{
		"about":     {"NixOS", "Packages"},
		"network":   {"Djoon", "enp7s0f4u1u4", "wg-vpn", "Known networks", "Free open wifi"},
		"bluetooth": {"WH-1000XM4", "Razer Mouse", "Paired", "Nearby"},
		"audio":     {"Ryzen HD Audio Controller Speaker", "Firefox", "Outputs", "Inputs", "Applications"},
		"inputs":    {"Mouse speed", "Touchpad speed", "Layout"},
		"monitors":  {"DP-9", "eDP-1", "amdgpu_bl1", "Brightness"},
		"netbird":   {"Exit nodes", "whiterabbit", "Peers"},
	}
	for i, p := range k.m.pages {
		k.m.cur = i
		screen := ansi.Strip(k.m.View().Content)
		for _, w := range want[p.Name()] {
			if !strings.Contains(screen, w) {
				t.Errorf("%s: %q is not on screen:\n%s", p.Name(), w, screen)
			}
		}
		checkClean(t, k.m.View().Content, p.Name())
	}
}

// the details popup, which only exists when the details do not fit beside the list
func TestInfoPopups(t *testing.T) {
	optIn(t)
	k := newMonkey(t)
	ws, hs := sizes()
	for i, p := range k.m.pages {
		k.m.cur = i
		k.feed(tea.WindowSizeMsg{Width: 60, Height: 20})
		k.feed(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
		k.feed(tea.KeyPressMsg(tea.Key{Code: 'j', Text: "j"}))
		if !enabled(p, p.Name()+".info") {
			if fakes() && os.Getenv("FAKE_EMPTY") == "" && os.Getenv("FAKE_NB") == "" && p.Name() != "about" && p.Name() != "power" {
				t.Errorf("%s: no details popup at 60x20", p.Name())
			}
			continue
		}
		k.feed(tea.KeyPressMsg(tea.Key{Code: 'i', Text: "i"}))
		k.settle(15 * time.Millisecond)
		if k.m.overlays.Len() != 1 {
			t.Fatalf("%s: i opened %d popups", p.Name(), k.m.overlays.Len())
		}
		for _, w := range ws {
			for _, h := range hs {
				k.feed(tea.WindowSizeMsg{Width: w, Height: h})
				k.check(w*1000 + h)
			}
		}
		k.feed(tea.WindowSizeMsg{Width: 60, Height: 20})
		k.closePopups()
	}
}

func enabled(p interface{ Commands() []commands.Command }, id string) bool {
	for _, c := range p.Commands() {
		if c.ID == id {
			return c.Enabled()
		}
	}
	return false
}

func (k *monkey) writes() string {
	b, _ := os.ReadFile(os.Getenv("FAKE_LOG"))
	return string(b)
}

func (k *monkey) goTo(page string) {
	for i, p := range k.m.pages {
		if p.Name() == page {
			k.m.cur = i
		}
	}
	k.feed(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
}

func (k *monkey) press(keys ...string) {
	for _, key := range keys {
		switch key {
		case "enter":
			k.feed(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
		case "esc":
			k.feed(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
		default:
			k.feed(tea.KeyPressMsg(tea.Key{Code: rune(key[0]), Text: key}))
		}
		k.settle(10 * time.Millisecond)
	}
}

// turning a display off asks to keep it, and goes back by itself
func TestKeepAndRevert(t *testing.T) {
	optIn(t)
	if !fakes() || os.Getenv("FAKE_FAIL") != "" || os.Getenv("FAKE_EMPTY") != "" || os.Getenv("FAKE_SWAY") != "" {
		t.Skip("needs the default fake tools")
	}
	k := newMonkey(t)
	k.goTo("monitors")
	// DP-9 is the first display

	// no: the displays come back
	k.press("o")
	k.settle(500 * time.Millisecond)
	if k.m.overlays.Len() != 1 {
		t.Fatalf("no popup after turning a display off: %d", k.m.overlays.Len())
	}
	if w := k.writes(); !strings.Contains(w, "output DP-9 disable") {
		t.Fatalf("the display was not turned off:\n%s", w)
	}
	screen := ansi.Strip(k.m.View().Content)
	if !strings.Contains(screen, "Keep these settings?") || !strings.Contains(screen, "Reverting in") {
		t.Fatalf("the popup does not ask:\n%s", screen)
	}
	k.press("n")
	k.settle(500 * time.Millisecond)
	w := k.writes()
	if !strings.Contains(w, "output DP-9 enable mode 3440x1440@99.982Hz pos 2560 0") || !strings.Contains(w, "output eDP-1 enable mode 2560x1440@240.000Hz pos 0 0") {
		t.Fatalf("saying no did not restore the displays:\n%s", w)
	}
	if k.m.overlays.Len() != 0 {
		t.Fatal("the popup stays after the answer")
	}

	// keep: nothing more happens, even after the delay
	before := strings.Count(k.writes(), "\n")
	k.press("o")
	k.settle(500 * time.Millisecond)
	k.press("y")
	k.settle(200 * time.Millisecond)
	afterKeep := strings.Count(k.writes(), "\n")
	if afterKeep != before+1 {
		t.Fatalf("keeping sent %d commands, want 1", afterKeep-before)
	}
	k.settle(11 * time.Second)
	if got := strings.Count(k.writes(), "\n"); got != afterKeep {
		t.Fatalf("a kept change was reverted later (%d new commands)", got-afterKeep)
	}

	// silence: it reverts by itself, and closes its own popup
	k.press("o")
	k.settle(500 * time.Millisecond)
	if k.m.overlays.Len() != 1 {
		t.Fatalf("no popup: %d", k.m.overlays.Len())
	}
	mark := strings.Count(k.writes(), "\n")
	k.settle(11 * time.Second)
	if k.m.overlays.Len() != 0 {
		t.Fatal("the popup is still there after the delay")
	}
	if got := k.writes(); strings.Count(got, "\n") <= mark {
		t.Fatal("nothing was restored after the delay")
	}
	tail := strings.Split(strings.TrimSpace(k.writes()), "\n")
	tail = tail[len(tail)-3:]
	if joined := strings.Join(tail, "\n"); !strings.Contains(joined, "output DP-9 enable mode") || !strings.Contains(joined, "output eDP-1 enable mode") {
		t.Fatalf("the delay did not restore the displays:\n%s", joined)
	}
	checkWrites(t, os.Getenv("FAKE_LOG"))
}

// an empty password is refused by the popup, and nothing reaches nmcli
func TestEmptyPasswordIsRefused(t *testing.T) {
	if !fakes() || os.Getenv("FAKE_EMPTY") != "" || os.Getenv("FAKE_FAIL") != "" {
		t.Skip("needs the default fake tools")
	}
	k := newMonkey(t)
	k.goTo("network")
	found := false
	for row := 0; row < 14 && !found; row++ {
		k.press("enter")
		if k.m.overlays.Len() == 1 && strings.Contains(ansi.Strip(k.m.View().Content), "Password") {
			found = true
			break
		}
		k.closePopups()
		k.press("j")
	}
	if !found {
		t.Fatal("no password popup on any row")
	}
	before := k.writes()
	k.press("enter")
	if k.m.overlays.Len() != 1 {
		t.Fatal("an empty password closed the popup")
	}
	if screen := ansi.Strip(k.m.View().Content); !strings.Contains(screen, "password is needed") {
		t.Fatalf("no explanation:\n%s", screen)
	}
	if after := k.writes(); after != before {
		t.Fatalf("a command went out with an empty password:\n%s", strings.TrimPrefix(after, before))
	}
	k.press("a", "enter")
	k.settle(300 * time.Millisecond)
	if w := k.writes(); !strings.Contains(w, "wifi connect") || !strings.Contains(w, "password a") {
		t.Fatalf("a typed password was not used:\n%s", w)
	}
}

// placing a display, and pairing a device that is not paired yet
func TestPlaceAndPair(t *testing.T) {
	if !fakes() || os.Getenv("FAKE_FAIL") != "" || os.Getenv("FAKE_EMPTY") != "" || os.Getenv("FAKE_SWAY") != "" {
		t.Skip("needs the default fake tools")
	}
	k := newMonkey(t)

	k.goTo("monitors")
	k.press("p")
	if k.m.overlays.Len() != 1 {
		t.Fatal("no placement popup")
	}
	k.press("j", "enter") // second choice: left of eDP-1
	k.settle(300 * time.Millisecond)
	if w := k.writes(); !strings.Contains(w, "output DP-9 pos -3440 0") {
		t.Fatalf("DP-9 was not placed left of eDP-1:\n%s", w)
	}
	k.goTo("monitors")
	k.press("p", "enter") // first choice: right of eDP-1
	k.settle(300 * time.Millisecond)
	if w := k.writes(); !strings.Contains(w, "output DP-9 pos 2560 0") {
		t.Fatalf("DP-9 was not placed right of eDP-1:\n%s", w)
	}

	k.goTo("bluetooth")
	k.press("G", "enter") // the last of the nearby devices: not paired
	k.settle(500 * time.Millisecond)
	w := k.writes()
	if _, err := os.Stat("/run/dbus/system_bus_socket"); err != nil {
		// pairing registers an agent on the system bus first: without one it stops there, and says so
		if strings.Contains(w, "bluetoothctl pair") {
			t.Fatalf("paired without an agent:\n%s", w)
		}
		if screen := ansi.Strip(k.m.View().Content); !strings.Contains(screen, "system bus") {
			t.Fatalf("no explanation when the agent cannot register:\n%s", screen)
		}
	} else {
		pair, trust, connect := strings.Index(w, "bluetoothctl pair 11:22:33:44:55:69"), strings.Index(w, "bluetoothctl trust 11:22:33:44:55:69"), strings.Index(w, "bluetoothctl connect 11:22:33:44:55:69")
		if pair < 0 || trust < pair || connect < trust {
			t.Fatalf("pair, trust, connect did not go out in this order:\n%s", w)
		}
	}
	checkWrites(t, os.Getenv("FAKE_LOG"))
}
