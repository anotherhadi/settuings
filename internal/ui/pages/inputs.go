package pages

import (
	"math"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/prompt"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/input"
)

type inputsMsg struct {
	err error
	st  input.State
}

type Inputs struct {
	list List
	st   input.State
	err  error
	busy bool
	load bool
	step float64
}

func NewInputs(step float64) *Inputs { return &Inputs{step: step} }

func (p *Inputs) Available() bool { return input.Available() == nil }
func (p *Inputs) Name() string    { return "inputs" }
func (p *Inputs) Label() string   { return "Inputs" }
func (p *Inputs) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Inputs) Leave()          {}
func (p *Inputs) Close()          {}

func (p *Inputs) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		if err := input.Available(); err != nil {
			return inputsMsg{err: err}
		}
		st, err := input.Read()
		return inputsMsg{st: st, err: err}
	}
}

func (p *Inputs) Update(msg tea.Msg) tea.Cmd {
	m, ok := msg.(inputsMsg)
	if !ok {
		return nil
	}
	p.busy, p.load = false, true
	p.st, p.err = m.st, m.err
	p.rebuild()
	return nil
}

func signed(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	if v > 0 {
		s = "+" + s
	}
	return s
}

func (p *Inputs) rebuild() {
	var items []Item
	if p.st.Mouse != nil || p.st.Touchpad != nil {
		items = append(items, Item{Header: true, Label: "Pointer"})
	}
	if m := p.st.Mouse; m != nil {
		items = append(items, Item{Key: "mouse", Icon: icon("\U000f037d", ""), Label: "Mouse speed", Note: theme.Subtle().Render(signed(m.Speed))})
	}
	if t := p.st.Touchpad; t != nil {
		items = append(items, Item{Key: "touchpad", Icon: icon("\U000f07f8", ""), Label: "Touchpad speed", Note: theme.Subtle().Render(signed(t.Speed))})
	}
	if k := p.st.Keyboard; k != nil {
		items = append(items, Item{Header: true, Label: "Keyboard"})
		items = append(items, Item{Key: "layout", Icon: icon("\uf11c", ""), Label: "Layout", Note: theme.Subtle().Render(k.Layout())})
	}
	p.list.Set(items)
}

func (p *Inputs) selected() string {
	it, ok := p.list.Selected()
	if !ok {
		return ""
	}
	return it.Key
}

func (p *Inputs) speed(sign float64) tea.Cmd {
	var cur *input.Pointer
	var set func(float64) error
	switch p.selected() {
	case "mouse":
		cur, set = p.st.Mouse, input.SetMouseSpeed
	case "touchpad":
		cur, set = p.st.Touchpad, input.SetTouchpadSpeed
	default:
		return nil
	}
	v := min(max(math.Round((cur.Speed+sign*p.step)*100)/100, -1), 1)
	if v == cur.Speed {
		return nil
	}
	cur.Speed = v
	p.rebuild()
	return act("inputs", "", func() error { return set(v) })
}

func (p *Inputs) askLayouts() tea.Cmd {
	pr := prompt.New("Keyboard layouts", "", func(s string) tea.Cmd {
		return act("inputs", "Layouts: "+s, func() error { return input.SetLayouts(s) })
	})
	pr.Hint = "xkb names separated by commas, e.g. fr,us"
	pr.Validate = input.ValidLayouts
	return overlay.Open(pr)
}

func (p *Inputs) Commands() []commands.Command {
	pointer := func() bool { k := p.selected(); return k == "mouse" || k == "touchpad" }
	layout := func() bool { return p.selected() == "layout" }
	cmds := []commands.Command{
		{ID: "inputs.faster", Title: "Pointer speed up", Help: "speed", When: pointer, Run: func() tea.Cmd { return p.speed(1) }},
		{ID: "inputs.slower", Title: "Pointer speed down", When: pointer, Run: func() tea.Cmd { return p.speed(-1) }},
		{ID: "inputs.layouts", Title: "Set the keyboard layouts", Help: "layouts", When: layout, Run: p.askLayouts},
		{ID: "inputs.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Inputs", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func devices(names []string) []string {
	lines := []string{blank(), theme.Subtle().Render("Applies to")}
	for _, n := range names {
		lines = append(lines, "  "+n)
	}
	return lines
}

func (p *Inputs) detailLines() []string {
	var pt *input.Pointer
	switch p.selected() {
	case "mouse":
		pt = p.st.Mouse
	case "touchpad":
		pt = p.st.Touchpad
	case "layout":
		k := p.st.Keyboard
		lines := []string{kv("Layout", k.Layout())}
		if len(k.Layouts) > 1 {
			lines = append(lines, kv("Layouts", strings.Join(k.Layouts, ", ")))
		}
		return append(lines, devices(k.Devices)...)
	default:
		return nil
	}
	lines := []string{kv("Speed", level(int((pt.Speed+1)*50), 16)+theme.Subtle().Render("  "+signed(pt.Speed)))}
	if pt.Profile != "" {
		lines = append(lines, kv("Profile", pt.Profile))
	}
	return append(lines, devices(pt.Devices)...)
}

func (p *Inputs) View(w, h int) string {
	f := Frame{Title: icon("\uf11c ", "") + "Inputs", DetailTitle: "Details", Empty: "no input device"}
	switch {
	case p.err != nil:
		f.Empty = p.err.Error()
	case !p.load:
		f.Empty = "loading"
	default:
		f.Right = "until sway reloads"
		f.Detail = p.detailLines()
	}
	return p.list.Render(w, h, f)
}

func (p *Inputs) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) && p.selected() == "layout" {
		return p.askLayouts()
	}
	return nil
}
