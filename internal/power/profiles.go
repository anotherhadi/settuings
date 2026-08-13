// Package power reads and controls the active power profile (via
// powerprofilesctl / power-profiles-daemon) and reports battery/AC status.
// Profile control shells out to a CLI tool, in the same spirit as the
// bluetooth and audio packages; battery status has no universally-installed
// CLI equivalent, so it's read directly from /sys/class/power_supply.
package power

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var ErrNotFound = errors.New("powerprofilesctl not found in PATH")

const defaultTimeout = 15 * time.Second

// Profile is one entry from `powerprofilesctl list`.
type Profile struct {
	Name   string
	Active bool
}

// CheckAvailable reports whether powerprofilesctl is installed and
// power-profiles-daemon is reachable. Call it before using the rest of
// this file's functions.
func CheckAvailable() error {
	if _, err := exec.LookPath("powerprofilesctl"); err != nil {
		return ErrNotFound
	}
	_, err := run("get")
	return err
}

// ListProfiles returns the available profiles in the order reported by
// powerprofilesctl, each flagged with whether it's currently active.
func ListProfiles() ([]Profile, error) {
	out, err := run("list")
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	for _, line := range strings.Split(out, "\n") {
		// Top-level profile entries are indented two spaces (or "* " when
		// active); their attribute lines (Driver:, Degraded:, ...) are
		// indented four, so those are skipped here.
		if line == "" || strings.HasPrefix(line, "    ") {
			continue
		}

		active := false
		trimmed := line
		switch {
		case strings.HasPrefix(trimmed, "* "):
			active = true
			trimmed = trimmed[2:]
		case strings.HasPrefix(trimmed, "  "):
			trimmed = trimmed[2:]
		default:
			continue
		}

		name := strings.TrimSuffix(trimmed, ":")
		if name == "" {
			continue
		}
		profiles = append(profiles, Profile{Name: name, Active: active})
	}
	return profiles, nil
}

// SetActive switches the active power profile.
func SetActive(name string) error {
	_, err := run("set", name)
	return err
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powerprofilesctl", args...) // #nosec G204 -- args are fixed powerprofilesctl subcommands or profile names read back from `powerprofilesctl list`, never shell-interpreted
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("powerprofilesctl %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
