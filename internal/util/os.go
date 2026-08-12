package util

import (
	"os"
	"runtime"
	"strings"
)

// OSName returns a human-readable OS name, e.g. "NixOS 25.05" or
// "Ubuntu 24.04.1 LTS", read from /etc/os-release. Falls back to
// runtime.GOOS if the file is missing or has no PRETTY_NAME.
func OSName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return runtime.GOOS
	}
	for _, line := range strings.Split(string(data), "\n") {
		name, ok := strings.CutPrefix(line, "PRETTY_NAME=")
		if !ok {
			continue
		}
		return strings.Trim(strings.TrimSpace(name), `"`)
	}
	return runtime.GOOS
}

// KernelVersion returns the running kernel release, e.g. "6.12.9-arch1-1".
func KernelVersion() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// Hostname returns the machine's hostname.
func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

// Arch returns the CPU architecture, e.g. "amd64".
func Arch() string {
	return runtime.GOARCH
}
