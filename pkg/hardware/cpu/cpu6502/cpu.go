package cpu6502

import (
	"fmt"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/bus"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu"
)

/*
 * 6502 CPU internal registers
 *
 * See: https://www.nesdev.org/wiki/CPU_registers
 */
type CPU6502Register struct {
	a uint8
	x uint8
	y uint8
	/* Stack pointer */
	s uint8
	/* Flags */
	p uint8
	/* Program counter */
	pc uint16
}

const (
	C6502_FLAG_CARRY      = (1 << 0)
	C6502_FLAG_ZERO       = (1 << 1)
	C6502_FLAG_INTDISABLE = (1 << 2)
	C6502_FLAG_DECIMAL    = (1 << 3)
	C6502_FLAG_BFLAG      = (1 << 4)
	C6502_FLAG_RESERVED   = (1 << 5)
	C6502_FLAG_OVERFLOW   = (1 << 6)
	C6502_FLAG_NEGATIVE   = (1 << 7)
)

type InstrImpl struct {
	Id   cpu.InstrID
	Impl func(c *CPU6502, i *cpu.Instr, opperand []byte) error
}

type CPU6502 struct {
	registers CPU6502Register
	sbus      *bus.SystemBus
	/* Current instruction we're executing */
	c_instr *cpu.Instr
	/* Amount of cycles we've spent at this instruction */
	n_cycles  byte
	impl_list []InstrImpl
}

func (cpu *CPU6502) Initialize() {
	var regs *CPU6502Register = &cpu.registers

	/* See: https://www.nesdev.org/wiki/CPU_power_up_state */
	regs.a = 0
	regs.x = 0
	regs.y = 0
	regs.s = 0xff
	regs.p = (C6502_FLAG_INTDISABLE | C6502_FLAG_RESERVED)

	a := uint8(0)
	b := uint8(0)

	cpu.sbus.Read(0xfffc, &a)
	cpu.sbus.Read(0xfffd, &b)

	debug.Log("a: 0x%x, b: 0x%x\n", a, b)

	regs.pc = uint16(a) | (uint16(b) << 8)

	// 31 e6 -> e631
	debug.Log("PC: 0x%x\n", regs.pc)
}

func (cpu *CPU6502) Reset() {
	var regs *CPU6502Register = &cpu.registers

	/* See: https://www.nesdev.org/wiki/CPU_power_up_state */
	regs.s -= 3
	regs.p |= C6502_FLAG_INTDISABLE
	regs.pc = 0xfffc
}

func (cpu *CPU6502) fetchOpperand(opperand []byte, len byte) error {
	var err error = nil

	for i := range len {
		err = cpu.sbus.Read(cpu.registers.pc+uint16(i), &opperand[i])

		if err != nil {
			break
		}
	}

	return err
}

func (c *CPU6502) SetFlag(flag byte) {
	c.registers.p |= flag
}

func (c *CPU6502) ClearFlag(flag byte) {
	c.registers.p &= ^(flag)
}

func (c *CPU6502) HasFlag(flag byte) bool {
	return (c.registers.p & flag) == flag
}

func (c *CPU6502) doJump(addr uint16) {
	c.registers.pc = addr
	// Tell DoCycle that we want to jump
	c.c_instr = nil
}

func (c *CPU6502) doRelativeJump(offset int8) {
	if offset < 0 {
		c.registers.pc -= uint16(-offset)
	} else {
		c.registers.pc += uint16(offset)
	}
}

func (c *CPU6502) doPush8(value uint8) {
	c.sbus.Write(0x0100+uint16(c.registers.s), uint8(value))
	c.registers.s--
}

func (c *CPU6502) doPop8(value *uint8) {
	c.registers.s++
	c.sbus.Read(0x0100+uint16(c.registers.s), value)
}

// 0xff00 => * 00 ff
// 00 ff =>
func (c *CPU6502) doPush16(value uint16) {
	c.doPush8(uint8(value >> 8))
	c.doPush8(uint8(value))
}

func (c *CPU6502) doPop16(value *uint16) {
	var a uint8
	var b uint8

	c.doPop8(&a)
	c.doPop8(&b)

	*value = uint16(a) | (uint16(b) << 8)
}

func (cpu *CPU6502) fetchInstrImpl(instid cpu.InstrID) *InstrImpl {
	var impl *InstrImpl
	if int(instid) >= len(cpu.impl_list) {
		return nil
	}

	impl = &cpu.impl_list[instid]

	if impl.Id != instid {
		debug.Error("Got id: %d\n", impl.Id)
		return nil
	}

	return impl
}

/*
 * Do a single 6502 clockcycle
 *
 * If we're still 'executing' an instruction, wait
 * the appropriate amount of time and skip this cycle
 *
 * Otherwise fetch a new instruction from the PC
 */
func (c *CPU6502) DoCycle() error {
	var err error
	var c_instr cpu.Instr
	var impl *InstrImpl

	if c.c_instr == nil {
		var c_opcode byte

		// Read the opcode from PC
		err = c.sbus.Read(c.registers.pc, &c_opcode)

		if err != nil {
			return err
		}

		// Try to get the instruction for this opcode
		c_instr, err = GetInstr(c_opcode)

		if err != nil {
			return err
		}

		c.c_instr = &c_instr
		c.n_cycles = 0
	}

	// Add a cycle
	c.n_cycles++

	// Reached the end of this instruction, execute the next intruction next cycle
	if c.n_cycles == c.c_instr.Cycles {

		var opperand []byte = make([]byte, c.c_instr.Len)

		err = c.fetchOpperand(opperand, c.c_instr.Len)

		if err != nil {
			return err
		}

		impl = c.fetchInstrImpl(c.c_instr.Instruction)

		if impl == nil {
			return fmt.Errorf("cpu6502: fetching unimplemented instruction: %d\n", c.c_instr.Instruction)
		}

		debug.Log("(0x%x Len:%d): ", c.registers.pc, c.c_instr.Len)

		err = impl.Impl(c, c.c_instr, opperand)

		if err != nil {
			return err
		}

		// Increment the pc to the next instruction, if the c_instr field isn't nil
		// NOTE: Instructions like jmp, that work directly on the PC may use this fact
		// to indicate we simply need to execute the next thing without adding shit to
		// the PC
		if c.c_instr != nil {
			c.registers.pc += uint16(c.c_instr.Len)
		}

		// Reset the current instruction to nil to prompt an instruction fetch on the next cycle
		c.c_instr = nil
	}

	return err
}

/*
 * Transfers CPU control to the Nmi handler
 *
 */
func (c *CPU6502) RaiseNmi() {

}

func New(sbus *bus.SystemBus) *CPU6502 {
	var c CPU6502 = CPU6502{
		registers: CPU6502Register{},
		sbus:      sbus,
		c_instr:   nil,
		n_cycles:  0,
		impl_list: cpu6502_imp,
	}

	return &c
}
