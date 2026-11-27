package pages

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/theme"
	"github.com/anotherhadi/foam/viewer"

	"github.com/anotherhadi/settuings/internal/about"
)

type aboutMsg struct {
	info about.Info
	err  error
}

type About struct {
	v     viewer.Model
	info  about.Info
	err   error
	busy  bool
	load  bool
	built [2]int
}

func NewAbout() *About { return &About{built: [2]int{-1, -1}} }

func (p *About) Available() bool { return about.Installed() }
func (p *About) Name() string    { return "about" }
func (p *About) Label() string   { return "About" }
func (p *About) Init() tea.Cmd   { return p.Refresh(true) }
func (p *About) Leave()          {}
func (p *About) Close()          {}

func (p *About) Refresh(manual bool) tea.Cmd {
	if p.busy || (p.load && !manual) {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		i, err := about.Fetch()
		return aboutMsg{info: i, err: err}
	}
}

func (p *About) Update(msg tea.Msg) tea.Cmd {
	m, ok := msg.(aboutMsg)
	if !ok {
		return nil
	}
	p.busy, p.load = false, true
	p.info, p.err = m.info, m.err
	p.built = [2]int{-1, -1}
	return nil
}

func (p *About) Commands() []commands.Command { return p.v.Commands() }

func (p *About) Mouse(msg tea.MouseMsg) tea.Cmd { return p.v.Mouse(msg) }

func (p *About) text(w int) []string {
	i := p.info
	keyW := 0
	for _, f := range i.Fields {
		keyW = max(keyW, ansi.StringWidth(f.Key))
	}
	keyW = min(keyW, max(w/3, 1))
	valW := max(w-keyW-2, 1)
	var out []string
	if i.Title != "" {
		out = append(out, strings.Split(theme.Accent().Style().Bold(true).Render(ansi.Wrap(i.Title, w, "")), "\n")...)
		out = append(out, draw.Rule(min(ansi.StringWidth(i.Title), w)))
	}
	key := theme.Primary().Style().Bold(true)
	for _, f := range i.Fields {
		k := draw.Pad(ansi.Truncate(f.Key, keyW, "…"), keyW)
		for n, v := range strings.Split(ansi.Wrap(f.Value, valW, ""), "\n") {
			if n > 0 {
				k = strings.Repeat(" ", keyW)
			}
			out = append(out, key.Render(k)+"  "+theme.Fg().Render(v))
		}
	}
	return out
}

func (p *About) lines(w, h int) []string {
	logo := p.info.Logo
	logoW := 0
	if len(logo) > 0 {
		logoW = ansi.StringWidth(logo[0])
	}
	if side := w - logoW - 3; logoW > 0 && side >= 28 {
		text := p.text(side)
		top := max((len(logo)-len(text))/2, 0)
		out := make([]string, max(len(logo), len(text)+top))
		for k := range out {
			l := strings.Repeat(" ", logoW)
			if k < len(logo) {
				l = logo[k]
			}
			if t := k - top; t >= 0 && t < len(text) {
				l += "   " + text[t]
			}
			out[k] = l
		}
		return out
	}
	text := p.text(w)
	if logoW > 0 && logoW <= w && h >= len(logo)+min(len(text), 4)+1 {
		return append(append(append([]string(nil), logo...), ""), text...)
	}
	return text
}

func (p *About) View(w, h int) string {
	iw, ih := max(w-2, 0), max(h-2, 0)
	o := draw.FrameOpts{Title: icon("\uf05a ", "") + "About", Active: true}
	switch {
	case p.err != nil:
		return draw.Frame(draw.Hatch(iw, ih, p.err.Error()), w, h, o)
	case !p.load:
		return draw.Frame(draw.Hatch(iw, ih, "loading"), w, h, o)
	}
	pad := 1
	if iw < 40 {
		pad = 0
	}
	cw := max(iw-2*pad, 0)
	if p.built != [2]int{cw, ih} {
		rows := p.lines(cw, ih)
		plain := make([]string, len(rows))
		for i, r := range rows {
			plain[i] = ansi.Strip(r)
		}
		p.v.SetContent(plain, rows)
		p.built = [2]int{cw, ih}
	}
	p.v.SetSize(cw, ih)
	o.Scroll = p.v.Scroll()
	body := lipgloss.NewStyle().PaddingLeft(pad).Render(p.v.View())
	return draw.Frame(draw.Fit(body, iw, ih), w, h, o)
}
