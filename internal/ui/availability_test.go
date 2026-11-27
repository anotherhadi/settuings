package ui

import (
	"testing"

	"github.com/anotherhadi/foam"
	fconfig "github.com/anotherhadi/foam/config"

	"github.com/anotherhadi/settuings/internal/config"
)

func TestPagesNeedTheirTool(t *testing.T) {
	cfg, err := fconfig.LoadBytes[config.Config](foam.Defaults, config.Defaults)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())
	t.Setenv("SWAYSOCK", "")
	m, err := build(cfg, true)
	if err == nil {
		defer m.Close()
		for _, p := range m.pages {
			if p.Name() != "power" {
				t.Errorf("page %s is shown without its tool", p.Name())
			}
		}
		return
	}
	if want := "no page available"; len(err.Error()) < len(want) || err.Error()[:len(want)] != want {
		t.Fatalf("err = %v", err)
	}
}
