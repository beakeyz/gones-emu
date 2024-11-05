package ppu

import (
	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/video"
)

type PPU struct {
    /* Videobackend used for actually drawing the PPU state to a screen */
    backend *video.VideoBackend
    /* Start address on the main system bus */
    start_addr uint16
    /* End address on the main system bus */
    end_addr uint16
    /* The current 'pixel' we're working on. This includes hblank and vblank */
    pixel_clock int
}

const (
    PPU_CTL = 0x2000
    PPU_MASK = 0x2001
    PPU_STATUS = 0x2002
    PPU_OAMADDR = 0x2003
    PPU_OAMDATA = 0x2004
    PPU_SCROLL = 0x2005
    PPU_ADDR = 0x2006
    PPU_DATA = 0x2007

    /*
     * We're gonna emulate the NTSC video signal, which gives us
     * this neat attribute
     */
    PPU_CYCLES_PER_CPU_CYCLE = 3
    /* Don't worry about where I got this from, just trust me */
    PPU_CYCLES_PER_SCANLINE = 341
    /* Don't worry about where I got this from, just trust me v2 */
    PPU_CYCLES_PER_SCREEN = 89342
)

func New(backend *video.VideoBackend) *PPU {
    return &PPU{
        backend: backend,
        /* This is register space (TODO: The other spaces (nametables, chr rom, pallet ram)) */
        start_addr: 0x2000,
        end_addr: 0x2007,
    }
}

/*!
 * Called when the ppu does a 'cycle'
 */
func (ppu *PPU) Execute(ticks int) error {

    /* We can execute multiple PPU ticks in a single Execute call, in order to obtain ez syncing */
    for range ticks {


        /* We sadly can't do fancy AND stuff, since PPU_CYCLES_PER_SCREEN isn't a power of 2 =( */
        ppu.pixel_clock = (ppu.pixel_clock + 1) % PPU_CYCLES_PER_SCREEN;
    }
    return nil
}

func (ppu *PPU) Read(addr uint16, value *uint8) error {
    debug.Log("(PPU) Reading at %x\n", addr)
	return nil
}

func (ppu *PPU) Write(addr uint16, value uint8) error {
    debug.Log("(PPU) Writing at %x\n", addr)
	return nil
}

func (ppu *PPU) StartAddr() uint16 {
	return ppu.start_addr
}

func (ppu *PPU) EndAddr() uint16 {
	return ppu.end_addr
}
