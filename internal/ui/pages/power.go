package pages

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/power"
)

type powerData struct {
	err       error
	profiles  []power.Profile
	batteries []power.Battery
	ac        power.ACStatus
}

type powerMsg powerData

type Power struct {
	pending
	list List
	d    powerData
	busy bool
	load bool
}

func NewPower() *Power { return &Power{} }

func (p *Power) Available() bool { return power.Installed() || batteries() }
func (p *Power) Name() string    { return "power" }
func (p *Power) Label() string   { return "Power" }
func (p *Power) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Power) Leave()          {}
func (p *Power) Close()          {}

func (p *Power) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		var d powerData
		if power.Installed() {
			d.profiles, _ = power.ListProfiles()
		}
		d.batteries, _ = power.ReadBatteries()
		d.ac, _ = power.ReadAC()
		if len(d.profiles) == 0 && len(d.batteries) == 0 {
			d.err = power.ErrNotFound
		}
		return powerMsg(d)
	}
}

func (p *Power) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "power")
	m, ok := msg.(powerMsg)
	if !ok {
		return nil
	}
	p.busy, p.load = false, true
	p.d = powerData(m)
	p.rebuild()
	return nil
}

func (p *Power) rebuild() {
	var items []Item
	if len(p.d.profiles) > 0 {
		items = append(items, Item{Header: true, Label: "Profiles"})
	}
	for _, pr := range p.d.profiles {
		note := ""
		if pr.Active {
			note = theme.Success().Render("active")
		}
		items = append(items, Item{Key: "profile:" + pr.Name, Icon: icon("\uf0e7", ""), Label: pr.Name, Note: note, Data: pr})
	}
	if len(p.d.batteries) > 0 {
		items = append(items, Item{Header: true, Label: "Batteries"})
	}
	for _, b := range p.d.batteries {
		note := theme.Subtle().Render(strings.ToLower(b.Status))
		if b.Capacity >= 0 {
			note = batteryColor(b.Capacity).Render(itoa(b.Capacity)+"%") + " " + note
		}
		items = append(items, Item{Key: "battery:" + b.Name, Icon: icon("\uf240", ""), Label: b.Name, Note: note, Data: b})
	}
	p.list.Set(items)
}

var profileHelp = []string{
	theme.Accent().Style().Bold(true).Render("performance"),
	theme.Fg().Render("Full speed: the CPU and GPU run as fast as they can. Fans get louder, the machine heats up and the battery drains faster. Best when plugged in."),
	"",
	theme.Accent().Style().Bold(true).Render("balanced"),
	theme.Fg().Render("The default: speed when something needs it, savings when nothing does. A good everyday compromise."),
	"",
	theme.Accent().Style().Bold(true).Render("power-saver"),
	theme.Fg().Render("Holds the performance back to stretch the battery and keep the machine cool and quiet. Things may feel slower."),
}

func batteryColor(pct int) theme.Color {
	switch {
	case pct <= 15:
		return theme.Danger()
	case pct <= 30:
		return theme.Warning()
	}
	return theme.Success()
}

func (p *Power) profile() (power.Profile, bool) {
	it, ok := p.list.Selected()
	if !ok {
		return power.Profile{}, false
	}
	pr, ok := it.Data.(power.Profile)
	return pr, ok
}

func (p *Power) use(name string) tea.Cmd {
	if p.on {
		return nil
	}
	for i := range p.d.profiles {
		p.d.profiles[i].Active = p.d.profiles[i].Name == name
	}
	p.rebuild()
	return p.do(act("power", "Profile: "+name, func() error { return power.SetActive(name) }))
}

func (p *Power) Commands() []commands.Command {
	cmds := []commands.Command{
		{ID: "power.set", Title: "Use this power profile", Help: "use", When: func() bool { pr, ok := p.profile(); return ok && !pr.Active }, Run: func() tea.Cmd {
			pr, _ := p.profile()
			return p.use(pr.Name)
		}},
		{ID: "power.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Power", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func (p *Power) detailLines() []string {
	it, ok := p.list.Selected()
	if !ok {
		return nil
	}
	switch v := it.Data.(type) {
	case power.Profile:
		return profileHelp
	case power.Battery:
		lines := []string{kv("Battery", v.Name), kv("Status", v.Status)}
		if v.Capacity >= 0 {
			lines = append(lines, kv("Charge", level(v.Capacity, 16)))
		}
		if v.Health >= 0 {
			lines = append(lines, kv("Health", level(v.Health, 16)))
		}
		add := func(k, s string) {
			if s != "" {
				lines = append(lines, kv(k, s))
			}
		}
		add("Time left", v.TimeToEmpty)
		add("Until full", v.TimeToFull)
		add("Technology", v.Technology)
		if v.CycleCount >= 0 {
			add("Cycles", itoa(v.CycleCount))
		}
		add("Maker", v.Manufacturer)
		add("Model", v.Model)
		if p.d.ac.Present {
			lines = append(lines, blank(), kv("AC power", yesNo(p.d.ac.Online)))
		}
		return lines
	}
	return nil
}

func (p *Power) View(w, h int) string {
	f := Frame{Title: icon("\U000f140b ", "") + "Power", DetailTitle: "Details", Empty: "no power info"}
	switch {
	case !p.load:
		f.Empty = "loading"
	case p.d.err != nil:
		f.Empty = "no battery, and powerprofilesctl not found"
	default:
		f.Detail = p.detailLines()
		if p.d.ac.Present {
			f.Right = map[bool]string{true: "plugged in", false: "on battery"}[p.d.ac.Online]
		}
	}
	if p.on {
		f.Right = "working…"
	}
	return p.list.Render(w, h, f)
}

func (p *Power) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) {
		if pr, ok := p.profile(); ok && !pr.Active {
			return p.use(pr.Name)
		}
	}
	return nil
}

func batteries() bool {
	b, _ := power.ReadBatteries()
	return len(b) > 0
}
