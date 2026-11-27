package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("sway is not running (swaymsg and SWAYSOCK needed)")

type Mode struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	Refresh int `json:"refresh"`
}

func (m Mode) Hz() float64 { return float64(m.Refresh) / 1000 }

func (m Mode) String() string {
	return fmt.Sprintf("%dx%d@%sHz", m.Width, m.Height, trimHz(m.Hz()))
}

func (m Mode) arg() string { return fmt.Sprintf("%dx%d@%.3fHz", m.Width, m.Height, m.Hz()) }

func trimHz(hz float64) string {
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", hz), "0"), ".")
}

type Rect struct{ X, Y, Width, Height int }

type Output struct {
	Name      string  `json:"name"`
	Make      string  `json:"make"`
	Model     string  `json:"model"`
	Serial    string  `json:"serial"`
	Active    bool    `json:"active"`
	Scale     float64 `json:"scale"`
	Transform string  `json:"transform"`
	Sync      string  `json:"adaptive_sync_status"`
	Workspace string  `json:"current_workspace"`
	Focused   bool    `json:"focused"`
	Rect      Rect    `json:"rect"`
	Mode      Mode    `json:"current_mode"`
	Modes     []Mode  `json:"modes"`
}

func (o Output) Label() string {
	if s := strings.TrimSpace(o.Make + " " + o.Model); s != "" {
		return s
	}
	return o.Name
}

func (o Output) SyncOn() bool { return o.Sync == "enabled" }

func (o Output) Logical() (w, h int) {
	if o.Active {
		return o.Rect.Width, o.Rect.Height
	}
	return o.Mode.Width, o.Mode.Height
}

func Available() error {
	if _, err := exec.LookPath("swaymsg"); err != nil || os.Getenv("SWAYSOCK") == "" {
		return ErrUnavailable
	}
	return nil
}

func swaymsg(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "swaymsg", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("swaymsg: %s", msg)
	}
	return stdout.Bytes(), nil
}

func List() ([]Output, error) {
	out, err := swaymsg("-r", "-t", "get_outputs")
	if err != nil {
		return nil, err
	}
	var outs []Output
	if err := json.Unmarshal(out, &outs); err != nil {
		return nil, err
	}
	return outs, nil
}

var safeName = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)

func apply(name string, args ...string) error {
	if !safeName.MatchString(name) {
		return fmt.Errorf("unsupported output name %q", name)
	}
	out, err := swaymsg(append([]string{"-r", "--", "output", name}, args...)...)
	if err != nil {
		return err
	}
	var res []struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(out, &res) == nil {
		for _, r := range res {
			if !r.Success {
				return errors.New(r.Error)
			}
		}
	}
	return nil
}

func SetEnabled(name string, on bool) error {
	if on {
		return apply(name, "enable")
	}
	return apply(name, "disable")
}

func SetMode(name string, m Mode) error { return apply(name, "mode", m.arg()) }

func SetScale(name string, scale float64) error {
	return apply(name, "scale", fmt.Sprintf("%.2f", scale))
}

func SetTransform(name, transform string) error { return apply(name, "transform", transform) }

func SetPosition(name string, x, y int) error {
	return apply(name, "pos", fmt.Sprint(x), fmt.Sprint(y))
}

func SetSync(name string, on bool) error {
	return apply(name, "adaptive_sync", map[bool]string{true: "on", false: "off"}[on])
}

func Restore(o Output) error {
	if !o.Active {
		return SetEnabled(o.Name, false)
	}
	return apply(o.Name, "enable", "mode", o.Mode.arg(), "pos", fmt.Sprint(o.Rect.X), fmt.Sprint(o.Rect.Y),
		"scale", fmt.Sprintf("%.2f", o.Scale), "transform", o.Transform,
		"adaptive_sync", map[bool]string{true: "on", false: "off"}[o.SyncOn()])
}

type Side int

const (
	LeftOf Side = iota
	RightOf
	Above
	Below
)

func (s Side) String() string { return [...]string{"Left of", "Right of", "Above", "Below"}[s] }

func Place(o, ref Output, side Side) (x, y int) {
	w, h := o.Logical()
	switch side {
	case LeftOf:
		return ref.Rect.X - w, ref.Rect.Y
	case RightOf:
		return ref.Rect.X + ref.Rect.Width, ref.Rect.Y
	case Above:
		return ref.Rect.X, ref.Rect.Y - h
	}
	return ref.Rect.X, ref.Rect.Y + ref.Rect.Height
}

func criteria(o Output) string {
	if strings.TrimSpace(o.Make) == "" && strings.TrimSpace(o.Model) == "" {
		return o.Name
	}
	return strings.TrimSpace(o.Make + " " + o.Model + " " + o.Serial)
}

func SwayConfig(outs []Output) string {
	var b strings.Builder
	for _, o := range outs {
		if !o.Active {
			fmt.Fprintf(&b, "output %s disable\n", o.Name)
			continue
		}
		fmt.Fprintf(&b, "output %s mode %s pos %d %d scale %s transform %s adaptive_sync %s\n",
			o.Name, o.Mode.arg(), o.Rect.X, o.Rect.Y, trimHz(o.Scale), o.Transform, map[bool]string{true: "on", false: "off"}[o.SyncOn()])
	}
	return strings.TrimRight(b.String(), "\n")
}

func Kanshi(outs []Output) string {
	var b strings.Builder
	b.WriteString("profile {\n")
	for _, o := range outs {
		if !o.Active {
			fmt.Fprintf(&b, "\toutput %q disable\n", criteria(o))
			continue
		}
		fmt.Fprintf(&b, "\toutput %q enable mode %s position %d,%d scale %s transform %s adaptive_sync %s\n",
			criteria(o), o.Mode.arg(), o.Rect.X, o.Rect.Y, trimHz(o.Scale), o.Transform, map[bool]string{true: "on", false: "off"}[o.SyncOn()])
	}
	b.WriteString("}")
	return b.String()
}
