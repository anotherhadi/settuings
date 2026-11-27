package main

import (
	"errors"
	"fmt"
	"os"

	fconfig "github.com/anotherhadi/foam/config"

	"github.com/anotherhadi/settuings/internal/config"
	"github.com/anotherhadi/settuings/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "settuings:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Args[1:])
	if errors.Is(err, fconfig.ErrExit) {
		return nil
	}
	if err != nil {
		return err
	}
	return ui.Run(cfg)
}
