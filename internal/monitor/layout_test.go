package monitor

import "testing"

func out(name string, x, y, w, h int) Output {
	return Output{Name: name, Active: true, Rect: Rect{x, y, w, h}}
}

func TestRepackMoves(t *testing.T) {
	// a laptop at 0,0, a monitor to its right, and a third below the laptop
	before := []Output{out("A", 0, 0, 2560, 1440), out("B", 2560, 0, 3440, 1440), out("C", 0, 1440, 1920, 1080)}
	moves := func(after []Output) map[string][2]int {
		m := map[string][2]int{}
		for _, mv := range repackMoves(before, after, "A") {
			m[mv.Name] = [2]int{mv.X, mv.Y}
		}
		return m
	}

	// A scaled to 2: smaller. B and C follow.
	got := moves([]Output{out("A", 0, 0, 1280, 720), out("B", 2560, 0, 3440, 1440), out("C", 0, 1440, 1920, 1080)})
	if len(got) != 2 || got["B"] != [2]int{1280, 0} || got["C"] != [2]int{0, 720} {
		t.Fatalf("shrink: %v", got)
	}
	// A scaled down to 0.75: bigger, B and C are pushed.
	got = moves([]Output{out("A", 0, 0, 3413, 1920), out("B", 2560, 0, 3440, 1440), out("C", 0, 1440, 1920, 1080)})
	if got["B"] != [2]int{3413, 0} || got["C"] != [2]int{0, 1920} {
		t.Fatalf("grow: %v", got)
	}
	// nothing changed, or the output is off: nothing to move
	if got = moves(before); len(got) != 0 {
		t.Fatalf("no change: %v", got)
	}
	off := []Output{{Name: "A"}, out("B", 2560, 0, 3440, 1440)}
	if len(repackMoves(before, off, "A")) != 0 {
		t.Fatal("moved around an output that is off")
	}
	// the changed output is the one on the right: the one on its left stays
	before = []Output{out("A", 0, 0, 2560, 1440), out("B", 2560, 0, 3440, 1440)}
	if got := repackMoves(before, []Output{out("A", 0, 0, 2560, 1440), out("B", 2560, 0, 2752, 1152)}, "B"); len(got) != 0 {
		t.Fatalf("a neighbour on the left moved: %v", got)
	}
	// a chain: C is right of B, which is right of A
	before = []Output{out("A", 0, 0, 100, 100), out("B", 100, 0, 100, 100), out("C", 200, 0, 100, 100)}
	got = map[string][2]int{}
	for _, mv := range repackMoves(before, []Output{out("A", 0, 0, 50, 100), out("B", 100, 0, 100, 100), out("C", 200, 0, 100, 100)}, "A") {
		got[mv.Name] = [2]int{mv.X, mv.Y}
	}
	if got["B"] != [2]int{50, 0} || got["C"] != [2]int{150, 0} {
		t.Fatalf("chain: %v", got)
	}
}

func TestProblems(t *testing.T) {
	if p := Problems([]Output{out("A", 0, 0, 100, 100), out("B", 100, 0, 100, 100)}); len(p) != 0 {
		t.Fatalf("touching displays: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100), out("B", 100, 50, 100, 100)}); len(p) != 0 {
		t.Fatalf("offset but touching: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100), out("B", 50, 0, 100, 100)}); len(p) != 1 || p[0] != "A and B overlap" {
		t.Fatalf("overlap: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100), out("B", 150, 0, 100, 100)}); len(p) != 2 {
		t.Fatalf("a gap: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100), out("B", 100, 100, 100, 100)}); len(p) != 2 {
		t.Fatalf("corners only: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100)}); len(p) != 0 {
		t.Fatalf("a single display: %v", p)
	}
	if p := Problems([]Output{out("A", 0, 0, 100, 100), {Name: "B"}}); len(p) != 0 {
		t.Fatalf("an off display: %v", p)
	}
}
