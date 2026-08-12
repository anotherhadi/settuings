package network

import (
	"errors"
	"strconv"
)

var ErrNoWifiDevice = errors.New("no Wi-Fi device found")

// WifiDevice returns the first Wi-Fi radio device, if any.
func WifiDevice() (Device, error) {
	devices, err := ListDevices()
	if err != nil {
		return Device{}, err
	}
	for _, d := range devices {
		if d.Type == "wifi" {
			return d, nil
		}
	}
	return Device{}, ErrNoWifiDevice
}

// WifiNetwork is one access point seen by `nmcli device wifi list`.
type WifiNetwork struct {
	Active   bool // this is the AP we're currently associated with
	SSID     string
	BSSID    string
	Signal   int // 0-100
	Security string
}

func (w WifiNetwork) Secured() bool {
	return w.Security != "" && w.Security != "--"
}

// ListWifi returns visible access points, deduplicated by SSID (keeping the
// strongest signal seen for each network), sorted strongest-first.
func ListWifi() ([]WifiNetwork, error) {
	rows, err := runTerse("IN-USE,SSID,BSSID,SIGNAL,SECURITY", "device", "wifi", "list")
	if err != nil {
		return nil, err
	}

	bySSID := make(map[string]WifiNetwork)
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		ssid := field(row, 1)
		if ssid == "" {
			continue // hidden network with no broadcast SSID
		}
		signal, _ := strconv.Atoi(field(row, 3))
		net := WifiNetwork{
			Active:   field(row, 0) == "*",
			SSID:     ssid,
			BSSID:    field(row, 2),
			Signal:   signal,
			Security: field(row, 4),
		}
		existing, ok := bySSID[ssid]
		if !ok || net.Active || net.Signal > existing.Signal {
			bySSID[ssid] = net
		}
		if !ok {
			order = append(order, ssid)
		}
	}

	networks := make([]WifiNetwork, 0, len(order))
	for _, ssid := range order {
		networks = append(networks, bySSID[ssid])
	}
	return networks, nil
}

// Rescan asks NetworkManager to trigger a fresh Wi-Fi scan. It returns once
// the scan is requested, not once it completes; call ListWifi shortly after.
func Rescan() error {
	_, err := run("device", "wifi", "rescan")
	return err
}

// ConnectWifi joins ssid, creating a new saved profile if none exists yet.
// Pass an empty password for open networks or when a saved profile already
// has one.
func ConnectWifi(ssid, password string) error {
	args := []string{"device", "wifi", "connect", ssid}
	if password != "" {
		args = append(args, "password", password)
	}
	_, err := run(args...)
	return err
}

// DisconnectWifi disconnects the given Wi-Fi device without forgetting it.
func DisconnectWifi(device string) error {
	_, err := run("device", "disconnect", device)
	return err
}

// SetWifiRadio turns the Wi-Fi radio on or off entirely.
func SetWifiRadio(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := run("radio", "wifi", state)
	return err
}

// RadioEnabled reports whether the Wi-Fi radio is currently switched on.
func RadioEnabled() (bool, error) {
	out, err := getField("WIFI", "general")
	if err != nil {
		return false, err
	}
	return out == "enabled", nil
}
