package network

type VPNConnection struct {
	Name   string
	Type   string
	Active bool
	Device string
}

func ListVPN() ([]VPNConnection, error) {
	rows, err := runTerse("NAME,TYPE,ACTIVE,DEVICE", "connection", "show")
	if err != nil {
		return nil, err
	}

	vpns := make([]VPNConnection, 0)
	for _, row := range rows {
		typ := field(row, 1)
		if typ != "vpn" && typ != "wireguard" {
			continue
		}
		vpns = append(vpns, VPNConnection{
			Name:   field(row, 0),
			Type:   typ,
			Active: field(row, 2) == "yes",
			Device: field(row, 3),
		})
	}
	return vpns, nil
}

func VPNUp(name string) error {
	_, err := run("connection", "up", "id", name)
	return err
}

func VPNDown(name string) error {
	_, err := run("connection", "down", "id", name)
	return err
}
