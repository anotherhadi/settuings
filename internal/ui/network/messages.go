package network

import "github.com/anotherhadi/settuings/internal/network"

// checkMsg carries the result of network.CheckAvailable.
type checkMsg struct{ err error }

// tickMsg triggers a periodic refresh of every tab while this page is active.
type tickMsg struct{}

type wifiListMsg struct {
	networks       []network.WifiNetwork
	ip             string
	device         string // wifi interface name, e.g. "wlo1"
	connectionName string // active connection profile name, for secret lookup
	radioOn        bool
	err            error
}

type wifiActionMsg struct {
	err error
}

type ethernetListMsg struct {
	devices []network.EthernetDevice
	err     error
}

type ethernetActionMsg struct {
	err error
}

type vpnListMsg struct {
	conns []network.VPNConnection
	err   error
}

type vpnActionMsg struct {
	err error
}

type knownListMsg struct {
	networks []network.KnownNetwork
	err      error
}

type knownSecretMsg struct {
	name   string
	secret string
	err    error
}

type knownActionMsg struct {
	err error
}

type deviceDetailMsg struct {
	detail network.DeviceDetail
	err    error
}

type radioActionMsg struct {
	err error
}
