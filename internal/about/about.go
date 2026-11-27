package about

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi"
)

type Field struct{ Key, Value string }

type Info struct {
	Logo   []string
	Title  string
	Fields []Field
}

var ErrNotFound = errors.New("fastfetch not found in PATH")

func Installed() bool {
	_, err := exec.LookPath("fastfetch")
	return err == nil
}

func run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "fastfetch", append(args, "--pipe", "false")...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return out.String(), nil
}

func Fetch() (Info, error) {
	var logo, info string
	var errLogo, errInfo error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		logo, errLogo = run("--logo-type", "builtin", "--logo-padding-left", "0", "--logo-padding-right", "0", "-s", "none")
	}()
	go func() {
		defer wg.Done()
		info, errInfo = run("-l", "none")
	}()
	wg.Wait()
	if errInfo != nil {
		return Info{}, errInfo
	}
	i := parseInfo(info)
	if errLogo == nil {
		i.Logo = parseLogo(logo)
	}
	return i, nil
}

func parseInfo(out string) Info {
	var i Info
	for n, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		plain := strings.TrimSpace(ansi.Strip(line))
		switch {
		case plain == "":
		case n == 0:
			i.Title = plain
		case strings.Trim(plain, "-─") == "":
		default:
			k, v, ok := strings.Cut(plain, ": ")
			if !ok {
				k, v = plain, ""
			}
			i.Fields = append(i.Fields, Field{Key: k, Value: v})
		}
	}
	return i
}

var sgr = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func parseLogo(out string) []string {
	lines := strings.Split(strings.TrimRight(out, "\n \t"), "\n")
	state := ""
	width := 0
	for _, l := range lines {
		width = max(width, ansi.StringWidth(l))
	}
	for k, l := range lines {
		lines[k] = state + l
		for _, s := range sgr.FindAllString(l, -1) {
			if s == "\x1b[0m" || s == "\x1b[m" {
				state = ""
			} else {
				state += s
			}
		}
		lines[k] += "\x1b[0m" + strings.Repeat(" ", width-ansi.StringWidth(l))
	}
	return lines
}
