package monitor

import (
	"strings"
	"testing"
)

var outs = []Output{
	{Name: "eDP-1", Make: "BOE", Model: "0x0B7D", Serial: "Unknown", Active: true, Scale: 1, Transform: "normal",
		Rect: Rect{0, 0, 2560, 1440}, Mode: Mode{2560, 1440, 240000}},
	{Name: "DP-9", Make: "AOC", Model: "U34G2G1", Serial: "0x00000E06", Active: true, Scale: 1.25, Transform: "normal", Sync: "enabled",
		Rect: Rect{2560, 0, 2752, 1152}, Mode: Mode{3440, 1440, 99982}},
	{Name: "HDMI-A-1", Active: false, Mode: Mode{1920, 1080, 60000}},
}

func TestMode(t *testing.T) {
	if got := (Mode{3440, 1440, 99982}).String(); got != "3440x1440@99.98Hz" {
		t.Errorf("String = %q", got)
	}
	if got := (Mode{2560, 1440, 240000}).String(); got != "2560x1440@240Hz" {
		t.Errorf("String = %q", got)
	}
	if got := (Mode{3440, 1440, 99982}).arg(); got != "3440x1440@99.982Hz" {
		t.Errorf("arg = %q", got)
	}
}

func TestConfigs(t *testing.T) {
	sway := SwayConfig(outs)
	for _, want := range []string{
		"output eDP-1 mode 2560x1440@240.000Hz pos 0 0 scale 1 transform normal adaptive_sync off",
		"output DP-9 mode 3440x1440@99.982Hz pos 2560 0 scale 1.25 transform normal adaptive_sync on",
		"output HDMI-A-1 disable",
	} {
		if !strings.Contains(sway, want) {
			t.Errorf("sway config misses %q:\n%s", want, sway)
		}
	}
	kanshi := Kanshi(outs)
	for _, want := range []string{
		"profile {",
		`output "BOE 0x0B7D Unknown" enable mode 2560x1440@240.000Hz position 0,0 scale 1 transform normal adaptive_sync off`,
		`output "AOC U34G2G1 0x00000E06" enable`,
		`output "HDMI-A-1" disable`,
	} {
		if !strings.Contains(kanshi, want) {
			t.Errorf("kanshi config misses %q:\n%s", want, kanshi)
		}
	}
}

func TestPlace(t *testing.T) {
	edp, dp := outs[0], outs[1]
	for side, want := range map[Side][2]int{
		RightOf: {2560, 0}, LeftOf: {-2752, 0}, Above: {0, -1152}, Below: {0, 1440},
	} {
		if x, y := Place(dp, edp, side); [2]int{x, y} != want {
			t.Errorf("%v: %d,%d, want %v", side, x, y, want)
		}
	}
}

func TestApplyRejectsOddNames(t *testing.T) {
	for _, n := range []string{`a b`, `a"b`, `a;b`, ``} {
		if err := apply(n, "enable"); err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("apply(%q) = %v", n, err)
		}
	}
}
