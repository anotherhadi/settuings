package audio

import (
	"fmt"
	"strings"
)

type StreamKind int

const (
	StreamOutput StreamKind = iota // a Stream/Output/Audio node, e.g. a browser tab playing sound
	StreamInput                    // a Stream/Input/Audio node, e.g. an app recording from the mic
)

// Stream is one running application audio stream (playback or capture), as
// listed under wpctl status's Streams section.
type Stream struct {
	ID     string
	App    string // human-readable application name
	Kind   StreamKind
	Volume int // 0-100
	Muted  bool
}

// ListStreams returns every currently running application audio stream,
// both playback and capture, output streams first.
func ListStreams() ([]Stream, error) {
	out, err := status()
	if err != nil {
		return nil, err
	}

	var streams []Stream
	for _, e := range parseSection(out, "Streams:") {
		props, err := inspect(e.ID)
		if err != nil {
			continue // the stream may have exited between `status` and `inspect`
		}
		var kind StreamKind
		switch props["media.class"] {
		case "Stream/Output/Audio":
			kind = StreamOutput
		case "Stream/Input/Audio":
			kind = StreamInput
		default:
			continue
		}

		app := props["application.name"]
		if app == "" {
			app = e.Name
		}

		vol, muted, err := GetVolume(e.ID)
		if err != nil {
			continue
		}

		streams = append(streams, Stream{ID: e.ID, App: app, Kind: kind, Volume: vol, Muted: muted})
	}

	// Output streams first, stable within each kind.
	out2 := make([]Stream, 0, len(streams))
	for _, s := range streams {
		if s.Kind == StreamOutput {
			out2 = append(out2, s)
		}
	}
	for _, s := range streams {
		if s.Kind == StreamInput {
			out2 = append(out2, s)
		}
	}
	return out2, nil
}

// inspect runs `wpctl inspect <id>` and parses its "key = value" property
// dump into a map. Values are unquoted; boolean/marker prefixes ("* ") are
// stripped from keys.
func inspect(id string) (map[string]string, error) {
	out, err := run("inspect", id)
	if err != nil {
		return nil, err
	}
	props := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "* ")
		key, value, ok := strings.Cut(line, " = ")
		if !ok {
			continue
		}
		props[key] = strings.Trim(value, `"`)
	}
	return props, nil
}

// nodeTarget resolves a wpctl global id (e.g. "57", as used by
// get-volume/set-volume/set-default/...) to the PipeWire node.name pw-play
// and pw-cat expect for their --target flag. The two tools use different id
// spaces: passing a wpctl id straight to --target is silently ignored and
// falls back to the default node instead of erroring.
func nodeTarget(id string) (string, error) {
	props, err := inspect(id)
	if err != nil {
		return "", err
	}
	name := props["node.name"]
	if name == "" {
		return "", fmt.Errorf("no node.name found for id %s", id)
	}
	return name, nil
}
