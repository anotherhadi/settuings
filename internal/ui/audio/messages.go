package audio

import "github.com/anotherhadi/settuings/internal/audio"

// checkMsg carries the result of audio.CheckAvailable.
type checkMsg struct{ err error }

// tickMsg triggers a periodic refresh of every tab while this page is active.
type tickMsg struct{}

type outputListMsg struct {
	devices []audio.Device
	err     error
}

type inputListMsg struct {
	devices []audio.Device
	err     error
}

type appsListMsg struct {
	streams []audio.Stream
	err     error
}

// outputActionMsg carries the result of a control action (set volume, mute,
// set default, play test tone) performed on the Output tab.
type outputActionMsg struct{ err error }

// inputActionMsg is the Input tab's equivalent of outputActionMsg.
type inputActionMsg struct{ err error }

// appsActionMsg is the Apps tab's equivalent of outputActionMsg.
type appsActionMsg struct{ err error }

// levelStartedMsg carries the result of starting a live input-level meter.
type levelStartedMsg struct {
	monitor *audio.LevelMonitor
	err     error
}

// levelMsg carries one peak reading from the live input-level meter.
type levelMsg struct {
	peak float64
	err  error
}
