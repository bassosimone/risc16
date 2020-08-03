package asm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseSpecificInstruction is the function parsing a specific instruction.
type ParseSpecificInstruction func(in <-chan LexerToken, label *string) []Instruction

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
			return
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
	return parser(in, label)
}

// ParseADD parses the ADD instruction
func ParseADD(in <-chan LexerToken, label *string) []Instruction {
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
	return []Instruction{InstructionADD{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		RC:         rc,
	}}
}

// ParseADDI parses the ADDI instruction
func ParseADDI(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionADDI{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseNAND parses the NAND instruction
func ParseNAND(in <-chan LexerToken, label *string) []Instruction {
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
	return []Instruction{InstructionNAND{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		RC:         rc,
	}}
}

// ParseLUI parses the LUI instruction
func ParseLUI(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionLUI{
		MaybeLabel: label,
		RA:         ra,
		Imm:        imm,
	}}
}

// ParseSW parses the SW instruction
func ParseSW(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionSW{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseLW parses the LW instruction
func ParseLW(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionLW{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseBEQ parses the BEQ instruction
func ParseBEQ(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionBEQ{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
		Imm:        imm,
	}}
}

// ParseJALR parses the JALR instruction
func ParseJALR(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	rb, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	return []Instruction{InstructionJALR{
		MaybeLabel: label,
		RA:         ra,
		RB:         rb,
	}}
}

// ParseNOP parses the NOP pseudo-instruction
func ParseNOP(in <-chan LexerToken, label *string) []Instruction {
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// NOP is mapped to ADD r0 r0 r0
	return []Instruction{InstructionADD{MaybeLabel: label}}
}

// ParseHALT parses the HALT pseudo-instruction
func ParseHALT(in <-chan LexerToken, label *string) []Instruction {
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// HALT is mapped to JALR r0 r0 <special-value>.
	return []Instruction{InstructionJALR{
		MaybeLabel: label,
		Imm:        ExceptionTypeEXCEPTION | ExceptionValueHALT,
	}}
}

// ParseLLI parses the LLI pseudo-instruction
func ParseLLI(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// LLI translates to ADDI RA RA (Imm & 0x3f)
	return []Instruction{InstructionLLI{
		MaybeLabel: label,
		RA:         ra,
		Imm:        imm,
	}}
}

// ParseMOVI parses the MOVI pseudo-instruction
func ParseMOVI(in <-chan LexerToken, label *string) []Instruction {
	ra, err := ParseRegisterOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	imm, err := ParseImmediateOrComma(in)
	if err != nil {
		return NewParseError(err)
	}
	if err := ParseEOL(in); err != nil {
		return NewParseError(err)
	}
	// MOVI translates to LUI and LLI
	return []Instruction{
		InstructionLUI{
			MaybeLabel: label,
			RA:         ra,
			Imm:        imm,
		},
		InstructionLLI{
			RA:  ra,
			Imm: imm,
		},
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
