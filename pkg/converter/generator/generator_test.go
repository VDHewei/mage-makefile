package generator

import (
	"strings"
	"testing"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateBasicTarget(t *testing.T) {
	input := `
.PHONY: build

build:
	go build -o bin/app ./...
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, code)

	assert.Contains(t, code, "package main")
	assert.Contains(t, code, "//go:build mage")
	assert.Contains(t, code, `"github.com/magefile/mage/sh"`)
	assert.Contains(t, code, "func Build() error {")
	assert.Contains(t, code, "sh.Run(")
	assert.Contains(t, code, `"go"`)
	assert.Contains(t, code, `"build"`)
}

func TestGenerateMultipleTargets(t *testing.T) {
	input := `
.PHONY: build test clean

build:
	go build -o bin/app ./...

test:
	go test ./...

clean:
	rm -rf bin/
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	assert.Contains(t, code, "func Build() error {")
	assert.Contains(t, code, "func Test() error {")
	assert.Contains(t, code, "func Clean() error {")
}

func TestGenerateVariableConstants(t *testing.T) {
	input := `
APP_NAME = myapp
BUILD_DIR = build

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./...
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	assert.Contains(t, code, "const APP_NAME =")
	assert.Contains(t, code, `"myapp"`)
	assert.Contains(t, code, "const BUILD_DIR =")
	assert.Contains(t, code, `"build"`)
}

func TestGenerateTargetWithDeps(t *testing.T) {
	input := `
.PHONY: all build test

all: build test
	echo "All done"

build:
	go build ./...

test:
	go test ./...
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	assert.Contains(t, code, `"github.com/magefile/mage/mg"`)
	assert.Contains(t, code, "mg.Deps(mg.F(Build))")
	assert.Contains(t, code, "mg.Deps(mg.F(Test))")
}

func TestGenerateComplexCommand(t *testing.T) {
	input := `
build:
	go build -o bin/app ./... && echo "done"
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Complex commands (with &&) use exec.Command
	assert.Contains(t, code, "exec.Command")
	assert.Contains(t, code, "cmd.Stdout = os.Stdout")
	assert.Contains(t, code, "cmd.Stderr = os.Stderr")
}

func TestGenerateEmptyMakefile(t *testing.T) {
	ir := &transformer.IR{
		PackageName: "main",
	}
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	assert.Contains(t, code, "package main")
}

func TestGenerateNilIR(t *testing.T) {
	gen := NewGenerator(nil)
	_, err := gen.Generate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGenerateCodeIsCompilableShape(t *testing.T) {
	input := `
BINARY = server
PKG = ./cmd/server

.PHONY: build test

build:
	go build -o bin/$(BINARY) $(PKG)

test:
	go test -v -race ./...
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	// Verify basic structure
	assert.True(t, strings.HasPrefix(code, "//go:build mage\n"), "missing build tag")
	assert.Contains(t, code, "package main")
	assert.Contains(t, code, "import (")

	// Should have one function per target
	assert.Equal(t, 2, strings.Count(code, "func "))

	// Should have const definitions
	assert.Contains(t, code, "const ")

	// Should return error from each function
	assert.Contains(t, code, ") error {")
	assert.Contains(t, code, "return nil")

	// Should close all braces properly
	openBraces := strings.Count(code, "{")
	closeBraces := strings.Count(code, "}")
	assert.Equal(t, openBraces, closeBraces, "unbalanced braces")
}

func TestGenerateShellCommands(t *testing.T) {
	input := `
build:
	@echo "building..."
	@mkdir -p bin
	@go build -o bin/app ./...
	@cp config.yaml bin/
`

	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)

	assert.Contains(t, code, "echo")
	assert.Contains(t, code, "mkdir")
	assert.Contains(t, code, "go")
}

func TestRenderTemplate(t *testing.T) {
	data := &TemplateData{
		PackageName: "main",
		Imports:     []string{"github.com/magefile/mage/sh"},
		Constants: []ConstDef{
			{Name: "AppName", Value: "myapp"},
		},
		Functions: []FuncDef{
			{
				Name:        "Build",
				Description: "builds the project",
				Body:        "\n\treturn sh.Run(\"go\", \"build\", \"./...\")\n\treturn nil\n",
			},
		},
	}

	code, err := renderTemplate(data)
	require.NoError(t, err)

	assert.Contains(t, code, "package main")
	assert.Contains(t, code, "const AppName = \"myapp\"")
	assert.Contains(t, code, "func Build() error {")
}

func TestToGoIdent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GOFLAGS", "GOFLAGS"},
		{"go-flags", "Go_flags"},
		{"app.name", "App_name"},
		{"123var", "_123var"},
		{"build-dir", "Build_dir"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToGoIdent(tt.input))
		})
	}
}

func TestCollectImports(t *testing.T) {
	// IR with simple commands only
	ir := &transformer.IR{
		PackageName: "main",
		Targets: []transformer.IRTarget{
			{
				Name:     "build",
				FuncName: "Build",
				Commands: []transformer.IRCommand{
					{CanUseSh: true, Args: []string{"go", "build"}, Transformed: "go build"},
				},
			},
		},
	}
	gen := NewGenerator(ir)
	imports := gen.collectImports()
	assert.Contains(t, imports, "github.com/magefile/mage/sh")
	assert.NotContains(t, imports, "os/exec")

	// IR with complex commands
	ir.Targets[0].Commands = []transformer.IRCommand{
		{CanUseSh: false, Args: []string{"go", "build", "-o", "app"}, Transformed: "go build -o app"},
	}
	imports = gen.collectImports()
	assert.Contains(t, imports, "os/exec")
}

func TestGenerateShellVariable(t *testing.T) {
	input := `GOHOSTOS:=$(shell go env GOHOSTOS)
build:
	@echo $(GOHOSTOS)
`
	ir := parseAndTransform(t, input)
	require.NotEmpty(t, ir.Variables, "should have at least one variable")

	// The shell command should have been executed
	assert.NotEmpty(t, ir.Variables[0].Value, "GOHOSTOS should be resolved from shell")

	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// The generated code should contain the resolved value
	assert.Contains(t, code, "const GOHOSTOS =")
	assert.NotContains(t, code, "shell", "generated code should not contain raw $(shell ...)")
}

func TestGenerateShellInRecipe(t *testing.T) {
	input := `build:
	echo "OS: $(shell go env GOHOSTOS)"
`
	ir := parseAndTransform(t, input)
	gen := NewGenerator(ir)
	code, err := gen.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// The generated code should be valid (no $(shell ...) leaks)
	assert.NotContains(t, code, "$(shell ", "generated code should not contain raw $(shell ...)")
	assert.Contains(t, code, "echo")
	assert.Contains(t, code, "OS:")
}

// parseAndTransform is a helper that parses a Makefile and transforms it to IR.
func parseAndTransform(t *testing.T, input string) *transformer.IR {
	t.Helper()

	p := parser.NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	tr := transformer.NewTransformer()
	return tr.Transform(m)
}
