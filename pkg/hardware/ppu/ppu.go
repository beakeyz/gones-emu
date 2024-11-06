package ppu

import (
	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/video"
)

type PPU struct {
    /* Bus for the PPU stuff (Pattern, Nametable, Pallets) */
    PpuBus *bus.SystemBus
    /* Videobackend used for actually drawing the PPU state to a screen */
    backend *video.VideoBackend
    /* Start address on the main system bus */
    start_addr uint16
    /* End address on the main system bus */
    end_addr uint16
    /* The current 'pixel' we're working on. This includes hblank and vblank */
    pixel_clock uint32
    /* Nes pallet array */
    nesPallet []video.Color
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
    _bus, err := bus.NewSystembus()

    if (err != nil) {
        return nil
    }

    pallet := make([]video.Color, 1)

    // Yuck, pallet generation
    pallet = append(pallet, video.NewColor(0x80, 0x80, 0x80, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x3D, 0xA6, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x12, 0xB0, 0xff))
    pallet = append(pallet, video.NewColor(0x44, 0x00, 0x96, 0xff))
    pallet = append(pallet, video.NewColor(0x1, 0x00, 0x5E, 0xff))
    pallet = append(pallet, video.NewColor(0xC7, 0x00, 0x28, 0xff))
    pallet = append(pallet, video.NewColor(0xBA, 0x06, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x8C, 0x17, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0xC, 0x2F, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x10, 0x45, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x05, 0x4A, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x47, 0x2E, 0xff))
    pallet = append(pallet, video.NewColor(0x0, 0x41, 0x66, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x00, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x05, 0x05, 0x05, 0xff))
    pallet = append(pallet, video.NewColor(0x05, 0x05, 0x05, 0xff))
    pallet = append(pallet, video.NewColor(0x7, 0xC7, 0xC7, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x77, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0x21, 0x55, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0x82, 0x37, 0xFA, 0xff))
    pallet = append(pallet, video.NewColor(0xB, 0x2F, 0xB5, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0x29, 0x50, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0x22, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0xD6, 0x32, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x4, 0x62, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x35, 0x80, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x05, 0x8F, 0x00, 0xff))
    pallet = append(pallet, video.NewColor(0x00, 0x8A, 0x55, 0xff))
    pallet = append(pallet, video.NewColor(0x0, 0x99, 0xCC, 0xff))
    pallet = append(pallet, video.NewColor(0x21, 0x21, 0x21, 0xff))
    pallet = append(pallet, video.NewColor(0x09, 0x09, 0x09, 0xff))
    pallet = append(pallet, video.NewColor(0x09, 0x09, 0x09, 0xff))
    pallet = append(pallet, video.NewColor(0xF, 0xFF, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0x0F, 0xD7, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0x69, 0xA2, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0xD4, 0x80, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0xF, 0x45, 0xF3, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0x61, 0x8B, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0x88, 0x33, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0x9C, 0x12, 0xff))
    pallet = append(pallet, video.NewColor(0xA, 0xBC, 0x20, 0xff))
    pallet = append(pallet, video.NewColor(0x9F, 0xE3, 0x0E, 0xff))
    pallet = append(pallet, video.NewColor(0x2B, 0xF0, 0x35, 0xff))
    pallet = append(pallet, video.NewColor(0x0C, 0xF0, 0xA4, 0xff))
    pallet = append(pallet, video.NewColor(0x5, 0xFB, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0x5E, 0x5E, 0x5E, 0xff))
    pallet = append(pallet, video.NewColor(0x0D, 0x0D, 0x0D, 0xff))
    pallet = append(pallet, video.NewColor(0x0D, 0x0D, 0x0D, 0xff))
    pallet = append(pallet, video.NewColor(0xF, 0xFF, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0xA6, 0xFC, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0xB3, 0xEC, 0xFF, 0xff))
    pallet = append(pallet, video.NewColor(0xDA, 0xAB, 0xEB, 0xff))
    pallet = append(pallet, video.NewColor(0xF, 0xA8, 0xF9, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0xAB, 0xB3, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0xD2, 0xB0, 0xff))
    pallet = append(pallet, video.NewColor(0xFF, 0xEF, 0xA6, 0xff))
    pallet = append(pallet, video.NewColor(0xF, 0xF7, 0x9C, 0xff))
    pallet = append(pallet, video.NewColor(0xD7, 0xE8, 0x95, 0xff))
    pallet = append(pallet, video.NewColor(0xA6, 0xED, 0xAF, 0xff))
    pallet = append(pallet, video.NewColor(0xA2, 0xF2, 0xDA, 0xff))
    pallet = append(pallet, video.NewColor(0x9, 0xFF, 0xFC, 0xff))
    pallet = append(pallet, video.NewColor(0xDD, 0xDD, 0xDD, 0xff))
    pallet = append(pallet, video.NewColor(0x11, 0x11, 0x11, 0xff))
    pallet = append(pallet, video.NewColor(0x11, 0x11, 0x11, 0xff))

    return &PPU{
        PpuBus: _bus,
        backend: backend,
        /* This is register space (TODO: The other spaces (nametables, chr rom, pallet ram)) */
        start_addr: 0x2000,
        end_addr: 0x2007,
        pixel_clock: 0,
        nesPallet: pallet,
    }
}

/*!
 * Called when the ppu does a 'cycle'
 */
func (ppu *PPU) Execute(ticks int) error {

    /* We can execute multiple PPU ticks in a single Execute call, in order to obtain ez syncing */
    for range ticks {

        ppu.backend.DrawNESPixel(int32(ppu.pixel_clock % PPU_CYCLES_PER_SCANLINE), int32(ppu.pixel_clock / PPU_CYCLES_PER_SCANLINE), ppu.nesPallet[ppu.pixel_clock % 0x40])

        /* We sadly can't do fancy AND stuff, since PPU_CYCLES_PER_SCREEN isn't a power of 2 =( */
        ppu.pixel_clock++;

        /* Beam the screen */
        if (ppu.pixel_clock >= PPU_CYCLES_PER_SCREEN) {
            ppu.pixel_clock = 0
            ppu.backend.Flush()
        }
    }
    return nil
}

/*
 * Handles CPU reads to the PPU
 */
func (ppu *PPU) Read(addr uint16, value *uint8) error {
    debug.Log("(PPU) Reading at %x\n", addr)

    // 
    switch (addr) {
    case 0x2000:
    case 0x2001:
    case 0x2002:
    case 0x2003:
    case 0x2004:
    case 0x2005:
    case 0x2006:
    case 0x2007:
        break;
    default:
        break;
    }
	return nil
}


/*
 * Handles CPU writes to the PPU
 */
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

/*
 * Read function for the PPU, on it's bus
 */
func (ppu *PPU) ppuRead(addr uint16, value *uint8) error {
    return ppu.PpuBus.Read(addr, value)
}

/*
 * Write function for the PPU, on it's bus
 */
func (ppu *PPU) ppuWrite(addr uint16, value uint8) error {
    return ppu.PpuBus.Write(addr, value)
}
