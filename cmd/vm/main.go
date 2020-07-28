package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/bassosimone/vmlang/pkg/vm"
)

func main() {
	log.SetFlags(0)
	debug := flag.Bool("d", false, "enable debugging")
	filename := flag.String("f", "", "file to run")
	verbose := flag.Bool("v", false, "be verbose")
	flag.Parse()
	if *filename == "" {
		log.Fatal("usage: vm [-d] [-v] -f <machine-code-file>")
	}
	fp, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	machine := new(vm.VM)
	scanner := bufio.NewScanner(fp)
	var addr uint16
	for scanner.Scan() {
		value, err := strconv.ParseUint(scanner.Text(), 16, 16)
		if err != nil {
			log.Fatal(err)
		}
		machine.M[addr] = uint16(value)
		addr++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	for {
		machine.Fetch()
		if *verbose {
			log.Printf("vm: %s\n", machine)
			log.Printf("vm: %#016b %s\n", machine.CI, vm.Disassemble(machine.CI))
		}
		if *debug {
			log.Printf("vm: paused...")
			fmt.Scanln()
		}
		if err := machine.Execute(); err != nil {
			if errors.Is(err, vm.ErrHalted) {
				break
			}
			log.Fatal(err)
		}
	}
}
