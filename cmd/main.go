package main

import (
	"fmt"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu/cpu6502"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/cartridge"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/ram"
)

func main() {
	var err error
	var sysbus *bus.SystemBus
	var r *ram.Ram
	var c *cpu6502.CPU6502

	// Enable debugging
	debug.Enable()

	sysbus, err = bus.NewSystembus()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Initialize the CPU
	c = cpu6502.New(sysbus)

	if c == nil {
		debug.Error("Failed to create CPU")
		return
	}

	// Add it's ram
	r = ram.New(0, 0x0800, 2048)

	if r == nil {
		debug.Error("Failed to create RAM")
		return
	}

	sysbus.AddComponent(r)

	// Load a cartridge
	err = cartridge.LoadCardridge(sysbus, "./res/Super Mario Bros (E).nes")

	if err != nil {
		debug.Error(err.Error())
		return
	}

	/* Initialize this bitch */
	c.Initialize()

	for true {
		err = c.DoCycle()

		if err != nil {
			debug.Error(err.Error())
			break
		}
	}
}
