package hardware

/*
 * Abstracts away the entire NES system
 * into a single struct.
 *
 * There are multiple chips which have their own clocks and work independently. Some of these also
 * have their own busses. For example, the PPU needs to raise an IRQ to the main 6502 CPU when it's
 * starting to vblank, so it needs access to the main CPU, but it also has it's own bus for accessing
 * hardware components.
 */
type NESSystem struct {
	/* 6502 CPU */
	/* PPU */
	/* */
}
