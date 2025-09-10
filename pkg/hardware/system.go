package hardware

import (
	"errors"
	"fmt"
	"time"

	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu/cpu6502"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/cartridge"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/ram"
	"github.com/beakeyz/gones-emu/pkg/hardware/mirror"
	"github.com/beakeyz/gones-emu/pkg/hardware/ppu"
	"github.com/beakeyz/gones-emu/pkg/video"
	"github.com/veandco/go-sdl2/sdl"
)

/*
 * Abstracts away the entire NES system
 * into a single struct.
 *
 * There are multiple chips which have their own clocks and work independently. Some of these also
 * have their own busses. For example, the PPU needs to raise an IRQ to the main 6502 CPU when it's
 * starting to vblank, so it needs access to the main CPU, but it also has it's own bus for accessing
 * hardware components.
 */
type NESSystem struct {
	/* 6502 CPU */
	MainCpu *cpu6502.CPU6502
	/* PPU */
	Ppu *ppu.PPU
	/* Regular system RAM */
	Ram *ram.Ram
	/* Bus */
	Bus *bus.SystemBus

	/* The backend */
	vbackend *video.VideoBackend

	/* How many system ticks have already been done */
	elapsedTicks uint64
}

func InitNesSystem(vidBackend *video.VideoBackend, cardridgePath string) (*NESSystem, error) {
	var err error
	var ret *NESSystem = nil
	var _cpu *cpu6502.CPU6502 = nil
	var _ppu *ppu.PPU = nil
	var _ram *ram.Ram = nil
	var _bus *bus.SystemBus = nil

	// Create the system bus for the CPU
	_bus, err = bus.NewSystembus()

	if err != nil {
		return nil, err
	}

	// Create a new CPU
	_cpu = cpu6502.New(_bus)

	if _cpu == nil {
		return nil, errors.New("Failed to create CPU for the system")
	}

	// Create the PPU
	_ppu = ppu.New(vidBackend)

	if _ppu == nil {
		return nil, errors.New("Failed to create PPU for the system")
	}

	// Add it's ram
	_ram = ram.New(0, 0x1fff, 0x0800)

	// Add the ram component
	_bus.AddComponent(_ram)

	// Add the PPU component
	_bus.AddComponent(_ppu)

	// Add the PPUs mirrors of the register space
	// TODO: Let the PPU module add its ranges on its own?
	for i := range 1023 {
		_bus.AddComponent(mirror.New(uint16(0x2008+(i*8)), uint16(0x200f+(i*8)), _ppu))
	}

	// Try to load the cardridge
	err = cartridge.LoadCardridge(_bus, _ppu.PpuBus, cardridgePath)

	// Fuck
	if err != nil {
		return nil, err
	}

	// Initialize the main CPU
	_cpu.Initialize()

	ret = &NESSystem{
		MainCpu:      _cpu,
		Ppu:          _ppu,
		Bus:          _bus,
		Ram:          _ram,
		vbackend:     vidBackend,
		elapsedTicks: 0,
	}

	return ret, nil
}

/*
 * Execute a single frame
 */
func (system *NESSystem) SystemFrame() error {

	var err error
	var cpuCyclesElapsed int

	/* Do a single CPU cycle */
	err = system.MainCpu.ExecuteFrame(&cpuCyclesElapsed)

	if err != nil {
		return err
	}

	/* Do three PPU cycles, to comply with relative component speed */
	system.Ppu.Execute(cpuCyclesElapsed * 3)

	/* Increment the system ticks */
	system.elapsedTicks++

	return nil
}

func (system *NESSystem) displayDebugInfo() {
	var b *video.VideoBackend = system.vbackend

	s := fmt.Sprintf(
		"A: 0x%x, Flags: 0x%x",
		system.MainCpu.GetAccumulator(),
		system.MainCpu.GetFlags(),
	)

	b.DrawText(500, 25, s, video.ColorWhite())
}

func (system *NESSystem) preDraw() {

	// Draw the background

	system.vbackend.UpdateBackground()

	// Draw the title text

	system.vbackend.DrawText(0, 0, "GONES!", video.ColorWhite())

	// Draw some debug info

	system.displayDebugInfo()
}

func (system *NESSystem) StartLoop() {

	running := true
	ran_tick := false

	for running {

		event := system.vbackend.CollectEvent()

		switch event.(type) {
		case *sdl.QuitEvent:
			running = false
		}

		system.preDraw()

		if system.vbackend.IsKeyPressed(sdl.K_RETURN) && !ran_tick {
			err := system.SystemFrame()

			if err != nil {
				break
			}

			ran_tick = true
		}

		if !system.vbackend.IsKeyPressed(sdl.K_RETURN) && ran_tick {
			ran_tick = false
		}

		system.vbackend.Flush()

		time.Sleep(time.Millisecond)
	}
}
