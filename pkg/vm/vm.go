// Package vm contains a RiSC-16 VM implementation.
//
// See https://user.eng.umd.edu/~blj/RiSC/.
package vm

import (
	"errors"
	"fmt"
)

// The following constants define RiSC-16 opcodes.
const (
	OpcodeADD = iota
	OpcodeADDI
	OpcodeNAND
	OpcodeLUI
	OpcodeSW
	OpcodeLW
	OpcodeBEQ
	OpcodeJALR
)

// The following constants define architecture properties.
const (
	MemorySize   = 1 << 16
	NumRegisters = 8
)

// The following constants define exception types.
const (
	ExceptionTypeNONE = iota << 4
	ExceptionTypeSYSCALL
	ExceptionTypeMFSPR
	ExceptionTypeMTSPR
	ExceptionTypeRFU1
	ExceptionTypeRFU2
	ExceptionTypeRFU3
	ExceptionTypeEXCEPTION
)

// The following constants define exception values.
const (
	ExceptionValueNONE = iota
	ExceptionValueHALT
	ExceptionValueTLBMISS
	ExceptionValueSIGSEGV
	ExceptionValueINVALID
)

// VM is a RiSC-16 virtual machine. The virtual machine is not
// goroutine safe; a single goroutine should manage it.
type VM struct {
	CI  uint16               // current instruction
	GPR [NumRegisters]uint16 // general purpose registers
	M   [MemorySize]uint16   // memory
	PC  uint16               // program counter
}

// Fetch fetches the next instruction, stores it in vm.CI, and increments
// the vm.PC program counter of the virtual machine.
func (vm *VM) Fetch() {
	vm.CI = vm.M[vm.PC]
	vm.PC++
}

// String generates a string representation of the VM state.
func (vm *VM) String() string {
	return fmt.Sprintf("{PC:%d GPR:%+v}", vm.PC, vm.GPR)
}

// Execute may return the following errors.
var (
	ErrHalted    = errors.New("vm: halted")
	ErrException = errors.New("vm: exception")
)

// Execute executes the current instruction vm.CI. This function will always
// clear vm.CI so that calling Execute again will execute a NOP. This function
// returns an error when the processor has halted or a fault has occurred.
func (vm *VM) Execute() error {
	// decode instruction
	opcode := (vm.CI >> 13)
	ra := (vm.CI >> 10) & 0b0111
	rb := (vm.CI >> 7) & 0b0111
	rc := vm.CI & 0b0111
	imm7 := SignExtend7(vm.CI & 0b111_1111)
	imm10 := vm.CI & 0b11_1111_1111
	// guarantee that r0 is always zero and next instruction is NOP
	defer func() {
		vm.GPR[0] = 0
		vm.CI = 0
	}()
	// execute instruction
	switch opcode {
	case OpcodeADD:
		vm.GPR[ra] = vm.GPR[rb] + vm.GPR[rc]
	case OpcodeADDI:
		vm.GPR[ra] = vm.GPR[rb] + imm7
	case OpcodeNAND:
		vm.GPR[ra] = ^(vm.GPR[rb] & vm.GPR[rc])
	case OpcodeLUI:
		vm.GPR[ra] = imm10 << 6
	case OpcodeSW:
		vm.M[vm.GPR[rb]+imm7] = vm.GPR[ra]
	case OpcodeLW:
		vm.GPR[ra] = vm.M[vm.GPR[rb]+imm7]
	case OpcodeBEQ:
		if vm.GPR[ra] == vm.GPR[rb] {
			vm.PC += imm7
		}
	case OpcodeJALR:
		if vm.GPR[ra] == 0 && vm.GPR[rb] == 0 {
			switch imm7 & 0b_0000_0000_0111_1111 {
			case ExceptionTypeEXCEPTION | ExceptionValueHALT:
				return ErrHalted
			default:
				return fmt.Errorf("%w with ID %d", ErrException, imm7)
			}
		}
		vm.GPR[ra] = vm.PC
		vm.PC = vm.GPR[rb]
	}
	return nil
}

// SignExtend7 extends the sign to negative values over 7 bit.
func SignExtend7(v uint16) uint16 {
	if (v & 0b0000_0000_0100_0000) != 0 {
		v |= 0b1111_1111_1000_0000
	}
	return v
}

// Disassemble disassembles a single instruction and returns valid
// assembly code implementing such instruction.
func Disassemble(instr uint16) string {
	// decode instruction
	opcode := (instr >> 13)
	ra := (instr >> 10) & 0b0111
	rb := (instr >> 7) & 0b0111
	rc := instr & 0b0111
	imm7 := SignExtend7(instr & 0b111_1111)
	imm10 := instr & 0b11_1111_1111
	// disassemble instruction
	switch opcode {
	case OpcodeADD:
		return fmt.Sprintf("add r%d r%d r%d", ra, rb, rc)
	case OpcodeADDI:
		return fmt.Sprintf("addi r%d r%d %d", ra, rb, int16(imm7))
	case OpcodeNAND:
		return fmt.Sprintf("nand r%d r%d r%d", ra, rb, rc)
	case OpcodeLUI:
		return fmt.Sprintf("lui r%d %d", ra, imm10)
	case OpcodeSW:
		return fmt.Sprintf("sw r%d r%d %d", ra, rb, int16(imm7))
	case OpcodeLW:
		return fmt.Sprintf("lw r%d r%d %d", ra, rb, int16(imm7))
	case OpcodeBEQ:
		return fmt.Sprintf("beq r%d r%d %d", ra, rb, int16(imm7))
	case OpcodeJALR:
		return fmt.Sprintf("jalr r%d r%d %d", ra, rb, int16(imm7))
	default:
		return fmt.Sprintf("# unknown instruction: %d", instr)
	}
}
