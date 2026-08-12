// Package audio wraps wpctl (and the pw-play/pw-cat helper tools that ship
// alongside it) to read and control PipeWire audio state: output/input
// devices, their volume and mute, the default device, and per-application
// streams. Every exported function shells out to a CLI tool; there is no
// direct PipeWire/WirePlumber library integration.
package audio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var ErrNotFound = errors.New("wpctl not found in PATH")

const defaultTimeout = 15 * time.Second

// CheckAvailable reports whether wpctl is installed and PipeWire is
// reachable. Call it before using the rest of the package.
func CheckAvailable() error {
	if _, err := exec.LookPath("wpctl"); err != nil {
		return ErrNotFound
	}
	if _, err := run("status"); err != nil {
		return err
	}
	return nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wpctl", args...) // #nosec G204 -- args are fixed wpctl subcommands or IDs read back from `wpctl status`, never shell-interpreted
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
		return "", fmt.Errorf("wpctl %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
