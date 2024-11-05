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
	_bus.AddComponent(ram.New(0, 0x0800, 2048))
    _bus.AddComponent(_ppu)

    // Add the PPUs mirrors of the register space
    // TODO: Let the PPU module add its ranges on its own?
    for i := range(1023) {
        _bus.AddComponent(mirror.New(uint16(0x2008 + (i * 8)), uint16(0x200f + (i * 8)), _ppu))
    }

    // Try to load the cardridge
    err = cartridge.LoadCardridge(_bus, cardridgePath)

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
        elapsedTicks: 0,
    }

    return ret, nil
}

func (system *NESSystem) SystemTick() {

    /* Do three PPU cycles, to comply with relative component speed */
    system.Ppu.Execute(3);

    /* Do a single CPU cycle */
    system.MainCpu.DoCycle()

    /* Increment the system ticks */
    system.elapsedTicks++
}
