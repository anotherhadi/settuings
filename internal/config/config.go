package config

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfig []byte

type Config struct {
	Version string `mapstructure:"-"`

	App struct {
		ExternalEditor string `mapstructure:"external_editor"`
	} `mapstructure:"app"`

	TUI struct {
		UseNerdfontIcons    bool   `mapstructure:"use_nerdfont_icons"`
		DefaultSidebarState string `mapstructure:"default_sidebar_state"`
	} `mapstructure:"tui"`

	Keybindings Keybindings `mapstructure:"keybindings"`
}

var Global *Config

func Load(path string) error {
	var defaults map[string]any
	if err := yaml.Unmarshal(defaultConfig, &defaults); err != nil {
		return fmt.Errorf("default config: %w", err)
	}
	for k, v := range flatten("", defaults) {
		viper.SetDefault(k, v)
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	Global = &Config{}
	if err := viper.Unmarshal(Global); err != nil {
		return err
	}
	return nil
}

func WriteDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, defaultConfig, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func flatten(prefix string, m map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]any); ok {
			for nk, nv := range flatten(key, nested) {
				out[nk] = nv
			}
		} else {
			out[key] = v
		}
	}
	return out
}
