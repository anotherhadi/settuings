package input

// checkMsg carries the result of input.CheckAvailable (hyprctl).
type checkMsg struct{ err error }

// tickMsg triggers a periodic refresh of layout + sensitivity while this
// page is active, so it reflects layout/sensitivity changes made elsewhere
// (hyprland.conf reload, another tool).
type tickMsg struct{}

type layoutMsg struct {
	layout string
	err    error
}

type setLayoutMsg struct{ err error }

type sensitivityMsg struct {
	value float64
	err   error
}

type setSensitivityMsg struct{ err error }

// resetMsg carries the result of resetting both the keyboard layout and
// mouse sensitivity to their hyprland.conf values.
type resetMsg struct{ err error }
