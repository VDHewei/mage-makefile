package script

import (
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// JSEngine implements ScriptEngine for JavaScript (ES5.1+) scripts using goja.
type JSEngine struct {
	Timeout time.Duration
}

// NewJSEngine creates a new JSEngine with the given timeout.
func NewJSEngine(timeout time.Duration) *JSEngine {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &JSEngine{Timeout: timeout}
}

// Name returns the human-readable name of this engine.
func (e *JSEngine) Name() string {
	return "JavaScript Engine (goja)"
}

// Type returns the engine type identifier.
func (e *JSEngine) Type() EngineType {
	return EngineJS
}

// Validate checks the JavaScript script syntax.
func (e *JSEngine) Validate(script string) error {
	vm := goja.New()

	// parseSrc validates syntax without executing
	_, err := goja.Compile("validate", script, true)
	if err != nil {
		return fmt.Errorf("js syntax error: %w", err)
	}
	_ = vm
	return nil
}

// Execute runs the JavaScript script in a sandboxed VM and returns the result.
func (e *JSEngine) Execute(script string, env map[string]string) (string, error) {
	vm := goja.New()

	// Set up console.log support
	var outputBuilder strings.Builder
	consoleObj := vm.NewObject()
	consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []string
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		line := strings.Join(args, " ")
		outputBuilder.WriteString(line)
		outputBuilder.WriteString("\n")
		return goja.Undefined()
	})
	consoleObj.Set("error", consoleObj.Get("log"))
	consoleObj.Set("warn", consoleObj.Get("log"))
	vm.Set("console", consoleObj)

	// Set environment variables
	envObj := vm.NewObject()
	for k, v := range env {
		envObj.Set(k, v)
	}
	vm.Set("env", envObj)

	for k, v := range env {
		vm.Set("ENV_"+k, v)
	}

	// Run the script
	v, err := vm.RunString(script)
	if err != nil {
		return "", fmt.Errorf("js execution error: %w", err)
	}

	// If there's console output, use it; otherwise use the return value
	output := strings.TrimSpace(outputBuilder.String())
	if output == "" {
		if v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
			output = v.String()
		}
	}

	_ = e.Timeout // reserved for future timeout implementation via context

	return output, nil
}

// Validate for goja: compile only, don't run.
func (e *JSEngine) compileCheck(script string) error {
	_, err := goja.Compile("script", script, true)
	return err
}
