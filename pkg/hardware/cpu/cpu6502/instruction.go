package cpu6502

import (
	"errors"

	"github.com/beakeyz/gones-emu/pkg/debug"
	"github.com/beakeyz/gones-emu/pkg/hardware/cpu"
)

/*
 * For the list of opcodes,
 * See: https://github.com/beevik/go6502/blob/284c57a62253fce2b37a2934deefe2fe761d90e3/cpu/instructions.go
 */

const (
	symADC cpu.InstrID = iota
	symAND
	symASL
	symBCC
	symBCS
	symBEQ
	symBIT
	symBMI
	symBNE
	symBPL
	symBRA
	symBRK
	symBVC
	symBVS
	symCLC
	symCLD
	symCLI
	symCLV
	symCMP
	symCPX
	symCPY
	symDEC
	symDEX
	symDEY
	symEOR
	symINC
	symINX
	symINY
	symJMP
	symJSR
	symLDA
	symLDX
	symLDY
	symLSR
	symNOP
	symORA
	symPHA
	symPHP
	symPHX
	symPHY
	symPLA
	symPLP
	symPLX
	symPLY
	symROL
	symROR
	symRTI
	symRTS
	symSBC
	symSEC
	symSED
	symSEI
	symSTA
	symSTZ
	symSTX
	symSTY
	symTAX
	symTAY
	symTRB
	symTSB
	symTSX
	symTXA
	symTXS
	symTYA
)

const (
	IMM cpu.AddrMode = iota // Immediate
	IMP                     // Implied (no operand)
	REL                     // Relative
	ZPG                     // Zero Page
	ZPX                     // Zero Page,X
	ZPY                     // Zero Page,Y
	ABS                     // Absolute
	ABX                     // Absolute,X
	ABY                     // Absolute,Y
	IND                     // (Indirect)
	IDX                     // (Indirect,X)
	IDY                     // (Indirect),Y
	ACC                     // Accumulator (no operand)
)

// All valid (opcode, mode) pairs
var cpuInstructions = []cpu.Instr{
	{Instruction: symLDA, Mode: IMM, Opcode: 0xa9, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symLDA, Mode: ZPG, Opcode: 0xa5, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symLDA, Mode: ZPX, Opcode: 0xb5, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDA, Mode: ABS, Opcode: 0xad, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDA, Mode: ABX, Opcode: 0xbd, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symLDA, Mode: ABY, Opcode: 0xb9, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symLDA, Mode: IDX, Opcode: 0xa1, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symLDA, Mode: IDY, Opcode: 0xb1, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symLDA, Mode: IND, Opcode: 0xb2, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symLDX, Mode: IMM, Opcode: 0xa2, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symLDX, Mode: ZPG, Opcode: 0xa6, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symLDX, Mode: ZPY, Opcode: 0xb6, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDX, Mode: ABS, Opcode: 0xae, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDX, Mode: ABY, Opcode: 0xbe, Len: 3, Cycles: 4, Pb_cross_cycles: 1},

	{Instruction: symLDY, Mode: IMM, Opcode: 0xa0, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symLDY, Mode: ZPG, Opcode: 0xa4, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symLDY, Mode: ZPX, Opcode: 0xb4, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDY, Mode: ABS, Opcode: 0xac, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symLDY, Mode: ABX, Opcode: 0xbc, Len: 3, Cycles: 4, Pb_cross_cycles: 1},

	{Instruction: symSTA, Mode: ZPG, Opcode: 0x85, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: ZPX, Opcode: 0x95, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: ABS, Opcode: 0x8d, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: ABX, Opcode: 0x9d, Len: 3, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: ABY, Opcode: 0x99, Len: 3, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: IDX, Opcode: 0x81, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: IDY, Opcode: 0x91, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symSTA, Mode: IND, Opcode: 0x92, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symSTX, Mode: ZPG, Opcode: 0x86, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symSTX, Mode: ZPY, Opcode: 0x96, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTX, Mode: ABS, Opcode: 0x8e, Len: 3, Cycles: 4, Pb_cross_cycles: 0},

	{Instruction: symSTY, Mode: ZPG, Opcode: 0x84, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symSTY, Mode: ZPX, Opcode: 0x94, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTY, Mode: ABS, Opcode: 0x8c, Len: 3, Cycles: 4, Pb_cross_cycles: 0},

	{Instruction: symSTZ, Mode: ZPG, Opcode: 0x64, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symSTZ, Mode: ZPX, Opcode: 0x74, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTZ, Mode: ABS, Opcode: 0x9c, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSTZ, Mode: ABX, Opcode: 0x9e, Len: 3, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symADC, Mode: IMM, Opcode: 0x69, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symADC, Mode: ZPG, Opcode: 0x65, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symADC, Mode: ZPX, Opcode: 0x75, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symADC, Mode: ABS, Opcode: 0x6d, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symADC, Mode: ABX, Opcode: 0x7d, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symADC, Mode: ABY, Opcode: 0x79, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symADC, Mode: IDX, Opcode: 0x61, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symADC, Mode: IDY, Opcode: 0x71, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symADC, Mode: IND, Opcode: 0x72, Len: 2, Cycles: 5, Pb_cross_cycles: 1},

	{Instruction: symSBC, Mode: IMM, Opcode: 0xe9, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symSBC, Mode: ZPG, Opcode: 0xe5, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symSBC, Mode: ZPX, Opcode: 0xf5, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSBC, Mode: ABS, Opcode: 0xed, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symSBC, Mode: ABX, Opcode: 0xfd, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symSBC, Mode: ABY, Opcode: 0xf9, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symSBC, Mode: IDX, Opcode: 0xe1, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symSBC, Mode: IDY, Opcode: 0xf1, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symSBC, Mode: IND, Opcode: 0xf2, Len: 2, Cycles: 5, Pb_cross_cycles: 1},

	{Instruction: symCMP, Mode: IMM, Opcode: 0xc9, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCMP, Mode: ZPG, Opcode: 0xc5, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symCMP, Mode: ZPX, Opcode: 0xd5, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symCMP, Mode: ABS, Opcode: 0xcd, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symCMP, Mode: ABX, Opcode: 0xdd, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symCMP, Mode: ABY, Opcode: 0xd9, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symCMP, Mode: IDX, Opcode: 0xc1, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symCMP, Mode: IDY, Opcode: 0xd1, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symCMP, Mode: IND, Opcode: 0xd2, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symCPX, Mode: IMM, Opcode: 0xe0, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCPX, Mode: ZPG, Opcode: 0xe4, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symCPX, Mode: ABS, Opcode: 0xec, Len: 3, Cycles: 4, Pb_cross_cycles: 0},

	{Instruction: symCPY, Mode: IMM, Opcode: 0xc0, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCPY, Mode: ZPG, Opcode: 0xc4, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symCPY, Mode: ABS, Opcode: 0xcc, Len: 3, Cycles: 4, Pb_cross_cycles: 0},

	{Instruction: symBIT, Mode: IMM, Opcode: 0x89, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symBIT, Mode: ZPG, Opcode: 0x24, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symBIT, Mode: ZPX, Opcode: 0x34, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symBIT, Mode: ABS, Opcode: 0x2c, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symBIT, Mode: ABX, Opcode: 0x3c, Len: 3, Cycles: 4, Pb_cross_cycles: 1},

	{Instruction: symCLC, Mode: IMP, Opcode: 0x18, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symSEC, Mode: IMP, Opcode: 0x38, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCLI, Mode: IMP, Opcode: 0x58, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symSEI, Mode: IMP, Opcode: 0x78, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCLD, Mode: IMP, Opcode: 0xd8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symSED, Mode: IMP, Opcode: 0xf8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symCLV, Mode: IMP, Opcode: 0xb8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symBCC, Mode: REL, Opcode: 0x90, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBCS, Mode: REL, Opcode: 0xb0, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBEQ, Mode: REL, Opcode: 0xf0, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBNE, Mode: REL, Opcode: 0xd0, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBMI, Mode: REL, Opcode: 0x30, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBPL, Mode: REL, Opcode: 0x10, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBVC, Mode: REL, Opcode: 0x50, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBVS, Mode: REL, Opcode: 0x70, Len: 2, Cycles: 2, Pb_cross_cycles: 1},
	{Instruction: symBRA, Mode: REL, Opcode: 0x80, Len: 2, Cycles: 2, Pb_cross_cycles: 1},

	{Instruction: symBRK, Mode: IMP, Opcode: 0x00, Len: 1, Cycles: 7, Pb_cross_cycles: 0},

	{Instruction: symAND, Mode: IMM, Opcode: 0x29, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symAND, Mode: ZPG, Opcode: 0x25, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symAND, Mode: ZPX, Opcode: 0x35, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symAND, Mode: ABS, Opcode: 0x2d, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symAND, Mode: ABX, Opcode: 0x3d, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symAND, Mode: ABY, Opcode: 0x39, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symAND, Mode: IDX, Opcode: 0x21, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symAND, Mode: IDY, Opcode: 0x31, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symAND, Mode: IND, Opcode: 0x32, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symORA, Mode: IMM, Opcode: 0x09, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symORA, Mode: ZPG, Opcode: 0x05, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symORA, Mode: ZPX, Opcode: 0x15, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symORA, Mode: ABS, Opcode: 0x0d, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symORA, Mode: ABX, Opcode: 0x1d, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symORA, Mode: ABY, Opcode: 0x19, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symORA, Mode: IDX, Opcode: 0x01, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symORA, Mode: IDY, Opcode: 0x11, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symORA, Mode: IND, Opcode: 0x12, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symEOR, Mode: IMM, Opcode: 0x49, Len: 2, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symEOR, Mode: ZPG, Opcode: 0x45, Len: 2, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symEOR, Mode: ZPX, Opcode: 0x55, Len: 2, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symEOR, Mode: ABS, Opcode: 0x4d, Len: 3, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symEOR, Mode: ABX, Opcode: 0x5d, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symEOR, Mode: ABY, Opcode: 0x59, Len: 3, Cycles: 4, Pb_cross_cycles: 1},
	{Instruction: symEOR, Mode: IDX, Opcode: 0x41, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symEOR, Mode: IDY, Opcode: 0x51, Len: 2, Cycles: 5, Pb_cross_cycles: 1},
	{Instruction: symEOR, Mode: IND, Opcode: 0x52, Len: 2, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symINC, Mode: ZPG, Opcode: 0xe6, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symINC, Mode: ZPX, Opcode: 0xf6, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symINC, Mode: ABS, Opcode: 0xee, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symINC, Mode: ABX, Opcode: 0xfe, Len: 3, Cycles: 7, Pb_cross_cycles: 0},
	{Instruction: symINC, Mode: ACC, Opcode: 0x1a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symDEC, Mode: ZPG, Opcode: 0xc6, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symDEC, Mode: ZPX, Opcode: 0xd6, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symDEC, Mode: ABS, Opcode: 0xce, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symDEC, Mode: ABX, Opcode: 0xde, Len: 3, Cycles: 7, Pb_cross_cycles: 0},
	{Instruction: symDEC, Mode: ACC, Opcode: 0x3a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symINX, Mode: IMP, Opcode: 0xe8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symINY, Mode: IMP, Opcode: 0xc8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symDEX, Mode: IMP, Opcode: 0xca, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symDEY, Mode: IMP, Opcode: 0x88, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symJMP, Mode: ABS, Opcode: 0x4c, Len: 3, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symJMP, Mode: ABX, Opcode: 0x7c, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symJMP, Mode: IND, Opcode: 0x6c, Len: 3, Cycles: 5, Pb_cross_cycles: 0},

	{Instruction: symJSR, Mode: ABS, Opcode: 0x20, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symRTS, Mode: IMP, Opcode: 0x60, Len: 1, Cycles: 6, Pb_cross_cycles: 0},

	{Instruction: symRTI, Mode: IMP, Opcode: 0x40, Len: 1, Cycles: 6, Pb_cross_cycles: 0},

	{Instruction: symNOP, Mode: IMP, Opcode: 0xea, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symTAX, Mode: IMP, Opcode: 0xaa, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symTXA, Mode: IMP, Opcode: 0x8a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symTAY, Mode: IMP, Opcode: 0xa8, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symTYA, Mode: IMP, Opcode: 0x98, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symTXS, Mode: IMP, Opcode: 0x9a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symTSX, Mode: IMP, Opcode: 0xba, Len: 1, Cycles: 2, Pb_cross_cycles: 0},

	{Instruction: symTRB, Mode: ZPG, Opcode: 0x14, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symTRB, Mode: ABS, Opcode: 0x1c, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symTSB, Mode: ZPG, Opcode: 0x04, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symTSB, Mode: ABS, Opcode: 0x0c, Len: 3, Cycles: 6, Pb_cross_cycles: 0},

	{Instruction: symPHA, Mode: IMP, Opcode: 0x48, Len: 1, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symPLA, Mode: IMP, Opcode: 0x68, Len: 1, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symPHP, Mode: IMP, Opcode: 0x08, Len: 1, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symPLP, Mode: IMP, Opcode: 0x28, Len: 1, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symPHX, Mode: IMP, Opcode: 0xda, Len: 1, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symPLX, Mode: IMP, Opcode: 0xfa, Len: 1, Cycles: 4, Pb_cross_cycles: 0},
	{Instruction: symPHY, Mode: IMP, Opcode: 0x5a, Len: 1, Cycles: 3, Pb_cross_cycles: 0},
	{Instruction: symPLY, Mode: IMP, Opcode: 0x7a, Len: 1, Cycles: 4, Pb_cross_cycles: 0},

	{Instruction: symASL, Mode: ACC, Opcode: 0x0a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symASL, Mode: ZPG, Opcode: 0x06, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symASL, Mode: ZPX, Opcode: 0x16, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symASL, Mode: ABS, Opcode: 0x0e, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symASL, Mode: ABX, Opcode: 0x1e, Len: 3, Cycles: 7, Pb_cross_cycles: 0},

	{Instruction: symLSR, Mode: ACC, Opcode: 0x4a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symLSR, Mode: ZPG, Opcode: 0x46, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symLSR, Mode: ZPX, Opcode: 0x56, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symLSR, Mode: ABS, Opcode: 0x4e, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symLSR, Mode: ABX, Opcode: 0x5e, Len: 3, Cycles: 7, Pb_cross_cycles: 0},

	{Instruction: symROL, Mode: ACC, Opcode: 0x2a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symROL, Mode: ZPG, Opcode: 0x26, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symROL, Mode: ZPX, Opcode: 0x36, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symROL, Mode: ABS, Opcode: 0x2e, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symROL, Mode: ABX, Opcode: 0x3e, Len: 3, Cycles: 7, Pb_cross_cycles: 0},

	{Instruction: symROR, Mode: ACC, Opcode: 0x6a, Len: 1, Cycles: 2, Pb_cross_cycles: 0},
	{Instruction: symROR, Mode: ZPG, Opcode: 0x66, Len: 2, Cycles: 5, Pb_cross_cycles: 0},
	{Instruction: symROR, Mode: ZPX, Opcode: 0x76, Len: 2, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symROR, Mode: ABS, Opcode: 0x6e, Len: 3, Cycles: 6, Pb_cross_cycles: 0},
	{Instruction: symROR, Mode: ABX, Opcode: 0x7e, Len: 3, Cycles: 7, Pb_cross_cycles: 0},
}

func GetInstr(opcode byte) (cpu.Instr, error) {
	/* Ew linear scan */
	for _, inst := range cpuInstructions {
		if inst.Opcode == opcode {
			return inst, nil
		}
	}

	debug.Error("Looking for: 0x%x\n", opcode)

	return cpu.Instr{}, errors.New("invalid opcode")
}
