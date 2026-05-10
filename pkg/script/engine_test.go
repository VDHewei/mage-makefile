package script

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Lua Engine Tests ---

func TestLuaEngine_Execute(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	result, err := engine.Execute(`print("hello from lua")`, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "hello from lua")
}

func TestLuaEngine_ExecuteWithEnv(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	env := map[string]string{
		"NAME":  "world",
		"COUNT": "42",
	}

	result, err := engine.Execute(`print("hello " .. env.NAME .. " count=" .. env.COUNT)`, env)
	require.NoError(t, err)
	assert.Contains(t, result, "hello world")
	assert.Contains(t, result, "count=42")
}

func TestLuaEngine_ExecuteArithmetic(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	result, err := engine.Execute(`print(2 + 3)`, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "5")
}

func TestLuaEngine_ValidateValid(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	err := engine.Validate(`print("valid")`)
	assert.NoError(t, err)
}

func TestLuaEngine_ValidateSyntaxError(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	err := engine.Validate(`print("missing quote)`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lua syntax error")
}

func TestLuaEngine_ExecuteSyntaxError(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	_, err := engine.Execute(`invalid syntax !!!`, nil)
	assert.Error(t, err)
}

func TestLuaEngine_TimeOut(t *testing.T) {
	engine := NewLuaEngine(100 * time.Millisecond)

	_, err := engine.Execute(`
		local x = 0
		for i = 1, 1000000000 do
			x = x + i
		end
		print("wont reach")
	`, nil)
	assert.Error(t, err)
}

func TestLuaEngine_Name(t *testing.T) {
	engine := NewLuaEngine(30 * time.Second)
	assert.Equal(t, "Lua Script Engine", engine.Name())
	assert.Equal(t, EngineLua, engine.Type())
}

func TestLuaEngine_SandboxNoIO(t *testing.T) {
	engine := NewLuaEngine(10 * time.Second)

	// Trying to access removed os module should fail
	_, err := engine.Execute(`os.exit(0)`, nil)
	assert.Error(t, err)
}

// --- JS Engine Tests ---

func TestJSEngine_Execute(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	result, err := engine.Execute(`console.log("hello from js")`, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "hello from js")
}

func TestJSEngine_ExecuteWithEnv(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	env := map[string]string{
		"NAME":  "world",
		"COUNT": "42",
	}

	result, err := engine.Execute(`console.log("hello " + env.NAME + " count=" + env.COUNT)`, env)
	require.NoError(t, err)
	assert.Contains(t, result, "hello world")
	assert.Contains(t, result, "count=42")
}

func TestJSEngine_ExecuteReturnValue(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	result, err := engine.Execute(`2 + 3`, nil)
	require.NoError(t, err)
	assert.Equal(t, "5", result)
}

func TestJSEngine_ValidateValid(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	err := engine.Validate(`console.log("valid")`)
	assert.NoError(t, err)
}

func TestJSEngine_ValidateSyntaxError(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	err := engine.Validate(`console.log("missing quote)`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "js syntax error")
}

func TestJSEngine_ExecuteSyntaxError(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	_, err := engine.Execute(`invalid syntax !!!`, nil)
	assert.Error(t, err)
}

func TestJSEngine_Name(t *testing.T) {
	engine := NewJSEngine(30 * time.Second)
	assert.Equal(t, "JavaScript Engine (goja)", engine.Name())
	assert.Equal(t, EngineJS, engine.Type())
}

func TestJSEngine_MultipleConsoleLogs(t *testing.T) {
	engine := NewJSEngine(10 * time.Second)

	result, err := engine.Execute(`
		console.log("line1");
		console.log("line2");
		console.log("line3");
	`, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line2")
	assert.Contains(t, result, "line3")
}

// --- Go Engine Tests ---

func TestGoEngine_Execute(t *testing.T) {
	engine := NewGoEngine()

	result, err := engine.Execute(`
		fmt.Println("hello from go")
	`, nil)
	require.NoError(t, err)
	assert.Contains(t, result, "package main")
	assert.Contains(t, result, "func main()")
	assert.Contains(t, result, "fmt.Println(\"hello from go\")")
}

func TestGoEngine_ExecuteWithEnv(t *testing.T) {
	engine := NewGoEngine()

	env := map[string]string{
		"NAME": "world",
	}

	result, err := engine.Execute(`print(NAME)`, env)
	require.NoError(t, err)
	assert.Contains(t, result, "NAMEEnv")
	assert.Contains(t, result, "world")
}

func TestGoEngine_ValidateValid(t *testing.T) {
	engine := NewGoEngine()

	err := engine.Validate(`{ valid code }`)
	assert.NoError(t, err)
}

func TestGoEngine_ValidateEmpty(t *testing.T) {
	engine := NewGoEngine()

	err := engine.Validate("")
	assert.Error(t, err)
}

func TestGoEngine_ValidateUnbalancedBraces(t *testing.T) {
	engine := NewGoEngine()

	err := engine.Validate(`{ unbalanced`)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unbalanced braces")
}

func TestGoEngine_Name(t *testing.T) {
	engine := NewGoEngine()
	assert.Equal(t, "Go Script Engine", engine.Name())
	assert.Equal(t, EngineGo, engine.Type())
}

func TestGoEngine_ExecuteEmpty(t *testing.T) {
	engine := NewGoEngine()

	_, err := engine.Execute("", nil)
	assert.Error(t, err)
}
