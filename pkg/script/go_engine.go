package script

import (
	"fmt"
	"strings"
)

// GoEngine implements ScriptEngine for native Go code snippets.
// It does not compile/run Go code directly; instead it validates
// and prepares Go code templates for embedding in magefile.go.
type GoEngine struct{}

// NewGoEngine creates a new GoEngine.
func NewGoEngine() *GoEngine {
	return &GoEngine{}
}

// Name returns the human-readable name of this engine.
func (e *GoEngine) Name() string {
	return "Go Script Engine"
}

// Type returns the engine type identifier.
func (e *GoEngine) Type() EngineType {
	return EngineGo
}

// Validate performs basic syntax validation on Go code.
// This is a simplified check; full Go syntax validation
// requires the go compiler toolchain.
func (e *GoEngine) Validate(script string) error {
	script = strings.TrimSpace(script)
	if script == "" {
		return fmt.Errorf("go script is empty")
	}

	// Basic checks for balanced braces
	if !e.balancedBraces(script) {
		return fmt.Errorf("go syntax error: unbalanced braces")
	}

	return nil
}

// Execute returns Go code as a string for embedding in magefile.go.
// It wraps the script with environment variable mappings.
func (e *GoEngine) Execute(script string, env map[string]string) (string, error) {
	script = strings.TrimSpace(script)
	if script == "" {
		return "", fmt.Errorf("go script is empty")
	}

	// Build environment variable setup
	var envSetup strings.Builder
	for k, v := range env {
		envSetup.WriteString(fmt.Sprintf(
			"\t%sEnv := \"%s\"\n",
			strings.ToUpper(k),
			strings.ReplaceAll(v, "\\", "\\\\"),
		))
		envSetup.WriteString(fmt.Sprintf(
			"\t_ = %sEnv\n",
			strings.ToUpper(k),
		))
	}

	// Wrap the script in a Go-compatible format
	var output strings.Builder
	output.WriteString("// Generated Go code from script\n")
	output.WriteString("package main\n\n")
	output.WriteString("func main() {\n")
	if envSetup.Len() > 0 {
		output.WriteString(envSetup.String())
		output.WriteString("\n")
	}
	output.WriteString(script)
	output.WriteString("\n}\n")

	return output.String(), nil
}

// balancedBraces checks if curly braces and parentheses are balanced.
func (e *GoEngine) balancedBraces(script string) bool {
	var braceStack []rune
	pairs := map[rune]rune{'}': '{', ')': '(', ']': '['}

	for _, ch := range script {
		switch ch {
		case '{', '(', '[':
			braceStack = append(braceStack, ch)
		case '}', ')', ']':
			if len(braceStack) == 0 {
				return false
			}
			last := braceStack[len(braceStack)-1]
			if last != pairs[ch] {
				return false
			}
			braceStack = braceStack[:len(braceStack)-1]
		}
	}
	return len(braceStack) == 0
}
