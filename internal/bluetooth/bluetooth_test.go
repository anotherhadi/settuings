package bluetooth

import "testing"

func TestLastLine(t *testing.T) {
	for in, want := range map[string]string{
		"": "",
		"Failed to connect: org.bluez.Error.Failed":                                "Failed to connect: org.bluez.Error.Failed",
		"Attempting to connect to AA\nFailed to connect: org.bluez.Error.Failed\n": "Failed to connect: org.bluez.Error.Failed",
		"  spaced  ": "spaced",
	} {
		if got := lastLine(in); got != want {
			t.Errorf("lastLine(%q) = %q, want %q", in, got, want)
		}
	}
}
