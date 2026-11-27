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

type Profile struct {
	Name   string
	Active bool
}

func CheckAvailable() error {
	if _, err := exec.LookPath("powerprofilesctl"); err != nil {
		return ErrNotFound
	}
	_, err := run("get")
	return err
}

func Installed() bool {
	_, err := exec.LookPath("powerprofilesctl")
	return err == nil
}

func ListProfiles() ([]Profile, error) {
	out, err := run("list")
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	for _, line := range strings.Split(out, "\n") {

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

func SetActive(name string) error {
	_, err := run("set", name)
	return err
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powerprofilesctl", args...)
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
