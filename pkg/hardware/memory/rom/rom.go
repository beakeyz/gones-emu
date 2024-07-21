package rom

import (
	"errors"
)

/*
 * Little endian ROM module
 */
type Rom struct {
	start_addr uint16
	end_addr   uint16
	size       uint32
	memory     []byte
}

func New(start uint16, end uint16, size uint32, data []byte) *Rom {
	var r Rom = Rom{
		start_addr: start,
		end_addr:   end,
		size:       size,
		memory:     data,
	}

	return &r
}

// 0x00ff -> ff 00
func (rom *Rom) Read(addr uint16, value *uint8) error {
	if value == nil {
		return errors.New("rom: null value buffer")
	}

	if addr > rom.end_addr || addr < rom.start_addr {
		return errors.New("rom: read out of range!")
	}

	*value = rom.memory[addr-rom.start_addr]
	return nil
}

func (rom *Rom) Write(addr uint16, value uint8) error {
	return errors.New("rom: tried to write to rom")
}

func (rom *Rom) StartAddr() uint16 {
	return rom.start_addr
}

func (rom *Rom) EndAddr() uint16 {
	return rom.end_addr
}
