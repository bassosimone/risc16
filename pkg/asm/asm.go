// Package asm contains the RiSC-16 assembler.
//
// See https://user.eng.umd.edu/~blj/RiSC/.
//
// Extentions
//
// This assembler features the following extensions:
//
// 1. it is possible to put a comma between the instruction name
// and the first register name, thus resulting in a language that
// would be rejected by the original parser written in C.
package asm

import "io"

// InstructionOrError contains either an assembled instruction
// or an error that occurred during the assemblation.
type InstructionOrError struct {
	Instruction uint16
	Error       error
	Lineno      int
}

// StartAssembler starts the assembler in a background goroutine an
// returns a sequence of InstructionOrError.
func StartAssembler(r io.Reader) <-chan InstructionOrError {
	out := make(chan InstructionOrError)
	go AssemblerAsync(r, out)
	return out
}

// AssemblerAsync runs the assembler. It reads from the input reader
// and it writes InstructionOrError on the output channel.
func AssemblerAsync(r io.Reader, out chan<- InstructionOrError) {
	defer close(out)
	var idx int64
	labels := make(map[string]int64)
	var instructions []Instruction
	for instr := range StartParsing(StartLexing(r)) {
		if instr.Label() != nil {
			labels[*instr.Label()] = idx
		}
		if instr.Err() != nil {
			out <- InstructionOrError{Error: instr.Err()}
			return
		}
		instructions = append(instructions, instr)
		idx++
	}
	for _, instr := range instructions {
		encoded, err := instr.Encode(labels)
		if err != nil {
			out <- InstructionOrError{Error: err}
			continue
		}
		out <- InstructionOrError{Instruction: encoded, Lineno: instr.Line()}
	}
}
