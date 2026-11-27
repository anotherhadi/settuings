package brightness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Device struct {
	Name    string
	Percent int
}

func Installed() bool {
	_, err := exec.LookPath("brightnessctl")
	return err == nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "brightnessctl", args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", errors.New(msg)
	}
	return stdout.String(), nil
}

func List() ([]Device, error) {
	out, err := run("-l", "-m", "-c", "backlight")
	if err != nil {
		return nil, err
	}
	return parse(out), nil
}

func parse(out string) []Device {
	var devs []Device
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Split(line, ",")
		if len(f) < 5 || f[1] != "backlight" {
			continue
		}
		pct, err := strconv.Atoi(strings.TrimSuffix(f[3], "%"))
		if err != nil {
			continue
		}
		devs = append(devs, Device{Name: f[0], Percent: pct})
	}
	return devs
}

func Set(name string, percent int) error {
	_, err := run("-d", name, "set", fmt.Sprintf("%d%%", min(max(percent, 1), 100)))
	return err
}
