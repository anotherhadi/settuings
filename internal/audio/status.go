package audio

import (
	"regexp"
	"strconv"
	"strings"
)

// statusEntry is one line parsed out of a `wpctl status` section (Sinks,
// Sources, or Streams).
type statusEntry struct {
	ID        string
	Default   bool
	Name      string
	Volume    int // 0-100, only meaningful when HasVolume is true
	Muted     bool
	HasVolume bool
}

// entryLineRe matches one tree entry, e.g.:
//
//	" │  *   57. Ryzen HD Audio Controller Speaker   [vol: 0.90 MUTED]"
//	"      110. pw-play"
//
// Leading tree-drawing characters and the optional "*" (default marker) are
// stripped; the remainder is the id and the free-form rest of the line.
var entryLineRe = regexp.MustCompile(`^[\s│├└─]*(\*)?\s*(\d+)\.\s+(.+)$`)

var volBracketRe = regexp.MustCompile(`^(.*?)\s*\[vol:\s*([0-9.]+)(\s+MUTED)?\]$`)

// status returns the raw output of `wpctl status`.
func status() (string, error) {
	return run("status")
}

// parseSection extracts the entries listed under headerLine (e.g. "Sinks:",
// "Sources:", "Streams:") within the "Audio" domain block of `wpctl status`
// output, stopping at the first blank line. Port/link sub-lines (which
// contain " > ") are tree children of a stream entry, not entries of their
// own, and are skipped.
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
		// A blank line ends the whole Audio block (e.g. after Streams:); a
		// line ending in ":" is the next section header (e.g. "Sources:").
		// Either marks the end of this section.
		if trimmed == "" || strings.HasSuffix(trimmed, ":") {
			break
		}
		m := entryLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id, isDefault, rest := m[2], m[1] == "*", strings.TrimSpace(m[3])
		if strings.Contains(rest, " > ") {
			continue // a port/link sub-line, not an entry
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
