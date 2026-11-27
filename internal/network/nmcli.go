package network

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var secretsRe = regexp.MustCompile(`(?i)secrets|no agents`)

var (
	ErrSecrets    = errors.New("password needed")
	ErrNotFound   = errors.New("nmcli not found in PATH")
	ErrNotRunning = errors.New("NetworkManager is not running")
)

const defaultTimeout = 15 * time.Second

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

func Installed() bool {
	_, err := exec.LookPath("nmcli")
	return err == nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nmcli", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		msg := strings.TrimPrefix(strings.TrimSpace(stderr.String()), "Error: ")
		if msg == "" {
			msg = err.Error()
		}
		if secretsRe.MatchString(msg) {
			return "", fmt.Errorf("%w: %s", ErrSecrets, msg)
		}
		return "", errors.New(msg)
	}
	return stdout.String(), nil
}

func getField(field string, args ...string) (string, error) {
	out, err := run(append([]string{"-g", field}, args...)...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

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

func field(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}
