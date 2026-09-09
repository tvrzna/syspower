package main

import (
	"fmt"
	"os"
	"slices"
	"time"

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
	x, y         int
	updateScreen bool
	lastErr      error
	statusMsg    string
}

func startKeyReader(keyChan chan<- keyAction) {
	go func() {
		buf := make([]byte, 3)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				keyChan <- QUIT
			}

			if n == 1 {
				switch buf[0] {
				case 'q', 3:
					keyChan <- QUIT
				case 'h':
					keyChan <- LEFT
				case 'j':
					keyChan <- UP
				case 'k':
					keyChan <- DOWN
				case 'l':
					keyChan <- RIGHT
				}
			} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
				switch buf[2] {
				case 68:
					keyChan <- LEFT
				case 65:
					keyChan <- UP
				case 66:
					keyChan <- DOWN
				case 67:
					keyChan <- RIGHT
				}
			}
		}
	}()
}

func runTui(registry *syspower.Registry) error {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set raw terminal: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(fd, oldState)

	selector := &selector{-1, 0, true, nil, ""}

	keyChan := make(chan keyAction, 1)
	updateChan := make(chan bool, 1)

	for _, target := range registry.List() {
		registry.Get(target, func(oldValue, newValue string) {
			if !selector.updateScreen {
				selector.updateScreen = true
			} else {
				updateChan <- true
			}
			selector.statusMsg = fmt.Sprintf("%s: %s -> %s", target, oldValue, newValue)
		})
	}

	go checkContent(registry)
	go startKeyReader(keyChan)

	for {
		if err := printContent(registry, selector); err != nil {
			return err
		}

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

func doMovement(registry *syspower.Registry, selector *selector, keyAction keyAction) {
	maxY := len(registry.ListCached())
	if maxY > 0 {
		maxY--
	}

	target := registry.ListCached()[selector.y]
	ctrl, err := registry.Get(target, nil)
	if err != nil {
		return
	}
	maxX := len(ctrl.Choices()) - 1
	selector.x = slices.Index(ctrl.Choices(), ctrl.Value())

	switch keyAction {
	case UP:
		selector.y--
		if selector.y < 0 {
			selector.y = maxY
		}
	case DOWN:
		selector.y++
		if selector.y > maxY {
			selector.y = 0
		}
	case LEFT:
		selector.x--
		if selector.x < 0 {
			selector.x = maxX
		}
	case RIGHT:
		selector.x++
		if selector.x > maxX {
			selector.x = 0
		}
	}

	if keyAction == LEFT || keyAction == RIGHT {
		selector.lastErr = ctrl.Set(ctrl.Choices()[selector.x])
		selector.updateScreen = false
		ctrl.UpdateValue()
	}
}

func checkContent(registry *syspower.Registry) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		for _, name := range registry.ListCached() {
			if ctrl, err := registry.Get(name, nil); err == nil {
				ctrl.UpdateValue()
			}
		}
	}
}

func printContent(registry *syspower.Registry, selector *selector) error {
	fmt.Print("\x1b[H\x1b[2J")

	fmt.Print("==========================================================================\r\n")
	fmt.Print("  syspower - System power control utility\r\n")
	fmt.Print("==========================================================================\r\n")

	fmt.Printf("   %-18s %-12s %s\r\n", "CONTROL", "VALUE", "CHOICES")
	fmt.Printf("   %-18s %-12s %s\r\n", "-------", "-----", "-------")
	for i, name := range registry.ListCached() {
		selected := "  "
		if i == selector.y {
			selected = "->"
		}

		ctrl, err := registry.Get(name, nil)
		if err != nil {
			continue
		}
		fmt.Printf("%s %-18s %-12s %s\r\n", selected, name, ctrl.Value(), stringifyChoices(ctrl.Choices(), ctrl.Value()))
	}

	statusMsg := ""
	if selector.lastErr != nil {
		statusMsg = fmt.Sprintf(" Error:  %v", selector.lastErr)
		selector.lastErr = nil
	} else if selector.statusMsg != "" {
		statusMsg = fmt.Sprintf(" Status: %s", selector.statusMsg)
		selector.statusMsg = ""
	}
	fmt.Print("--------------------------------------------------------------------------\r\n")
	fmt.Printf("%s\r\n", statusMsg)
	fmt.Print(" Keys:   [↑/↓][j/k] Select Row  |  [←/→][h/l] Change Option  |  [q] Quit\r\n")
	fmt.Print("==========================================================================\r\n")

	return nil
}
