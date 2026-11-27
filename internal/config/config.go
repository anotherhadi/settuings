package config

import (
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/anotherhadi/foam"
	"github.com/anotherhadi/foam/commands"
	fconfig "github.com/anotherhadi/foam/config"
	"github.com/anotherhadi/foam/theme"
)

//go:embed default.yaml
var Defaults []byte

var PageNames = []string{"about", "network", "bluetooth", "audio", "power", "inputs", "monitors", "netbird"}

type Config struct {
	Theme theme.Theme       `koanf:",squash"`
	Keys  commands.Bindings `koanf:"keys"`
	App   App               `koanf:",squash"`
}

type App struct {
	Page       string   `koanf:"page"`
	Pages      []string `koanf:"pages"`
	Refresh    int      `koanf:"refresh" validate:"gte=1"`
	VolumeStep int      `koanf:"volumeStep" validate:"gte=1,lte=100"`
	SpeedStep  float64  `koanf:"speedStep" validate:"gt=0,lte=1"`
	LightStep  int      `koanf:"brightnessStep" validate:"gte=1,lte=100"`
}

func (c *Config) Validate() error {
	if len(c.App.Pages) == 0 {
		return errors.New("pages: at least one page")
	}
	for i, p := range c.App.Pages {
		if !slices.Contains(PageNames, p) {
			return fmt.Errorf("pages[%d]: %q is not one of %s", i, p, strings.Join(PageNames, ", "))
		}
		if slices.Index(c.App.Pages, p) != i {
			return fmt.Errorf("pages[%d]: %q twice", i, p)
		}
	}
	if !slices.Contains(c.App.Pages, c.App.Page) {
		return fmt.Errorf("page: %q is not in pages", c.App.Page)
	}
	return nil
}

var Version string

func getVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}

const usage = `settuings: system settings in the terminal

Usage: settuings [options]`

func Load(args []string) (*Config, error) {
	var version bool
	cfg, _, err := fconfig.Load[Config](fconfig.Options{
		Name:  "settuings",
		Usage: usage,
		Base:  [][]byte{foam.Defaults, Defaults},
		Files: []string{
			filepath.Join(fconfig.Dir("foam"), "config.yaml"),
			filepath.Join(fconfig.Dir("settuings"), "config.yaml"),
		},
		Flags: func(fs *flag.FlagSet) {
			fs.StringP("page", "p", "", "page to open: "+strings.Join(PageNames, ", "))
			fs.Int("refresh", 0, "seconds between refreshes")
			fs.BoolVarP(&version, "version", "v", false, "show the version")
		},
		FlagKeys: map[string]string{"page": "page", "refresh": "refresh"},
	}, args)
	if version {
		fmt.Println("settuings", getVersion())
		return nil, fconfig.ErrExit
	}
	return cfg, err
}
