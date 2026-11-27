package ui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// optIn keeps the long hunts (random keys, every size, every popup) out of the
// usual run: SETTUINGS_MONKEY=1 starts them.
func optIn(t *testing.T) {
	t.Helper()
	if os.Getenv("SETTUINGS_MONKEY") == "" {
		t.Skip("long hunt: SETTUINGS_MONKEY=1 to run it")
	}
}

// isolate keeps the tests away from the real machine: the tools are looked up
// in SETTUINGS_FAKES (see the fakes used for development), or nowhere.
func isolate(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("SETTUINGS_FAKES")
	if dir == "" {
		dir = t.TempDir()
		t.Setenv("SWAYSOCK", "")
	}
	t.Setenv("PATH", dir)
	log := filepath.Join(t.TempDir(), "writes.log")
	t.Setenv("FAKE_LOG", log)
	return log
}

var sgr = regexp.MustCompile(`\x1b\[[0-9;:]*m`)

// checkClean fails when the screen holds anything but text and colors: a name
// from a device must not be able to drive the terminal.
func checkClean(t *testing.T, content, state string) {
	t.Helper()
	left := sgr.ReplaceAllString(content, "")
	for i, r := range left {
		if r == '\n' {
			continue
		}
		if r < 0x20 || r == 0x7f || r >= 0x80 && r <= 0x9f {
			t.Fatalf("%s: control character %U at %d: %q", state, r, i, left[max(i-20, 0):min(i+20, len(left))])
		}
	}
	if strings.Contains(left, "PWNED") && strings.Contains(content, "\x1b]") {
		t.Fatalf("%s: an escape sequence from the data reached the screen", state)
	}
}

type monkey struct {
	t       *testing.T
	m       *Model
	rng     *rand.Rand
	history []string
	queue   chan tea.Msg
}

func (k *monkey) note(format string, a ...any) {
	k.history = append(k.history, fmt.Sprintf(format, a...))
	if len(k.history) > 40 {
		k.history = k.history[1:]
	}
}

var cmdType = reflect.TypeOf((tea.Cmd)(nil))

// exec runs a command the way the program does: on its own, its message coming
// back through the queue.
func (k *monkey) exec(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	go func() { k.expand(cmd()) }()
}

func (k *monkey) expand(msg tea.Msg) {
	if msg == nil {
		return
	}
	if v := reflect.ValueOf(msg); v.Kind() == reflect.Slice && v.Type().Elem() == cmdType {
		sequence := v.Type().Name() == "sequenceMsg"
		for i := 0; i < v.Len(); i++ {
			c, _ := v.Index(i).Interface().(tea.Cmd)
			switch {
			case c == nil:
			case sequence:
				k.expand(c())
			default:
				k.exec(c)
			}
		}
		return
	}
	if _, quit := msg.(tea.QuitMsg); quit {
		return
	}
	k.queue <- msg
}

func (k *monkey) feed(msg tea.Msg) { k.exec(k.m.update(msg)) }

// pump delivers what the commands answered since the last time.
func (k *monkey) pump() {
	for {
		select {
		case msg := <-k.queue:
			k.feed(msg)
		default:
			return
		}
	}
}

func (k *monkey) settle(d time.Duration) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		time.Sleep(5 * time.Millisecond)
		k.pump()
	}
}

const printable = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 ?:/[]+-,.;'\"\\|!@#$%^&*()_=<>~`é日☕"

var specialKeys = []tea.Key{
	{Code: tea.KeyEnter}, {Code: tea.KeyEnter}, {Code: tea.KeyEscape}, {Code: tea.KeyTab}, {Code: tea.KeyTab, Mod: tea.ModShift},
	{Code: tea.KeyUp}, {Code: tea.KeyDown}, {Code: tea.KeyLeft}, {Code: tea.KeyRight}, {Code: tea.KeyPgUp}, {Code: tea.KeyPgDown},
	{Code: tea.KeyHome}, {Code: tea.KeyEnd}, {Code: tea.KeyBackspace}, {Code: tea.KeyDelete}, {Code: tea.KeySpace, Text: " "},
}

const useful = "jjjkkhlyndxctpawsoermi?:fQ+-"

func (k *monkey) key() tea.Msg {
	switch k.rng.Intn(10) {
	case 0, 1, 2, 3:
		r := rune(useful[k.rng.Intn(len(useful))])
		return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)})
	case 4, 5:
		return tea.KeyPressMsg(specialKeys[k.rng.Intn(len(specialKeys))])
	case 6:
		r := rune("abcdefghijklmnopqrstuvwxyz"[k.rng.Intn(26)])
		mod := []tea.KeyMod{tea.ModCtrl, tea.ModAlt}[k.rng.Intn(2)]
		return tea.KeyPressMsg(tea.Key{Code: r, Mod: mod})
	}
	rs := []rune(printable)
	r := rs[k.rng.Intn(len(rs))]
	return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)})
}

func (k *monkey) mouse() tea.Msg {
	w, h := max(k.m.w, 1), max(k.m.h, 1)
	mm := tea.Mouse{X: k.rng.Intn(w+6) - 3, Y: k.rng.Intn(h+6) - 3}
	switch k.rng.Intn(8) {
	case 0, 1, 2:
		mm.Button = tea.MouseLeft
		return tea.MouseClickMsg(mm)
	case 3:
		mm.Button = tea.MouseRight
		return tea.MouseClickMsg(mm)
	case 4, 5:
		mm.Button = []tea.MouseButton{tea.MouseWheelUp, tea.MouseWheelDown, tea.MouseWheelLeft, tea.MouseWheelRight}[k.rng.Intn(4)]
		return tea.MouseWheelMsg(mm)
	case 6:
		return tea.MouseMotionMsg(mm)
	}
	return tea.MouseReleaseMsg(mm)
}

func (k *monkey) size() tea.Msg {
	switch k.rng.Intn(5) {
	case 0:
		return tea.WindowSizeMsg{Width: 1 + k.rng.Intn(40), Height: 1 + k.rng.Intn(14)}
	case 1:
		return tea.WindowSizeMsg{Width: 1 + k.rng.Intn(250), Height: 1 + k.rng.Intn(80)}
	}
	return tea.WindowSizeMsg{Width: 30 + k.rng.Intn(120), Height: 8 + k.rng.Intn(40)}
}

var pastes = []string{"hello", "line1\nline2", "tab\there", "\x1b[2Jboom", "\x1b]0;PWNED\x07", strings.Repeat("long ", 80), "日本語", "", "fr,us", "-1"}

func (k *monkey) step() {
	switch n := k.rng.Intn(100); {
	case n < 66:
		msg := k.key()
		k.note("key %v", msg)
		k.feed(msg)
	case n < 80:
		msg := k.mouse()
		k.note("mouse %T %+v", msg, msg)
		k.feed(msg)
	case n < 90:
		msg := k.size()
		k.note("size %+v", msg)
		k.feed(msg)
	case n < 94:
		p := pastes[k.rng.Intn(len(pastes))]
		k.note("paste %q", p)
		k.feed(tea.PasteMsg{Content: p})
	default:
		k.note("tick")
		k.feed(tickMsg{})
	}
}

func (k *monkey) check(step int) {
	k.t.Helper()
	m := k.m
	if m.w == 0 || m.h == 0 {
		return
	}
	content := m.View().Content
	state := fmt.Sprintf("step %d, page %s, overlays %d", step, m.page().Name(), m.overlays.Len())
	lines := strings.Split(content, "\n")
	if len(lines) != m.h {
		k.t.Fatalf("%s at %dx%d: %d lines", state, m.w, m.h, len(lines))
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got > m.w || got < m.w && strings.TrimSpace(ansi.Strip(l)) != "" {
			k.t.Fatalf("%s at %dx%d: line %d is %d wide: %q", state, m.w, m.h, i, got, ansi.Strip(l))
		}
	}
	checkClean(k.t, content, state)
}

func (k *monkey) loadPages() {
	for _, p := range k.m.pages {
		k.exec(p.Refresh(true))
	}
	k.settle(600 * time.Millisecond)
}

func monkeyRun(t *testing.T, seed int64, steps int) {
	m := newModel(t)
	log := os.Getenv("FAKE_LOG")
	k := &monkey{t: t, m: m, rng: rand.New(rand.NewSource(seed)), queue: make(chan tea.Msg, 1<<16)}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("seed %d: panic: %v\n%s\nlast actions:\n  %s", seed, r, debug.Stack(), strings.Join(k.history, "\n  "))
		}
	}()
	k.feed(tea.WindowSizeMsg{Width: 120, Height: 36})
	k.loadPages()
	for i := 0; i < steps; i++ {
		k.step()
		k.pump()
		k.check(i)
		if i%25 == 0 {
			time.Sleep(20 * time.Millisecond)
		}
	}
	k.settle(500 * time.Millisecond)
	k.check(steps)
	checkWrites(t, log)
}

func TestMonkey(t *testing.T) {
	optIn(t)
	seeds, steps := envInt("SETTUINGS_SEEDS", 4), envInt("SETTUINGS_STEPS", 300)
	for seed := int64(1); seed <= int64(seeds); seed++ {
		t.Run(fmt.Sprint("seed", seed), func(t *testing.T) { monkeyRun(t, seed, steps) })
	}
}

// ----------------------------------------------------------------- the writes

var writeShapes = map[string][]*regexp.Regexp{
	"nmcli": rx(
		`^device wifi rescan$`, `^radio wifi (on|off)$`, `^device wifi connect .+$`, `^device (connect|disconnect) .+$`,
		`^connection (up|down|delete) id .+$`,
		`^connection modify id .+ connection\.autoconnect (yes|no)$`,
		`^connection modify id .+ wifi-sec\.psk-flags 0 wifi-sec\.psk .+$`),
	"bluetoothctl":     rx(`^power (on|off)$`, `^(pair|connect|disconnect|trust|untrust|remove) [0-9A-F]{2}(:[0-9A-F]{2}){5}$`, `^--timeout \d+ scan on$`),
	"wpctl":            rx(`^set-volume \d+ (\d|[1-9]\d|100)%$`, `^set-mute \d+ (0|1|toggle)$`, `^set-default \d+$`),
	"powerprofilesctl": rx(`^set (performance|balanced|power-saver)$`),
	"brightnessctl":    rx(`^-d \S+ set (\d|[1-9]\d|100)%$`),
	"netbird":          rx(`^up --no-browser$`, `^down$`, `^networks select( -a)?( [\w.-]+)+$`, `^networks deselect( [\w.-]+)+$`),
	"pw-play":          rx(`^--target \S+ -$`),
	"swaymsg": rx(
		`^-r -- input type:(pointer|touchpad) pointer_accel -?\d\.\d\d$`,
		`^-r -- input type:keyboard xkb_layout "[A-Za-z0-9_,()-]+"$`,
		`^-r -- output [A-Za-z0-9._:-]+ (enable|disable)$`,
		`^-r -- output [A-Za-z0-9._:-]+ mode \d+x\d+@\d+\.\d{3}Hz$`,
		`^-r -- output [A-Za-z0-9._:-]+ scale \d\.\d\d$`,
		`^-r -- output [A-Za-z0-9._:-]+ transform (normal|90|180|270)$`,
		`^-r -- output [A-Za-z0-9._:-]+ pos -?\d+ -?\d+$`,
		`^-r -- output [A-Za-z0-9._:-]+ adaptive_sync (on|off)$`,
		`^-r -- output [A-Za-z0-9._:-]+ enable mode \d+x\d+@\d+\.\d{3}Hz pos -?\d+ -?\d+ scale \d\.\d\d transform (normal|90|180|270) adaptive_sync (on|off)$`),
}

func rx(patterns ...string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, p := range patterns {
		out = append(out, regexp.MustCompile(p))
	}
	return out
}

var number = regexp.MustCompile(`pointer_accel (-?\d\.\d\d)$|scale (\d\.\d\d)$`)

func checkWrites(t *testing.T, path string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if keep := os.Getenv("SETTUINGS_WRITES"); keep != "" {
		if f, err := os.OpenFile(keep, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			f.Write(b)
			f.Close()
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		tool, args, _ := strings.Cut(line, " ")
		ok := false
		for _, re := range writeShapes[tool] {
			if re.MatchString(args) {
				ok = true
			}
		}
		if !ok {
			t.Errorf("unexpected command: %s", line)
			continue
		}
		if m := number.FindStringSubmatch(args); m != nil {
			v, _ := strconv.ParseFloat(m[1]+m[2], 64)
			switch {
			case m[1] != "" && (v < -1 || v > 1):
				t.Errorf("pointer speed out of range: %s", line)
			case m[2] != "" && (v < 0.5 || v > 4):
				t.Errorf("scale out of range: %s", line)
			}
		}
	}
}

func envInt(name string, def int) int {
	if n, err := strconv.Atoi(os.Getenv(name)); err == nil && n > 0 {
		return n
	}
	return def
}
