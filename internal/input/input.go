package input

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("sway is not running (swaymsg and SWAYSOCK needed)")

type Pointer struct {
	Speed   float64
	Profile string
	Devices []string
}

type Keyboard struct {
	Layouts []string
	Index   int
	Devices []string
}

func (k Keyboard) Layout() string {
	if k.Index >= 0 && k.Index < len(k.Layouts) {
		return k.Layouts[k.Index]
	}
	return ""
}

type State struct {
	Mouse    *Pointer
	Touchpad *Pointer
	Keyboard *Keyboard
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
	cmd := exec.CommandContext(ctx, "swaymsg", append([]string{"-r"}, args...)...)
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

type rawDevice struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	LayoutName []string `json:"xkb_layout_names"`
	LayoutIdx  int      `json:"xkb_active_layout_index"`
	Libinput   struct {
		AccelSpeed   *float64 `json:"accel_speed"`
		AccelProfile string   `json:"accel_profile"`
	} `json:"libinput"`
}

var notKeyboard = regexp.MustCompile(`(?i)button|switch|hotkeys|video bus|audio|avrcp|control|dock|speaker|headset|headphone`)

func Read() (State, error) {
	out, err := swaymsg("-t", "get_inputs")
	if err != nil {
		return State{}, err
	}
	var raw []rawDevice
	if err := json.Unmarshal(out, &raw); err != nil {
		return State{}, err
	}
	return parse(raw), nil
}

func parse(raw []rawDevice) State {
	var st State
	keyboards := map[string]bool{}
	for _, r := range raw {
		if r.Type != "keyboard" || len(r.LayoutName) == 0 || notKeyboard.MatchString(r.Name) {
			continue
		}
		keyboards[r.Name] = true
		if st.Keyboard == nil {
			st.Keyboard = &Keyboard{Layouts: r.LayoutName, Index: r.LayoutIdx}
		}
		st.Keyboard.Devices = append(st.Keyboard.Devices, r.Name)
	}
	for _, r := range raw {
		if (r.Type != "pointer" && r.Type != "touchpad") || r.Libinput.AccelSpeed == nil || keyboards[r.Name] {
			continue
		}
		slot := &st.Mouse
		if r.Type == "touchpad" {
			slot = &st.Touchpad
		}
		if *slot == nil {
			*slot = &Pointer{Speed: *r.Libinput.AccelSpeed, Profile: r.Libinput.AccelProfile}
		}
		(*slot).Devices = append((*slot).Devices, r.Name)
	}
	if st.Keyboard != nil {
		slices.Sort(st.Keyboard.Devices)
	}
	return st
}

func command(args ...string) error {
	out, err := swaymsg(append([]string{"--"}, args...)...)
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

func SetMouseSpeed(v float64) error    { return setSpeed("type:pointer", v) }
func SetTouchpadSpeed(v float64) error { return setSpeed("type:touchpad", v) }

func setSpeed(target string, v float64) error {
	return command("input", target, "pointer_accel", fmt.Sprintf("%.2f", min(max(v, -1), 1)))
}

var layoutSpec = regexp.MustCompile(`^[A-Za-z0-9_,()-]+$`)

func ValidLayouts(s string) error {
	if !layoutSpec.MatchString(s) {
		return errors.New("use xkb names separated by commas, e.g. fr,us")
	}
	return nil
}

func SetLayouts(layouts string) error {
	if err := ValidLayouts(layouts); err != nil {
		return err
	}
	return command("input", "type:keyboard", "xkb_layout", `"`+layouts+`"`)
}
