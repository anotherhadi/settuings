package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/geom"
	"github.com/anotherhadi/foam/help"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/palette"

	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/ui/pages"
)

type Model struct {
	cfg *config.Config
	reg *commands.Registry

	overlays overlay.Stack
	notes    notify.Center

	pages []pages.Page
	names []string
	cur   int

	w, h       int
	toastRects []toastRect
	tabRects   []tabRect
}

type toastRect struct {
	geom.Rect
	id int
}

type tabRect struct {
	x0, x1, page int
}

type tickMsg struct{}

func New(cfg *config.Config) (*Model, error) { return build(cfg, true) }

func build(cfg *config.Config, onlyAvailable bool) (*Model, error) {
	reg, err := commands.New(cfg.Keys)
	if err != nil {
		return nil, fmt.Errorf("invalid config:\n%w", err)
	}
	m := &Model{cfg: cfg, reg: reg}
	pages.KeyLabel = func(id string) string {
		first, _, _ := strings.Cut(reg.KeysOf(commands.Command{ID: id}).String(), " / ")
		return first
	}
	m.notes.Max = 100
	var names []string
	for _, name := range cfg.App.Pages {
		if p := newPage(name, cfg); !onlyAvailable || p.Available() {
			m.pages = append(m.pages, p)
			names = append(names, name)
		}
	}
	if len(m.pages) == 0 {
		return nil, errors.New("no page available: none of the tools they need is installed")
	}
	m.names = names
	m.cur = max(slices.Index(names, cfg.App.Page), 0)
	if err := reg.Check(m.allScopes()); err != nil {
		return nil, fmt.Errorf("invalid config:\n%w", err)
	}
	return m, nil
}

func newPage(name string, cfg *config.Config) pages.Page {
	switch name {
	case "network":
		return pages.NewNetwork()
	case "bluetooth":
		return pages.NewBluetooth()
	case "audio":
		return pages.NewAudio(cfg.App.VolumeStep)
	case "power":
		return pages.NewPower()
	case "inputs":
		return pages.NewInputs(cfg.App.SpeedStep)
	case "monitors":
		return pages.NewMonitors(cfg.App.LightStep)
	case "netbird":
		return pages.NewNetBird()
	}
	return pages.NewAbout()
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(time.Duration(m.cfg.App.Refresh)*time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *Model) Init() tea.Cmd { return tea.Batch(m.page().Init(), m.tick()) }

func (m *Model) Close() {
	for _, p := range m.pages {
		p.Close()
	}
}

func (m *Model) page() pages.Page { return m.pages[m.cur] }

func (m *Model) setPage(i int) tea.Cmd {
	n := len(m.pages)
	i = (i%n + n) % n
	if i == m.cur {
		return nil
	}
	m.page().Leave()
	m.cur = i
	return m.page().Refresh(false)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, m.update(msg) }

func (m *Model) update(msg tea.Msg) tea.Cmd {
	if m.overlays.Handle(msg) {
		return nil
	}
	if cmd, ok := m.notes.Update(msg); ok {
		return cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return nil
	case tea.KeyPressMsg:
		return m.key(msg)
	case tea.PasteMsg:
		if o := m.overlays.Top(); o != nil {
			return o.Update(msg)
		}
		return nil
	case tea.MouseMsg:
		return m.mouse(msg)
	case tickMsg:
		return tea.Batch(m.page().Refresh(false), m.tick())
	case pages.DoneMsg:
		return m.done(msg)
	case pages.CloseIfTop:
		if m.overlays.Top() == msg.O {
			return overlay.Close
		}
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.pages))
	for _, p := range m.pages {
		cmds = append(cmds, p.Update(msg))
	}
	return tea.Batch(cmds...)
}

func (m *Model) done(msg pages.DoneMsg) tea.Cmd {
	var cmds []tea.Cmd
	switch {
	case msg.Err != nil:
		cmds = append(cmds, notify.Err(msg.Err))
	case msg.OK != "":
		cmds = append(cmds, notify.Success(msg.OK))
	}
	for _, p := range m.pages {
		cmds = append(cmds, p.Update(msg))
		if p.Name() == msg.Page {
			cmds = append(cmds, p.Refresh(false))
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) key(msg tea.KeyPressMsg) tea.Cmd {
	if !minSize.Fits(m.w, m.h) {
		if b, ok := m.reg.Match(m.path()[:1], msg); ok && b.ID == "app.quit" {
			return b.Run()
		}
		return nil
	}
	if m.overlays.Top() != nil {
		return m.overlays.Key(m.reg, msg)
	}
	if b, ok := m.reg.Match(m.path(), msg); ok {
		return b.Run()
	}
	return nil
}

func (m *Model) mouse(msg tea.MouseMsg) tea.Cmd {
	if !minSize.Fits(m.w, m.h) {
		return nil
	}
	if m.overlays.Top() != nil {
		return m.overlays.Mouse(msg)
	}
	mm := msg.Mouse()
	_, click := msg.(tea.MouseClickMsg)
	if click {
		for _, t := range m.toastRects {
			if t.Contains(mm.X, mm.Y) {
				m.notes.DismissID(t.id)
				return nil
			}
		}
		if mm.Y == 0 {
			for _, t := range m.tabRects {
				if mm.X >= t.x0 && mm.X < t.x1 {
					return m.setPage(t.page)
				}
			}
			return nil
		}
	}
	body := m.body()
	if !body.Contains(mm.X, mm.Y) {
		return nil
	}
	mm.X, mm.Y = body.Local(mm.X, mm.Y)
	switch msg.(type) {
	case tea.MouseClickMsg:
		return m.page().Mouse(tea.MouseClickMsg(mm))
	case tea.MouseWheelMsg:
		return m.page().Mouse(tea.MouseWheelMsg(mm))
	}
	return nil
}

func (m *Model) global() []commands.Command {
	cmds := []commands.Command{
		{ID: "app.quit", Title: "Quit", Run: func() tea.Cmd { return tea.Quit }},
		{ID: "app.help", Title: "Show the shortcuts", Help: "help", Run: func() tea.Cmd {
			return overlay.Open(help.New(m.reg.Groups(m.path())))
		}},
		{ID: "app.palette", Title: "Command palette", Run: func() tea.Cmd {
			return overlay.Open(palette.New(m.reg.Active(m.path()), "app.palette"))
		}},
		{ID: "app.refresh", Title: "Refresh", Help: "refresh", Run: func() tea.Cmd { return m.page().Refresh(true) }},
		{ID: "page.next", Title: "Next page", Help: "page", When: func() bool { return len(m.pages) > 1 }, Run: func() tea.Cmd { return m.setPage(m.cur + 1) }},
		{ID: "page.prev", Title: "Previous page", When: func() bool { return len(m.pages) > 1 }, Run: func() tea.Cmd { return m.setPage(m.cur - 1) }},
	}
	for _, name := range config.PageNames {
		i := slices.Index(m.names, name)
		label := fmt.Sprintf("Go to %s", name)
		cmds = append(cmds, commands.Command{ID: "page." + name, Title: label, When: func() bool { return i >= 0 }, Run: func() tea.Cmd { return m.setPage(i) }})
	}
	return cmds
}

func (m *Model) path() []commands.Scope {
	p := m.page()
	return []commands.Scope{
		{Name: "settuings", Commands: m.global()},
		{Name: p.Label(), Commands: p.Commands()},
	}
}

func (m *Model) allScopes() []commands.Scope {
	s := []commands.Scope{{Name: "settuings", Commands: m.global()}}
	for _, p := range m.pages {
		s = append(s, commands.Scope{Name: p.Label(), Commands: p.Commands()})
	}
	return append(s, pages.OverlayScopes()...)
}
