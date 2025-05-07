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
	/* PPU registers */
	ctl_register    byte
	mask_register   byte
	status_register byte

	/* Start address on the main system bus */
	start_addr uint16
	/* End address on the main system bus */
	end_addr uint16
	/* The current 'pixel' we're working on. This includes hblank and vblank */
	pixel_clock uint32
	pixel_x     int32
	pixel_y     int32
	/* Nes pallet array */
	nesPallet []video.Color
}

const (
	PPU_CTL     = 0x2000
	PPU_MASK    = 0x2001
	PPU_STATUS  = 0x2002
	PPU_OAMADDR = 0x2003
	PPU_OAMDATA = 0x2004
	PPU_SCROLL  = 0x2005
	PPU_ADDR    = 0x2006
	PPU_DATA    = 0x2007

	/* PPU CTL bit fields */
	PPU_CTL_NTADDR_MASK             = 0x01 | 0x02
	PPU_CTL_ADDR_INC                = 0x04
	PPU_CTL_SPRITE_PATTERN_TBL_ADDR = 0x08
	PPU_CTL_BG_PATTERN_TBL_ADDR     = 0x10
	PPU_CTL_SPRITE_SZ_8x16          = 0x20
	PPU_CTL_MSTR_SLV_SELECT         = 0x40 // 0 = Master, 1 = Slave
	PPU_CTL_NMI_ON_VBLANK           = 0x80

	/* PPU MASK bit fields */
	PPU_MASK_DISPLAY_TYPE       = 0x01
	PPU_MASK_BG_SHOW_LEFT8      = 0x02
	PPU_MASK_SPRITES_SHOW_LEFT8 = 0x04
	PPU_MASK_RENDER_BG          = 0x08
	PPU_MASK_RENDER_SPRITES     = 0x10
	PPU_MASK_CLR_INTENSITY_MASK = 0x20 | 0x40 | 0x80
	PPU_MASK_FULL_BG_CLR_MASK   = 0x20 | 0x40 | 0x80

	/* PPU STATUS bit fields */
	PPU_STATUS_IGNORE_VRAM_WRITES = 0x10
	PPU_STATUS_SPRITE_OVERFLOW    = 0x20
	PPU_STATUS_HIT_SPRITE0        = 0x40
	PPU_STATUS_IN_VBLANK          = 0x80

	/* 16 Kilobytes of addressable PPU memory */
	PPU_MEM_SZ = 16 * 1024

	PPU_CHR_ROM_BASE = 0x0000
	PPU_CHR_ROM_SZ   = 8 * 1024
	PPU_CHR_ROM_END  = PPU_CHR_ROM_BASE + PPU_CHR_ROM_SZ

	PPU_VRAM_BASE = 0x2000
	PPU_VRAM_SZ   = 4 * 1024
	PPU_VRAM_END  = PPU_VRAM_BASE + 8*1024 - 256

	PPU_PALETTE_BASE = 0x3F00
	PPU_PALETTE_SZ   = 32
	PPU_PALETTE_END  = PPU_PALETTE_BASE + PPU_PALETTE_SZ*8

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

	if err != nil {
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
		PpuBus:          _bus,
		backend:         backend,
		ctl_register:    0,
		mask_register:   0,
		status_register: 0,
		/* This is register space (TODO: The other spaces (nametables, chr rom, pallet ram)) */
		start_addr:  0x2000,
		end_addr:    0x2007,
		pixel_clock: 0,
		pixel_x:     0,
		pixel_y:     0,
		nesPallet:   pallet,
	}
}

func (ppu *PPU) EnteredVBlank() bool {
	return (ppu.pixel_y == 241 && ppu.pixel_x == 1)
}

func (ppu *PPU) IsVBlank() bool {
	return (ppu.pixel_y >= 240 && ppu.pixel_y != 261)
}

func (ppu *PPU) IsHBlank() bool {
	return (ppu.pixel_x >= 257 && ppu.pixel_x <= 320)
}

/*!
 * Called when the ppu does a 'cycle'
 */
func (ppu *PPU) Execute(ticks int) error {

	/* We can execute multiple PPU ticks in a single Execute call, in order to obtain ez syncing */
	for range ticks {
		/* Set the current pixel coordinates */
		ppu.pixel_x = int32(ppu.pixel_clock % PPU_CYCLES_PER_SCANLINE)
		ppu.pixel_y = int32(ppu.pixel_clock / PPU_CYCLES_PER_SCANLINE)

		// If we're inside the visible vertical region
		if !ppu.IsVBlank() {

			if ppu.IsHBlank() {

			} else {

				ppu.backend.DrawNESPixel(
					ppu.pixel_x,
					ppu.pixel_y,
					ppu.nesPallet[(ppu.pixel_clock/PPU_CYCLES_PER_SCANLINE)%0x40])
			}

		} else { // Post-render and VBlank
			if ppu.EnteredVBlank() {
				// Set vblank flag
				ppu.SetStatusBits(PPU_STATUS_IN_VBLANK)

				// TODO: Signal CPU NMI if that is set in the PPU control word
			}
		}

		/* We sadly can't do fancy AND stuff, since PPU_CYCLES_PER_SCREEN isn't a power of 2 =( */
		ppu.pixel_clock++

		/* Beam the screen */
		if ppu.pixel_clock >= PPU_CYCLES_PER_SCREEN {
			ppu.pixel_clock = 0
			ppu.backend.Flush()
		}
	}
	return nil
}

func (ppu *PPU) PostFrame() {

}

func (ppu *PPU) ClearStatusBits(bits uint8) {
	ppu.status_register &= ^bits
}

func (ppu *PPU) SetStatusBits(bits uint8) {
	ppu.status_register |= bits
}

/*
 * Handles CPU reads to the PPU
 */
func (ppu *PPU) Read(addr uint16, value *uint8) error {
	debug.Log("(PPU) Reading at 0x%x\n", addr)

	//
	switch addr {
	case 0x2000:
		*value = ppu.ctl_register
		break
	case 0x2001:
		*value = ppu.mask_register
		break
	case 0x2002:
		*value = ppu.status_register

		ppu.ClearStatusBits(PPU_STATUS_IN_VBLANK)
		break
	case 0x2003:
	case 0x2004:
	case 0x2005:
	case 0x2006:
	case 0x2007:
		break
	default:
		break
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
