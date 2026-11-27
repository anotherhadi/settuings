package audio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Device struct {
	ID      string
	Name    string
	Default bool
	Volume  int
	Muted   bool
}

func ListSinks() ([]Device, error) {
	return listDevices("Sinks:")
}

func ListSources() ([]Device, error) {
	return listDevices("Sources:")
}

func listDevices(header string) ([]Device, error) {
	out, err := status()
	if err != nil {
		return nil, err
	}
	entries := parseSection(out, header)
	devices := make([]Device, 0, len(entries))
	for _, e := range entries {
		devices = append(devices, Device{
			ID:      e.ID,
			Name:    e.Name,
			Default: e.Default,
			Volume:  e.Volume,
			Muted:   e.Muted,
		})
	}
	return devices, nil
}

var volumeLineRe = regexp.MustCompile(`^Volume:\s+([0-9.]+)(\s+\[MUTED\])?`)

func GetVolume(id string) (int, bool, error) {
	out, err := run("get-volume", id)
	if err != nil {
		return 0, false, err
	}
	m := volumeLineRe.FindStringSubmatch(strings.TrimSpace(out))
	if m == nil {
		return 0, false, fmt.Errorf("unexpected wpctl get-volume output: %q", out)
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false, err
	}
	return int(v*100 + 0.5), m[2] != "", nil
}

func SetVolume(id string, pct int) error {
	if pct < 0 {
		pct = 0
	} else if pct > 100 {
		pct = 100
	}
	_, err := run("set-volume", id, strconv.Itoa(pct)+"%")
	return err
}

func SetMute(id string, mute bool) error {
	state := "0"
	if mute {
		state = "1"
	}
	_, err := run("set-mute", id, state)
	return err
}

func ToggleMute(id string) error {
	_, err := run("set-mute", id, "toggle")
	return err
}

func SetDefault(id string) error {
	_, err := run("set-default", id)
	return err
}
