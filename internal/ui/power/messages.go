package power

import "github.com/anotherhadi/settuings/internal/power"

// checkMsg carries the result of power.CheckAvailable (powerprofilesctl).
type checkMsg struct{ err error }

// tickMsg triggers a periodic refresh of profiles + battery while this page
// is active.
type tickMsg struct{}

type profilesMsg struct {
	profiles []power.Profile
	err      error
}

type setProfileMsg struct{ err error }

type batteryMsg struct {
	batteries []power.Battery
	ac        power.ACStatus
	err       error
}
