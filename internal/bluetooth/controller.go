package bluetooth

import "strings"

// ControllerInfo is the state of the default Bluetooth adapter, as reported
// by `bluetoothctl show`.
type ControllerInfo struct {
	Address      string
	Name         string
	Powered      bool
	Discoverable bool
	Pairable     bool
	Discovering  bool
}

// ShowController returns the state of the default Bluetooth controller.
func ShowController() (ControllerInfo, error) {
	out, err := run(defaultTimeout, "show")
	if err != nil {
		return ControllerInfo{}, err
	}

	var c ControllerInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if rest, ok := strings.CutPrefix(line, "Controller "); ok {
			if fields := strings.Fields(rest); len(fields) > 0 {
				c.Address = fields[0]
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "Name":
			c.Name = value
		case "Powered":
			c.Powered = value == "yes"
		case "Discoverable":
			c.Discoverable = value == "yes"
		case "Pairable":
			c.Pairable = value == "yes"
		case "Discovering":
			c.Discovering = value == "yes"
		}
	}
	return c, nil
}

// SetPower turns the Bluetooth controller on or off entirely.
func SetPower(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := run(defaultTimeout, "power", state)
	return err
}
