package audio

import (
	"regexp"
	"strconv"
	"strings"
)

type statusEntry struct {
	ID        string
	Default   bool
	Name      string
	Volume    int
	Muted     bool
	HasVolume bool
}

var entryLineRe = regexp.MustCompile(`^[\s│├└─]*(\*)?\s*(\d+)\.\s+(.+)$`)

var volBracketRe = regexp.MustCompile(`^(.*?)\s*\[vol:\s*([0-9.]+)(\s+MUTED)?\]$`)

func status() (string, error) {
	return run("status")
}

func parseSection(out, headerLine string) []statusEntry {
	lines := strings.Split(out, "\n")

	audioStart := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "Audio" {
			audioStart = i
			break
		}
	}
	if audioStart == -1 {
		return nil
	}

	sectionStart := -1
	for i := audioStart; i < len(lines); i++ {
		if strings.HasSuffix(strings.TrimSpace(lines[i]), headerLine) {
			sectionStart = i + 1
			break
		}
	}
	if sectionStart == -1 {
		return nil
	}

	var entries []statusEntry
	for i := sectionStart; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasSuffix(trimmed, ":") {
			break
		}
		m := entryLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id, isDefault, rest := m[2], m[1] == "*", strings.TrimSpace(m[3])
		if strings.Contains(rest, " > ") {
			continue
		}

		entry := statusEntry{ID: id, Default: isDefault, Name: rest}
		if vm := volBracketRe.FindStringSubmatch(rest); vm != nil {
			entry.Name = strings.TrimSpace(vm[1])
			if v, err := strconv.ParseFloat(vm[2], 64); err == nil {
				entry.Volume = int(v*100 + 0.5)
				entry.HasVolume = true
			}
			entry.Muted = vm[3] != ""
		}
		entries = append(entries, entry)
	}
	return entries
}
