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
	UP keyAction = iota + 1
	DOWN
	LEFT
	RIGHT
	NONE
	QUIT
)

type selector struct {
	y          int
	skipUpdate bool
	statusMsg  string
}

func runTui(registry *syspower.Registry) error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw terminal: %v\n", err)
	}
	defer term.Restore(fd, oldState)

	selector := &selector{y: 0}

	keyChan := make(chan keyAction, 1)
	updateChan := make(chan bool, 1)

	for _, target := range registry.List() {
		targetName := target
		registry.Get(targetName, func(oldValue, newValue string) {
			if !selector.skipUpdate {
				updateChan <- true
			}
			selector.skipUpdate = false
			selector.statusMsg = fmt.Sprintf(" Status: %s: %s -> %s", target, oldValue, newValue)
		})
	}

	go registry.WatchChanges()
	go startKeyReader(keyChan)

	writer := bufio.NewWriter(os.Stdout)
	for {
		printContent(writer, registry, selector)

		select {
		case keyAction := <-keyChan:
			if keyAction == QUIT {
				return nil
			}
			doMovement(registry, selector, keyAction)
		case <-updateChan:
			continue
		}
	}
}

func startKeyReader(keyChan chan<- keyAction) {
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			keyChan <- QUIT
			return
		}

		switch string(buf[:n]) {
		case "q", "\x03":
			keyChan <- QUIT
		case "h", "\x1b[D":
			keyChan <- LEFT
		case "j", "\x1b[B":
			keyChan <- DOWN
		case "k", "\x1b[A":
			keyChan <- UP
		case "l", "\x1b[C":
			keyChan <- RIGHT
		}
	}
}

func doMovement(registry *syspower.Registry, selector *selector, keyAction keyAction) {
	list := registry.ListCached()
	listLen := len(list)
	if listLen == 0 {
		return
	}

	switch keyAction {
	case UP:
		selector.y = (selector.y - 1 + listLen) % listLen
	case DOWN:
		selector.y = (selector.y + 1) % listLen
	case LEFT, RIGHT:
		ctrl, err := registry.Get(list[selector.y], nil)
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

		if keyAction == LEFT {
			currIdx = (currIdx - 1 + choicesLen) % choicesLen
		} else {
			currIdx = (currIdx + 1) % choicesLen
		}

		if err := ctrl.Set(choices[currIdx]); err != nil {
			selector.statusMsg = fmt.Sprintf(" Error: %v", err)
		} else {
			selector.skipUpdate = true
		}
		ctrl.UpdateValue()
	}
}

func printContent(w *bufio.Writer, registry *syspower.Registry, selector *selector) {
	w.WriteString("\x1b[H\x1b[2J")
	w.WriteString("==========================================================================\r\n")
	w.WriteString("  syspower - System power control utility\r\n")
	w.WriteString("==========================================================================\r\n")

	fmt.Fprintf(w, "   %-18s %-12s %s\r\n", "CONTROL", "VALUE", "CHOICES")
	fmt.Fprintf(w, "   %-18s %-12s %s\r\n", "-------", "-----", "-------")
	for i, name := range registry.ListCached() {
		selected := "  "
		if i == selector.y {
			selected = "->"
		}

		ctrl, err := registry.Get(name, nil)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "%s %-18s %-12s %s\r\n", selected, name, ctrl.Value(), stringifyChoices(ctrl.Choices(), ctrl.Value()))
	}

	w.WriteString("--------------------------------------------------------------------------\r\n")
	fmt.Fprintf(w, "%s\r\n", selector.statusMsg)
	w.WriteString(" Keys:   [↑/↓][k/j] Select Row  |  [←/→][h/l] Change Option  |  [q] Quit\r\n")
	w.WriteString("==========================================================================\r\n")
	selector.statusMsg = ""
	w.Flush()
}
