package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bassosimone/risc16/pkg/asm"
)

func main() {
	log.SetFlags(0)
	filename := flag.String("f", "", "file to process")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()
	if *filename == "" {
		log.Fatal("usage: asm -f <assmebly-code-file>")
	}
	fp, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	for instr := range asm.StartAssembler(fp) {
		if instr.Error != nil {
			log.Fatal(instr.Error)
		}
		fmt.Printf("%04x", instr.Instruction)
		if *debug {
			fmt.Printf("  # %d", instr.Lineno)
		}
		fmt.Println("")
	}
}
