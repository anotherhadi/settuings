package pages

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/audio"
)

type audioKind int

const (
	audSink audioKind = iota
	audSource
	audStream
)

type audioRow struct {
	kind   audioKind
	id     string
	name   string
	volume int
	muted  bool
	def    bool
	input  bool
}

type audioData struct {
	err     error
	sinks   []audio.Device
	sources []audio.Device
	streams []audio.Stream
}

type audioMsg audioData

type levelMsg struct {
	id   string
	peak float64
	err  error
}

type Audio struct {
	pending
	list  List
	d     audioData
	busy  bool
	load  bool
	step  int
	mon   *audio.LevelMonitor
	monID string
	peak  float64
}

func NewAudio(step int) *Audio { return &Audio{step: step} }

func (p *Audio) Available() bool { return audio.Installed() }
func (p *Audio) Name() string    { return "audio" }
func (p *Audio) Label() string   { return "Audio" }
func (p *Audio) Init() tea.Cmd   { return p.Refresh(true) }
func (p *Audio) Leave()          { p.stopMonitor() }
func (p *Audio) Close()          { p.stopMonitor() }

func (p *Audio) Refresh(bool) tea.Cmd {
	if p.busy {
		return nil
	}
	p.busy = true
	return func() tea.Msg {
		var d audioData
		if d.err = audio.CheckAvailable(); d.err != nil {
			return audioMsg(d)
		}
		if d.sinks, d.err = audio.ListSinks(); d.err != nil {
			return audioMsg(d)
		}
		d.sources, _ = audio.ListSources()
		d.streams, _ = audio.ListStreams()
		return audioMsg(d)
	}
}

func (p *Audio) Update(msg tea.Msg) tea.Cmd {
	p.clear(msg, "audio")
	switch m := msg.(type) {
	case audioMsg:
		p.busy, p.load = false, true
		p.d = audioData(m)
		p.rebuild()
	case levelMsg:
		if p.mon == nil || m.id != p.monID {
			return nil
		}
		if m.err != nil {
			p.stopMonitor()
			return nil
		}
		p.peak = m.peak
		return p.nextLevel()
	}
	return nil
}

func (p *Audio) stopMonitor() {
	if p.mon != nil {
		p.mon.Stop()
		p.mon, p.monID, p.peak = nil, "", 0
	}
}

func (p *Audio) nextLevel() tea.Cmd {
	mon, id := p.mon, p.monID
	return func() tea.Msg {
		v, err := mon.Next()
		return levelMsg{id: id, peak: v, err: err}
	}
}

func volNote(r audioRow) string {
	switch {
	case r.muted:
		return theme.Warning().Render("muted")
	}
	s := theme.Subtle().Render(itoa(r.volume) + "%")
	if r.def {
		s = theme.Success().Render("●") + " " + s
	}
	return s
}

func (p *Audio) rebuild() {
	var items []Item
	add := func(title string, devs []audio.Device, kind audioKind, ic string) {
		if len(devs) == 0 {
			return
		}
		items = append(items, Item{Header: true, Label: title})
		for _, d := range devs {
			r := audioRow{kind: kind, id: d.ID, name: d.Name, volume: d.Volume, muted: d.Muted, def: d.Default, input: kind == audSource}
			items = append(items, Item{Key: fmt.Sprint(kind, d.ID), Icon: ic, Label: d.Name, Note: volNote(r), Data: r})
		}
	}
	add("Outputs", p.d.sinks, audSink, icon("\U000f04c3", ""))
	add("Inputs", p.d.sources, audSource, icon("\uf130", ""))
	var apps []Item
	for _, s := range p.d.streams {
		r := audioRow{kind: audStream, id: s.ID, name: s.App, volume: s.Volume, muted: s.Muted, input: s.Kind == audio.StreamInput}
		ic := icon("\U000f04c3", "")
		if r.input {
			ic = icon("\uf130", "")
		}
		apps = append(apps, Item{Key: "stream" + s.ID, Icon: ic, Label: s.App, Note: volNote(r), Data: r})
	}
	if len(apps) > 0 {
		items = append(items, Item{Header: true, Label: "Applications"})
		items = append(items, apps...)
	}
	p.list.Set(items)
}

func (p *Audio) row() (audioRow, bool) {
	it, ok := p.list.Selected()
	if !ok {
		return audioRow{}, false
	}
	r, ok := it.Data.(audioRow)
	return r, ok
}

func (p *Audio) setVolume(d int) tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	v := min(max(r.volume+d, 0), 100)
	if v == r.volume {
		return nil
	}
	r.volume = v
	p.patch(r)
	return act("audio", "", func() error { return audio.SetVolume(r.id, v) })
}

func (p *Audio) patch(r audioRow) {
	for i := range p.d.sinks {
		if p.d.sinks[i].ID == r.id && r.kind == audSink {
			p.d.sinks[i].Volume, p.d.sinks[i].Muted = r.volume, r.muted
		}
	}
	for i := range p.d.sources {
		if p.d.sources[i].ID == r.id && r.kind == audSource {
			p.d.sources[i].Volume, p.d.sources[i].Muted = r.volume, r.muted
		}
	}
	for i := range p.d.streams {
		if p.d.streams[i].ID == r.id && r.kind == audStream {
			p.d.streams[i].Volume, p.d.streams[i].Muted = r.volume, r.muted
		}
	}
	p.rebuild()
}

func (p *Audio) test() tea.Cmd {
	r, ok := p.row()
	if !ok {
		return nil
	}
	switch r.kind {
	case audSink:
		return tea.Batch(notify.Info("Playing a test tone"), act("audio", "", func() error { return audio.PlayTestTone(r.id) }))
	case audSource:
		if p.monID == r.id {
			p.stopMonitor()
			return nil
		}
		p.stopMonitor()
		mon, err := audio.StartLevelMonitor(r.id)
		if err != nil {
			return notify.Err(err)
		}
		p.mon, p.monID = mon, r.id
		return p.nextLevel()
	}
	return nil
}

func (p *Audio) Commands() []commands.Command {
	sel := func(f func(audioRow) bool) func() bool {
		return func() bool { r, ok := p.row(); return ok && f(r) }
	}
	device := func(r audioRow) bool { return r.kind != audStream }
	cmds := []commands.Command{
		{ID: "audio.default", Title: "Set as default", Help: "default", When: sel(func(r audioRow) bool { return device(r) && !r.def }), Run: func() tea.Cmd {
			r, _ := p.row()
			return p.do(act("audio", r.name+" is the default", func() error { return audio.SetDefault(r.id) }))
		}},
		{ID: "audio.mute", Title: "Toggle mute", Help: "mute", When: sel(func(audioRow) bool { return true }), Run: func() tea.Cmd {
			r, _ := p.row()
			r.muted = !r.muted
			p.patch(r)
			return act("audio", "", func() error { return audio.SetMute(r.id, r.muted) })
		}},
		{ID: "audio.louder", Title: "Volume up", Help: "volume", When: sel(func(audioRow) bool { return true }), Run: func() tea.Cmd { return p.setVolume(p.step) }},
		{ID: "audio.quieter", Title: "Volume down", When: sel(func(audioRow) bool { return true }), Run: func() tea.Cmd { return p.setVolume(-p.step) }},
		{ID: "audio.test", Title: "Test tone, or input level meter", Help: "test", When: sel(device), Run: p.test},
		{ID: "audio.info", Title: "Show the details", Help: "info", When: func() bool { return p.list.DetailHidden() && len(p.detailLines()) > 0 }, Run: func() tea.Cmd {
			return openInfo("Audio", p.detailLines())
		}},
	}
	return append(cmds, p.list.Commands()...)
}

func (p *Audio) detailLines() []string {
	r, ok := p.row()
	if !ok {
		return nil
	}
	kind := map[audioKind]string{audSink: "output", audSource: "input"}[r.kind]
	if r.kind == audStream {
		kind = "playback"
		if r.input {
			kind = "recording"
		}
	}
	lines := []string{kv("Name", r.name), kv("Type", kind), kv("ID", r.id), kv("Volume", level(r.volume, 16)), kv("Muted", yesNo(r.muted))}
	if device(r) {
		lines = append(lines, kv("Default", yesNo(r.def)))
	}
	if r.kind == audSource {
		meter := theme.Subtle().Render("press t to monitor")
		if p.monID == r.id {
			meter = level(int(p.peak*100), 16)
		}
		lines = append(lines, blank(), kv("Level", meter))
	}
	return lines
}

func device(r audioRow) bool { return r.kind != audStream }

func (p *Audio) View(w, h int) string {
	f := Frame{Title: icon("\U000f057e ", "") + "Audio", DetailTitle: "Details", Empty: "no audio device"}
	switch {
	case p.d.err != nil:
		f.Empty = p.d.err.Error()
	case !p.load:
		f.Empty = "loading"
	default:
		f.Detail = p.detailLines()
	}
	return p.list.Render(w, h, f)
}

func (p *Audio) Mouse(msg tea.MouseMsg) tea.Cmd {
	if p.list.Mouse(msg) {
		if r, ok := p.row(); ok && device(r) && !r.def {
			return act("audio", r.name+" is the default", func() error { return audio.SetDefault(r.id) })
		}
	}
	return nil
}
