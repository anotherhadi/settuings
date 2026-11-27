package network

import "strings"

type Device struct {
	Name       string
	Type       string
	State      string
	Connection string
}

func (d Device) Connected() bool {
	return strings.HasPrefix(d.State, "connected")
}

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

func DeviceConnect(device string) error {
	_, err := run("device", "connect", device)
	return err
}

func DeviceIP4(device string) (string, error) {
	out, err := getField("IP4.ADDRESS", "device", "show", device)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", nil
	}

	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = out[:i]
	}
	if i := strings.IndexByte(out, '/'); i >= 0 {
		out = out[:i]
	}
	return out, nil
}

type DeviceDetail struct {
	HWAddr   string
	MTU      string
	IPv4     string
	Gateway4 string
	DNS4     []string
	IPv6     []string
	Gateway6 string
	DNS6     []string
}

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
			key = key[:i]
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
