package util

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

// ScrollbarGutterWidth is the fixed width, in terminal columns, that
// RefreshScrollbar reserves for its gutter — whether or not a scrollbar is
// actually needed right now. Content that pre-wraps itself to a viewport's
// width (e.g. glamour markdown) must subtract this, otherwise it overflows
// the moment the scrollbar's thumb/track appear and the gutter eats into
// the width the content assumed it had.
const ScrollbarGutterWidth = 2

// SetYOffset sets vp's Y offset and refreshes its scrollbar gutter to match.
// Every scroll site should go through this (or [ScrollLines]/[ScrollViewport])
// instead of calling vp.SetYOffset directly, so the scrollbar never goes
// stale.
func SetYOffset(vp *viewport.Model, y int) {
	vp.SetYOffset(y)
	RefreshScrollbar(vp)
}

// ScrollLines scrolls vp by n lines (negative n scrolls up) and refreshes
// its scrollbar gutter.
func ScrollLines(vp *viewport.Model, n int) {
	SetYOffset(vp, vp.YOffset()+n)
}

// ScrollViewport scrolls vp vertically by half its height, then refreshes
// its scrollbar gutter. delta should be -1 for up, +1 for down.
func ScrollViewport(vp *viewport.Model, delta int) {
	step := vp.Height() / 2
	if step < 1 {
		step = 1
	}
	ScrollLines(vp, delta*step)
}

// HandleMouseWheel applies standard mouse wheel scrolling to vp, refreshing
// its scrollbar gutter on vertical scrolls.
// Vertical: one line at a time. Shift+vertical or horizontal: scroll 6 columns.
func HandleMouseWheel(msg tea.MouseWheelMsg, vp *viewport.Model) {
	switch msg.Button {
	case tea.MouseWheelUp:
		if msg.Mod.Contains(tea.ModShift) {
			vp.ScrollLeft(6)
		} else {
			ScrollLines(vp, -1)
		}
	case tea.MouseWheelDown:
		if msg.Mod.Contains(tea.ModShift) {
			vp.ScrollRight(6)
		} else {
			ScrollLines(vp, 1)
		}
	case tea.MouseWheelLeft:
		vp.ScrollLeft(6)
	case tea.MouseWheelRight:
		vp.ScrollRight(6)
	}
}

// RefreshScrollbar rebuilds vp's left-gutter scrollbar to reflect its
// current size, content and scroll position. Call it after anything that
// changes vp's width, height or content; YOffset changes are already
// covered if they go through [SetYOffset], [ScrollLines], [ScrollViewport]
// or [HandleMouseWheel].
//
// The gutter is always [ScrollbarGutterWidth] columns wide, even when
// there's nothing to scroll (it's just blank then). That keeps vp's usable
// content width constant regardless of scroll state, so content sized
// against vp.Width() doesn't need to know whether a scrollbar happens to be
// showing right now.
func RefreshScrollbar(vp *viewport.Model) {
	height := vp.Height()
	total := vp.TotalLineCount()
	blank := strings.Repeat(" ", ScrollbarGutterWidth)

	if height <= 0 || total <= height {
		vp.LeftGutterFunc = func(viewport.GutterContext) string { return blank }
		return
	}

	yOffset := vp.YOffset()
	thumbSize := max(1, height*height/total)
	thumbStart := 0
	if maxOffset := total - height; maxOffset > 0 {
		thumbStart = yOffset * (height - thumbSize) / maxOffset
	}

	trackStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Subtle)
	thumbStyle := lipgloss.NewStyle().Foreground(ilovetui.S.Primary)

	vp.LeftGutterFunc = func(ctx viewport.GutterContext) string {
		pos := ctx.Index - yOffset
		if pos >= thumbStart && pos < thumbStart+thumbSize {
			return thumbStyle.Render("█") + " "
		}
		return trackStyle.Render("│") + " "
	}
}
