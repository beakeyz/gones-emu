package hardware

type Component interface {
	Read(addr uint16, value *uint8) error
	Write(addr uint16, value uint8) error
	StartAddr() uint16
	EndAddr() uint16
}
