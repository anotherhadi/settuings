package network

import (
	"errors"
	"testing"
)

func TestEmptyPasswordIsRefused(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := ConnectKnownWithSecret("home", ""); !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("err = %v", err)
	}
}
