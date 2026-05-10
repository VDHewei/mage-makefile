package transformer

import (
	"strings"
	"testing"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformBasicTarget(t *testing.T) {
	input := `
.PHONY: build

build:
	go build -o bin/app ./...
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, m)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	assert.Equal(t, "build", ir.Targets[0].Name)
	assert.Equal(t, "Build", ir.Targets[0].FuncName)
	assert.True(t, ir.Targets[0].IsPhony)
	assert.True(t, ir.Targets[0].IsDefault)
	require.Len(t, ir.Targets[0].Commands, 1)
	assert.Contains(t, ir.Targets[0].Commands[0].Transformed, "go build")
}

func TestTransformMultipleTargets(t *testing.T) {
	input := `
.PHONY: build test clean

build:
	go build -o bin/app ./...

test:
	go test ./...

clean:
	rm -rf bin/
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 3)

	names := make(map[string]bool)
	for _, tgt := range ir.Targets {
		names[tgt.Name] = true
	}
	assert.True(t, names["build"])
	assert.True(t, names["test"])
	assert.True(t, names["clean"])

	for _, tgt := range ir.Targets {
		assert.True(t, tgt.IsPhony, "target %s should be phony", tgt.Name)
	}
}

func TestTransformVariableExpansion(t *testing.T) {
	input := `
APP_NAME = myapp
BUILD_DIR = bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./...
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	cmd := ir.Targets[0].Commands[0].Transformed
	assert.Contains(t, cmd, "bin/myapp")
}

func TestTransformVariables(t *testing.T) {
	input := `
GOFLAGS = -v -race
PKG = ./...

build:
	go build $(GOFLAGS) $(PKG)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Variables, 2)
	flags := ir.Variables[0]
	assert.Equal(t, "GOFLAGS", flags.Name)
	assert.Equal(t, "-v -race", flags.Value)

	cmd := ir.Targets[0].Commands[0].Transformed
	assert.Contains(t, cmd, "-v -race")
	assert.Contains(t, cmd, "./...")
}

func TestTransformAutoVars(t *testing.T) {
	input := `
output.txt: input.txt
	cp $< $@
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	target := ir.Targets[0]
	assert.Equal(t, "output.txt", target.Name)
	assert.Equal(t, "OutputTxt", target.FuncName)
	assert.Equal(t, []string{"input.txt"}, target.Prerequisites)

	cmd := target.Commands[0].Transformed
	assert.Contains(t, cmd, "input.txt")
	assert.Contains(t, cmd, "output.txt")
}

func TestTransformFuncNameConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"build", "Build"},
		{"serve-dev", "ServeDev"},
		{"build.all", "BuildAll"},
		{"my_target", "MyTarget"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toFuncName(tt.input))
		})
	}
}

func TestPlatformMappingUnix(t *testing.T) {
	pm := NewPlatformMapping("linux")
	assert.Equal(t, "linux", pm.TargetOS)

	// No mapping for Unix -> Unix
	result := pm.MapCommand("rm -rf bin")
	assert.Equal(t, "rm -rf bin", result)

	result = pm.MapCommand("cp file1 file2")
	assert.Equal(t, "cp file1 file2", result)
}

func TestPlatformMappingWindows(t *testing.T) {
	pm := NewPlatformMapping("windows")
	assert.Equal(t, "windows", pm.TargetOS)

	tests := []struct {
		input    string
		expected string
	}{
		{"rm -rf bin", "del -rf bin"},
		{"cp file1 file2", "copy file1 file2"},
		{"mv old new", "move old new"},
		{"cat file.txt", "type file.txt"},
		{"mkdir -p dir", "mkdir -p dir"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := pm.MapCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformCommandPrefix(t *testing.T) {
	input := `
build:
	@echo "Building..."
	-go build -o bin/app ./...
	+go vet ./...
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	require.Len(t, ir.Targets[0].Commands, 3)

	// @ prefix stripped, command preserved
	assert.Contains(t, ir.Targets[0].Commands[0].Transformed, "echo")
	assert.True(t, ir.Targets[0].Commands[0].CanUseSh)

	// - prefix stripped, command preserved, CanUseSh disabled
	assert.Contains(t, ir.Targets[0].Commands[1].Transformed, "go build")
	assert.False(t, ir.Targets[0].Commands[1].CanUseSh)

	// + prefix stripped
	assert.Contains(t, ir.Targets[0].Commands[2].Transformed, "go vet")
	assert.True(t, ir.Targets[0].Commands[2].CanUseSh)
}

func TestTransformEmptyMakefile(t *testing.T) {
	input := `# Just a comment
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	assert.Empty(t, ir.Targets)
	assert.Empty(t, ir.Variables)
}

func TestShellSplit(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"go build -o bin/app", []string{"go", "build", "-o", "bin/app"}},
		{`echo "hello world"`, []string{"echo", "hello world"}},
		{"echo 'single quoted'", []string{"echo", "single quoted"}},
		{"  ls -la  ", []string{"ls", "-la"}},
		{"", nil},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellSplit(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformShellVariable(t *testing.T) {
	input := `GOHOSTOS:=$(shell go env GOHOSTOS)
build:
	@echo $(GOHOSTOS)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, m)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Variables, 1)
	assert.Equal(t, "GOHOSTOS", ir.Variables[0].Name)
	assert.NotEmpty(t, ir.Variables[0].Value, "GOHOSTOS should be resolved from $(shell go env GOHOSTOS)")

	require.Len(t, ir.Targets, 1)
	cmd := ir.Targets[0].Commands[0].Transformed
	assert.Contains(t, cmd, ir.Variables[0].Value)
}

func TestTransformShellFailure(t *testing.T) {
	// Non-existent command should return empty string
	input := `RESULT:=$(shell nonexistent_cmd_xyz_123)
build:
	@echo $(RESULT)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Variables, 1)
	assert.Equal(t, "RESULT", ir.Variables[0].Name)
	assert.Empty(t, ir.Variables[0].Value, "Failed shell command should return empty string")
}

func TestTransformShellWithNestedVar(t *testing.T) {
	input := `GO_BINARY=go
build:
	@$(shell $(GO_BINARY) version)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	cmd := ir.Targets[0].Commands[0].Transformed
	// The inner $(GO_BINARY) should be expanded to "go", then executed as "go version"
	// The output should contain "go" (version string like "go version go1.x ...")
	assert.Contains(t, cmd, "go")
}

func TestTransformShellInRecipe(t *testing.T) {
	input := `build:
	echo "Current OS: $(shell go env GOHOSTOS)"
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	require.Len(t, ir.Targets, 1)
	cmd := ir.Targets[0].Commands[0].Transformed
	// The $(shell ...) call should be resolved in-place in the recipe
	assert.Contains(t, cmd, "Current OS:")
	assert.NotContains(t, cmd, "$(shell ", "shell call should be resolved")
}

func TestResolveShellCallsUnit(t *testing.T) {
	tr := NewTransformer()
	varMap := map[string]string{}

	// No shell calls - should return unchanged
	result := tr.resolveShellCalls("echo hello", varMap)
	assert.Equal(t, "echo hello", result)

	// Failed shell command - should return empty output
	result = tr.resolveShellCalls("$(shell nonexistent_cmd_xyz_123)", varMap)
	assert.Empty(t, result)

	// Multiple shell calls
	result = tr.resolveShellCalls("$(shell echo a)$(shell echo b)", varMap)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
}

// ==== Conditional tests ====

func TestTransformIfeqTrue(t *testing.T) {
	// ifeq with matching arguments should use the true branch
	input := `PLATFORM=linux
ifeq ($(PLATFORM), linux)
BIN_DIR=bin/linux
else
BIN_DIR=bin/other
endif
build:
	@echo $(BIN_DIR)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m.Conditionals, 1)

	tr := NewTransformer()
	ir := tr.Transform(m)

	// Should pick the true branch: BIN_DIR=bin/linux
	found := false
	for _, v := range ir.Variables {
		if v.Name == "BIN_DIR" {
			assert.Equal(t, "bin/linux", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found, "BIN_DIR should exist in IR variables")

	// Verify it's resolved in the recipe
	require.Len(t, ir.Targets, 1)
	assert.Contains(t, ir.Targets[0].Commands[0].Transformed, "bin/linux")
}

func TestTransformIfeqFalse(t *testing.T) {
	// ifeq with non-matching arguments should use the false branch
	input := `PLATFORM=windows
ifeq ($(PLATFORM), linux)
BIN_DIR=bin/linux
else
BIN_DIR=bin/other
endif
build:
	@echo $(BIN_DIR)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "BIN_DIR" {
			assert.Equal(t, "bin/other", v.Value, "false branch should be taken")
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformIfneqTrue(t *testing.T) {
	// ifneq with non-matching arguments should use the true branch
	input := `PLATFORM=windows
ifneq ($(PLATFORM), linux)
MSG=non-linux-platform
else
MSG=linux-platform
endif
build:
	@echo $(MSG)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "MSG" {
			assert.Equal(t, "non-linux-platform", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformIfdefTrue(t *testing.T) {
	// ifdef with a defined, non-empty variable should use the true branch
	input := `TOOLCHAIN=gcc
ifdef TOOLCHAIN
CC=gcc
else
CC=clang
endif
build:
	@echo $(CC)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "CC" {
			assert.Equal(t, "gcc", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformIfdefFalse(t *testing.T) {
	// ifdef with undefined variable should use the false branch
	input := `ifdef UNDEFINED_VAR_XYZ
CC=gcc
else
CC=clang
endif
build:
	@echo $(CC)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "CC" {
			assert.Equal(t, "clang", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformIfdefEmptyVar(t *testing.T) {
	// GNU Make: ifdef treats a variable defined with empty value as undefined
	input := `FOO=
ifdef FOO
RESULT=defined
else
RESULT=undefined
endif
build:
	@echo $(RESULT)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "RESULT" {
			assert.Equal(t, "undefined", v.Value, "empty var should be treated as undefined by ifdef")
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformIfndefTrue(t *testing.T) {
	// ifndef with undefined variable should use the true branch
	input := `ifndef UNDEFINED_VAR_XYZ
CC=default-cc
else
CC=custom-cc
endif
build:
	@echo $(CC)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "CC" {
			assert.Equal(t, "default-cc", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformConditionalNoElse(t *testing.T) {
	// Conditional without else branch - true branch should be used, false branch should be empty
	input := `DEBUG=1
ifeq ($(DEBUG), 1)
CFLAGS=-g -O0
endif
build:
	@echo $(CFLAGS)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "CFLAGS" {
			assert.Equal(t, "-g -O0", v.Value)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformConditionalMultipleAssignments(t *testing.T) {
	// Multiple variable assignments in a conditional branch
	input := `PLATFORM=linux
ifeq ($(PLATFORM), linux)
CC=gcc
CXX=g++
AR=ar
else
CC=clang
CXX=clang++
endif
build:
	@echo $(CC) $(CXX) $(AR)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	varMap := make(map[string]string)
	for _, v := range ir.Variables {
		varMap[v.Name] = v.Value
	}

	assert.Equal(t, "gcc", varMap["CC"])
	assert.Equal(t, "g++", varMap["CXX"])
	assert.Equal(t, "ar", varMap["AR"])
}

func TestTransformConditionalShellExpansion(t *testing.T) {
	// Conditional with $(shell ...) expansion in variables inside the body
	input := `ifeq (a, a)
GOHOSTOS:=$(shell go env GOHOSTOS)
endif
build:
	@echo $(GOHOSTOS)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "GOHOSTOS" {
			assert.NotEmpty(t, v.Value, "GOHOSTOS should be resolved from $(shell go env GOHOSTOS)")
			assert.False(t, strings.Contains(v.Value, "$(shell "), "shell call should be resolved")
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformConditionalVarInCondition(t *testing.T) {
	// Conditional where the condition uses a variable reference
	input := `GOHOSTOS:=$(shell go env GOHOSTOS)
BUILD_TARGET=linux
ifeq ($(GOHOSTOS), $(BUILD_TARGET))
IS_TARGET=true
else
IS_TARGET=false
endif
build:
	@echo $(IS_TARGET)
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	found := false
	for _, v := range ir.Variables {
		if v.Name == "IS_TARGET" {
			// This test determines the outcome at runtime based on current platform
			// Both GOHOSTOS and BUILD_TARGET are compared after expansion
			assert.True(t, v.Value == "true" || v.Value == "false",
				"IS_TARGET should be true or false based on platform match")
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransformConditionalKratosRealWorld(t *testing.T) {
	// Real-world kratos Makefile pattern: ifeq with platform check and static assignment
	input := `GOHOSTOS:=$(shell go env GOHOSTOS)
ifeq ($(GOHOSTOS), windows)
Git_Bash=C:/Program\\ Files/Git/bin/bash
INTERNAL_PROTO_FILES=internal/standalone/internal/compute/v1/compute.proto
API_PROTO_FILES=api/standalone/v1/standalone.proto
else
Git_Bash=/bin/bash
INTERNAL_PROTO_FILES=internal/standalone/internal/compute/v1/compute.proto
API_PROTO_FILES=api/standalone/v1/standalone.proto
endif
build:
	@$(Git_Bash) -c 'echo build'
`
	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := NewTransformer()
	ir := tr.Transform(m)

	// INTERNAL_PROTO_FILES and API_PROTO_FILES must be resolved regardless of platform
	varMap := make(map[string]string)
	for _, v := range ir.Variables {
		varMap[v.Name] = v.Value
	}

	assert.Equal(t, "internal/standalone/internal/compute/v1/compute.proto", varMap["INTERNAL_PROTO_FILES"],
		"INTERNAL_PROTO_FILES should be resolved from conditional")
	assert.Equal(t, "api/standalone/v1/standalone.proto", varMap["API_PROTO_FILES"],
		"API_PROTO_FILES should be resolved from conditional")
	assert.NotEmpty(t, varMap["Git_Bash"],
		"Git_Bash should be resolved (win or unix path based on current OS)")
}
