package main

import (
	"fmt"
	"os"
	"time"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

func main() {
	input := `.PHONY: build
build:
	echo hello`

	p := parser.NewParser(input)
	mf, _ := p.Parse()
	fmt.Printf("Parse: %d targets\n", len(mf.Targets))

	tr := transformer.NewTransformerWithPlatform("linux")
	done := make(chan bool)
	go func() {
		tr.Transform(mf)
		close(done)
	}()
	select {
	case <-done:
		fmt.Println("OK")
	case <-time.After(3 * time.Second):
		fmt.Println("TIMEOUT even on simple Makefile!")
		os.Exit(1)
	}
	fmt.Println("All good")
}
