// Package bluetooth wraps bluetoothctl to read and control Bluetooth state
// (controller power, device pairing/connection). Every exported function
// shells out to bluetoothctl in its non-interactive single-command mode;
// there is no direct D-Bus/BlueZ integration.
package bluetooth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

var (
	ErrNotFound     = errors.New("bluetoothctl not found in PATH")
	ErrNoController = errors.New("no Bluetooth controller found")
)

const defaultTimeout = 15 * time.Second

// CheckAvailable reports whether bluetoothctl is installed and a Bluetooth
// controller is reachable. Call it before using the rest of the package.
func CheckAvailable() error {
	if _, err := exec.LookPath("bluetoothctl"); err != nil {
		return ErrNotFound
	}
	out, err := run(defaultTimeout, "list")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNoController, err)
	}
	if strings.TrimSpace(out) == "" {
		return ErrNoController
	}
	return nil
}

// run executes bluetoothctl in single-command (non-interactive) mode.
// bluetoothctl writes both normal output and error messages to stdout
// (colored with ANSI escapes), so errors are extracted from there rather
// than stderr.
func run(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bluetoothctl", args...) // #nosec G204 -- args are fixed bluetoothctl subcommands or user-supplied device addresses, never shell-interpreted
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := ansi.Strip(stdout.String())
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		msg := strings.TrimSpace(out)
		if msg == "" {
			msg = strings.TrimSpace(stderr.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("bluetoothctl %s: %s", strings.Join(args, " "), msg)
	}
	return out, nil
}
