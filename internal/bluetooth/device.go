package bluetooth

import (
	"strconv"
	"strings"
	"time"
)

const (
	// actionTimeout bounds connect/pair calls, which can take longer than a
	// plain query while BlueZ negotiates with the remote device.
	actionTimeout = 30 * time.Second
	scanBuffer    = 10 * time.Second
)

// Device is a Bluetooth device known to BlueZ: either paired, or seen
// during a recent scan.
type Device struct {
	Address   string
	Name      string
	Paired    bool
	Bonded    bool
	Trusted   bool
	Connected bool
	Blocked   bool
	Icon      string // freedesktop icon name, e.g. "input-mouse"; empty if unknown
	Battery   int    // 0-100, -1 if unknown
}

// DeviceDetail extends Device with the extra fields shown on the "device
// info" screen.
type DeviceDetail struct {
	Device
	Modalias string
	UUIDs    []string // human-readable service labels
}

// parseDeviceLines parses the "Device <address> <name>" lines printed by
// `bluetoothctl devices` into address -> name.
func parseDeviceLines(out string) map[string]string {
	names := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		rest, ok := strings.CutPrefix(line, "Device ")
		if !ok {
			continue
		}
		addr, name, ok := strings.Cut(rest, " ")
		if !ok {
			continue
		}
		names[addr] = name
	}
	return names
}

func deviceSet(filter string) (map[string]bool, error) {
	args := []string{"devices"}
	if filter != "" {
		args = append(args, filter)
	}
	out, err := run(defaultTimeout, args...)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool)
	for addr := range parseDeviceLines(out) {
		set[addr] = true
	}
	return set, nil
}

// ListDevices returns every device BlueZ currently knows about: paired
// devices plus anything seen during a recent scan. Paired devices get an
// extra `info` lookup for icon/battery/blocked state; unpaired ones don't,
// to keep a large scan result snappy.
func ListDevices() ([]Device, error) {
	out, err := run(defaultTimeout, "devices")
	if err != nil {
		return nil, err
	}
	names := parseDeviceLines(out)

	paired, err := deviceSet("Paired")
	if err != nil {
		return nil, err
	}
	connected, err := deviceSet("Connected")
	if err != nil {
		return nil, err
	}
	trusted, err := deviceSet("Trusted")
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(names))
	for addr, name := range names {
		d := Device{
			Address:   addr,
			Name:      name,
			Paired:    paired[addr],
			Connected: connected[addr],
			Trusted:   trusted[addr],
			Battery:   -1,
		}
		if d.Paired {
			if detail, err := Info(addr); err == nil {
				d.Icon = detail.Icon
				d.Battery = detail.Battery
				d.Bonded = detail.Bonded
				d.Blocked = detail.Blocked
				if detail.Name != "" {
					d.Name = detail.Name
				}
			}
		}
		if d.Name == "" {
			d.Name = d.Address
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// Info fetches full detail for one device via `bluetoothctl info`.
func Info(address string) (DeviceDetail, error) {
	out, err := run(defaultTimeout, "info", address)
	if err != nil {
		return DeviceDetail{}, err
	}

	d := DeviceDetail{Device: Device{Address: address, Battery: -1}}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Device ") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "Name":
			d.Name = value
		case "Icon":
			d.Icon = value
		case "Paired":
			d.Paired = value == "yes"
		case "Bonded":
			d.Bonded = value == "yes"
		case "Trusted":
			d.Trusted = value == "yes"
		case "Blocked":
			d.Blocked = value == "yes"
		case "Connected":
			d.Connected = value == "yes"
		case "Modalias":
			d.Modalias = value
		case "UUID":
			if i := strings.Index(value, "("); i >= 0 {
				value = strings.TrimSpace(value[:i])
			}
			if value != "" {
				d.UUIDs = append(d.UUIDs, value)
			}
		case "Battery Percentage":
			if i, j := strings.Index(value, "("), strings.Index(value, ")"); i >= 0 && j > i {
				if n, err := strconv.Atoi(value[i+1 : j]); err == nil {
					d.Battery = n
				}
			}
		}
	}
	return d, nil
}

// Scan puts the controller into discovery mode for the given number of
// seconds, then returns. Newly seen devices show up in a subsequent
// ListDevices call.
func Scan(seconds int) error {
	_, err := run(time.Duration(seconds)*time.Second+scanBuffer, "--timeout", strconv.Itoa(seconds), "scan", "on")
	return err
}

// Pair initiates pairing with a device. Because non-interactive
// bluetoothctl has no agent to answer PIN/confirmation prompts, this only
// succeeds for "just works" devices (most mice, keyboards, headsets).
func Pair(address string) error {
	_, err := run(actionTimeout, "pair", address)
	return err
}

// Connect brings up all profiles for an already-paired (or just-paired)
// device.
func Connect(address string) error {
	_, err := run(actionTimeout, "connect", address)
	return err
}

// Disconnect disconnects a device without forgetting it.
func Disconnect(address string) error {
	_, err := run(defaultTimeout, "disconnect", address)
	return err
}

// SetTrusted marks a device as trusted (allowed to reconnect without
// confirmation) or not.
func SetTrusted(address string, trust bool) error {
	verb := "untrust"
	if trust {
		verb = "trust"
	}
	_, err := run(defaultTimeout, verb, address)
	return err
}

// Remove deletes a paired device's profile, or drops an unpaired device
// from BlueZ's scan cache.
func Remove(address string) error {
	_, err := run(defaultTimeout, "remove", address)
	return err
}
