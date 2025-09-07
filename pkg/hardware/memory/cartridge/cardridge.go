package cartridge

import (
	"fmt"
	"os"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/memory/rom"
	"github.com/beakeyz/gones-emu/pkg/hardware/mirror"
)

type NESFileHeader struct {
	sig       [4]byte
	prgrom_sz int
	chrrom_sz int
	flags     [5]byte
	reserved  [5]byte
}

func newNesHeader(data []byte) NESFileHeader {
	ret := NESFileHeader{}

	ret.sig[0] = data[0]
	ret.sig[1] = data[1]
	ret.sig[2] = data[2]
	ret.sig[3] = data[3]

	ret.prgrom_sz = int(data[4]) * 16384

	if ret.prgrom_sz == 0 {
		ret.prgrom_sz = 16384
	}

	ret.chrrom_sz = int(data[5]) * 8192

	ret.flags[0] = data[6]
	ret.flags[1] = data[7]

	return ret
}

func LoadCardridge(cpuBus *bus.SystemBus, ppuBus *bus.SystemBus, filepath string) error {
	var f *os.File
	var err error
	var buffer []byte = make([]byte, 16)
	var prg_buffer []byte
	var chr_buffer []byte

	// Try to open the (hopefully) ROM file
	f, err = os.Open(filepath)

	if err != nil {
		return err
	}

	// Murder the file
	defer f.Close()

	_, err = f.Read(buffer)

	if err != nil {
		return err
	}

	if buffer[0] != 'N' || buffer[1] != 'E' || buffer[2] != 'S' {
		return fmt.Errorf("Invalid file loaded : %c%c%c", buffer[0], buffer[1], buffer[2])
	}

	header := newNesHeader(buffer)

	mapperNumber := (header.flags[1] & 0xF0) | ((header.flags[0] & 0xF0) >> 4)

	debug.Log("Got mapper number: %d\n", mapperNumber)

	read_off := 16

	if (header.flags[0] & 0x3) == 0x3 {
		read_off += 512
	}

	// Make the buffer
	prg_buffer = make([]byte, header.prgrom_sz)

	// Seek the offset for the ROM section
	_, err = f.ReadAt(prg_buffer, int64(read_off))

	if err != nil {
		return err
	}

	read_off += header.prgrom_sz
	chr_buffer = make([]byte, header.chrrom_sz)

	_, err = f.ReadAt(chr_buffer, int64(read_off))

	if err != nil {
		return err
	}

	switch mapperNumber {
	case 0:
		var prg_rom *rom.Rom
		var chr_rom *rom.Rom

		prgbase := 0x8000
		prg_size := header.prgrom_sz

		prg_rom = rom.New(uint16(prgbase), uint16(prgbase+prg_size)-1, uint32(prg_size), prg_buffer)
		chr_rom = rom.New(0, uint16(header.chrrom_sz)-1, uint32(header.chrrom_sz), chr_buffer)

		// Add an extra mirror, to account for the 16k ROM
		if prg_size == 0x4000 {
			mirror := mirror.New(uint16(prgbase+prg_size), uint16(prgbase+2*prg_size), prg_rom)

			cpuBus.AddComponent(mirror)
		}

		// Add the actual program rom component
		cpuBus.AddComponent(prg_rom)

		debug.Log("PRG size: 0x%x (0x%x -> 0x%x)\n", header.prgrom_sz, prg_rom.StartAddr(), prg_rom.EndAddr())

		// Add the character rom to the PPU bus
		ppuBus.AddComponent(chr_rom)
	default:
		return fmt.Errorf("found unimplemented mapper")

	}

	return nil
}
