package script

import (
	"context"
	"fmt"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// LuaEngine implements ScriptEngine for Lua scripts.
type LuaEngine struct {
	Timeout time.Duration
}

// NewLuaEngine creates a new LuaEngine with the given timeout.
func NewLuaEngine(timeout time.Duration) *LuaEngine {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &LuaEngine{Timeout: timeout}
}

// Name returns the human-readable name of this engine.
func (e *LuaEngine) Name() string {
	return "Lua Script Engine"
}

// Type returns the engine type identifier.
func (e *LuaEngine) Type() EngineType {
	return EngineLua
}

// Validate checks the Lua script syntax.
func (e *LuaEngine) Validate(script string) error {
	L := lua.NewState()
	defer L.Close()

	// Use LoadString to validate syntax without executing
	fn, err := L.LoadString(script)
	if err != nil {
		return fmt.Errorf("lua syntax error: %w", err)
	}
	_ = fn
	return nil
}

// Execute runs the Lua script in a sandboxed VM and returns the result.
func (e *LuaEngine) Execute(script string, env map[string]string) (string, error) {
	L := lua.NewState()
	defer L.Close()

	// Restrict dangerous functions for sandboxing
	e.sandbox(L)

	// Set environment variables
	for k, v := range env {
		L.SetGlobal("ENV_"+k, lua.LString(v))
	}

	// Set environment table
	envTable := L.NewTable()
	for k, v := range env {
		L.SetField(envTable, k, lua.LString(v))
	}
	L.SetGlobal("env", envTable)

	// Set timeout via Go context
	if e.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
		defer cancel()
		L.SetContext(ctx)
	}

	// Execute the script
	if err := L.DoString(script); err != nil {
		return "", fmt.Errorf("lua execution error: %w", err)
	}

	// Collect output from Lua's print statements
	output := extractLuaOutput(L)
	if output == "" {
		// If no output was printed, try to get the return value
		if L.GetTop() > 0 {
			output = L.ToString(-1)
		}
	}

	return strings.TrimSpace(output), nil
}

// sandbox removes potentially dangerous functions from the Lua VM.
func (e *LuaEngine) sandbox(L *lua.LState) {
	// Remove dangerous modules
	L.SetGlobal("os", lua.LNil)
	L.SetGlobal("io", lua.LNil)
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)

	// Provide a safe print function that captures output
	printTable := L.NewTable()
	L.SetField(printTable, "print", L.NewFunction(func(L *lua.LState) int {
		var parts []string
		n := L.GetTop()
		for i := 1; i <= n; i++ {
			parts = append(parts, L.ToString(i))
		}
		output := strings.Join(parts, "\t")
		// Store captured output in a global variable
		captured := L.GetGlobal("__captured_output")
		if captured == lua.LNil {
			L.SetGlobal("__captured_output", lua.LString(output))
		} else {
			L.SetGlobal("__captured_output", lua.LString(captured.String()+"\n"+output))
		}
		L.SetGlobal("print", lua.LString(output))
		return 0
	}))
	L.SetGlobal("print", printTable.RawGetString("print"))
}

// extractLuaOutput retrieves captured output from the Lua state.
func extractLuaOutput(L *lua.LState) string {
	output := L.GetGlobal("print")
	if output == lua.LNil {
		return ""
	}
	if s, ok := output.(lua.LString); ok {
		return string(s)
	}
	return output.String()
}
