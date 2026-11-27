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

func CheckAvailable() error {
	if _, err := exec.LookPath("wpctl"); err != nil {
		return ErrNotFound
	}
	if _, err := run("status"); err != nil {
		return err
	}
	return nil
}

func Installed() bool {
	_, err := exec.LookPath("wpctl")
	return err == nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wpctl", args...)
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
