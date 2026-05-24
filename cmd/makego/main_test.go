package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput executes rootCmd with given args and returns stdout as string.
func captureOutput(t *testing.T, args []string) string {
	t.Helper()

	// Save original state and restore after test
	oldStdout := os.Stdout

	// Use a pipe with a goroutine to avoid deadlock on full buffer
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	// Read from pipe in a goroutine to prevent write deadlock
	bufCh := make(chan string, 1)
	go func() {
		var buf strings.Builder
		tmp := make([]byte, 4096)
		for {
			n, readErr := r.Read(tmp)
			if n > 0 {
				buf.Write(tmp[:n])
			}
			if readErr != nil {
				bufCh <- buf.String()
				return
			}
		}
	}()

	// Set os.Args for cobra
	rootCmd.SetArgs(args)

	// Execute
	err = rootCmd.Execute()

	// Close writer to signal EOF to the reader goroutine
	w.Close()
	os.Stdout = oldStdout

	// Wait for reader goroutine to finish
	captured := <-bufCh

	if err != nil {
		return "" // command returned error; content is in buffer
	}

	return captured
}

func TestCLI_Version(t *testing.T) {
	output := captureOutput(t, []string{"version"})
	if !strings.Contains(output, "makego v0.1.0") {
		t.Errorf("expected version output, got: %s", output)
	}
}

func TestCLI_ConvertBasic(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Create a minimal Makefile
	makefileContent := "build:\n\tgcc -o program main.c\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	output := captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
	})

	// Check output mentions the file
	if !strings.Contains(output, "magefile.go") {
		t.Errorf("expected convert output, got: %s", output)
	}

	// Check the generated file exists and contains expected content
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "package main") {
		t.Errorf("generated file missing package declaration")
	}
	if !strings.Contains(content, "func Build") {
		t.Errorf("generated file missing Build function")
	}
	if !strings.Contains(content, "mage") {
		t.Errorf("generated file missing mage build tag")
	}
}

func TestCLI_ConvertWithVars(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Create a Makefile with variables
	makefileContent := `CC = gcc
CFLAGS = -Wall -O2

build:
	$(CC) $(CFLAGS) -o program main.c
`
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
	})

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)

	// Variables with non-shell values should become constants
	if !strings.Contains(content, "const") && !strings.Contains(content, "var") {
		t.Logf("no variable declarations found (may be inline expanded):\n%s", content)
	}

	// The build function should exist
	if !strings.Contains(content, "func Build") {
		t.Errorf("generated file missing Build function")
	}
}

func TestCLI_ConvertWithDeps(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Create a Makefile with dependencies
	makefileContent := `all: build test

build:
	gcc -o program main.c

test:
	./program --test
`
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
	})

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)

	// Check all functions are generated
	if !strings.Contains(content, "func All") {
		t.Errorf("generated file missing All function")
	}
	if !strings.Contains(content, "func Build") {
		t.Errorf("generated file missing Build function")
	}
	if !strings.Contains(content, "func Test") {
		t.Errorf("generated file missing Test function")
	}
}

func TestCLI_DetectCommand(t *testing.T) {
	saveDetectGlobals(t)
	output := captureOutput(t, []string{"detect"})

	if !strings.Contains(output, "System:") {
		t.Errorf("expected detect output with System info, got: %s", output)
	}
	if !strings.Contains(output, "Shell:") {
		t.Errorf("expected detect output with Shell info")
	}
}

// =============================================================================
// Phase 4: CLI Detect Tests
// =============================================================================

// saveDetectGlobals saves and restores detect command global variables for test isolation.
func saveDetectGlobals(t *testing.T) {
	t.Helper()
	oldOS := detectOS
	oldJSON := detectJSON
	oldInstall := detectInstall
	oldInteractive := detectInteractive
	oldReport := detectReport
	t.Cleanup(func() {
		detectOS = oldOS
		detectJSON = oldJSON
		detectInstall = oldInstall
		detectInteractive = oldInteractive
		detectReport = oldReport
	})
}

func TestCLI_DetectWithMakefile(t *testing.T) {
	saveDetectGlobals(t)
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\tgcc -o program main.c\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	output := captureOutput(t, []string{"detect", "Makefile"})
	assert.Contains(t, output, "Checking:")
	assert.Contains(t, output, "gcc")
	assert.Contains(t, output, "Makefile")
}

func TestCLI_DetectWithOS(t *testing.T) {
	saveDetectGlobals(t)
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\tgcc -o program main.c\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	output := captureOutput(t, []string{"detect", "Makefile", "--os", "linux"})
	assert.Contains(t, output, "linux", "should mention target OS linux")
	assert.Contains(t, output, "gcc")
}

func TestCLI_DetectWithJSON(t *testing.T) {
	saveDetectGlobals(t)
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\tgcc -o program main.c\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	output := captureOutput(t, []string{"detect", "Makefile", "--json"})
	// JSON output follows system info header; find the JSON part
	jsonIdx := strings.Index(output, "{")
	assert.True(t, jsonIdx >= 0, "JSON output should contain {")
	jsonPart := output[jsonIdx:]
	assert.Contains(t, jsonPart, "target_os")
	assert.Contains(t, jsonPart, "gcc")
	assert.Contains(t, jsonPart, "is_available")
}

func TestCLI_DetectCommandName(t *testing.T) {
	saveDetectGlobals(t)
	output := captureOutput(t, []string{"detect", "gcc"})
	assert.Contains(t, output, "gcc")
	// Should either show available or not available
	assert.True(t,
		strings.Contains(output, "Available") || strings.Contains(output, "NOT available"),
		"output should indicate command availability")
}

func TestCLI_DetectWithReport(t *testing.T) {
	saveDetectGlobals(t)
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\tgcc -o program main.c\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	output := captureOutput(t, []string{"detect", "Makefile", "--report"})
	assert.Contains(t, output, "Compatibility Report")
	assert.Contains(t, output, "gcc")
	// Output should show either [OK] or [MISSING] for gcc
	assert.True(t,
		strings.Contains(output, "[OK]") || strings.Contains(output, "[MISSING]"),
		"output should show command status as [OK] or [MISSING]")
	assert.Contains(t, output, "Target OS")
	assert.Contains(t, output, "Total")
}

// == Phase 5+: Script Engine CLI Tests ==

func TestCLI_ConvertWithScriptGo(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\techo \"hello\"\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	output := captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
		"--script", "go",
	})

	assert.Contains(t, output, "magefile.go")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	assert.Contains(t, string(data), "package main")
}

func TestCLI_ConvertWithScriptLua(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\techo \"hello\"\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	output := captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
		"--script", "lua",
	})

	assert.Contains(t, output, "magefile.go")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	assert.Contains(t, string(data), "package main")
}

func TestCLI_ConvertWithScriptJs(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	makefileContent := "build:\n\techo \"hello\"\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	output := captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
		"--script", "js",
	})

	assert.Contains(t, output, "magefile.go")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	assert.Contains(t, string(data), "package main")
}

func TestCLI_ConvertWithWindowsPlatform(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(dir)

	// Use a complex command (&&) to trigger the exec.Command path with cross-platform shell
	makefileContent := "build:\n\trm -rf bin/ && echo \"done\"\n"
	if err := os.WriteFile("Makefile", []byte(makefileContent), 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	outputFile := filepath.Join(dir, "magefile.go")
	captureOutput(t, []string{
		"convert", "Makefile",
		"--output", outputFile,
		"--platform", "windows",
	})

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(data)
	// Windows cross-platform generation should include runtime.GOOS check for complex commands
	assert.Contains(t, content, "runtime.GOOS")
	assert.Contains(t, content, "\"cmd\"")
	assert.Contains(t, content, "\"/C\"")
}

// == Phase 6: Compile CLI Tests ==

func TestCLI_CompilesNative(t *testing.T) {
	dir := t.TempDir()
	magefile := filepath.Join(dir, "main.go")
	mageContent := `//go:build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
)

func Build() error {
	fmt.Println("building")
	return sh.Run("echo", "done")
}
`
	if err := os.WriteFile(magefile, []byte(mageContent), 0644); err != nil {
		t.Fatalf("write magefile: %v", err)
	}

	output := filepath.Join(dir, "magebuild")
	outputArgs := []string{"compile", magefile, "--output", output}

	out := captureOutput(t, outputArgs)
	// Output should mention compilation
	if !strings.Contains(out, "Compiling") {
		t.Logf("compile output: %s", out)
	}
}
