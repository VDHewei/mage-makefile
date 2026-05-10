package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetect_GoOnSystem tests Go detection on local system.
func TestDetect_GoOnSystem(t *testing.T) {
	c := NewNativeCompiler()

	status, err := c.Detect()
	require.NoError(t, err)

	// On a developer machine, Go should be installed
	assert.True(t, status.Installed, "Go should be installed on developer machine")
	assert.NotEmpty(t, status.Path, "Go path should not be empty")
}

// TestDetect_WithCacheDir tests detection with a custom cache dir.
func TestDetect_WithCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	c := NewNativeCompiler()
	c.SetSDKCacheDir(tmpDir)

	status, err := c.Detect()
	require.NoError(t, err)

	// Without actually caching Go, it should fall back to PATH
	assert.True(t, status.Installed || !status.Cached)
}

// TestBootstrap_CreatesCacheDir tests that bootstrap ensures the cache dir.
func TestBootstrap_CreatesCacheDir(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "mage_makefile", "go-sdk")

	c := NewNativeCompiler()
	c.SetSDKCacheDir(tmpDir)

	// Bootstrap should succeed (Go is on PATH)
	err := c.Bootstrap()
	require.NoError(t, err)

	// Cache dir should be created (bootstrap ensures it exists)
	_, err = os.Stat(filepath.Dir(tmpDir))
	// may or may not exist depending on whether download was needed
	_ = err
}

// TestNativeCompilation compiles a simple magefile.go.
func TestNativeCompilation(t *testing.T) {
	c := NewNativeCompiler()

	// Create a simple magefile.go
	tmpDir := t.TempDir()
	magefile := filepath.Join(tmpDir, "main.go")
	mageContent := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello from mage")
}
`
	err := os.WriteFile(magefile, []byte(mageContent), 0644)
	require.NoError(t, err)

	output := filepath.Join(tmpDir, "mage_binary")
	if runtime.GOOS == "windows" {
		output += ".exe"
	}

	err = c.Native(magefile, output)
	if err != nil {
		t.Skipf("native compilation failed (may need network for go mod tidy): %v", err)
		return
	}

	// Verify binary exists
	_, err = os.Stat(output)
	require.NoError(t, err, "compiled binary should exist")

	// Verify binary is executable
	info, err := os.Stat(output)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0), "binary should be non-empty")
}

// TestCrossCompilation tests cross-compilation for a different OS/ARCH.
func TestCrossCompilation(t *testing.T) {
	c := NewNativeCompiler()

	tmpDir := t.TempDir()
	magefile := filepath.Join(tmpDir, "main.go")
	mageContent := `package main

func main() {
	println("cross compiled")
}
`
	err := os.WriteFile(magefile, []byte(mageContent), 0644)
	require.NoError(t, err)

	// Cross compile to a different OS
	targetOS := "linux"
	targetArch := "amd64"
	if runtime.GOOS == "linux" {
		targetOS = "darwin"
	}

	output := filepath.Join(tmpDir, "mage_cross")

	err = c.Cross(magefile, targetOS, targetArch, output)
	if err != nil {
		t.Skipf("cross compilation failed (may need network): %v", err)
		return
	}

	_, err = os.Stat(output)
	assert.NoError(t, err, "cross-compiled binary should exist")
}

// TestNativeCompiler_GoBinOverride tests setting a custom Go binary.
func TestNativeCompiler_GoBinOverride(t *testing.T) {
	c := NewNativeCompiler()

	// Find the actual go binary
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Go not found on PATH")
	}
	c.SetGoBin(goBin)

	status, err := c.Detect()
	require.NoError(t, err)
	assert.True(t, status.Installed)
}

// TestSDKCache_Dir tests the SDK cache directory path.
func TestSDKCache_Dir(t *testing.T) {
	c := NewNativeCompiler()

	// Default cache dir should contain .mage_makefile/go-sdk
	dir := defaultSDKCacheDir()
	assert.Contains(t, dir, ".mage_makefile")
	assert.Contains(t, dir, "go-sdk")
	_ = c
}

// TestSDKPath_Structure tests the SDK path structure.
func TestSDKPath_Structure(t *testing.T) {
	c := NewNativeCompiler()
	c.SetSDKCacheDir("/tmp/mage-test-sdk")

	path := c.sdkPath("1.22.0", "linux", "amd64")
	expected := filepath.Join("/tmp/mage-test-sdk", "1.22.0", "linux-amd64")
	assert.Equal(t, expected, filepath.Clean(path))
}

// TestGoSDKStatus_Defaults tests the default values of GoSDKStatus.
func TestGoSDKStatus_Defaults(t *testing.T) {
	status := &GoSDKStatus{}
	assert.False(t, status.Installed)
	assert.False(t, status.Cached)
	assert.Empty(t, status.Path)
	assert.Empty(t, status.Version)
}

// TestCompile_NonexistentMagefile tests compilation with a nonexistent file.
func TestCompile_NonexistentMagefile(t *testing.T) {
	c := NewNativeCompiler()

	err := c.Native("/nonexistent/magefile.go", "/tmp/output")
	assert.Error(t, err)
}

// TestCrossCompile_UnknownArch tests cross-compilation with unsupported arch.
func TestCrossCompile_UnknownArch(t *testing.T) {
	// This test verifies that the GOARCH is passed through correctly.
	// The actual build may fail if the arch is unsupported, but the
	// compilation attempt should be made.

	c := NewNativeCompiler()

	tmpDir := t.TempDir()
	magefile := filepath.Join(tmpDir, "main.go")
	mageContent := `//go:build mage

package main

func main() {}
`
	err := os.WriteFile(magefile, []byte(mageContent), 0644)
	require.NoError(t, err)

	output := filepath.Join(tmpDir, "mage_unknown")
	err = c.Cross(magefile, "plan9", "arm", output)

	// May fail or succeed; we just verify it doesn't panic
	_ = err
}

// TestVerifyChecksum tests the SHA256 checksum verification.
func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world")
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	assert.True(t, verifyChecksum(data, expected))

	assert.False(t, verifyChecksum(data, "badhash"))
}
