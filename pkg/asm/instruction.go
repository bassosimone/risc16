package asm

import (
	"fmt"
	"strconv"
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

// Instruction is a parsed instruction.
type Instruction interface {
	// Err returns the error occurred processing the instruction. If this
	// function returns nil, then the instruction is valid.
	Err() error

	// Label returns the label associated with the instruction. If this
	// function returns nil, then there is no label.
	Label() *string

	// Encode encodes the instruction. The table passed in input maps each
	// label to the corresponding offset in memory.
	Encode(labels map[string]int64) (uint16, error)
}

// InstructionErr is an error
type InstructionErr struct {
	Error error
}

// Err implements Instruction.Err
func (ie InstructionErr) Err() error {
	return ie.Error
}

// Label implements Instruction.Label
func (ie InstructionErr) Label() *string {
	return nil
}

// Encode implements Instruction.Encode
func (ie InstructionErr) Encode(labels map[string]int64) (uint16, error) {
	return 0, fmt.Errorf("%w because this is an error", ErrCannotEncode)
}

// NewParseError constructs a new parsed instruction
// that actually wraps a parsing error.
func NewParseError(err error) []Instruction {
	return []Instruction{InstructionErr{Error: err}}
}

var _ Instruction = InstructionErr{}

// InstructionADD is the ADD instruction
type InstructionADD struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	RC         uint16
}

// Err implements Instruction.Err
func (ia InstructionADD) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionADD) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionADD) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeADD & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	out |= ia.RC & 0b111
	return out, nil
}

var _ Instruction = InstructionADD{}

// InstructionADDI is the ADDI instruction
type InstructionADDI struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionADDI) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionADDI) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionADDI) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeADDI & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	imm, err := ResolveImmediate(labels, ia.Imm, 7)
	if err != nil {
		return 0, err
	}
	out |= imm
	return out, nil
}

var _ Instruction = InstructionADDI{}

// InstructionNAND is the NAND instruction
type InstructionNAND struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	RC         uint16
}

// Err implements Instruction.Err
func (ia InstructionNAND) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionNAND) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionNAND) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeNAND & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	out |= ia.RC & 0b111
	return out, nil
}

var _ Instruction = InstructionNAND{}

// InstructionLUI is the LUI instruction
type InstructionLUI struct {
	MaybeLabel *string
	RA         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionLUI) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionLUI) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionLUI) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeLUI & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	imm, err := ResolveImmediate(labels, ia.Imm, 16)
	if err != nil {
		return 0, err
	}
	out |= (imm >> 6)
	return out, nil
}

var _ Instruction = InstructionLUI{}

// InstructionSW is the SW instruction
type InstructionSW struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionSW) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionSW) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionSW) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeSW & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	imm, err := ResolveImmediate(labels, ia.Imm, 7)
	if err != nil {
		return 0, err
	}
	out |= imm
	return out, nil
}

var _ Instruction = InstructionSW{}

// InstructionLW is the LW instruction
type InstructionLW struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionLW) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionLW) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionLW) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeLW & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	imm, err := ResolveImmediate(labels, ia.Imm, 7)
	if err != nil {
		return 0, err
	}
	out |= imm
	return out, nil
}

var _ Instruction = InstructionLW{}

// InstructionBEQ is the BEQ instruction
type InstructionBEQ struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionBEQ) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionBEQ) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionBEQ) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeBEQ & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	imm, err := ResolveImmediate(labels, ia.Imm, 7)
	if err != nil {
		return 0, err
	}
	out |= imm
	return out, nil
}

var _ Instruction = InstructionBEQ{}

// InstructionJALR is the JALR instruction
type InstructionJALR struct {
	MaybeLabel *string
	RA         uint16
	RB         uint16
	Imm        uint16
}

// Err implements Instruction.Err
func (ia InstructionJALR) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionJALR) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionJALR) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeJALR & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RB & 0b111) << 7
	out |= ia.Imm & 0b111_1111
	return out, nil
}

var _ Instruction = InstructionJALR{}

// InstructionLLI is the LLI instruction
type InstructionLLI struct {
	MaybeLabel *string
	RA         uint16
	Imm        string
}

// Err implements Instruction.Err
func (ia InstructionLLI) Err() error {
	return nil
}

// Label implements Instruction.Label
func (ia InstructionLLI) Label() *string {
	return ia.MaybeLabel
}

// Encode implements Instruction.Encode
func (ia InstructionLLI) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= (OpcodeADDI & 0b111) << 13
	out |= (ia.RA & 0b111) << 10
	out |= (ia.RA & 0b111) << 7
	imm, err := ResolveImmediate(labels, ia.Imm, 16)
	if err != nil {
		return 0, err
	}
	out |= (imm & 0b11_1111)
	return out, nil
}

var _ Instruction = InstructionLLI{}

// ResolveImmediate resolves the value of an immediate
func ResolveImmediate(labels map[string]int64, name string, bits int) (uint16, error) {
	if bits < 1 || bits > 16 {
		panic("bits value out of range")
	}
	value, err := strconv.ParseInt(name, 0, 64)
	if err != nil {
		var found bool
		value, found = labels[name]
		if !found {
			return 0, fmt.Errorf("%w because label '%s' is missing", ErrCannotEncode, name)
		}
		// fallthrough
	}
	var mask int64 = (1 << bits) - 1
	if (value & ^mask) != 0 {
		return 0, fmt.Errorf("%w for immediate '%s'", ErrOutOrRange, name)
	}
	return uint16(value & mask), nil
}
