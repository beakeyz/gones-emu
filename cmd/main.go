package main

import (
	"fmt"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu/cpu6502"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/cartridge"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/ram"
	"github.com/beakeyz/gones-emu/pkg/hardware/mirror"
	"github.com/beakeyz/gones-emu/pkg/hardware/ppu"
	"github.com/beakeyz/gones-emu/pkg/video"
)

func main() {
	var err error
    // System bus for handling generic read and write operations
	var sysbus *bus.SystemBus
    // Video backend for drawing what the PPU wants
    var vidBackend video.VideoBackend
    // Main 6502 CPU for executing cartridge code
	var c *cpu6502.CPU6502
    // Pixel processing unit
    var p *ppu.PPU

	// Enable debugging
	debug.Enable()

    // Initialize the video backend
	err = video.InitVideo(&vidBackend)

    if (err != nil) {
        debug.Error("Failed to initialize video")
        return
    }

    // Create the system bus
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
	sysbus.AddComponent(ram.New(0, 0x0800, 2048))

    // Create the ppu
    p = ppu.New(&vidBackend, 0x2000, 0x2007)

    // Try to add the ppu
    sysbus.AddComponent(p)

    for i := range(1023) {
        sysbus.AddComponent(mirror.New(uint16(0x2008 + (i * 8)), uint16(0x200f + (i * 8)), p))
    }

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
