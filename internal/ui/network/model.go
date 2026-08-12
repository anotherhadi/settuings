// Package network provides the "Network" TUI page: a thin, tabbed wrapper
// around nmcli covering Wi-Fi, Ethernet, VPN, and known Wi-Fi networks.
package network

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
	"github.com/anotherhadi/settuings/internal/icons"
)

type tab int

const (
	tabWifi tab = iota
	tabEthernet
	tabVPN
	tabKnown // hidden from the tab bar; reachable only via its own shortcut
)

// visibleTabs are the tabs shown in the tab bar and reachable by cycling.
var visibleTabs = []tab{tabWifi, tabEthernet, tabVPN}

func (t tab) label() string {
	switch t {
	case tabWifi:
		return "Wifi"
	case tabEthernet:
		return "Ethernet"
	case tabVPN:
		return "VPN"
	case tabKnown:
		return "Known"
	default:
		return ""
	}
}

func (t tab) icon() string {
	switch t {
	case tabWifi:
		return icons.I.Wifi
	case tabEthernet:
		return icons.I.Ethernet
	case tabVPN:
		return icons.I.VPN
	case tabKnown:
		return icons.I.Lock
	default:
		return ""
	}
}

type Model struct {
	active      tab
	previousTab tab // where to return to when leaving the hidden Known tab
	width       int
	height      int
	help        help.Model

	checkErr error // reason nmcli/NetworkManager is unusable; nil once confirmed OK
	checked  bool

	wifi         wifiState
	ethernet     ethernetState
	vpn          vpnState
	known        knownState
	deviceDetail deviceDetailState
}

func New() Model {
	return Model{
		active:   tabWifi,
		help:     ilovetui.NewHelp(),
		wifi:     newWifiState(),
		ethernet: newEthernetState(),
		vpn:      newVPNState(),
		known:    newKnownState(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Activate is called every time the user navigates onto this page. It
// re-checks nmcli availability, kicks off a fetch for every tab, and starts
// the periodic auto-refresh loop.
func (m *Model) Activate() tea.Cmd {
	return tea.Batch(
		checkCmd(),
		fetchWifiCmd(),
		fetchEthernetCmd(),
		fetchVPNCmd(),
		fetchKnownCmd(),
		tickCmd(),
	)
}

func (m Model) IsEditing() bool {
	return m.wifi.connecting || m.known.confirmForget
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.help.SetWidth(w - 2)
	m.resizeDetails()
}
