package main

import (
	"fmt"
	"os"

	"github.com/tvrzna/syspower/internal/syspower"
	"golang.org/x/term"
)

const cliVersion = "0.0.2"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "syspower: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
			args = append(args, "tui")
		} else {
			printHelp()
			return nil
		}
	}

	registry := syspower.NewRegistry()
	switch args[0] {
	case "version", "-v", "--version":
		fmt.Printf("syspower %s (cli %s)\nhttps://github.com/tvrzna/syspower\n\nReleased under the MIT License.\n", syspower.GetVersion(), cliVersion)
		return nil
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "tui", "-t", "--tui":
		return runTui(registry)
	}
	return runCli(args, registry)
}

func printHelp() {
	fmt.Printf(`syspower - System power control utility

Usage:
	syspower <command>
	syspower <control> <action> [value]

Global Commands:
	help, -h, --help		Print this help
	list, -l, --list		List all available sysfs controls
	tui, -t, --tui			Start TUI
	version, -v, --version		Print version

Control Actions:
	get				Get active control value
	choices, list			Show available control choices
	set <value>			Write a new value to the control
`)
}
