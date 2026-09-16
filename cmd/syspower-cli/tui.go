package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"

	"github.com/tvrzna/syspower/internal/syspower"
	"golang.org/x/term"
)

type keyAction byte

const (
	NONE keyAction = iota
	UP
	DOWN
	LEFT
	RIGHT
	QUIT
)

type event struct {
	action   keyAction
	target   string
	oldValue string
	newValue string
}

type selector struct {
	y         int
	statusMsg string
}

func runTui(registry *syspower.Registry) error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw terminal: %w", err)
	}
	defer term.Restore(fd, oldState)

	eventChan := make(chan event, 16)

	for _, target := range registry.List() {
		targetName := target
		registry.Get(targetName, func(oldValue, newValue string) {
			select {
			case eventChan <- event{NONE, targetName, oldValue, newValue}:
			default:
			}
		})
	}

	go registry.WatchChanges()
	go startKeyReader(eventChan)

	sel := &selector{y: 0}
	writer := bufio.NewWriter(os.Stdout)
	for {
		printContent(writer, registry, sel)

		e := <-eventChan
		switch e.action {
		case QUIT:
			return nil
		case NONE:
			if e.target != "" {
				sel.statusMsg = fmt.Sprintf(" Status: %s: %s -> %s", e.target, e.oldValue, e.newValue)
			}
		default:
			doMovement(registry, sel, e.action)
		}
	}
}

func startKeyReader(eventChan chan<- event) {
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			eventChan <- event{action: QUIT}
			return
		}

		var key keyAction
		switch string(buf[:n]) {
		case "q", "\x03":
			key = QUIT
		case "h", "\x1b[D":
			key = LEFT
		case "j", "\x1b[B":
			key = DOWN
		case "k", "\x1b[A":
			key = UP
		case "l", "\x1b[C":
			key = RIGHT
		default:
			continue
		}
		eventChan <- event{action: key}
	}
}

func doMovement(registry *syspower.Registry, sel *selector, action keyAction) {
	list := registry.ListCached()
	listLen := len(list)
	if listLen == 0 {
		return
	}

	switch action {
	case UP:
		sel.y = (sel.y - 1 + listLen) % listLen
	case DOWN:
		sel.y = (sel.y + 1) % listLen
	case LEFT, RIGHT:
		ctrl, err := registry.Get(list[sel.y], nil)
		if err != nil {
			return
		}

		choices := ctrl.Choices()
		choicesLen := len(choices)
		if choicesLen == 0 {
			return
		}

		currIdx := slices.Index(choices, ctrl.Value())
		if currIdx == -1 {
			currIdx = 0
		}

		if action == LEFT {
			currIdx = (currIdx - 1 + choicesLen) % choicesLen
		} else {
			currIdx = (currIdx + 1) % choicesLen
		}

		if err := ctrl.Set(choices[currIdx]); err != nil {
			sel.statusMsg = fmt.Sprintf(" Error: %v", err)
		}
		ctrl.UpdateValue()
	}
}

func printContent(w *bufio.Writer, registry *syspower.Registry, sel *selector) {
	w.WriteString("\x1b[H\x1b[2J")
	w.WriteString("==========================================================================\r\n")
	w.WriteString("  syspower - System power control utility\r\n")
	w.WriteString("==========================================================================\r\n")

	fmt.Fprintf(w, "   %-18s %-12s %s\r\n", "CONTROL", "VALUE", "CHOICES")
	fmt.Fprintf(w, "   %-18s %-12s %s\r\n", "-------", "-----", "-------")
	for i, name := range registry.ListCached() {
		selected := "  "
		if i == sel.y {
			selected = "->"
		}

		ctrl, err := registry.Get(name, nil)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "%s %-18s %-12s %s\r\n", selected, name, ctrl.Value(), stringifyChoices(ctrl.Choices(), ctrl.Value()))
	}

	w.WriteString("--------------------------------------------------------------------------\r\n")
	fmt.Fprintf(w, "%s\r\n", sel.statusMsg)
	w.WriteString(" Keys:   [↑/↓][k/j] Select Row  |  [←/→][h/l] Change Option  |  [q] Quit\r\n")
	w.WriteString("==========================================================================\r\n")
	sel.statusMsg = ""
	w.Flush()
}
