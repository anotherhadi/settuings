<div align="center">
    <img alt="logo" src="./.github/assets/logo.png" width="120px" />
</div>

<br>

# SetTUIngs

> A TUI to manage your Linux system settings like wifi, bluetooth, and more, without leaving the terminal.

![GitHub Stars](https://www.shieldcn.dev/github/stars/anotherhadi/settuings.svg?variant=outline&theme=violet)
![Release](https://www.shieldcn.dev/github/release/anotherhadi/settuings.svg?variant=outline&theme=violet)
![CI](https://www.shieldcn.dev/github/ci/anotherhadi/settuings.svg?variant=outline&theme=violet)
[![Ko-fi](https://www.shieldcn.dev/badge/Ko--fi-sponsor-FF5E5B.svg?logo=kofi&variant=secondary&theme=violet)](https://ko-fi.com/anotherhadi)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [What is Spilltea?](#what-is-settuings)
- [Legal Disclaimer](#legal-disclaimer)
- [Features](#features)
- [Installation](#installation)
- [Project Management](#project-management)
- [Configuration](#configuration)
  - [Per-project configuration](#per-project-configuration)
- [CLI Flags](#cli-flags)
- [Plugin System](#plugin-system)
- [Vim / Neovim Integration](#vim--neovim-integration)
- [Deployment](#deployment)
- [Tech Stack](#tech-stack)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## What is SetTUIngs?

SetTUIngs is a keyboard-driven TUI for the everyday Linux system settings you'd otherwise dig through a GUI panel for: WiFi, Ethernet, VPN, Bluetooth, and audio. Every page is a thin, readable wrapper around the CLI tools you already have installed (`nmcli`, `bluetoothctl`, and `wpctl`), so there's no background daemon or D-Bus client library involved, and it always reflects whatever changed things elsewhere (a NetworkManager applet, `pavucontrol`, another terminal).

<img alt="demo" src="./.github/assets/demo.gif" width="700" />

## Features

- **WiFi**: Scan, connect, and manage networks, including signal strength, security, saved passwords, and known networks, with the radio toggled on/off in one keystroke.
- **Ethernet & VPN**: Bring interfaces up or down and connect to VPN profiles from the same Network page.
- **Bluetooth**: Discover, pair, connect, trust, and forget devices, and power the adapter on/off.
- **Audio**: Adjust output/input volume, mute, and pick the default speaker or mic; preview a device with a test tone or a live input-level meter; control the volume of individual apps (browser tab, Spotify, etc.) independently.
- **Vim-like Navigation**: The entire interface is keyboard-driven with Vim-inspired shortcuts. Use `h/j/k/l` to move, `gg`/`G` to jump to the top/bottom, `/` to search, `q` to close panels, and more. All keybindings are fully customizable via the config file.

## Installation

<details>
<summary>Go install</summary>

```sh
go install github.com/anotherhadi/settuings/cmd/settuings@latest
```

Requires Go 1.22+. The binary will be placed in `$GOPATH/bin` (or `~/go/bin`).

</details>

<details>
<summary>Nix (temporary run, no install)</summary>

```sh
nix run github:anotherhadi/settuings
```

</details>

<details>
<summary>NixOS (flake)</summary>

Add settuings to your flake inputs:

```nix
inputs.settuings.url = "github:anotherhadi/settuings";
```

Then add the package to your system or home-manager packages:

```nix
environment.systemPackages = [ inputs.settuings.packages.${pkgs.system}.default ];
```

</details>

## CLI Flags

<!-- exec: echo '```' && go run ./cmd/settuings -h && echo '```' -->

```
A minimal, terminal-based HTTP(S) proxy for pentesters and CTF players.

Usage:
  spilltea [flags]

Flags:
      --add-default-config      copy the default config file to the config path and exit
      --add-default-plugins     copy built-in example plugins into the plugins dir and exit
  -c, --config string           path to config file
  -h, --help                    help for spilltea
      --host string             proxy host (overrides config)
      --plugins-dir string      path to plugins dir (overrides config)
  -p, --port int                proxy port (overrides config)
  -P, --project string          project name to open directly, or "tmp" for a temporary session
      --ssl-insecure            skip TLS certificate verification (overrides config)
      --upstream-proxy string   upstream proxy URL, e.g. http://user:pass@host:8888 (overrides config)
  -v, --version                 version for spilltea
```

<!-- endexec -->

## Tech Stack

| Component | Library                                                 |
| --------- | ------------------------------------------------------- |
| TUI       | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styles    | [lipgloss](https://github.com/charmbracelet/lipgloss)   |
| Config    | [viper](https://github.com/spf13/viper)                 |
| Plugins   | [gopher-lua](https://github.com/yuin/gopher-lua)        |

---

<div align="center">
  <a href="https://github.com/anotherhadi/settuings">github</a> |
  <a href="https://gitlab.com/anotherhadi_mirror/settuings">gitlab (mirror)</a> |
  <a href="https://git.hadi.icu/anotherhadi/settuings">gitea (mirror)</a>
</div
