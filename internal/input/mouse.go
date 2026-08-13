package input

import (
	"encoding/json"
	"fmt"
)

const (
	MinSensitivity = -1.0
	MaxSensitivity = 1.0
)

func CurrentSensitivity() (float64, error) {
	out, err := run("getoption", "input:sensitivity", "-j")
	if err != nil {
		return 0, err
	}
	var resp getOptionResp
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return 0, fmt.Errorf("parse getoption input:sensitivity: %w", err)
	}
	return resp.Float, nil
}

func SetSensitivity(v float64) error {
	switch {
	case v < MinSensitivity:
		v = MinSensitivity
	case v > MaxSensitivity:
		v = MaxSensitivity
	}
	_, err := run("keyword", "input:sensitivity", fmt.Sprintf("%.2f", v))
	return err
}
