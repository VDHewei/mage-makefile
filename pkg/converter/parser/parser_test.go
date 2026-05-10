package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_Comments(t *testing.T) {
	input := "# This is a comment\n#Another comment"
	tokens := TokenizeAll(input)

	assert.True(t, hasToken(tokens, TokenComment, "This is a comment"))
	assert.True(t, hasToken(tokens, TokenComment, "Another comment"))
}

func TestLexer_Target(t *testing.T) {
	input := "build:\n\techo building"
	tokens := TokenizeAll(input)

	targets := filterTokens(tokens, TokenTarget)
	require.Len(t, targets, 1)
	assert.Equal(t, "build:", targets[0].Value)
}

func TestLexer_TargetWithPrerequisites(t *testing.T) {
	input := "build: dep1 dep2\n\techo done"
	tokens := TokenizeAll(input)

	targets := filterTokens(tokens, TokenTarget)
	require.Len(t, targets, 1)
	assert.Equal(t, "build: dep1 dep2", targets[0].Value)
}

func TestLexer_Recipe(t *testing.T) {
	input := "build:\n\techo hello\n\tgo build"
	tokens := TokenizeAll(input)

	recipes := filterTokens(tokens, TokenRecipe)
	require.Len(t, recipes, 2)
	assert.Equal(t, "echo hello", recipes[0].Value)
	assert.Equal(t, "go build", recipes[1].Value)
}

func TestLexer_VariableAssign(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"CC = gcc", "CC = gcc"},
		{"CC := gcc", "CC := gcc"},
		{"CFLAGS += -Wall", "CFLAGS += -Wall"},
		{"FOO ?= bar", "FOO ?= bar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := TokenizeAll(tt.input)
			vars := filterTokens(tokens, TokenVariableAssign)
			require.Len(t, vars, 1, "expected 1 variable token")
			assert.Equal(t, tt.expected, vars[0].Value)
		})
	}
}

func TestLexer_VariableRef(t *testing.T) {
	input := "$(CC) ${FOO} $@ $<"
	tokens := TokenizeAll(input)

	refs := filterTokens(tokens, TokenVariableRef)
	require.Len(t, refs, 4)
	assert.Equal(t, "(CC)", refs[0].Value)
	assert.Equal(t, "{FOO}", refs[1].Value)
	assert.Equal(t, "@", refs[2].Value)
	assert.Equal(t, "<", refs[3].Value)
}

func TestLexer_Include(t *testing.T) {
	input := "include foo.mk\n-include bar.mk"
	tokens := TokenizeAll(input)

	includes := filterTokens(tokens, TokenInclude)
	require.Len(t, includes, 2)
}

func TestLexer_Conditional(t *testing.T) {
	input := "ifeq ($(OS),Windows_NT)\nendif"
	tokens := TokenizeAll(input)

	conds := filterTokens(tokens, TokenConditional)
	require.Len(t, conds, 2) // ifeq + endif
	assert.Contains(t, conds[0].Value, "ifeq")
	assert.Contains(t, conds[1].Value, "endif")
}

func TestLexer_DefineEndef(t *testing.T) {
	input := "define HELP\nUsage: make help\nendef"
	tokens := TokenizeAll(input)

	defs := filterTokens(tokens, TokenDefine)
	require.Len(t, defs, 1)
	assert.Equal(t, "HELP", defs[0].Value)

	endefs := filterTokens(tokens, TokenEndef)
	require.Len(t, endefs, 1)
}

func TestParser_BasicTarget(t *testing.T) {
	input := "build:\n\techo hello"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Targets, 1)

	target := m.Targets[0]
	assert.Equal(t, "build", target.Name)
	require.Len(t, target.Recipes, 1)
	assert.Equal(t, "echo hello", target.Recipes[0])
}

func TestParser_TargetWithPrerequisites(t *testing.T) {
	input := "app: main.go lib.go\n\tgo build -o app main.go"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Targets, 1)

	target := m.Targets[0]
	assert.Equal(t, "app", target.Name)
	require.Len(t, target.Prerequisites, 2)
	assert.Equal(t, "main.go", target.Prerequisites[0])
	assert.Equal(t, "lib.go", target.Prerequisites[1])
	assert.Len(t, target.Recipes, 1)
}

func TestParser_Variables(t *testing.T) {
	input := "CC = gcc\nCFLAGS := -O2\nLIBS += -lm"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Variables, 3)

	assert.Equal(t, "CC", m.Variables[0].Name)
	assert.Equal(t, "=", m.Variables[0].Operator)
	assert.Equal(t, "gcc", m.Variables[0].Value)

	assert.Equal(t, "CFLAGS", m.Variables[1].Name)
	assert.Equal(t, ":=", m.Variables[1].Operator)

	assert.Equal(t, "LIBS", m.Variables[2].Name)
	assert.Equal(t, "+=", m.Variables[2].Operator)
}

func TestParser_MultipleTargets(t *testing.T) {
	input := "build:\n\techo build\n\nclean:\n\trm -rf dist"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Targets, 2)

	assert.Equal(t, "build", m.Targets[0].Name)
	assert.Equal(t, "clean", m.Targets[1].Name)
}

func TestParser_Include(t *testing.T) {
	input := "include common.mk\n-include optional.mk"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Includes, 2)

	assert.False(t, m.Includes[0].IsOptional)
	assert.True(t, m.Includes[1].IsOptional)
}

func TestParser_PhonyTarget(t *testing.T) {
	input := ".PHONY: clean\n\nclean:\n\trm -rf dist"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	assert.True(t, m.Targets[0].IsPhony)
}

func TestParser_Conditional(t *testing.T) {
	input := "ifeq ($(OS),Windows_NT)\n\tCC = cl\nelse\n\tCC = gcc\nendif"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, m.Conditionals, 1)

	cond := m.Conditionals[0]
	assert.Equal(t, "ifeq", cond.Kind)
	assert.Contains(t, cond.Cond, "(OS),Windows_NT")
}

func TestParser_DefineVariable(t *testing.T) {
	input := "define HELP\nThis is help text\nMore help\nendef"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)

	// Find the multi-line variable
	var foundHelpVar bool
	for _, v := range m.Variables {
		if v.Name == "HELP" {
			foundHelpVar = true
			assert.Contains(t, v.Value, "help text")
			break
		}
	}
	assert.True(t, foundHelpVar, "HELP multi-line variable not found")
}

func TestParser_Comments(t *testing.T) {
	input := "# Top comment\n\nbuild:\n\t# Build comment\n\techo build"
	p := NewParser(input)
	m, err := p.Parse()
	require.NoError(t, err)
	assert.Contains(t, m.Comments, "Top comment")
}

func TestExpandVariable(t *testing.T) {
	vars := map[string]string{
		"CC":     "gcc",
		"CFLAGS": "-Wall -O2",
		"SRC":    "main.c",
	}

	result := ExpandVariable("$(CC) $(CFLAGS) -c $(SRC)", vars)
	assert.Equal(t, "gcc -Wall -O2 -c main.c", result)

	result = ExpandVariable("${CC}", vars)
	assert.Equal(t, "gcc", result)

	result = ExpandVariable("no vars here", vars)
	assert.Equal(t, "no vars here", result)

	// Unknown variable reference should expand to empty string
	result = ExpandVariable("$(UNKNOWN_VAR)", vars)
	assert.Equal(t, "", result)
}

func TestExpandVariableGnuMakeFuncs(t *testing.T) {
	vars := map[string]string{
		"CC": "gcc",
	}

	// GNU Make $(subst ...) evaluates: replace \ with / in CC="gcc" → no change
	result := ExpandVariable(`OUT:=$(subst \,/,$(CC))`, vars)
	assert.Equal(t, `OUT:=gcc`, result, "subst \\ with / in gcc = gcc")

	// GNU Make $(dir ...) evaluates: dir of "gcc" = "./"
	result = ExpandVariable(`DIR:=$(dir $(CC))`, vars)
	assert.Equal(t, `DIR:=./`, result, "dir of gcc = ./")

	// GNU Make $(patsubst ...) is not supported yet, so preserved as-is
	result = ExpandVariable(`$(patsubst %.c,%.o,$(CC))`, vars)
	assert.Equal(t, `$(patsubst %.c,%.o,$(CC))`, result, "patsubst function should be preserved")

	// Mixed: $(CC) expands, $(shell ...) is not supported yet so preserved
	result = ExpandVariable(`echo $(CC) $(shell uname)`, vars)
	assert.Equal(t, `echo gcc $(shell uname)`, result, "CC should expand, shell should be preserved")

	// Nested subst + dir with CC="gcc":
	// dir(gcc) = ./, subst cmd\→bin\bash.exe in ./ = ./, subst \→/ in ./ =./
	result = ExpandVariable(`$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(CC))))`, vars)
	assert.Equal(t, `./`, result, "nested GNU Make functions fully evaluated")

	// Real-world: subst \ with / in "src/foo.c" = "src/foo.c"
	vars["SRC"] = "src/foo.c"
	result = ExpandVariable(`$(subst \,/,$(SRC))`, vars)
	assert.Equal(t, `src/foo.c`, result, "subst \\ with / in unix path = unchanged")

	// subst a with b in "abc" = "bbc"
	result = ExpandVariable(`$(subst a,b,abc)`, vars)
	assert.Equal(t, `bbc`, result, "subst a with b in abc = bbc")

	// dir of "src/foo.c" = "src/"
	result = ExpandVariable(`$(dir src/foo.c)`, vars)
	assert.Equal(t, `src/`, result, "dir of src/foo.c = src/")

	// dir of multiple paths
	result = ExpandVariable(`$(dir src/foo.c src/bar.c)`, vars)
	assert.Equal(t, `src/ src/`, result, "dir of multiple paths")
}

func TestExpandVariableDepthAware(t *testing.T) {
	vars := map[string]string{
		"VAR": "value",
	}

	// Nested parens within a variable reference (not a GNU Make function)
	// This should NOT happen in practice, but depth-aware matching handles it
	result := ExpandVariable(`$(call fn,$(VAR))`, vars)
	assert.Equal(t, `$(call fn,$(VAR))`, result, "call function should be preserved")

	// Multiple nested function calls
	result = ExpandVariable(`$(sort $(VAR))`, vars)
	assert.Equal(t, `$(sort $(VAR))`, result, "sort function should be preserved")

	// Self-reference guard should still work
	vars["PATH"] = "/usr/bin:$(PATH)"
	result = ExpandVariable("$(PATH)", vars)
	assert.Equal(t, "/usr/bin:$(PATH)", result, "self-reference should not recurse infinitely")

	// Simple variable still works
	result = ExpandVariable("prefix_$(VAR)_suffix", vars)
	assert.Equal(t, "prefix_value_suffix", result)
}

func TestExtractFuncName(t *testing.T) {
	assert.Equal(t, "subst", extractFuncName("subst a,b,c"))
	assert.Equal(t, "shell", extractFuncName("shell echo hi"))
	assert.Equal(t, "dir", extractFuncName("dir $(CC)"))
	assert.Equal(t, "CC", extractFuncName("CC"))
	assert.Equal(t, "", extractFuncName(""))
}

func TestIsGnuMakeFunc(t *testing.T) {
	assert.True(t, isGnuMakeFunc("subst"))
	assert.True(t, isGnuMakeFunc("shell"))
	assert.True(t, isGnuMakeFunc("dir"))
	assert.True(t, isGnuMakeFunc("patsubst"))
	assert.True(t, isGnuMakeFunc("wildcard"))
	assert.False(t, isGnuMakeFunc("CC"))
	assert.False(t, isGnuMakeFunc("CFLAGS"))
	assert.False(t, isGnuMakeFunc(""))
	assert.False(t, isGnuMakeFunc("SUBST")) // uppercase - not a function name
}

func TestIsWindowsDrivePrefix(t *testing.T) {
	assert.True(t, isWindowsDrivePrefix(`C:\`))
	assert.True(t, isWindowsDrivePrefix(`C:\Program`))
	assert.True(t, isWindowsDrivePrefix(`D:/path`))
	assert.True(t, isWindowsDrivePrefix(`Z:\`))
	assert.False(t, isWindowsDrivePrefix(`C:`))           // too short
	assert.False(t, isWindowsDrivePrefix(`CC:\`))         // no colon after letter
	assert.False(t, isWindowsDrivePrefix(`src/`))         // unix path
	assert.False(t, isWindowsDrivePrefix(``))             // empty
	assert.False(t, isWindowsDrivePrefix(`/usr/bin/`))    // no drive letter
}

func TestSplitDirArgs(t *testing.T) {
	// Unix-style: space-separated paths
	result := splitDirArgs(`src/foo.c src/bar.c`)
	assert.Equal(t, []string{`src/foo.c`, `src/bar.c`}, result)

	// Windows paths with spaces on single line
	result = splitDirArgs(`C:\Program Files\Git\cmd\git.exe`)
	assert.Equal(t, []string{`C:\Program Files\Git\cmd\git.exe`}, result, "Windows path with spaces preserved")

	// Multi-line input (from multi-line shell output)
	multiLine := "C:\\Program Files\\Git\\mingw64\\bin\\git.exe\nC:\\Program Files\\Git\\cmd\\git.exe"
	result = splitDirArgs(multiLine)
	assert.Equal(t, []string{
		`C:\Program Files\Git\mingw64\bin\git.exe`,
		`C:\Program Files\Git\cmd\git.exe`,
	}, result, "multi-line Windows paths preserved")

	// Empty
	result = splitDirArgs(``)
	assert.Nil(t, result)

	// Single path on each line (multi-line from shell)
	result = splitDirArgs("/usr/bin/git\n/bin/sh")
	assert.Equal(t, []string{"/usr/bin/git", "/bin/sh"}, result)
}

func TestEvalDirWindowsPaths(t *testing.T) {
	vars := map[string]string{}
	// Single Windows path with spaces
	result := ExpandVariable(`$(dir C:\Program Files\Git\cmd\git.exe)`, vars)
	assert.Equal(t, `C:\Program Files\Git\cmd\`, result, "dir of Windows path with spaces")

	// Multi-line Windows paths (simulating $(shell where git) output)
	// Note: actual CRLF from shell is normalized to LF by evalDir
	result = ExpandVariable("$(dir C:\\Program Files\\Git\\mingw64\\bin\\git.exe\nC:\\Program Files\\Git\\cmd\\git.exe)", vars)
	assert.Equal(t, "C:\\Program Files\\Git\\mingw64\\bin\\ C:\\Program Files\\Git\\cmd\\",
		result, "dir of multi-line Windows paths")

	// Kratos-style: $(subst \,/,$(subst cmd\,bin\bash.exe,$(dir ...)))
	// with Windows paths. After outer subst: / replaces \ in dir result.
	// C:\Program Files\Git\cmd\ → cmd\ replaced with bin\bash.exe → C:\Program Files\Git\bin\bash.exe
	// Then \ → /: C:/Program Files/Git/bin/bash.exe
	result = ExpandVariable("$(subst \\,/,$(subst cmd\\,bin\\bash.exe,$(dir C:\\Program Files\\Git\\cmd\\git.exe)))", vars)
	assert.Equal(t, "C:/Program Files/Git/bin/bash.exe",
		result, "kratos Git_Bash pattern with Windows path")
}

// Helper functions

func hasToken(tokens []Token, tp TokenType, value string) bool {
	for _, t := range tokens {
		if t.Type == tp && strings.Contains(t.Value, value) {
			return true
		}
	}
	return false
}

func filterTokens(tokens []Token, tp TokenType) []Token {
	var result []Token
	for _, t := range tokens {
		if t.Type == tp {
			result = append(result, t)
		}
	}
	return result
}
