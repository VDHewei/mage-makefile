package main

import (
	"fmt"
	"os"
	"time"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
)

func main() {
	path := "testdata/realworld/ai-fox/Makefile"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	fmt.Println("File:", path, "-", len(data), "bytes")

	done := make(chan struct{}, 1)
	var result interface{}
	go func() {
		p := parser.NewParser(string(data))
		mf, err := p.Parse()
		result = []interface{}{mf, err}
		close(done)
	}()

	select {
	case <-done:
		mf := result.([]interface{})[0].(*parser.Makefile)
		err := result.([]interface{})[1]
		if err != nil {
			fmt.Println("Parse error:", err)
			return
		}
		fmt.Printf("Parsed OK: %d targets, %d variables\n", len(mf.Targets), len(mf.Variables))
		fmt.Println("Variables:")
		for _, v := range mf.Variables {
			fmt.Printf("  %s %s %q\n", v.Name, v.Operator, v.Value)
		}
		fmt.Println("Targets:")
		for _, t := range mf.Targets {
			fmt.Printf("  %q (line %d, prereqs=%v)\n", t.Name, t.Line, t.Prerequisites)
			for _, r := range t.Recipes {
				fmt.Printf("    recipe: %q\n", r)
			}
		}
	case <-time.After(5 * time.Second):
		fmt.Println("TIMEOUT - parser hung!")
	}
	os.Exit(0)
}
