package netbird

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const networksOut = `Available Networks:

  - ID: ch
    Network: 0.0.0.0/0, ::/0
    Status: Not Selected

  - ID: home
    Network: 0.0.0.0/0, ::/0
    Status: Selected

  - ID: lab
    Network: 10.1.0.0/16
    Status: Selected

  - ID: dns
    Domains: a.example.com, b.example.com
    Resolved IPs:
      a.example.com: 1.2.3.4
    Status: Not Selected
`

func TestParseNetworks(t *testing.T) {
	nets := parseNetworks(networksOut)
	if len(nets) != 4 {
		t.Fatalf("%d networks: %+v", len(nets), nets)
	}
	want := []struct {
		id   string
		exit bool
		sel  bool
	}{{"ch", true, false}, {"home", true, true}, {"lab", false, true}, {"dns", false, false}}
	for i, w := range want {
		if n := nets[i]; n.ID != w.id || n.ExitNode() != w.exit || n.Selected != w.sel {
			t.Errorf("%d: %+v, want %+v", i, n, w)
		}
	}
	if d := nets[3].Domains; len(d) != 2 || d[1] != "b.example.com" {
		t.Errorf("domains = %v", d)
	}
	if got := parseNetworks("No networks available.\n"); len(got) != 0 {
		t.Errorf("empty list: %+v", got)
	}
}

func TestParseStatus(t *testing.T) {
	st, err := parseStatus([]byte(`{"daemonStatus":"Connected","daemonVersion":"v0.71.4","profileName":"default",
	"fqdn":"h-laptop.nb","netbirdIp":"100.126.115.18/16","management":{"connected":true},"signal":{"connected":false},
	"peers":{"details":[
	 {"fqdn":"a.nb","netbirdIp":"100.1.1.1","status":"Connected","connectionType":"P2P","latency":5108168,"lastWireguardHandshake":"2026-11-27T21:50:32+01:00","transferReceived":10,"transferSent":20},
	 {"fqdn":"b.nb","netbirdIp":"100.1.1.2","status":"Idle","connectionType":"-","lastWireguardHandshake":"0001-01-01T00:00:00Z"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	if !st.Up() || st.NeedsLogin() || !st.Management || st.Signal || st.IP != "100.126.115.18/16" || len(st.Peers) != 2 {
		t.Fatalf("state = %+v", st)
	}
	if p := st.Peers[0]; !p.Connected() || p.Latency != 5108168*time.Nanosecond || p.Handshake.IsZero() || p.Sent != 20 {
		t.Fatalf("peer = %+v", p)
	}
	if !st.Peers[1].Handshake.IsZero() {
		t.Fatal("a zero handshake should stay zero")
	}
	for status, login := range map[string]bool{NeedsLogin: true, LoginFailed: true, SessionExpired: true, Idle: false, Connected: false} {
		if (State{Daemon: status}).NeedsLogin() != login {
			t.Errorf("NeedsLogin(%s) != %v", status, login)
		}
	}
	if _, err := parseStatus([]byte("not json")); err == nil {
		t.Fatal("bad json accepted")
	}
}

func TestClassify(t *testing.T) {
	if err := classify("Error: failed to connect to daemon error: dial unix /var/run/netbird/sock: connect: no such file or directory"); !errors.Is(err, ErrNoDaemon) {
		t.Errorf("daemon down = %v", err)
	}
	if err := classify("Error: something else\nusage..."); err == nil || err.Error() != "something else" {
		t.Errorf("other = %v", err)
	}
}

func TestSelectionRejectsOddIDs(t *testing.T) {
	for _, id := range []string{"", "-a", "--all"} {
		if err := Select(true, id); err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("Select(%q) = %v", id, err)
		}
	}
	if err := Deselect(); err == nil {
		t.Error("Deselect() with nothing accepted")
	}
}

func TestTogglePlan(t *testing.T) {
	nets := parseNetworks(networksOut)
	cases := []struct {
		id      string
		verb    string
		replace bool
		ids     string
	}{
		{"ch", "select", true, "ch,lab"},
		{"home", "deselect", false, "home"},
		{"lab", "deselect", false, "lab"},
		{"dns", "select", false, "dns"},
	}
	for _, c := range cases {
		p, ok := togglePlan(c.id, nets)
		if !ok || p.verb != c.verb || p.replace != c.replace || strings.Join(p.ids, ",") != c.ids {
			t.Errorf("%s: %+v", c.id, p)
		}
	}
	if _, ok := togglePlan("nope", nets); ok {
		t.Error("unknown network planned")
	}
}
