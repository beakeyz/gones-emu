package hardware

import (
	"errors"

	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu/cpu6502"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/cartridge"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/ram"
	"github.com/beakeyz/gones-emu/pkg/hardware/mirror"
	"github.com/beakeyz/gones-emu/pkg/hardware/ppu"
	"github.com/beakeyz/gones-emu/pkg/video"
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
    _bus, err = bus.NewSystembus();

    if (err != nil) {
        return nil, err
    }

    // Create a new CPU
    _cpu = cpu6502.New(_bus)

    if (_cpu == nil) {
        return nil, errors.New("Failed to create CPU for the system")
    }

    // Create the PPU
    _ppu = ppu.New(vidBackend)

    if (_ppu == nil) {
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
    for i := range(1023) {
        _bus.AddComponent(mirror.New(uint16(0x2008 + (i * 8)), uint16(0x200f + (i * 8)), _ppu))
    }

    // Try to load the cardridge
    err = cartridge.LoadCardridge(_bus, _ppu.PpuBus, cardridgePath)

    // Fuck
    if (err != nil) {
        return nil, err
    }

    // Initialize the main CPU
    _cpu.Initialize()

    ret = &NESSystem{
        MainCpu: _cpu,
        Ppu: _ppu,
        Bus: _bus,
        Ram: _ram,
        elapsedTicks: 0,
    }

    return ret, nil
}

func (system *NESSystem) SystemTick() {

    /* Do three PPU cycles, to comply with relative component speed */
    system.Ppu.Execute(1);

    /* Do a single CPU cycle */
    if (system.elapsedTicks % (ppu.PPU_CYCLES_PER_CPU_CYCLE) == 0) {
        system.MainCpu.DoCycle()
    }

    /* Increment the system ticks */
    system.elapsedTicks++
}
