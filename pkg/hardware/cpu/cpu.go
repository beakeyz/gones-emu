package cpu

type InstrID byte
type AddrMode byte

/*
 * Represent a single instruction
 *
 * When the CPU emulator reads an instruction, we find this
 * struct in it's implementation list, based by op
 */
type Instr struct {
	Instruction     InstrID
	Mode            AddrMode
	Opcode          byte
	Len             byte
	Cycles          byte
	Pb_cross_cycles byte
}

type CPU interface {
	/* Initializes its internal registers to their initial values */
	Initialize()
	/* Performs a CPU reset */
	Reset()
	/* Executes a single CPU cycle */
	DoCycle() error
}
