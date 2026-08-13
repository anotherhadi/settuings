package icons

import "github.com/anotherhadi/settuings/internal/config"

type Icons struct {
	Edit   string
	Reload string
	Add    string

	About     string
	Network   string
	Bluetooth string
	Audio     string
	Power     string
	Inputs    string

	Wifi     string
	Ethernet string
	VPN      string
	Check    string
	Lock     string

	Output  string
	Input   string
	Apps    string
	Battery string
}

var I *Icons

func Init(cfg *config.Config) {
	if cfg.TUI.UseNerdfontIcons {
		I = &Icons{
			Edit:   "󰏫 ",
			Reload: "󰑙 ",
			Add:    "󰐕 ",

			About:     " ",
			Network:   "󰛳 ",
			Bluetooth: "󰂯 ",
			Audio:     "󰓃 ",
			Power:     "󰐥 ",
			Inputs:    "󰌌 ",

			Wifi:     "󰤨 ",
			Ethernet: "󰈀 ",
			VPN:      "󰖂 ",
			Check:    "󰄬 ",
			Lock:     "󰌾 ",

			Output:  "󰕾 ",
			Input:   "󰍬 ",
			Apps:    "󰀻 ",
			Battery: "󰁹 ",
		}
	} else {
		I = &Icons{}
	}
}
