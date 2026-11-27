package pages

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/settuings/internal/monitor"
)

const (
	cellPlain = iota
	cellBorder
	cellSelected
	cellLabel
	cellLabelSelected
)

type cell struct {
	r     rune
	style int
}

func layoutPreview(outs []monitor.Output, selected string, w, h int) []string {
	var act []monitor.Output
	for _, o := range outs {
		if o.Active {
			act = append(act, o)
		}
	}
	if len(act) == 0 || w < 16 || h < 5 {
		return nil
	}
	minX, minY, maxX, maxY := act[0].Rect.X, act[0].Rect.Y, 0, 0
	for _, o := range act {
		minX, minY = min(minX, o.Rect.X), min(minY, o.Rect.Y)
		maxX, maxY = max(maxX, o.Rect.X+o.Rect.Width), max(maxY, o.Rect.Y+o.Rect.Height)
	}
	totalW, totalH := float64(maxX-minX), float64(maxY-minY)
	scale := max(totalW/float64(w), totalH/float64(h*2))
	grid := make([][]cell, h)
	for y := range grid {
		grid[y] = make([]cell, w)
		for x := range grid[y] {
			grid[y][x] = cell{' ', cellPlain}
		}
	}
	cellX := func(v int) int { return int(math.Round(float64(v-minX) / scale)) }
	cellY := func(v int) int { return int(math.Round(float64(v-minY) / (scale * 2))) }
	offX, offY := max((w-cellX(maxX))/2, 0), max((h-cellY(maxY))/2, 0)
	draw := func(o monitor.Output, sel bool) {
		x0, x1 := offX+cellX(o.Rect.X), offX+cellX(o.Rect.X+o.Rect.Width)
		y0, y1 := offY+cellY(o.Rect.Y), offY+cellY(o.Rect.Y+o.Rect.Height)
		bw, bh := min(max(x1-x0, 3), w-x0), min(max(y1-y0, 3), h-y0)
		if bw < 3 || bh < 3 {
			return
		}
		bs, ls := cellBorder, cellLabel
		if sel {
			bs, ls = cellSelected, cellLabelSelected
		}
		set := func(x, y int, r rune, st int) {
			if x >= 0 && x < w && y >= 0 && y < h {
				grid[y][x] = cell{r, st}
			}
		}
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				r := ' '
				switch {
				case y == 0 && x == 0:
					r = '╭'
				case y == 0 && x == bw-1:
					r = '╮'
				case y == bh-1 && x == 0:
					r = '╰'
				case y == bh-1 && x == bw-1:
					r = '╯'
				case y == 0 || y == bh-1:
					r = '─'
				case x == 0 || x == bw-1:
					r = '│'
				}
				st := bs
				if r == ' ' {
					st = cellPlain
				}
				set(x0+x, y0+y, r, st)
			}
		}
		name := []rune(ansi.Truncate(o.Name, bw-2, "…"))
		for i, r := range name {
			set(x0+1+(bw-2-len(name))/2+i, y0+bh/2, r, ls)
		}
	}
	for _, o := range act {
		if o.Name != selected {
			draw(o, false)
		}
	}
	for _, o := range act {
		if o.Name == selected {
			draw(o, true)
		}
	}
	styles := map[int]lipgloss.Style{
		cellPlain:         lipgloss.NewStyle(),
		cellBorder:        theme.Muted().Style(),
		cellSelected:      theme.Primary().Style(),
		cellLabel:         theme.Subtle().Style(),
		cellLabelSelected: theme.Primary().Style().Bold(true),
	}
	lines := make([]string, h)
	for y, row := range grid {
		var b strings.Builder
		for x := 0; x < w; {
			end := x
			var run strings.Builder
			for end < w && row[end].style == row[x].style {
				run.WriteRune(row[end].r)
				end++
			}
			b.WriteString(styles[row[x].style].Render(run.String()))
			x = end
		}
		lines[y] = b.String()
	}
	return lines
}
