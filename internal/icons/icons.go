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

	Wifi     string
	Ethernet string
	VPN      string
	Check    string
	Lock     string

	Output string
	Input  string
	Apps   string
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

			Wifi:     "󰤨 ",
			Ethernet: "󰈀 ",
			VPN:      "󰖂 ",
			Check:    "󰄬 ",
			Lock:     "󰌾 ",

			Output: "󰕾 ",
			Input:  "󰍬 ",
			Apps:   "󰀻 ",
		}
	} else {
		I = &Icons{}
	}
}
