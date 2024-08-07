package mirror

import (
	"fmt"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware"
)

/*
 */
type Mirror struct {
	c     *hardware.Component
	start uint16
	end   uint16
}

func New(start uint16, end uint16, c hardware.Component) *Mirror {
	var r Mirror = Mirror{
		start: start,
		end:   end,
		c:     &c,
	}

	return &r
}

func (m *Mirror) Read(addr uint16, value *uint8) error {
	var delta uint16
	var mirror_addr uint16

	if addr < m.start || addr > m.end {
		return fmt.Errorf("mirror: tried to access out of range")
	}

	delta = addr - m.start
	mirror_addr = (*m.c).StartAddr() + delta

	if mirror_addr > (*m.c).EndAddr() {
		return fmt.Errorf("mirror: read outside child components reach")
	}

    debug.Log("Reading at %x -> %x\n", addr, mirror_addr)

	// Redirect the read command
	return (*m.c).Read(mirror_addr, value)
}

func (m *Mirror) Write(addr uint16, value uint8) error {
	var delta uint16
	var mirror_addr uint16

	if addr < m.start || addr > m.end {
		return fmt.Errorf("mirror: tried to access out of range")
	}

	delta = addr - m.start
	mirror_addr = (*m.c).StartAddr() + delta

	if mirror_addr > (*m.c).EndAddr() {
		return fmt.Errorf("mirror: write outside child components reach")
	}

	// Redirect the read command
	return (*m.c).Write(mirror_addr, value)
}

func (m *Mirror) StartAddr() uint16 {
	return m.start
}

func (m *Mirror) EndAddr() uint16 {
	return m.end
}
