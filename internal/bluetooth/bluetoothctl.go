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

func Installed() bool {
	_, err := exec.LookPath("bluetoothctl")
	return err == nil
}

func run(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bluetoothctl", args...)
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
