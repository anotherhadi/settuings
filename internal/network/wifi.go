package network

import (
	"errors"
	"strconv"
)

var ErrNoWifiDevice = errors.New("no Wi-Fi device found")

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

type WifiNetwork struct {
	Active   bool
	SSID     string
	BSSID    string
	Signal   int
	Security string
}

func (w WifiNetwork) Secured() bool {
	return w.Security != "" && w.Security != "--"
}

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
			continue
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

func Rescan() error {
	_, err := run("device", "wifi", "rescan")
	return err
}

func ConnectWifi(ssid, password string) error {
	args := []string{"device", "wifi", "connect", ssid}
	if password != "" {
		args = append(args, "password", password)
	}
	_, err := run(args...)
	return err
}

func DisconnectWifi(device string) error {
	_, err := run("device", "disconnect", device)
	return err
}

func SetWifiRadio(on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	_, err := run("radio", "wifi", state)
	return err
}

func RadioEnabled() (bool, error) {
	out, err := getField("WIFI", "general")
	if err != nil {
		return false, err
	}
	return out == "enabled", nil
}

func ConnectKnown(name string) error {
	_, err := run("connection", "up", "id", name)
	return err
}

var ErrEmptyPassword = errors.New("a password is needed")

func ConnectKnownWithSecret(name, password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	if _, err := run("connection", "modify", "id", name, "wifi-sec.psk-flags", "0", "wifi-sec.psk", password); err != nil {
		return err
	}
	return ConnectKnown(name)
}
