package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam"
	fconfig "github.com/anotherhadi/foam/config"
	"github.com/anotherhadi/foam/confirm"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/ui/pages"
)

func newModel(t *testing.T) *Model {
	t.Helper()
	isolate(t)
	cfg, err := fconfig.LoadBytes[config.Config](foam.Defaults, config.Defaults)
	if err != nil {
		t.Fatal(err)
	}
	theme.Set(cfg.Theme)
	m, err := build(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(m.Close)
	return m
}

func checkScreen(t *testing.T, m *Model, w, h int, state string) {
	t.Helper()
	lines := strings.Split(m.View().Content, "\n")
	if len(lines) != h {
		t.Fatalf("%s at %dx%d: %d lines", state, w, h, len(lines))
	}
	for i, l := range lines {
		if got := ansi.StringWidth(l); got > w || got < w && strings.TrimSpace(ansi.Strip(l)) != "" {
			t.Fatalf("%s at %dx%d: line %d is %d wide: %q", state, w, h, i, got, ansi.Strip(l))
		}
	}
}

func TestResponsive(t *testing.T) {
	m := newModel(t)
	for i, p := range m.pages {
		m.cur = i
		if cmd := p.Refresh(true); cmd != nil {
			m.update(cmd())
		}
		for _, size := range [][2]int{{200, 60}, {120, 40}, {100, 30}, {89, 25}, {80, 24}, {60, 20}, {45, 14}, {40, 12}, {33, 10}, {32, 10}, {20, 5}} {
			w, h := size[0], size[1]
			m.update(tea.WindowSizeMsg{Width: w, Height: h})
			checkScreen(t, m, w, h, p.Name())
		}
	}
}

func TestKeys(t *testing.T) {
	m := newModel(t)
	m.update(tea.WindowSizeMsg{Width: 100, Height: 30})
	for i := range m.pages {
		m.cur = i
		if cmd := m.page().Refresh(true); cmd != nil {
			m.update(cmd())
		}
		for _, k := range []string{"j", "j", "k", "G", "g", "i", "esc", "?", "esc", ":", "esc", "tab"} {
			m.update(tea.KeyPressMsg(tea.Key{Text: k, Code: []rune(k)[0]}))
			checkScreen(t, m, 100, 30, "key "+k)
		}
	}
}

func TestCloseIfTop(t *testing.T) {
	m := newModel(t)
	m.update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a, b := confirm.New("A", "a?", nil), confirm.New("B", "b?", nil)
	for _, o := range []*confirm.Model{a, b} {
		m.update(overlay.Open(o)())
	}
	if cmd := m.update(pages.CloseIfTop{O: a}); cmd != nil || m.overlays.Len() != 2 {
		t.Fatal("closed an overlay that is not on top")
	}
	m.update(m.update(pages.CloseIfTop{O: b})())
	if m.overlays.Len() != 1 || m.overlays.Top() != overlay.Overlay(a) {
		t.Fatalf("overlays left: %d", m.overlays.Len())
	}
}

func TestHelpBarFollowsThePopup(t *testing.T) {
	m := newModel(t)
	m.update(tea.WindowSizeMsg{Width: 100, Height: 30})
	page := ansi.Strip(m.helpBar())
	if !strings.Contains(page, "refresh") {
		t.Fatalf("page help bar: %q", page)
	}
	m.update(overlay.Open(confirm.New("Q", "sure?", nil))())
	popup := ansi.Strip(m.helpBar())
	if !strings.Contains(popup, "yes") || !strings.Contains(popup, "no") || strings.Contains(popup, "refresh") {
		t.Fatalf("popup help bar: %q", popup)
	}
}
