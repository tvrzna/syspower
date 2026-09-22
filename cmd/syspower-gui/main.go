package main

import (
	"fmt"
	"os"

	"github.com/tvrzna/syspower/internal/syspower"
	. "modernc.org/tk9.0"
	_ "modernc.org/tk9.0/themes/azure"
)

const guiVersion = "0.0.1"

func main() {
	processArgs(os.Args[1:])

	registry := syspower.NewRegistry()
	Pack(buildApp(registry))
	go registry.WatchChanges()
	ActivateTheme("azure dark")
	App.WmTitle("syspower-gui")
	App.Center().Wait()
}

func buildApp(registry *syspower.Registry) *TFrameWidget {
	r := TFrame()

	variables := make(map[string]*VariableOpt)

	for i, target := range registry.List() {
		if ctrl, err := registry.Get(target, func(oldValue, newValue string) {
			PostEvent(func() {
				if v, ok := variables[target]; ok {
					v.Set(newValue)
				}
			}, false)
		}); err == nil {
			radioFrame := r.TLabelframe(Txt(target), Padding(0))
			Grid(radioFrame, Row(i), Column(0), Padx(5), Pady(5), Sticky("nsew"))

			rv := Variable(ctrl.Value())
			variables[target] = rv
			for j, choice := range ctrl.Choices() {
				radio := radioFrame.TRadiobutton(Txt(choice), rv, Value(choice), Command(func() {
					if err := ctrl.Set(rv.Get()); err != nil {
						PostEvent(func() {
							rv.Set(ctrl.Value())
						}, false)
					}
				}))
				Grid(radio, Row(0), Column(j), Padx(5), Pady(5), Sticky("nsew"))
			}
		}
	}

	return r
}

func processArgs(args []string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Printf("syspower-gui %s (gui %s)\nhttps://github.com/tvrzna/syspower\n\nReleased under the MIT License.\n", syspower.GetVersion(), guiVersion)
		os.Exit(0)
	case "help", "-h", "--help":
		printHelp()
		os.Exit(0)
	}
}

func printHelp() {
	fmt.Printf(`syspower-gui - System power control utility with GUI

Usage:
	syspower-gui <command>

Global Commands:
	help, -h, --help		Print this help
	version, -v, --version		Print version
`)
}
