# syspower

A lightweight, dependency-minimal Linux tool to view and manage system power attributes via `sysfs` (e.g., `cpuboost`, `platform_profile`).

Provides both terminal (CLI/TUI) and desktop (GUI) interfaces built around a shared Go core engine.

## Features

- **CLI & TUI (`syspower`)**: Instant command-line queries/updates and an interactive terminal UI.
- **GUI (`syspower-gui`)**: Graphical desktop interface built with Tk.
- **Minimal Dependencies**: Core logic relies purely on the Go standard library and Linux `sysfs`.

## Requirements

- **OS**: Linux (with populated `/sys` power attributes)
- **Go**: 1.26.6 or newer
- **Permissions**: Root or `sudo`/`doas` privileges to write to sysfs attributes

## Building

```bash
# Build both CLI and GUI binaries
make build

# Build individual targets
make build-cli    # Outputs dist/syspower-cli
make build-gui    # Outputs dist/syspower-gui

# Clean build directory
make clean

```

## Usage

### 1. CLI & TUI (`syspower`)

**Interactive TUI** (launches automatically when run without arguments in an interactive TTY):
```bash
sudo ./dist/syspower-cli tui
```
* **Navigation**: `Up`(`j`) / `Down`(`k`) to select control, `Left`(`h`) / `Right`(`l`) to change choice.
* **Quit**: Press `q` or `Ctrl+C`.

**Direct CLI Commands**:
```bash
# List all available controls and current values
syspower-cli list

# Query a specific control
syspower-cli get cpuboost

# Set a value (requires root/sudo)
sudo syspower-cli set cpuboost 1
```

### 2. GUI (`syspower-gui`)

Launch the desktop UI to interactively select power profiles using radio buttons:

```bash
sudo ./dist/syspower-gui
```