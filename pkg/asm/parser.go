package asm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseSpecificInstruction is the function parsing a specific instruction.
type ParseSpecificInstruction func(in <-chan LexerToken, label *string) ParsedInstruction

// InstructionParsers maps an instruction to its parser.
var InstructionParsers = map[string]ParseSpecificInstruction{
	"add":    ParseADD,
	"addi":   ParseADDI,
	"nand":   nil,
	"lui":    nil,
	"sw":     nil,
	"lw":     nil,
	"beq":    nil,
	"jalr":   nil,
	"nop":    nil,
	"halt":   nil,
	"lli":    nil,
	"movi":   nil,
	".fill":  nil,
	".space": nil,
}

// The following errors may occur during parsing.
var (
	ErrExpectedNameOrNumber = errors.New("asm: expected name or number")
	ErrUnknownInstruction   = errors.New("asm: unknown instruction")
	ErrExpectedComma        = errors.New("asm: expected comma")
	ErrExpectedEOL          = errors.New("asm: expected end of line")
	ErrInvalidRegisterName  = errors.New("asm: invalid register name")
	ErrOutOrRange           = errors.New("asm: immediate value out of range")
	ErrCannotEncode         = errors.New("ams: can't encode instruction")
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

// ParsedInstruction is a parsed instruction.
type ParsedInstruction interface {
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

// StartParsing starts parsing in a backend goroutine.
func StartParsing(in <-chan LexerToken) <-chan ParsedInstruction {
	out := make(chan ParsedInstruction)
	go ParseAsync(in, out)
	return out
}

// ParsedGenericInstruction is a generic parsed instruction.
type ParsedGenericInstruction struct {
	Error      error
	MaybeLabel *string
	Opcode     uint16
	RA         uint16
}

// Err implements ParsedInstruction.Err.
func (pi *ParsedGenericInstruction) Err() error {
	return pi.Error
}

// Label implements ParsedInstruction.Label.
func (pi *ParsedGenericInstruction) Label() *string {
	return pi.MaybeLabel
}

// Encode implements ParsedInstruction.Encode.
func (pi *ParsedGenericInstruction) Encode(labels map[string]int64) (uint16, error) {
	return 0, fmt.Errorf("%w because this is not a specific instruction", ErrCannotEncode)
}

// EncodeCommon encodes the Opcode and RA.
func (pi *ParsedGenericInstruction) EncodeCommon() (out uint16) {
	out |= (pi.Opcode & 0b111) << 13
	out |= (pi.RA & 0b111) << 10
	return
}

// NewParseError constructs a new parsed instruction
// that actually wraps a parsing error.
func NewParseError(err error) ParsedInstruction {
	return &ParsedGenericInstruction{Error: err}
}

// ParsedInstructionRRR is a parsed RRR instruction.
type ParsedInstructionRRR struct {
	ParsedGenericInstruction
	RB uint16
	RC uint16
}

// Encode implements ParsedInstruction.Encode
func (pi *ParsedInstructionRRR) Encode(labels map[string]int64) (out uint16, err error) {
	out |= pi.ParsedGenericInstruction.EncodeCommon()
	out |= (pi.RB & 0b111) << 7
	out |= pi.RC & 0b111
	return
}

// ParsedInstructionRRI is a parsed RRI instruction.
type ParsedInstructionRRI struct {
	ParsedGenericInstruction
	RB   uint16
	Imm7 string
}

// Encode implements ParsedInstruction.Encode
func (pi *ParsedInstructionRRI) Encode(labels map[string]int64) (uint16, error) {
	var out uint16
	out |= pi.ParsedGenericInstruction.EncodeCommon()
	out |= (pi.RB & 0b111) << 7
	n, err := strconv.ParseInt(pi.Imm7, 0, 64)
	if err != nil {
		var found bool
		n, found = labels[pi.Imm7]
		if !found {
			return 0, fmt.Errorf("%w because label '%s' is missing",
				ErrCannotEncode, pi.Imm7)
		}
	}
	if n < -64 || n > 63 {
		return 0, fmt.Errorf("%w for immediate '%s'", ErrOutOrRange, pi.Imm7)
	}
	out |= uint16(n & 0b111_1111)
	return out, nil
}

// ParseAsync is the async instructions parser.
func ParseAsync(in <-chan LexerToken, out chan<- ParsedInstruction) {
	defer func() {
		for range in {
			// drain channel (for robustness)
		}
		close(out)
	}()
	for {
		instr := ParseSingleInstruction(in)
		if instr == nil {
			return
		}
		out <- instr
		if instr.Err() != nil {
			return
		}
	}
}

// ParseSingleInstruction parses an instruction.
func ParseSingleInstruction(in <-chan LexerToken) ParsedInstruction {
again:
	// 1. parse optional label
	var label *string
	token := <-in
	switch token.Type {
	case LexerEOF:
		return nil // end of lexing and parsing
	case LexerEOL:
		goto again // empty line
	case LexerLabel:
		v := strings.TrimSuffix(token.Value, ":")
		label = &v
		token = <-in
	default:
		// fallthrough
	}
	// 2. parse the instruction
	switch token.Type {
	case LexerNameOrNumber:
	default:
		return NewParseError(fmt.Errorf("%w while parsing instruction name on line %d",
			ErrExpectedNameOrNumber, token.Lineno))
	}
	parser := InstructionParsers[token.Value]
	if parser == nil {
		return NewParseError(fmt.Errorf("%w while processing instruction name on line %d",
			ErrUnknownInstruction, token.Lineno))
	}
	return parser(in, label)
}

// ParseADD parses the ADD instruction
func ParseADD(in <-chan LexerToken, label *string) ParsedInstruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rc, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return &ParsedInstructionRRR{
		ParsedGenericInstruction: ParsedGenericInstruction{
			MaybeLabel: label,
			Opcode:     OpcodeADD,
			RA:         ra,
		},
		RB: rb,
		RC: rc,
	}
}

// ParseADDI parses the ADDI instruction
func ParseADDI(in <-chan LexerToken, label *string) ParsedInstruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm7, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return &ParsedInstructionRRI{
		ParsedGenericInstruction: ParsedGenericInstruction{
			MaybeLabel: label,
			Opcode:     OpcodeADDI,
			RA:         ra,
		},
		RB:   rb,
		Imm7: imm7,
	}
}

// ParseRegisterOrComma parses a register ignoring a comma
// that may or may not appear before the register.
func ParseRegisterOrComma(in <-chan LexerToken) (uint16, error) {
	token := <-in
	switch token.Type {
	case LexerNameOrNumber:
	case LexerComma:
		// skip the optional comma
		token = <-in
		switch token.Type {
		case LexerNameOrNumber:
		default:
			return 0, fmt.Errorf("%w while parsing register name on line %d",
				ErrExpectedNameOrNumber, token.Lineno)
		}
	default:
		return 0, fmt.Errorf("%w while parsing register name on line %d",
			ErrExpectedNameOrNumber, token.Lineno)
	}
	switch v := strings.TrimPrefix(token.Value, "r"); v {
	case "0", "1", "2", "3", "4", "5", "6", "7":
		n, _ := strconv.Atoi(v)
		return uint16(n), nil
	default:
		return 0, fmt.Errorf("%w while parsing register name '%s' on line %d",
			ErrInvalidRegisterName, token.Value, token.Lineno)
	}
}

// ParseImmediateOrComma parses an immediate ignoring a comma
// that may or may not appear before the register.
func ParseImmediateOrComma(in <-chan LexerToken) (string, error) {
	token := <-in
	switch token.Type {
	case LexerNameOrNumber:
	case LexerComma:
		// skip the optional comma
		token = <-in
		switch token.Type {
		case LexerNameOrNumber:
		default:
			return "", fmt.Errorf("%w while parsing register name on line %d",
				ErrExpectedNameOrNumber, token.Lineno)
		}
	default:
		return "", fmt.Errorf("%w while parsing immediate on line %d",
			ErrExpectedNameOrNumber, token.Lineno)
	}
	return token.Value, nil
}

// ParseEOL expects to find the end of line token.
func ParseEOL(in <-chan LexerToken) error {
	token := <-in
	switch token.Type {
	case LexerEOL:
		return nil
	default:
		return fmt.Errorf("%w while processing instruction on line %d",
			ErrExpectedEOL, token.Lineno)
	}
}
