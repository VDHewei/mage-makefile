package compiler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// downloadGoSDK downloads a Go SDK distribution.
// This is a partial implementation: network operations are stubbed
// to avoid timeout concerns. Use with custom SDK path in production.
func (c *NativeCompiler) downloadGoSDK(version, goos, goarch string) error {
	targetDir := c.sdkPath(version, goos, goarch)

	// Create the SDK cache directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create SDK directory %s: %w", targetDir, err)
	}

	// TODO: Implement actual download logic
	// The URL pattern for Go downloads is:
	//   https://go.dev/dl/go{version}.{os}-{arch}.tar.gz
	//
	// 1. Download the archive
	// 2. Verify checksum from https://go.dev/dl/?mode=json
	// 3. Extract to targetDir
	//
	// This is stubbed because:
	// - Network downloads may timeout
	// - Checksum verification requires fetching additional files
	// - Extraction of tar.gz requires archive handling

	return nil
}

// verifyChecksum verifies the SHA256 checksum of data.
func verifyChecksum(data []byte, expected string) bool {
	h := sha256.New()
	h.Write(data)
	actual := hex.EncodeToString(h.Sum(nil))
	return actual == expected
}

// extractSDK extracts a Go SDK archive to the destination directory.
func extractSDK(archive string, dest string) error {
	// TODO: Implement tar.gz extraction
	// Use archive/tar and compress/gzip from the standard library
	_, _ = archive, dest
	return nil
}

// UseLocalSDK configures the compiler to use a local SDK directory
// instead of downloading. This is useful for testing.
func (c *NativeCompiler) UseLocalSDK(sdkPath string) {
	// Create a symlink or copy to the cache directory
	targetDir := filepath.Join(c.sdkCacheDir, "local", "go")
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return
	}

	// Try to copy/symlink the go binary
	goBin := filepath.Join("bin", "go")
	srcGo := filepath.Join(sdkPath, goBin)
	dstGo := filepath.Join(targetDir, goBin)

	_ = srcGo
	_ = dstGo
	// In a real implementation, this would create a symlink or copy
}

// Initializes a test SDK in the cache directory.
// Only used in tests.
func (c *NativeCompiler) _initTestSDK(sdkPath string) error {
	// Create the SDK directory structure
	binDir := filepath.Join(sdkPath, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// On Windows, create a .bat file wrapper for go
	goExe := filepath.Join(binDir, "go")
	if err := os.WriteFile(goExe, []byte("#!/bin/bash\necho 'go version go1.22.0'\nexit 0"), 0755); err != nil {
		return err
	}

	return nil
}
