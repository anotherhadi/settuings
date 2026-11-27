package audio

import (
	"fmt"
	"strings"
)

type StreamKind int

const (
	StreamOutput StreamKind = iota
	StreamInput
)

type Stream struct {
	ID     string
	App    string
	Kind   StreamKind
	Volume int
	Muted  bool
}

func ListStreams() ([]Stream, error) {
	out, err := status()
	if err != nil {
		return nil, err
	}

	var streams []Stream
	for _, e := range parseSection(out, "Streams:") {
		props, err := inspect(e.ID)
		if err != nil {
			continue
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
