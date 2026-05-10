package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// NativeCompiler implements the Compiler interface using local Go toolchain.
type NativeCompiler struct {
	sdkCacheDir string
	goBin       string
}

// NewNativeCompiler creates a new NativeCompiler.
func NewNativeCompiler() *NativeCompiler {
	return &NativeCompiler{
		sdkCacheDir: defaultSDKCacheDir(),
	}
}

// SetGoBin overrides the Go binary path (useful for testing).
func (c *NativeCompiler) SetGoBin(path string) {
	c.goBin = path
}

// SetSDKCacheDir overrides the SDK cache directory (useful for testing).
func (c *NativeCompiler) SetSDKCacheDir(dir string) {
	c.sdkCacheDir = dir
}

// getGo returns the Go binary path, with override support.
func (c *NativeCompiler) getGo() (string, error) {
	if c.goBin != "" {
		if _, err := os.Stat(c.goBin); err != nil {
			return "", fmt.Errorf("go binary not found at %s: %w", c.goBin, err)
		}
		return c.goBin, nil
	}
	return detectGo()
}

// Native compiles a magefile.go to a native binary for the current platform.
func (c *NativeCompiler) Native(magefile string, output string) error {
	return c.compile(magefile, output, runtime.GOOS, runtime.GOARCH)
}

// Cross compiles a magefile.go to a binary for the specified target.
func (c *NativeCompiler) Cross(magefile string, targetOS, targetArch, output string) error {
	return c.compile(magefile, output, targetOS, targetArch)
}

// compile performs the actual compilation in a temporary workspace.
func (c *NativeCompiler) compile(magefilePath string, output string, goos, goarch string) error {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "mage-compile-*")
	if err != nil {
		return fmt.Errorf("failed to create temp workspace: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy magefile to workspace
	mageContent, err := os.ReadFile(magefilePath)
	if err != nil {
		return fmt.Errorf("failed to read magefile: %w", err)
	}

	destFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(destFile, mageContent, 0644); err != nil {
		return fmt.Errorf("failed to write magefile to workspace: %w", err)
	}

	// Create go.mod
	modContent := `module mage_build

go 1.21

require github.com/magefile/mage v1.15.0

require (
	github.com/magefile/mage v1.15.0
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	// Run go mod tidy to resolve dependencies
	goBin, err := c.getGo()
	if err != nil {
		return fmt.Errorf("failed to find Go: %w", err)
	}

	tidyCmd := exec.Command(goBin, "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Env = os.Environ()
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s: %w", string(output), err)
	}

	// Run go build
	buildCmd := exec.Command(goBin, "build", "-o", output, ".")
	buildCmd.Dir = tmpDir
	buildCmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)

	if compileOutput, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %s: %w", string(compileOutput), err)
	}

	return nil
}

// detectGo finds the Go binary on the system PATH or in the cached SDK directory.
func detectGo() (string, error) {
	// First check PATH
	if path, err := exec.LookPath("go"); err == nil {
		return path, nil
	}

	// Then check cached SDK directory
	cacheDir := defaultSDKCacheDir()
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return "", fmt.Errorf("Go not found on PATH and no cached SDK: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			sdkPath := filepath.Join(cacheDir, entry.Name(), "bin", "go")
			if runtime.GOOS == "windows" {
				sdkPath += ".exe"
			}
			if _, err := os.Stat(sdkPath); err == nil {
				return sdkPath, nil
			}
		}
	}

	return "", fmt.Errorf("Go not found")
}

// defaultSDKCacheDir returns the default SDK cache directory.
func defaultSDKCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".mage_makefile", "go-sdk")
}
