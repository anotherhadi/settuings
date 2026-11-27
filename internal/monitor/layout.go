package monitor

import "fmt"

type Move struct {
	Name string
	X, Y int
}

func find(outs []Output, name string) *Output {
	for i := range outs {
		if outs[i].Name == name {
			return &outs[i]
		}
	}
	return nil
}

func repackMoves(before, after []Output, name string) []Move {
	old, cur := find(before, name), find(after, name)
	if old == nil || cur == nil || !old.Active || !cur.Active {
		return nil
	}
	dw, dh := cur.Rect.Width-old.Rect.Width, cur.Rect.Height-old.Rect.Height
	if dw == 0 && dh == 0 {
		return nil
	}
	right, bottom := old.Rect.X+old.Rect.Width, old.Rect.Y+old.Rect.Height
	var moves []Move
	for _, r := range after {
		if r.Name == name || !r.Active {
			continue
		}
		x, y := r.Rect.X, r.Rect.Y
		if r.Rect.X >= right {
			x += dw
		}
		if r.Rect.Y >= bottom && r.Rect.X < right && r.Rect.X+r.Rect.Width > old.Rect.X {
			y += dh
		}
		if x != r.Rect.X || y != r.Rect.Y {
			moves = append(moves, Move{r.Name, x, y})
		}
	}
	return moves
}

func Repack(name string, before []Output) error {
	after, err := List()
	if err != nil {
		return err
	}
	for _, m := range repackMoves(before, after, name) {
		if err := SetPosition(m.Name, m.X, m.Y); err != nil {
			return err
		}
	}
	return nil
}

func Overlap(a, b Output) bool {
	return a.Active && b.Active &&
		a.Rect.X < b.Rect.X+b.Rect.Width && b.Rect.X < a.Rect.X+a.Rect.Width &&
		a.Rect.Y < b.Rect.Y+b.Rect.Height && b.Rect.Y < a.Rect.Y+a.Rect.Height
}

func Touching(a, b Output) bool {
	if !a.Active || !b.Active || Overlap(a, b) {
		return false
	}
	vertical := a.Rect.Y < b.Rect.Y+b.Rect.Height && b.Rect.Y < a.Rect.Y+a.Rect.Height
	horizontal := a.Rect.X < b.Rect.X+b.Rect.Width && b.Rect.X < a.Rect.X+a.Rect.Width
	return vertical && (a.Rect.X+a.Rect.Width == b.Rect.X || b.Rect.X+b.Rect.Width == a.Rect.X) ||
		horizontal && (a.Rect.Y+a.Rect.Height == b.Rect.Y || b.Rect.Y+b.Rect.Height == a.Rect.Y)
}

func Problems(outs []Output) []string {
	var act []Output
	for _, o := range outs {
		if o.Active {
			act = append(act, o)
		}
	}
	var out []string
	for i, a := range act {
		touches := false
		for j, b := range act {
			if i == j {
				continue
			}
			if Overlap(a, b) {
				touches = true
				if i < j {
					out = append(out, fmt.Sprintf("%s and %s overlap", a.Name, b.Name))
				}
			}
			touches = touches || Touching(a, b)
		}
		if len(act) > 1 && !touches {
			out = append(out, a.Name+" does not touch another display")
		}
	}
	return out
}
