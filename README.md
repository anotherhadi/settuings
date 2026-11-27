```
█▀▀ █▀▀ ▀█▀ ▀█▀ █ █ █ █▄ █ █▀▀ █▀▀
▄▄█ ██▄  █   █  █▄█ █ █ ▀█ █▄█ ▄▄█
```

# Settuings

> A TUI for your system settings: **Wi-Fi, Bluetooth, audio, power, mouse and keyboard**, without leaving the terminal.

Every page is a thin wrapper around the tools you already have (`nmcli`, `bluetoothctl`, `wpctl`, `powerprofilesctl`, `swaymsg`, `fastfetch`): no daemon, and it always reflects what changed elsewhere.

## Installation

With go:

```sh
go install github.com/anotherhadi/settuings/cmd/settuings@latest
```

With Nix, from the [anotherhadi NUR](https://github.com/anotherhadi/nur-packages):

```sh
nix run github:anotherhadi/nur-packages#settuings
```

```nix
# flake: inputs.nur-anotherhadi.url = "github:anotherhadi/nur-packages";
environment.systemPackages = [ nur-anotherhadi.packages.${system}.settuings ];

# or through NUR
environment.systemPackages = [ pkgs.nur.repos.anotherhadi.settuings ];
```

Each [GitHub release](https://github.com/anotherhadi/settuings/releases) has a Linux archive (amd64 and arm64) with the `settuings` binary.

## Usage

```
settuings [options]

  -p, --page string   page to open: about, network, bluetooth, audio, power, inputs, monitors, netbird
      --refresh int   seconds between refreshes
  -c, --set key=value set a config key (repeatable)
      --print-config  print the merged config, then exit
  -v, --version       show the version
```

`?` lists the shortcuts of the page, `:` opens the command palette. The mouse works too: click a tab or a row, scroll with the wheel.

## Configuration

Put only what you change in `~/.config/settuings/config.yaml`. The look (theme, borders, nerd fonts) and the shared keys are in `~/.config/foam/config.yaml`. Every key is in [`default.yaml`](internal/config/default.yaml); environment variables work too (`SETTUINGS_REFRESH=2`).

```yaml
page: audio
pages: [network, bluetooth, audio]
volumeStep: 2
brightnessStep: 10
keys:
  app:
    quit: [q]
  audio:
    mute: [m, x]
```
