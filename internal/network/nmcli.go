// Package network wraps nmcli to read and control network state
// (Wi-Fi, Ethernet, VPN). Every exported function shells out to nmcli;
// there is no direct D-Bus/NetworkManager integration.
package network

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("nmcli not found in PATH")
	ErrNotRunning = errors.New("NetworkManager is not running")
)

const defaultTimeout = 15 * time.Second

// CheckAvailable reports whether nmcli is installed and NetworkManager is
// reachable. Call it before using the rest of the package.
func CheckAvailable() error {
	if _, err := exec.LookPath("nmcli"); err != nil {
		return ErrNotFound
	}
	out, err := run("-t", "-f", "RUNNING", "general")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotRunning, err)
	}
	if strings.TrimSpace(out) != "running" {
		return ErrNotRunning
	}
	return nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nmcli", args...) // #nosec G204 -- args are fixed nmcli subcommands or user-supplied SSID/password, never shell-interpreted
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
		return "", fmt.Errorf("nmcli %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// getField runs `nmcli -g <field> ...` and returns the single trimmed value.
func getField(field string, args ...string) (string, error) {
	out, err := run(append([]string{"-g", field}, args...)...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// runTerse runs nmcli with -t -f <fields> and parses each output line into
// its colon-separated columns, unescaping nmcli's `\:` and `\\`.
func runTerse(fields string, args ...string) ([][]string, error) {
	out, err := run(append([]string{"-t", "-f", fields}, args...)...)
	if err != nil {
		return nil, err
	}
	out = strings.TrimRight(out, "\n")
	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, parseTerseLine(line))
	}
	return rows, nil
}

func parseTerseLine(line string) []string {
	var fields []string
	var cur strings.Builder
	escaped := false
	for _, r := range line {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		switch r {
		case '\\':
			escaped = true
		case ':':
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	fields = append(fields, cur.String())
	return fields
}

// field is a small helper to safely index a parsed terse row.
func field(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}
