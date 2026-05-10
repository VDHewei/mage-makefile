package main

import (
	"fmt"
	"os"

	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

func main() {
	path := "testdata/realworld/kratos/Makefile"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	fmt.Println("File:", path, "-", len(data), "bytes")

	p := parser.NewParser(string(data))
	mf, err := p.Parse()
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	fmt.Printf("Parse OK: %d targets, %d vars, %d conditionals\n",
		len(mf.Targets), len(mf.Variables), len(mf.Conditionals))

	tr := transformer.NewTransformerWithPlatform("linux")
	ir := tr.Transform(mf)
	fmt.Printf("Transform OK: %d targets\n", len(ir.Targets))

	gen := generator.NewGenerator(ir)
	code, err := gen.Generate()
	if err != nil {
		fmt.Println("Generate error:", err)
		return
	}
	fmt.Printf("Generate OK: %d bytes\n", len(code))
	fmt.Println("--- FULL CODE ---")
	fmt.Println(code)
}
