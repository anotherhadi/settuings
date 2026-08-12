package network

import "strings"

// Device is one row of `nmcli device status`.
type Device struct {
	Name       string
	Type       string // e.g. "wifi", "ethernet", "wireguard", "loopback"
	State      string // e.g. "connected", "disconnected", "unavailable"
	Connection string // active connection profile name, empty if none
}

func (d Device) Connected() bool {
	return strings.HasPrefix(d.State, "connected")
}

// ListDevices returns every network device known to NetworkManager.
func ListDevices() ([]Device, error) {
	rows, err := runTerse("DEVICE,TYPE,STATE,CONNECTION", "device", "status")
	if err != nil {
		return nil, err
	}
	devices := make([]Device, 0, len(rows))
	for _, row := range rows {
		devices = append(devices, Device{
			Name:       field(row, 0),
			Type:       field(row, 1),
			State:      field(row, 2),
			Connection: field(row, 3),
		})
	}
	return devices, nil
}

// DeviceConnect asks NetworkManager to bring a device up using whichever
// saved connection profile fits it best. Unlike VPNUp (which reactivates a
// specific, already-known connection by name), this works even right after
// a disconnect, when `device status` no longer reports which connection
// profile was last active on that device.
func DeviceConnect(device string) error {
	_, err := run("device", "connect", device)
	return err
}

// DeviceIP4 returns the device's primary private IPv4 address without its
// CIDR suffix (e.g. "192.168.1.54"), or "" if the device has none.
func DeviceIP4(device string) (string, error) {
	out, err := getField("IP4.ADDRESS", "device", "show", device)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", nil
	}
	// nmcli -g can return multiple addresses newline-separated; keep the first.
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = out[:i]
	}
	if i := strings.IndexByte(out, '/'); i >= 0 {
		out = out[:i]
	}
	return out, nil
}

// DeviceDetail holds everything settuings shows on the "network info" screen
// for a connected device: addressing, gateway, DNS and hardware details.
type DeviceDetail struct {
	HWAddr   string
	MTU      string
	IPv4     string // address + CIDR, e.g. "192.168.1.54/24"
	Gateway4 string
	DNS4     []string
	IPv6     []string
	Gateway6 string
	DNS6     []string
}

// DeviceDetails fetches full addressing/hardware info for device via
// `nmcli device show`. Unlike the tabular nmcli commands elsewhere in this
// package, `device show` prints one "KEY:VALUE" line per property (with
// "KEY[n]" for multi-valued ones) and does not escape colons inside values,
// so it's parsed differently from runTerse.
func DeviceDetails(device string) (DeviceDetail, error) {
	out, err := run("-t", "-f",
		"GENERAL.HWADDR,GENERAL.MTU,IP4.ADDRESS,IP4.GATEWAY,IP4.DNS,IP6.ADDRESS,IP6.GATEWAY,IP6.DNS",
		"device", "show", device)
	if err != nil {
		return DeviceDetail{}, err
	}

	var d DeviceDetail
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if i := strings.IndexByte(key, '['); i >= 0 {
			key = key[:i] // strip the "[n]" multi-value index
		}
		switch key {
		case "GENERAL.HWADDR":
			d.HWAddr = value
		case "GENERAL.MTU":
			d.MTU = value
		case "IP4.ADDRESS":
			if d.IPv4 == "" {
				d.IPv4 = value
			}
		case "IP4.GATEWAY":
			d.Gateway4 = value
		case "IP4.DNS":
			d.DNS4 = append(d.DNS4, value)
		case "IP6.ADDRESS":
			d.IPv6 = append(d.IPv6, value)
		case "IP6.GATEWAY":
			d.Gateway6 = value
		case "IP6.DNS":
			d.DNS6 = append(d.DNS6, value)
		}
	}
	return d, nil
}
