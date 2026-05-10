package main

import (
	"fmt"
	"time"

	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
)

func main() {
	tr := transformer.NewTransformerWithPlatform("linux")
	
	// Test with a string that might cause issues
	testCmds := []string{
		"echo \"Targets:\"",
		"echo hello",
		"command -v cargo >/dev/null 2>&1 && echo ok",
		"docker run --rm -e TEST=1 -v \"$(PWD)/dir:/data\" image sh -c \"echo hi\"",
	}
	
	for _, cmd := range testCmds {
		done := make(chan bool)
		var result string
		go func() {
			result = tr.MapCommand(cmd)
			close(done)
		}()
		select {
		case <-done:
			fmt.Printf("  %q -> %q\n", cmd, result)
		case <-time.After(2 * time.Second):
			fmt.Printf("  HANG on: %q\n", cmd)
		}
	}
	fmt.Println("Done")
}
