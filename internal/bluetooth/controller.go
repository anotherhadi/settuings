package bluetooth

import "strings"

type ControllerInfo struct {
	Address      string
	Name         string
	Powered      bool
	Discoverable bool
	Pairable     bool
	Discovering  bool
}

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

func SetPower(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := run(defaultTimeout, "power", state)
	return err
}
