package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/tvrzna/syspower/internal/syspower"
)

const cliVersion = "0.0.1"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "syspower: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	registry := syspower.NewRegistry()

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Printf("syspower %s (cli %s)\n", syspower.GetVersion(), cliVersion)
		return nil
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "tui", "-t", "--tui":
		return runTui(registry)
	case "list", "-l", "--list":
		printTable(registry)
		return nil
	}

	target := args[0]
	if len(args) < 2 {
		return fmt.Errorf("undefined action for control '%s'", target)
	}

	sysfsval, err := registry.Get(target, func(oldValue, newValue string) {
		fmt.Printf("%s: %s -> %s\n", target, oldValue, newValue)
	})
	if err != nil {
		return err
	}

	switch args[1] {
	case "choices", "list":
		fmt.Println(stringifyChoices(sysfsval.Choices(), sysfsval.Value()))
		return nil

	case "get":
		fmt.Println(sysfsval.Value())
		return nil

	case "set":
		if len(args) < 3 {
			return fmt.Errorf("undefined value to be set for %s", target)
		}

		valueToSet := args[2]

		if err := sysfsval.Set(valueToSet); err != nil {
			return err
		}

		if err := sysfsval.UpdateValue(); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("missing action for control '%[1]s'\nUsage syspower %[1]s <get|set|choices>", target)
	}
}

func stringifyChoices(choices []string, activeVal string) string {
	var formatted []string
	for _, choice := range choices {
		if choice == activeVal {
			formatted = append(formatted, fmt.Sprintf("[%s]", choice))
		} else {
			formatted = append(formatted, choice)
		}
	}
	return strings.Join(formatted, "  ")
}

func printTable(registry *syspower.Registry) {
	fmt.Printf("%-18s %-12s %s\n", "CONTROL", "VALUE", "CHOICES")
	fmt.Printf("%-18s %-12s %s\n", "-------", "-----", "-------")

	for _, name := range registry.List() {
		ctrl, err := registry.Get(name, nil)
		if err != nil {
			continue
		}
		fmt.Printf("%-18s %-12s %s\n", name, ctrl.Value(), stringifyChoices(ctrl.Choices(), ctrl.Value()))
	}
}

func printHelp() {
	fmt.Printf(`syspower - System power control utility

Usage:
	syspower <command>
	syspower <control> <action> [value]

Global Commands:
	help, -h, --help		Print this help
	list, -l, --list		List all available sysfs controls
	version, -v, --version		Print version

Control Actions:
	get				Get active control value
	choices, list			Show available control choices
	set <value>			Write a new value to the control
	`)
}
