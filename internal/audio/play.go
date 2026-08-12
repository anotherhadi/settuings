package audio

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"os/exec"
)

const sampleRate = 44100

// toneStep is one segment of the test chime: a tone at freqHz lasting dur,
// with a short fade in/out to avoid clicks.
type toneStep struct {
	freqHz float64
	dur    float64 // seconds
}

// testChime is a short two-tone "ding-dong" beep, the same shape used by
// most desktop sound settings panels to preview an output device.
var testChime = []toneStep{
	{freqHz: 880, dur: 0.15},
	{freqHz: 660, dur: 0.22},
}

// PlayTestTone plays a short chime through the given sink, without changing
// the system default output.
func PlayTestTone(sinkID string) error {
	target, err := nodeTarget(sinkID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pw-play", "--target", target, "-") // #nosec G204 -- target is a node.name read back from `wpctl inspect`, never shell-interpreted
	cmd.Stdin = bytes.NewReader(synthChime(testChime))
	return cmd.Run()
}

// synthChime renders steps as a 16-bit PCM mono WAV file in memory.
func synthChime(steps []toneStep) []byte {
	const amplitude = 0.35 * 32767
	const fade = 0.015 // seconds of fade in/out per step, to avoid clicks

	var samples []int16
	for _, step := range steps {
		n := int(step.dur * sampleRate)
		for i := 0; i < n; i++ {
			t := float64(i) / sampleRate
			env := 1.0
			if t < fade {
				env = t / fade
			} else if remaining := step.dur - t; remaining < fade {
				env = remaining / fade
			}
			v := amplitude * env * math.Sin(2*math.Pi*step.freqHz*t)
			samples = append(samples, int16(v))
		}
	}

	pcm := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(s))
	}

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	writeU32(&buf, uint32(36+len(pcm)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	writeU32(&buf, 16)
	writeU16(&buf, 1) // PCM
	writeU16(&buf, 1) // mono
	writeU32(&buf, sampleRate)
	writeU32(&buf, sampleRate*2)
	writeU16(&buf, 2)  // block align
	writeU16(&buf, 16) // bits per sample
	buf.WriteString("data")
	writeU32(&buf, uint32(len(pcm)))
	buf.Write(pcm)
	return buf.Bytes()
}

func writeU32(buf *bytes.Buffer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	buf.Write(b[:])
}

func writeU16(buf *bytes.Buffer, v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	buf.Write(b[:])
}
