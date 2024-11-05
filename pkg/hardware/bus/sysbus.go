package bus

import (
	"errors"

	"github.com/beakeyz/gones-emu/pkg/hardware/comp"
)

type SystemBus struct {
	components []comp.Component
}

func NewSystembus() (*SystemBus, error) {
	bus := SystemBus{
		components: make([]comp.Component, 0),
	}

	return &bus, nil
}

func (bus *SystemBus) AddComponent(comp comp.Component) error {
    if (comp == nil) {
        return errors.New("tried to add a nil component!")
    }

	bus.components = append(bus.components, comp)
    return nil
}

func (bus *SystemBus) getComponent(addr uint16) *comp.Component {
	for _, comp := range bus.components {
		if addr >= comp.StartAddr() && addr <= comp.EndAddr() {
			return &comp
		}
	}

	return nil
}

func (bus *SystemBus) Read(addr uint16, value *uint8) error {
	var comp *comp.Component = bus.getComponent(addr)

	if comp == nil {
		return errors.New("failed to get component for address")
	}

	return (*comp).Read(addr, value)
}

func (bus *SystemBus) Write(addr uint16, value uint8) error {
	var comp *comp.Component = bus.getComponent(addr)

	if comp == nil {
		return errors.New("failed to get component for address")
	}

	return (*comp).Write(addr, value)
}
