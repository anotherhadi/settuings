package pages

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/confirm"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/brightness"
	"github.com/anotherhadi/settuings/internal/monitor"
)

const revertAfter = 10

type monData struct {
	err    error
	outs   []monitor.Output
	lights []brightness.Device
}

type monMsg monData

type appliedMsg struct {
	prev []monitor.Output
	desc string
	err  error
}

type keepMsg struct{ gen int }
type revertMsg struct{ gen int }
type keepTickMsg struct{ gen int }

type Monitors struct {
	pending
	list List
	d    monData
	busy bool
	load bool
	step int

	gen  int
	prev []monitor.Output
	desc string
	left int
	keep *confirm.Model
}

func NewMonitors(step int) *Monitors { return &Monitors{step: step} }

func (p *Monitors) Available() bool { return monitor.Available() == nil }
func (p *Monitors) Name() string    { return "monitors" }
func (p *Monitors) Label() string   { return "Monitors" }
func (p *Monitors) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Monitors) Leave()          {}
func (p *Monitors) Close()          {}

func (p *Monitors) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		var d monData
		if d.err = monitor.Available(); d.err != nil {
			return monMsg(d)
		}
		d.outs, d.err = monitor.List()
		if brightness.Installed() {
			d.lights, _ = brightness.List()
		}
		return monMsg(d)
	}
}

func (p *Monitors) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "monitors")
	switch m := msg.(type) {
	case monMsg:
		p.busy, p.load = false, true
		p.d = monData(m)
		p.rebuild()
	case appliedMsg:
		p.on = false
		if m.err != nil {
			return tea.Batch(notify.Err(m.err), p.Refresh(false))
		}
		return tea.Batch(p.askKeep(m), p.Refresh(false))
	case keepMsg:
		if m.gen == p.gen {
			p.gen++
			return notify.Success("Kept " + p.desc)
		}
	case revertMsg:
		if m.gen == p.gen {
			return p.revert()
		}
	case keepTickMsg:
		if m.gen != p.gen {
			return nil
		}
		p.left--
		if p.left > 0 {
			p.keep.Question = p.question()
			return p.tick()
		}
		box := p.keep
		return tea.Batch(func() tea.Msg { return CloseIfTop{O: box} }, p.revert())
	}
	return nil
}

func (p *Monitors) question() string {
	return fmt.Sprintf("%s. Reverting in %ds.", p.desc, p.left)
}

func (p *Monitors) tick() tea.Cmd {
	gen := p.gen
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return keepTickMsg{gen} })
}

func (p *Monitors) askKeep(m appliedMsg) tea.Cmd {
	p.gen++
	gen := p.gen
	p.prev, p.desc, p.left = m.prev, m.desc, revertAfter
	c := confirm.New("Keep these settings?", p.question(), func() tea.Msg { return keepMsg{gen} })
	c.Yes, c.No = "Keep", "Revert"
	c.OnNo = func() tea.Msg { return revertMsg{gen} }
	c.Focus(true)
	p.keep = c
	return tea.Batch(overlay.Open(c), p.tick())
}

func (p *Monitors) revert() tea.Cmd {
	p.gen++
	prev := p.prev
	return act("monitors", "Reverted to the previous settings", func() error {
		for _, o := range prev {
			if err := monitor.Restore(o); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Monitors) rebuild() {
	var items []Item
	if len(p.d.outs) > 0 {
		items = append(items, Item{Header: true, Label: "Displays"})
	}
	for _, o := range p.d.outs {
		note := theme.Subtle().Render(o.Mode.String())
		if !o.Active {
			note = theme.Subtle().Render("off")
		}
		items = append(items, Item{Key: "out:" + o.Name, Icon: icon("\U000f0379", ""), Label: o.Name + "  " + o.Label(), Note: note, Dim: !o.Active, Data: o})
	}
	if len(p.d.lights) > 0 {
		items = append(items, Item{Header: true, Label: "Brightness"})
	}
	for _, l := range p.d.lights {
		items = append(items, Item{Key: "bl:" + l.Name, Icon: icon("\U000f00df", ""), Label: l.Name, Note: theme.Subtle().Render(itoa(l.Percent) + "%"), Data: l})
	}
	p.list.Set(items)
}

func (p *Monitors) selected() (any, bool) {
	it, ok := p.list.Selected()
	return it.Data, ok
}

func (p *Monitors) output() (monitor.Output, bool) {
	d, ok := p.selected()
	o, isOut := d.(monitor.Output)
	return o, ok && isOut
}

func (p *Monitors) light() (brightness.Device, bool) {
	d, ok := p.selected()
	l, isLight := d.(brightness.Device)
	return l, ok && isLight
}

func (p *Monitors) activeCount() int {
	n := 0
	for _, o := range p.d.outs {
		if o.Active {
			n++
		}
	}
	return n
}

func (p *Monitors) snapshot() []monitor.Output { return append([]monitor.Output(nil), p.d.outs...) }

func (p *Monitors) change(desc string, fn func() error) tea.Cmd {
	prev := p.snapshot()
	return p.do(func() tea.Msg { return appliedMsg{prev: prev, desc: desc, err: fn()} })
}

func (p *Monitors) resize(name string, apply func() error) func() error {
	before := p.snapshot()
	return func() error {
		if err := apply(); err != nil {
			return err
		}
		return monitor.Repack(name, before)
	}
}

func (p *Monitors) modes() tea.Cmd {
	o, ok := p.output()
	if !ok || !o.Active {
		return nil
	}
	var modes []monitor.Mode
	seen := map[string]bool{}
	for _, m := range o.Modes {
		if !seen[m.String()] {
			seen[m.String()] = true
			modes = append(modes, m)
		}
	}
	slices.SortFunc(modes, func(a, b monitor.Mode) int {
		if d := b.Width*b.Height - a.Width*a.Height; d != 0 {
			return d
		}
		return b.Refresh - a.Refresh
	})
	var opts []string
	sel := 0
	for i, m := range modes {
		label := m.String()
		if m.String() == o.Mode.String() {
			label += "  ●"
			sel = i
		}
		opts = append(opts, label)
	}
	return overlay.Open(newChoice("Mode · "+o.Name, opts, sel, func(i int) tea.Cmd {
		m := modes[i]
		return p.change("Mode "+m.String()+" on "+o.Name, p.resize(o.Name, func() error { return monitor.SetMode(o.Name, m) }))
	}))
}

func (p *Monitors) rotate() tea.Cmd {
	o, ok := p.output()
	if !ok || !o.Active {
		return nil
	}
	opts := []string{"normal", "90", "180", "270"}
	sel := slices.Index(opts, o.Transform)
	return overlay.Open(newChoice("Rotation · "+o.Name, opts, max(sel, 0), func(i int) tea.Cmd {
		return p.change("Rotation "+opts[i]+" on "+o.Name, p.resize(o.Name, func() error { return monitor.SetTransform(o.Name, opts[i]) }))
	}))
}

func (p *Monitors) place() tea.Cmd {
	o, ok := p.output()
	if !ok || !o.Active {
		return nil
	}
	type spot struct {
		ref  monitor.Output
		side monitor.Side
	}
	var spots []spot
	var opts []string
	for _, ref := range p.d.outs {
		if ref.Name == o.Name || !ref.Active {
			continue
		}
		for _, s := range []monitor.Side{monitor.RightOf, monitor.LeftOf, monitor.Above, monitor.Below} {
			spots = append(spots, spot{ref, s})
			opts = append(opts, s.String()+" "+ref.Name)
		}
	}
	return overlay.Open(newChoice("Place · "+o.Name, opts, 0, func(i int) tea.Cmd {
		x, y := monitor.Place(o, spots[i].ref, spots[i].side)
		return p.do(act("monitors", opts[i], func() error { return monitor.SetPosition(o.Name, x, y) }))
	}))
}

func (p *Monitors) power() tea.Cmd {
	o, ok := p.output()
	if !ok {
		return nil
	}
	if !o.Active {
		return p.do(act("monitors", o.Name+" on", func() error { return monitor.SetEnabled(o.Name, true) }))
	}
	if p.activeCount() < 2 {
		return notify.Warning("Can't turn off the only active display")
	}
	return p.change(o.Name+" turned off", func() error { return monitor.SetEnabled(o.Name, false) })
}

func (p *Monitors) adjust(sign int) tea.Cmd {
	if o, ok := p.output(); ok && o.Active {
		v := math.Round((o.Scale+float64(sign)*0.25)*100) / 100
		if v < 0.5 || v > 4 {
			return nil
		}
		return p.do(act("monitors", "", p.resize(o.Name, func() error { return monitor.SetScale(o.Name, v) })))
	}
	l, ok := p.light()
	if !ok {
		return nil
	}
	v := min(max(l.Percent+sign*p.step, 1), 100)
	if v == l.Percent {
		return nil
	}
	for i := range p.d.lights {
		if p.d.lights[i].Name == l.Name {
			p.d.lights[i].Percent = v
		}
	}
	p.rebuild()
	return act("monitors", "", func() error { return brightness.Set(l.Name, v) })
}

func (p *Monitors) Commands() []commands.Command {
	active := func() bool { o, ok := p.output(); return ok && o.Active }
	out := func() bool { _, ok := p.output(); return ok }
	cmds := []commands.Command{
		{ID: "monitors.mode", Title: "Choose the resolution and refresh rate", Help: "mode", When: active, Run: p.modes},
		{ID: "monitors.power", Title: "Turn the display on or off", Help: "on/off", When: out, Run: p.power},
		{ID: "monitors.more", Title: "Scale or brightness up", Help: "adjust", When: func() bool { _, l := p.light(); return active() || l }, Run: func() tea.Cmd { return p.adjust(1) }},
		{ID: "monitors.less", Title: "Scale or brightness down", When: func() bool { _, l := p.light(); return active() || l }, Run: func() tea.Cmd { return p.adjust(-1) }},
		{ID: "monitors.rotate", Title: "Rotate the display", Help: "rotate", When: active, Run: p.rotate},
		{ID: "monitors.place", Title: "Place the display next to another", Help: "place", When: func() bool { return active() && p.activeCount() > 1 }, Run: p.place},
		{ID: "monitors.sync", Title: "Toggle adaptive sync", When: active, Run: func() tea.Cmd {
			o, _ := p.output()
			return p.do(act("monitors", "Adaptive sync "+onOff(!o.SyncOn())+" on "+o.Name, func() error { return monitor.SetSync(o.Name, !o.SyncOn()) }))
		}},
		{ID: "monitors.config", Title: "Show the config to keep this layout (sway, kanshi)", Help: "config", When: func() bool { return len(p.d.outs) > 0 }, Run: func() tea.Cmd {
			return overlay.Open(&configBox{outs: p.d.outs})
		}},
		{ID: "monitors.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Monitors", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func (p *Monitors) detailLines() []string {
	if l, ok := p.light(); ok {
		return []string{kv("Device", l.Name), kv("Brightness", level(l.Percent, 16))}
	}
	o, ok := p.output()
	if !ok {
		return nil
	}
	state := theme.Subtle().Render("off")
	if o.Active {
		state = theme.Success().Render("on")
	}
	lines := []string{kv("Name", o.Name), kv("Model", o.Label()), kv("State", state)}
	if o.Serial != "" && o.Serial != "Unknown" {
		lines = append(lines, kv("Serial", o.Serial))
	}
	if !o.Active {
		return append(lines, kv("Modes", itoa(len(o.Modes))))
	}
	w, h := o.Logical()
	lines = append(lines,
		kv("Mode", o.Mode.String()),
		kv("Scale", strconv.FormatFloat(o.Scale, 'f', -1, 64)),
		kv("Rotation", o.Transform),
		kv("Position", fmt.Sprintf("%d,%d", o.Rect.X, o.Rect.Y)),
		kv("Size", fmt.Sprintf("%dx%d", w, h)),
		kv("Adaptive sync", onOff(o.SyncOn())),
	)
	if o.Workspace != "" {
		lines = append(lines, kv("Workspace", o.Workspace))
	}
	if pv := layoutPreview(p.d.outs, o.Name, 36, 8); pv != nil {
		lines = append(lines, blank())
		lines = append(lines, pv...)
	}
	if probs := monitor.Problems(p.d.outs); len(probs) > 0 {
		lines = append(lines, blank())
		for _, pr := range probs {
			lines = append(lines, theme.Warning().Render(pr))
		}
		if p.activeCount() > 1 {
			lines = append(lines, theme.Subtle().Render("p places this display next to another"))
		}
	}
	return lines
}

func (p *Monitors) View(w, h int) string {
	f := Frame{Title: icon("\U000f0379 ", "") + "Monitors", DetailTitle: "Details", Empty: "no display"}
	switch {
	case p.d.err != nil:
		f.Empty = p.d.err.Error()
	case !p.load:
		f.Empty = "loading"
	default:
		f.Right = "until sway reloads"
		f.Detail = p.detailLines()
	}
	return p.list.Render(w, h, f)
}

func (p *Monitors) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) {
		return p.modes()
	}
	return nil
}
