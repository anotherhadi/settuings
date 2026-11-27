package audio

import (
	"encoding/binary"
	"io"
	"math"
	"os/exec"
	"strconv"
)

const (
	levelSampleRate  = 16000
	levelFrameFrames = levelSampleRate / 10
	levelFrameBytes  = levelFrameFrames * 4
)

type LevelMonitor struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func StartLevelMonitor(sourceID string) (*LevelMonitor, error) {
	target, err := nodeTarget(sourceID)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("pw-cat",
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

func (l *LevelMonitor) Stop() {
	if l.cmd.Process != nil {
		_ = l.cmd.Process.Kill()
	}
	_ = l.cmd.Wait()
}
