package main

import (
	"fmt"
	"strings"

	"github.com/tvrzna/syspower/internal/syspower"
)

func runCli(args []string, registry *syspower.Registry) error {
	switch args[0] {
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
