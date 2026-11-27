package pages

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestWifiPayload(t *testing.T) {
	cases := []struct{ ssid, sec, pw, want string }{
		{"home", "wpa-psk", "secret", "WIFI:T:WPA;S:home;P:secret;;"},
		{"home", "WPA2", "secret", "WIFI:T:WPA;S:home;P:secret;;"},
		{"home", "sae", "secret", "WIFI:T:SAE;S:home;P:secret;;"},
		{"cafe", "", "", "WIFI:T:nopass;S:cafe;;"},
		{`a;b:c`, "wpa-psk", `p"w\`, `WIFI:T:WPA;S:a\;b\:c;P:p\"w\\;;`},
	}
	for _, c := range cases {
		if got := wifiPayload(c.ssid, c.sec, c.pw); got != c.want {
			t.Errorf("wifiPayload(%q, %q) = %q, want %q", c.ssid, c.sec, got, c.want)
		}
	}
}

func TestQRFits(t *testing.T) {
	q := newQR("home", wifiPayload("home", "wpa-psk", "secret"))
	for _, size := range [][2]int{{200, 60}, {70, 34}, {30, 10}, {10, 4}} {
		v := q.View(size[0], size[1])
		for _, l := range strings.Split(v, "\n") {
			if w := ansi.StringWidth(l); w > size[0] {
				t.Fatalf("%v: line is %d wide", size, w)
			}
		}
		if h := strings.Count(v, "\n") + 1; h > size[1] {
			t.Fatalf("%v: %d lines", size, h)
		}
	}
}
