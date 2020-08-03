package asm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseSpecificInstruction is the function parsing a specific instruction.
type ParseSpecificInstruction func(
	in <-chan LexerToken, label *string, lineno int) []Instruction

// InstructionParsers maps an instruction to its parser.
var InstructionParsers = map[string]ParseSpecificInstruction{
	"add":    ParseADD,
	"addi":   ParseADDI,
	"nand":   ParseNAND,
	"lui":    ParseLUI,
	"sw":     ParseSW,
	"lw":     ParseLW,
	"beq":    ParseBEQ,
	"jalr":   ParseJALR,
	"nop":    ParseNOP,
	"halt":   ParseHALT,
	"lli":    ParseLLI,
	"movi":   ParseMOVI,
	".fill":  ParseFILL,
	".space": ParseSPACE,
}

// The following errors may occur when assembling.
var (
	ErrExpectedNameOrNumber = errors.New("asm: expected name or number")
	ErrUnknownInstruction   = errors.New("asm: unknown instruction")
	ErrExpectedComma        = errors.New("asm: expected comma")
	ErrExpectedEOL          = errors.New("asm: expected end of line")
	ErrInvalidRegisterName  = errors.New("asm: invalid register name")
	ErrOutOfRange           = errors.New("asm: immediate value out of range")
	ErrCannotEncode         = errors.New("asm: can't encode instruction")
	ErrTooManyInstructions  = errors.New("asm: too many instructions")
)

// StartParsing starts parsing in a backend goroutine.
func StartParsing(in <-chan LexerToken) <-chan Instruction {
	out := make(chan Instruction)
	go ParseAsync(in, out)
	return out
}

// ParseAsync is the async instructions parser.
func ParseAsync(in <-chan LexerToken, out chan<- Instruction) {
	defer func() {
		for range in {
			// drain channel (for robustness)
		}
		close(out)
	}()
	for {
		instr := ParseSingleInstruction(in)
		if instr == nil {
			return // this is end of lexing
		}
		for _, i := range instr {
			out <- i
			if i.Err() != nil {
				return
			}
		}
	}
}

// ParseSingleInstruction parses an instruction.
func ParseSingleInstruction(in <-chan LexerToken) []Instruction {
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
	return parser(in, label, token.Lineno)
}

// ParseADD parses the ADD instruction
func ParseADD(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rc, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionADD{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		RC:         rc,
	}}
}

// ParseADDI parses the ADDI instruction
func ParseADDI(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionADDI{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseNAND parses the NAND instruction
func ParseNAND(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rc, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionNAND{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		RC:         rc,
	}}
}

// ParseLUI parses the LUI instruction
func ParseLUI(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionLUI{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		Imm:        imm,
	}}
}

// ParseSW parses the SW instruction
func ParseSW(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionSW{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseLW parses the LW instruction
func ParseLW(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionLW{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseBEQ parses the BEQ instruction
func ParseBEQ(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionBEQ{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseJALR parses the JALR instruction
func ParseJALR(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionJALR{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
	}}
}

// ParseNOP parses the NOP pseudo-instruction
func ParseNOP(in <-chan LexerToken, label *string, lineno int) []Instruction {
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// NOP is mapped to ADD r0 r0 r0
	return []Instruction{InstructionADD{Lineno: lineno, MaybeLabel: label}}
}

// ParseHALT parses the HALT pseudo-instruction
func ParseHALT(in <-chan LexerToken, label *string, lineno int) []Instruction {
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// HALT is mapped to JALR r0 r0 <special-value>.
	return []Instruction{InstructionJALR{
		Lineno:     lineno,
		MaybeLabel: label,
		Imm:        ExceptionTypeEXCEPTION | ExceptionValueHALT,
	}}
}

// ParseLLI parses the LLI pseudo-instruction
func ParseLLI(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// LLI translates to ADDI RA RA (Imm & 0x3f)
	return []Instruction{InstructionLLI{
		Lineno:     lineno,
		MaybeLabel: label,
		RA:         ra,
		Imm:        imm,
	}}
}

// ParseMOVI parses the MOVI pseudo-instruction
func ParseMOVI(in <-chan LexerToken, label *string, lineno int) []Instruction {
	ra, err := MaybeSkipCommaThenParseRegister(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// MOVI translates to LUI and LLI
	return []Instruction{
		InstructionLUI{
			Lineno:     lineno,
			MaybeLabel: label,
			RA:         ra,
			Imm:        imm,
		},
		InstructionLLI{
			Lineno:     lineno,
			MaybeLabel: nil, // no label for second instruction
			RA:         ra,
			Imm:        imm,
		},
	}
}

// ParseFILL parses the .FILL pseudo-instruction
func ParseFILL(in <-chan LexerToken, label *string, lineno int) []Instruction {
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	value, err := strconv.ParseInt(imm, 0, 16)
	if err != nil {
		return NewParseError(fmt.Errorf("%w for data", ErrOutOfRange))
	}
	return []Instruction{InstructionDATA{
		Lineno:     lineno,
		MaybeLabel: label,
		Value:      uint16(value),
	}}
}

// ParseSPACE parses the .SPACE pseudo-instruction
func ParseSPACE(in <-chan LexerToken, label *string, lineno int) (out []Instruction) {
	imm, err := MaybeSkipCommaThenParseImmediate(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	count, err := strconv.ParseUint(imm, 0, 16)
	if err != nil || count <= 0 {
		return NewParseError(fmt.Errorf("%w for data", ErrOutOfRange))
	}
	for i := uint64(0); i < count; i++ {
		out = append(out, InstructionDATA{Lineno: lineno, MaybeLabel: label})
		label = nil
	}
	return
}

// ParseRegisterOrComma parses a register ignoring a comma
// that may or may not appear before the register.
func MaybeSkipCommaThenParseRegister(in <-chan LexerToken) (uint16, error) {
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
func MaybeSkipCommaThenParseImmediate(in <-chan LexerToken) (string, error) {
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
