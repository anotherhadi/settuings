package network

type EthernetDevice struct {
	Device     string
	Connection string
	State      string
	IP4        string
}

func (e EthernetDevice) Connected() bool {
	return e.State == "connected"
}

func ListEthernet() ([]EthernetDevice, error) {
	devices, err := ListDevices()
	if err != nil {
		return nil, err
	}

	eth := make([]EthernetDevice, 0)
	for _, d := range devices {
		if d.Type != "ethernet" {
			continue
		}
		e := EthernetDevice{
			Device:     d.Name,
			Connection: d.Connection,
			State:      d.State,
		}
		if e.Connected() {
			e.IP4, _ = DeviceIP4(d.Name)
		}
		eth = append(eth, e)
	}
	return eth, nil
}

func EthernetDown(device string) error {
	_, err := run("device", "disconnect", device)
	return err
}
