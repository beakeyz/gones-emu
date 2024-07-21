package ram

import (
	"errors"
)

/*
 * Little endian RAM module
 */
type Ram struct {
	start_addr uint16
	end_addr   uint16
	size       uint32
	memory     []byte
}

func New(start uint16, end uint16, size uint32) *Ram {
	var r Ram = Ram{
		start_addr: start,
		end_addr:   end,
		size:       size,
		memory:     make([]byte, size),
	}

	return &r
}

// 0x00ff -> ff 00
func (ram *Ram) Read(addr uint16, value *uint8) error {
	if value == nil {
		return errors.New("ram: null value buffer")
	}

	if addr > ram.end_addr || addr < ram.start_addr {
		return errors.New("ram: read out of range!")
	}

	*value = ram.memory[addr-ram.start_addr]
	return nil
}

func (ram *Ram) Write(addr uint16, value uint8) error {
	if addr > ram.end_addr || addr < ram.start_addr {
		return errors.New("ram: write out of range!")
	}

	ram.memory[addr-ram.start_addr] = byte(value)
	ram.memory[addr] = byte(value)
	return nil
}

func (ram *Ram) StartAddr() uint16 {
	return ram.start_addr
}

func (ram *Ram) EndAddr() uint16 {
	return ram.end_addr
}
