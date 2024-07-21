package bus

import (
	"errors"

	"github.com/beakeyz/gones-emu/pkg/hardware"
)

type SystemBus struct {
	components []hardware.Component
}

func NewSystembus() (*SystemBus, error) {
	bus := SystemBus{
		components: make([]hardware.Component, 0),
	}

	return &bus, nil
}

func (bus *SystemBus) AddComponent(comp hardware.Component) {
	bus.components = append(bus.components, comp)
}

func (bus *SystemBus) getComponent(addr uint16) *hardware.Component {
	for _, comp := range bus.components {
		if addr >= comp.StartAddr() && addr <= comp.EndAddr() {
			return &comp
		}
	}

	return nil
}

func (bus *SystemBus) Read(addr uint16, value *uint8) error {
	var comp *hardware.Component = bus.getComponent(addr)

	if comp == nil {
		return errors.New("failed to get component for address")
	}

	return (*comp).Read(addr, value)
}

func (bus *SystemBus) Write(addr uint16, value uint8) error {
	var comp *hardware.Component = bus.getComponent(addr)

	if comp == nil {
		return errors.New("failed to get component for address")
	}

	return (*comp).Write(addr, value)
}
