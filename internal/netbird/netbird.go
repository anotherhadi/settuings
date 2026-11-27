package netbird

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	Connected      = "Connected"
	Connecting     = "Connecting"
	Idle           = "Idle"
	NeedsLogin     = "NeedsLogin"
	LoginFailed    = "LoginFailed"
	SessionExpired = "SessionExpired"
)

var (
	ErrNotFound  = errors.New("netbird not found in PATH")
	ErrNoDaemon  = errors.New("the netbird daemon is not running")
	LoginCommand = "netbird up"
	StartCommand = "sudo systemctl start netbird"
)

type Peer struct {
	FQDN      string
	IP        string
	Status    string
	Type      string
	Latency   time.Duration
	Handshake time.Time
	Received  int64
	Sent      int64
}

func (p Peer) Connected() bool { return p.Status == Connected }

type State struct {
	Daemon     string
	Version    string
	Profile    string
	FQDN       string
	IP         string
	Management bool
	Signal     bool
	Peers      []Peer
}

func (s State) NeedsLogin() bool {
	return s.Daemon == NeedsLogin || s.Daemon == LoginFailed || s.Daemon == SessionExpired
}

func (s State) Up() bool { return s.Daemon == Connected || s.Daemon == Connecting }

type Network struct {
	ID       string
	Ranges   []string
	Domains  []string
	Selected bool
}

func (n Network) ExitNode() bool {
	for _, r := range n.Ranges {
		if r == "0.0.0.0/0" || r == "::/0" {
			return true
		}
	}
	return false
}

func Installed() bool {
	_, err := exec.LookPath("netbird")
	return err == nil
}

func run(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "netbird", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		if ctx.Err() != nil {
			msg = "timed out"
		}
		return "", classify(msg)
	}
	return stdout.String(), nil
}

func classify(msg string) error {
	msg = strings.TrimSpace(strings.TrimPrefix(msg, "Error: "))
	if strings.Contains(msg, "connect to daemon") || strings.Contains(msg, "no such file") || strings.Contains(msg, "connection refused") {
		return ErrNoDaemon
	}
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	return errors.New(msg)
}

type rawStatus struct {
	DaemonStatus  string `json:"daemonStatus"`
	DaemonVersion string `json:"daemonVersion"`
	ProfileName   string `json:"profileName"`
	FQDN          string `json:"fqdn"`
	IP            string `json:"netbirdIp"`
	Management    struct {
		Connected bool `json:"connected"`
	} `json:"management"`
	Signal struct {
		Connected bool `json:"connected"`
	} `json:"signal"`
	Peers struct {
		Details []struct {
			FQDN      string    `json:"fqdn"`
			IP        string    `json:"netbirdIp"`
			Status    string    `json:"status"`
			Type      string    `json:"connectionType"`
			Latency   int64     `json:"latency"`
			Handshake time.Time `json:"lastWireguardHandshake"`
			Received  int64     `json:"transferReceived"`
			Sent      int64     `json:"transferSent"`
		} `json:"details"`
	} `json:"peers"`
}

func Read() (State, error) {
	out, err := run(10*time.Second, "status", "--json")
	if err != nil {
		return State{}, err
	}
	return parseStatus([]byte(out))
}

func parseStatus(b []byte) (State, error) {
	var r rawStatus
	if err := json.Unmarshal(b, &r); err != nil {
		return State{}, fmt.Errorf("unexpected netbird status: %w", err)
	}
	st := State{Daemon: r.DaemonStatus, Version: r.DaemonVersion, Profile: r.ProfileName, FQDN: r.FQDN, IP: r.IP,
		Management: r.Management.Connected, Signal: r.Signal.Connected}
	for _, p := range r.Peers.Details {
		peer := Peer{FQDN: p.FQDN, IP: p.IP, Status: p.Status, Type: p.Type, Latency: time.Duration(p.Latency), Received: p.Received, Sent: p.Sent}
		if p.Handshake.Year() > 1 {
			peer.Handshake = p.Handshake
		}
		st.Peers = append(st.Peers, peer)
	}
	return st, nil
}

func Networks() ([]Network, error) {
	out, err := run(10*time.Second, "networks", "list")
	if err != nil {
		return nil, err
	}
	return parseNetworks(out), nil
}

func parseNetworks(out string) []Network {
	var nets []Network
	split := func(s string) []string {
		var l []string
		for _, f := range strings.Split(s, ",") {
			if f = strings.TrimSpace(f); f != "" {
				l = append(l, f)
			}
		}
		return l
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if id, ok := strings.CutPrefix(line, "- ID:"); ok {
			nets = append(nets, Network{ID: strings.TrimSpace(id)})
			continue
		}
		if len(nets) == 0 {
			continue
		}
		n := &nets[len(nets)-1]
		switch {
		case strings.HasPrefix(line, "Network:"):
			n.Ranges = split(strings.TrimPrefix(line, "Network:"))
		case strings.HasPrefix(line, "Domains:"):
			n.Domains = split(strings.TrimPrefix(line, "Domains:"))
		case strings.HasPrefix(line, "Status:"):
			n.Selected = strings.TrimSpace(strings.TrimPrefix(line, "Status:")) == "Selected"
		}
	}
	return nets
}

func Up() error {
	_, err := run(60*time.Second, "up", "--no-browser")
	return err
}

func Down() error {
	_, err := run(20*time.Second, "down")
	return err
}

func Select(replace bool, ids ...string) error {
	return selection("select", replace, ids)
}

func Deselect(ids ...string) error { return selection("deselect", false, ids) }

func selection(verb string, replace bool, ids []string) error {
	if len(ids) == 0 {
		return errors.New("no network")
	}
	args := []string{"networks", verb}
	if verb == "select" && !replace {
		args = append(args, "-a")
	}
	for _, id := range ids {
		if id == "" || strings.HasPrefix(id, "-") {
			return fmt.Errorf("unsupported network id %q", id)
		}
	}
	_, err := run(15*time.Second, append(args, ids...)...)
	return err
}

type plan struct {
	verb    string
	replace bool
	ids     []string
}

func togglePlan(id string, nets []Network) (plan, bool) {
	var target *Network
	for i := range nets {
		if nets[i].ID == id {
			target = &nets[i]
		}
	}
	if target == nil {
		return plan{}, false
	}
	switch {
	case target.Selected:
		return plan{verb: "deselect", ids: []string{id}}, true
	case target.ExitNode():
		ids := []string{id}
		for _, n := range nets {
			if n.Selected && !n.ExitNode() {
				ids = append(ids, n.ID)
			}
		}
		return plan{verb: "select", replace: true, ids: ids}, true
	}
	return plan{verb: "select", ids: []string{id}}, true
}

func Toggle(id string, nets []Network) error {
	p, ok := togglePlan(id, nets)
	if !ok {
		return fmt.Errorf("unknown network %q", id)
	}
	if p.verb == "deselect" {
		return Deselect(p.ids...)
	}
	return Select(p.replace, p.ids...)
}
