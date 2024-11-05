package main

import (
	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware"
	"github.com/beakeyz/gones-emu/pkg/video"
)

func main() {
	var err error
    // Video backend for drawing what the PPU wants
    var vidBackend video.VideoBackend
    var nes *hardware.NESSystem

	// Enable debugging
	debug.Enable()

    // Initialize the video backend
	err = video.InitVideo(&vidBackend)

    if (err != nil) {
        debug.Error("Failed to initialize video")
        return
    }

    nes, err = hardware.InitNesSystem(&vidBackend, "res/Super Mario Bros (E).nes")

	for true {
        nes.SystemTick()
	}
}
