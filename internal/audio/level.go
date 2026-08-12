package audio

import (
	"encoding/binary"
	"io"
	"math"
	"os/exec"
	"strconv"
)

// levelSampleRate and levelFrameSamples pick a small chunk size (~100ms) so
// the UI can poll for a fresh peak reading at a responsive rate without
// reading much audio data.
const (
	levelSampleRate  = 16000
	levelFrameFrames = levelSampleRate / 10 // 100ms
	levelFrameBytes  = levelFrameFrames * 4 // 4 bytes per float32 sample
)

// LevelMonitor streams raw audio from a source and reports its peak
// amplitude in short chunks, for a live input-level meter. Callers must call
// Stop when done to release the underlying pw-cat process.
type LevelMonitor struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

// StartLevelMonitor begins capturing raw audio from sourceID.
func StartLevelMonitor(sourceID string) (*LevelMonitor, error) {
	target, err := nodeTarget(sourceID)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("pw-cat", // #nosec G204 -- target is a node.name read back from `wpctl inspect`, never shell-interpreted
		"--record",
		"--target", target,
		"--format", "f32",
		"--rate", strconv.Itoa(levelSampleRate),
		"--channels", "1",
		"--raw",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &LevelMonitor{cmd: cmd, stdout: stdout}, nil
}

// Next blocks until one ~100ms chunk of audio has been captured and returns
// its peak amplitude as a value in [0, 1]. It returns an error once the
// monitor is stopped or the source disappears.
func (l *LevelMonitor) Next() (float64, error) {
	buf := make([]byte, levelFrameBytes)
	n, err := io.ReadFull(l.stdout, buf)
	if n == 0 {
		return 0, err
	}

	var peak float32
	for i := 0; i+4 <= n; i += 4 {
		v := math.Float32frombits(binary.LittleEndian.Uint32(buf[i : i+4]))
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	if peak > 1 {
		peak = 1
	}
	return float64(peak), nil
}

// Stop terminates the capture process. Safe to call more than once.
func (l *LevelMonitor) Stop() {
	if l.cmd.Process != nil {
		_ = l.cmd.Process.Kill()
	}
	_ = l.cmd.Wait()
}
