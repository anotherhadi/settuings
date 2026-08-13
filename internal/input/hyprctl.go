package input

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var ErrNotFound = errors.New("hyprctl not found in PATH")

const defaultTimeout = 15 * time.Second

func CheckAvailable() error {
	if _, err := exec.LookPath("hyprctl"); err != nil {
		return ErrNotFound
	}
	_, err := run("version")
	return err
}

// Installed reports whether hyprctl is present in PATH, without probing
// Hyprland. Cheap enough to call when deciding whether to show the Inputs
// page in the sidebar.
func Installed() bool {
	_, err := exec.LookPath("hyprctl")
	return err == nil
}

func Reset() error {
	_, err := run("reload")
	return err
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "hyprctl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

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
		return "", fmt.Errorf("hyprctl %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}
