package bluetooth

import "github.com/anotherhadi/settuings/internal/bluetooth"

// checkMsg carries the result of bluetooth.CheckAvailable.
type checkMsg struct{ err error }

// tickMsg triggers a periodic refresh of the controller + device list while
// this page is active.
type tickMsg struct{}

type controllerMsg struct {
	info bluetooth.ControllerInfo
	err  error
}

type deviceListMsg struct {
	devices []bluetooth.Device
	err     error
}

type deviceActionMsg struct{ err error }

type powerActionMsg struct{ err error }

type scanDoneMsg struct{ err error }

type detailMsg struct {
	detail bluetooth.DeviceDetail
	err    error
}
