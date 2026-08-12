package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/icons"
	"github.com/anotherhadi/settuings/internal/keys"
	"github.com/anotherhadi/settuings/internal/style"
	appUI "github.com/anotherhadi/settuings/internal/ui/app"
	"github.com/spf13/cobra"
)

// Version is overwritten at build time by goreleaser/ldflag with the current version tag, or "dev" if not set.
var version = "dev"

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

func main() {
	var (
		flagConfig           string
		flagAddDefaultConfig bool
		flagPage             string
	)

	rootCmd := &cobra.Command{
		Use:           "settuings",
		Short:         "A TUI to manage your Linux system settings like wifi, bluetooth, and more, without leaving the terminal.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			cfgPath := filepath.Join(home, ".config", "settuings", "config.yaml")
			if flagConfig != "" {
				cfgPath = flagConfig
			}

			if flagAddDefaultConfig {
				if err := config.WriteDefaultConfig(cfgPath); err != nil {
					return fmt.Errorf("add-default-config: %w", err)
				}
				fmt.Printf("default config written to %s\n", cfgPath)
				return nil
			}

			if err := config.Load(cfgPath); err != nil {
				return fmt.Errorf("config: %w", err)
			}
			config.Global.Version = version

			style.Init()
			icons.Init(config.Global)
			keys.Init(config.Global)

			m, err := appUI.New(flagPage)
			if err != nil {
				return err
			}

			final, err := tea.NewProgram(m).Run()
			if err != nil {
				return fmt.Errorf("tui: %w", err)
			}
			if app, ok := final.(appUI.Model); ok {
				if ferr := app.FatalErr(); ferr != nil {
					return ferr
				}
			}
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "path to config file")
	rootCmd.Flags().BoolVar(&flagAddDefaultConfig, "add-default-config", false, "copy the default config file to the config path and exit")
	rootCmd.Flags().StringVarP(&flagPage, "page", "p", "", fmt.Sprintf("page to open at launch (%s)", strings.Join(appUI.PageNames(), ", ")))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
