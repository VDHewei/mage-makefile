package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Bootstrap ensures a Go SDK is available for compilation.
func (c *NativeCompiler) Bootstrap() error {
	// First try to detect Go
	status, err := c.Detect()
	if err != nil {
		return fmt.Errorf("bootstrap detect failed: %w", err)
	}

	if status.Installed {
		// Go is available on the system
		return nil
	}

	// Try to use cached SDK
	if status.Cached {
		return nil
	}

	// Need to download Go SDK
	if err := c.downloadGoSDK("1.22.0", runtime.GOOS, runtime.GOARCH); err != nil {
		return fmt.Errorf("bootstrap download failed: %w", err)
	}

	return nil
}

// Detect checks whether Go is available (installed or cached).
func (c *NativeCompiler) Detect() (*GoSDKStatus, error) {
	status := &GoSDKStatus{}

	// Check PATH for Go
	goPath, err := detectGo()
	if err == nil {
		status.Installed = true
		status.Path = goPath
		status.Version = "detected"
		return status, nil
	}

	// Check cached SDK
	status.Installed = false
	cacheDir := c.sdkCacheDir
	if cacheDir == "" {
		cacheDir = defaultSDKCacheDir()
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		// Cache dir doesn't exist yet
		if os.IsNotExist(err) {
			return status, nil
		}
		return nil, fmt.Errorf("failed to read SDK cache dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			sdkPath := filepath.Join(cacheDir, entry.Name(), "go")
			if _, err := os.Stat(sdkPath); err == nil {
				status.Cached = true
				status.Path = filepath.Join(cacheDir, entry.Name())
				status.Version = entry.Name()
				return status, nil
			}
		}
	}

	return status, nil
}

// sdkPath returns the path where a Go SDK version would be cached.
func (c *NativeCompiler) sdkPath(version, goos, goarch string) string {
	cacheDir := c.sdkCacheDir
	if cacheDir == "" {
		cacheDir = defaultSDKCacheDir()
	}
	dirName := fmt.Sprintf("%s/%s-%s", version, goos, goarch)
	return filepath.Join(cacheDir, dirName)
}
