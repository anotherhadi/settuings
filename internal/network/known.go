package network

import "strings"

type KnownNetwork struct {
	Name        string
	Security    string
	AutoConnect bool
	Active      bool
}

func ListKnownWifi() ([]KnownNetwork, error) {
	rows, err := runTerse("NAME,TYPE,AUTOCONNECT,ACTIVE", "connection", "show")
	if err != nil {
		return nil, err
	}

	known := make([]KnownNetwork, 0)
	for _, row := range rows {
		if field(row, 1) != "802-11-wireless" {
			continue
		}
		name := field(row, 0)
		security, _ := getField("802-11-wireless-security.key-mgmt", "connection", "show", "id", name)
		known = append(known, KnownNetwork{
			Name:        name,
			Security:    security,
			AutoConnect: field(row, 2) == "yes",
			Active:      field(row, 3) == "yes",
		})
	}
	return known, nil
}

func Secret(name string) (string, error) {
	out, err := run("--show-secrets", "-g", "802-11-wireless-security.psk", "connection", "show", "id", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func SetAutoConnect(name string, enabled bool) error {
	value := "no"
	if enabled {
		value = "yes"
	}
	_, err := run("connection", "modify", "id", name, "connection.autoconnect", value)
	return err
}

func Forget(name string) error {
	_, err := run("connection", "delete", "id", name)
	return err
}
