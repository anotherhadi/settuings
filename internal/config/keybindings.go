package config

type GlobalKeys struct {
	Quit          string `mapstructure:"quit"`
	Escape        string `mapstructure:"escape"`
	OpenLogs      string `mapstructure:"open_logs"`
	ToggleSidebar string `mapstructure:"toggle_sidebar"`
	Help          string `mapstructure:"help"`
	Up            string `mapstructure:"up"`
	Down          string `mapstructure:"down"`
	Left          string `mapstructure:"left"`
	Right         string `mapstructure:"right"`
	CycleFocus    string `mapstructure:"cycle_focus"`
	ScrollUp      string `mapstructure:"scroll_up"`
	ScrollDown    string `mapstructure:"scroll_down"`
	GotoTop       string `mapstructure:"goto_top"`
	GotoBottom    string `mapstructure:"goto_bottom"`
	PrevPage      string `mapstructure:"prev_page"`
	NextPage      string `mapstructure:"next_page"`
}

type DocsKeys struct {
	Search      string `mapstructure:"search"`
	SearchReset string `mapstructure:"search_reset"`
	SearchNext  string `mapstructure:"search_next"`
	SearchPrev  string `mapstructure:"search_prev"`
}

type NetworkKeys struct {
	Select       string `mapstructure:"select"`
	Refresh      string `mapstructure:"refresh"`
	Forget       string `mapstructure:"forget"`
	ToggleAuto   string `mapstructure:"toggle_auto"`
	RevealSecret string `mapstructure:"reveal_secret"`
	Known        string `mapstructure:"known"`
	Info         string `mapstructure:"info"`
	Disconnect   string `mapstructure:"disconnect"`
	ToggleRadio  string `mapstructure:"toggle_radio"`
}

type BluetoothKeys struct {
	Select      string `mapstructure:"select"`
	Refresh     string `mapstructure:"refresh"`
	Forget      string `mapstructure:"forget"`
	ToggleTrust string `mapstructure:"toggle_trust"`
	Info        string `mapstructure:"info"`
	TogglePower string `mapstructure:"toggle_power"`
}

type AudioKeys struct {
	Select  string `mapstructure:"select"`
	Refresh string `mapstructure:"refresh"`
	Mute    string `mapstructure:"mute"`
	Test    string `mapstructure:"test"`
}

type Keybindings struct {
	Global    GlobalKeys    `mapstructure:"global"`
	Docs      DocsKeys      `mapstructure:"docs"`
	Network   NetworkKeys   `mapstructure:"network"`
	Bluetooth BluetoothKeys `mapstructure:"bluetooth"`
	Audio     AudioKeys     `mapstructure:"audio"`
}
