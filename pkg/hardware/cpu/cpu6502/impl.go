package cpu6502

import (
	"fmt"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu"
)

func get16BitAddressLE(opperand []byte) uint16 {
	// For example:
	// f0 0f -> 0x0ff0
	return uint16(opperand[1]) | (uint16(opperand[2]) << 8)
}

func getValueBasedOnOpperand(c *CPU6502, i *cpu.Instr, opperand []byte) (uint8, uint16) {
	var value uint8 = 0
	var address uint16 = 0

	if i.Mode == IMM {
		value = opperand[1]
	} else {
		switch i.Mode {
		case ZPG:
			address = uint16(opperand[1]) % 256
			break
		case ZPX:
			address = (uint16(opperand[1]) + uint16(c.registers.x)) % 256
			break
		case ZPY:
			address = (uint16(opperand[1]) + uint16(c.registers.y)) % 256
			break
		case ABS:
			// Grab the 16 bit address
			address = get16BitAddressLE(opperand)
			break
		case ABX:
			address = get16BitAddressLE(opperand) + uint16(c.registers.x)
			break
		case ABY:
			address = get16BitAddressLE(opperand) + uint16(c.registers.y)
			break
		case IDX:
			var placeholder_1 uint8 = 0
			var placeholder_2 uint8 = 0
			// First index into the zero page
			address = (uint16(opperand[1]) + uint16(c.registers.x)) % 256

			c.sbus.Read(address, &placeholder_1)

			// Index one more to get the second byte of the absolute address
			address = (uint16(opperand[1]) + (uint16(c.registers.x) + 1)) % 256

			c.sbus.Read(address, &placeholder_2)

			// Construct the full address
			address = uint16(placeholder_1) | (uint16(placeholder_2) << 8)
			break
		case IDY:
			var placeholder_1 uint8 = 0
			var placeholder_2 uint8 = 0
			// First index into the zero page
			address = uint16(opperand[1]) % 256

			c.sbus.Read(address, &placeholder_1)

			// Index one more to get the second byte of the absolute address
			address = (uint16(opperand[1]) + 1) % 256

			c.sbus.Read(address, &placeholder_2)

			// Construct the full address
			address = uint16(placeholder_1) | (uint16(placeholder_2) << 8)
			// Add the conents of the y register
			address += uint16(c.registers.y)
			break
		}

		// Read from the retrived address following the correct mode
		c.sbus.Read(address, &value)
	}

	return value, address
}

func doNegativeCheck(c *CPU6502, value uint8) {
	if (value & 0x7) == 0x7 {
		c.SetFlag(C6502_FLAG_NEGATIVE)
	} else {
		c.ClearFlag(C6502_FLAG_NEGATIVE)
	}
}

// Overflow check needs to happen before the accumulator is set
func doOverflowCheck(c *CPU6502, prev_value uint16, value uint16) {
	if prev_value <= 127 && (prev_value+value) >= 128 {
		c.SetFlag(C6502_FLAG_OVERFLOW)
	} else {
		c.ClearFlag(C6502_FLAG_OVERFLOW)
	}
}

func doZeroCheck(c *CPU6502, value uint8) {
	if value == 0 {
		c.SetFlag(C6502_FLAG_ZERO)
	} else {
		c.ClearFlag(C6502_FLAG_ZERO)
	}
}

func doCarryCheck(c *CPU6502, value uint16) {
	if value > 256 {
		c.SetFlag(C6502_FLAG_CARRY)
	} else {
		c.ClearFlag(C6502_FLAG_CARRY)
	}
}

func doRelativeJmp(c *CPU6502, opperand []byte) {
	c.doRelativeJump(int8(opperand[1]))

	debug.Log("Jumping to: pc + %d = 0x%x\n", int8(opperand[1]), c.registers.pc)
}

var cpu6502_imp = []InstrImpl{
	// Add memory to regs.a with cary
	// A + M + C -> A, C
	{Id: symADC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("Executing ADC instruction: ")

		// Grab the value for this intruction
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		// Add the two fuckers like they're u16 integers
		sum := uint16(c.registers.a) + uint16(value)

		carry_val := 0

		if c.HasFlag(C6502_FLAG_CARRY) {
			sum++
			carry_val = 1
		}

		new_carry_val := 0

		doCarryCheck(c, sum)

		if c.HasFlag(C6502_FLAG_CARRY) {
			new_carry_val = 1
		}

		debug.Log("a:%d + m:%d + c:%d => a:%d, c:%d\n", c.registers.a, value, carry_val, uint8(sum), new_carry_val)

		// Overflow check needs to happen before the accumulator is set
		doOverflowCheck(c, uint16(c.registers.a), sum)

		// Set the registers correctly
		c.registers.a = uint8(sum)

		doNegativeCheck(c, c.registers.a)
		doZeroCheck(c, c.registers.a)

		return nil
	}},
	// And shit together
	// A & M -> A
	{Id: symAND, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("Executing AND instruction")

		// Get the value
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("Result: %d & %d => %d\n", c.registers.a, value, (c.registers.a & value))

		// Do the opperation
		c.registers.a &= value

		doZeroCheck(c, c.registers.a)
		doNegativeCheck(c, c.registers.a)

		return nil
	}},
	{Id: symASL, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("ASL: mode=0x%x\n", i.Mode)

		if i.Mode == ACC {
			if (c.registers.a & 0x7) == 0x7 {
				c.SetFlag(C6502_FLAG_CARRY)
			}

			c.registers.a <<= 1

			// Do other checks
			doNegativeCheck(c, c.registers.a)
			doZeroCheck(c, c.registers.a)
		} else {
			address := get16BitAddressLE(opperand)
			value, _ := getValueBasedOnOpperand(c, i, opperand)

			// Do the cary check
			if (value & 0x7) == 0x7 {
				c.SetFlag(C6502_FLAG_CARRY)
			}

			// Shift the value
			value <<= 1

			// Do other checks
			doNegativeCheck(c, value)
			doZeroCheck(c, value)

			// Write back to disk
			c.sbus.Write(address, value)

		}

		return nil
	}},
	{Id: symBCC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("Executing BCC instruction\n")

		// No need to do shit if there is no carry flag
		if c.HasFlag(C6502_FLAG_CARRY) {
			return nil
		}

		if int8(opperand[1]) < 0 {
			c.doJump(c.registers.pc - uint16(-int8(opperand[1])))

			debug.Log("Jumping to: pc-%d\n", uint16(-int8(opperand[1])))
		} else {
			c.doJump(c.registers.pc + uint16(opperand[1]))

			debug.Log("Jumping to: pc+%d\n", uint16(opperand[1]))
		}
		return nil
	}},
	{Id: symBCS, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("Executing BCS instruction\n")

		// No need to do shit if there is a carry flag
		if !c.HasFlag(C6502_FLAG_CARRY) {
			return nil
		}

		if int8(opperand[1]) < 0 {
			c.doJump(c.registers.pc - uint16(-int8(opperand[1])))

			debug.Log("Jumping to: pc-%d\n", uint16(-int8(opperand[1])))
		} else {
			c.doJump(c.registers.pc + uint16(opperand[1]))

			debug.Log("Jumping to: pc+%d\n", uint16(opperand[1]))
		}
		return nil
	}},
	{Id: symBEQ, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("executing BEQ\n")

		// No need to do shit if there is a carry flag
		if !c.HasFlag(C6502_FLAG_ZERO) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symBIT, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symBMI, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("executing BMI\n")

		if !c.HasFlag(C6502_FLAG_NEGATIVE) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symBNE, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("executing BNE\n")

		if c.HasFlag(C6502_FLAG_ZERO) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symBPL, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("executing BPL\n")

		if c.HasFlag(C6502_FLAG_NEGATIVE) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symBRA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symBRK, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symBVC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("executing BVC\n")

		if c.HasFlag(C6502_FLAG_OVERFLOW) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symBVS, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("executing BVS\n")

		if !c.HasFlag(C6502_FLAG_NEGATIVE) {
			return nil
		}

		doRelativeJmp(c, opperand)
		return nil
	}},
	{Id: symCLC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("Clearing carry\n")
		c.ClearFlag(C6502_FLAG_CARRY)
		return nil
	}},
	{Id: symCLD, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("Clearing decimal\n")
		c.ClearFlag(C6502_FLAG_DECIMAL)
		return nil
	}},
	{Id: symCLI, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("Clearing interupt disable\n")
		c.ClearFlag(C6502_FLAG_INTDISABLE)
		return nil
	}},
	{Id: symCLV, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("Clearing overflow\n")
		c.ClearFlag(C6502_FLAG_OVERFLOW)
		return nil
	}},
	{Id: symCMP, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		newval := int8(c.registers.a) - int8(value)

		debug.Log("Executing CMP: a:%d - m:%d = %d\n", c.registers.a, value, newval)

		doNegativeCheck(c, uint8(newval))

		c.ClearFlag(C6502_FLAG_CARRY)
		c.ClearFlag(C6502_FLAG_ZERO)

		if newval <= 0 {
			c.SetFlag(C6502_FLAG_CARRY)

			if newval == 0 {
				c.SetFlag(C6502_FLAG_ZERO)
			}
		}

		return nil
	}},
	{Id: symCPX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		newval := int8(c.registers.x) - int8(value)

		debug.Log("Executing CPX: x:%d - m:%d = %d\n", c.registers.y, value, newval)

		doNegativeCheck(c, uint8(newval))

		c.ClearFlag(C6502_FLAG_CARRY)
		c.ClearFlag(C6502_FLAG_ZERO)

		if newval <= 0 {
			c.SetFlag(C6502_FLAG_CARRY)

			if newval == 0 {
				c.SetFlag(C6502_FLAG_ZERO)
			}
		}

		return nil
	}},
	{Id: symCPY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		newval := int8(c.registers.y) - int8(value)

		debug.Log("Executing CPY: y:%d - m:%d = %d\n", c.registers.y, value, newval)

		doNegativeCheck(c, uint8(newval))

		c.ClearFlag(C6502_FLAG_CARRY)
		c.ClearFlag(C6502_FLAG_ZERO)

		if newval <= 0 {
			c.SetFlag(C6502_FLAG_CARRY)

			if newval == 0 {
				c.SetFlag(C6502_FLAG_ZERO)
			}
		}

		return nil
	}},
	{Id: symDEC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, addr := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("DEC: 0x%x - 1 -> 0x%x\n", value, addr)

		// Do the decrement
		value--

		doNegativeCheck(c, value)
		doZeroCheck(c, value)

		c.sbus.Write(addr, value)
		return nil
	}},
	{Id: symDEX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("DEX: x:0x%x - 1 -> x\n", c.registers.x)

		// Do the decrement
		c.registers.x--

		doNegativeCheck(c, c.registers.x)
		doZeroCheck(c, c.registers.x)
		return nil
	}},
	{Id: symDEY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("DEY: y:0x%x - 1 -> y\n", c.registers.y)

		// Do the decrement
		c.registers.y--

		doNegativeCheck(c, c.registers.y)
		doZeroCheck(c, c.registers.y)
		return nil
	}},
	{Id: symEOR, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symINC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, addr := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("INC: 0x%x + 1 -> 0x%x\n", value, addr)

		// Do the decrement
		value++

		doNegativeCheck(c, value)
		doZeroCheck(c, value)

		c.sbus.Write(addr, value)
		return nil
	}},
	{Id: symINX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("INX: x:0x%x + 1 -> x\n", c.registers.x)

		// Do the decrement
		c.registers.x++

		doNegativeCheck(c, c.registers.x)
		doZeroCheck(c, c.registers.x)
		return nil
	}},
	{Id: symINY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		debug.Log("INY: y:0x%x + 1 -> y\n", c.registers.y)

		// Do the decrement
		c.registers.y++

		doNegativeCheck(c, c.registers.y)
		doZeroCheck(c, c.registers.y)
		return nil
	}},
	{Id: symJMP, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		address := get16BitAddressLE(opperand)

		if i.Mode == IND {
			var a uint8
			var b uint8
			c.sbus.Read(address, &a)
			c.sbus.Read(address+1, &b)

			debug.Log(" (Old: 0x%x) ", address)

			address = uint16(a) | (uint16(b) << 8)
		}

		debug.Log("JMP: to addr: 0x%x\n", address)

		c.doJump(address)
		return nil
	}},
	{Id: symJSR, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		next_pc := c.registers.pc + 2
		target_addr := get16BitAddressLE(opperand)

		debug.Log("Doing JSR: retaddr=0x%x, targetaddr=0x%x\n", next_pc, target_addr)

		// Push the return address
		c.doPush16(next_pc)

		// Do the CPU push
		c.doJump(target_addr)
		return nil
	}},
	{Id: symLDA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		doZeroCheck(c, value)
		doNegativeCheck(c, value)

		debug.Log("Executing LDA v=0x%x\n", value)

		c.registers.a = value
		return nil
	}},
	{Id: symLDX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		doZeroCheck(c, value)
		doNegativeCheck(c, value)

		debug.Log("Executing LDX v: %d\n", value)

		c.registers.x = value
		return nil
	}},
	{Id: symLDY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		value, _ := getValueBasedOnOpperand(c, i, opperand)

		doZeroCheck(c, value)
		doNegativeCheck(c, value)

		debug.Log("Executing LDY v: %d\n", value)

		c.registers.y = value
		return nil
	}},
	{Id: symLSR, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symNOP, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symORA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPHA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPHP, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPHX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		// Illigal?
		return fmt.Errorf("(PHX) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPHY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		// Illigal?
		return fmt.Errorf("(PHY) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPLA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPLP, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPLX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("(PLX) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symPLY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("(PLY) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symROL, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symROR, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symRTI, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symRTS, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("(RTS) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symSBC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symSEC, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		c.SetFlag(C6502_FLAG_CARRY)
		return nil
	}},
	{Id: symSED, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		c.SetFlag(C6502_FLAG_DECIMAL)
		return nil
	}},
	{Id: symSEI, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		c.SetFlag(C6502_FLAG_INTDISABLE)
		return nil
	}},
	{Id: symSTA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		_, addr := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("Executing STA(%d): putting a:%d into 0x%x\n", i.Mode, c.registers.a, addr)

		c.sbus.Write(addr, c.registers.a)
		return nil
	}},
	{Id: symSTZ, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("(STZ) executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symSTX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		_, addr := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("Executing STX: putting x:%d into 0x%x\n", c.registers.x, addr)

		c.sbus.Write(addr, c.registers.x)
		return nil
	}},
	{Id: symSTY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		_, addr := getValueBasedOnOpperand(c, i, opperand)

		debug.Log("Executing STY: putting y:%d into 0x%x\n", c.registers.y, addr)

		c.sbus.Write(addr, c.registers.y)
		return nil
	}},
	{Id: symTAX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTAY, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTRB, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTSB, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTSX, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTXA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
	{Id: symTXS, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {

		debug.Log("TSX: pushing x=0x%x on the stack\n", c.registers.x)

		c.doPush8(c.registers.x)
		return nil
	}},
	{Id: symTYA, Impl: func(c *CPU6502, i *cpu.Instr, opperand []byte) error {
		return fmt.Errorf("executing unimplemented instruction: %d", i.Instruction)
	}},
}