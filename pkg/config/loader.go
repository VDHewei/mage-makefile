package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Loader reads configuration using the priority hierarchy:
//
//	./.mage_makefile.toml -> $MAGE_MAKEFILE_CONFIG env -> $HOME/.mage_makefile.toml -> embedded base.toml
type Loader struct {
	cfg *Config
}

// NewLoader creates a new config loader applying the hierarchy.
func NewLoader() (*Loader, error) {
	cfg := DefaultConfig()

	paths := configPaths()
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := loadFile(p, cfg); err != nil {
			return nil, fmt.Errorf("load config %s: %w", p, err)
		}
	}

	return &Loader{cfg: cfg}, nil
}

// Config returns the loaded configuration.
func (l *Loader) Config() *Config {
	return l.cfg
}

// configPaths returns config file paths in priority order (lowest first).
func configPaths() []string {
	home, _ := os.UserHomeDir()

	paths := []string{
		// Lowest priority: embedded base config (handled by DefaultConfig)
		homeConfig(home), // $HOME/.mage_makefile.toml
		envConfig(),      // $MAGE_MAKEFILE_CONFIG
		localConfig(),    // ./.mage_makefile.toml
	}
	return paths
}

func homeConfig(home string) string {
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".mage_makefile.toml")
}

func envConfig() string {
	return os.Getenv("MAGE_MAKEFILE_CONFIG")
}

func localConfig() string {
	return ".mage_makefile.toml"
}

// loadFile reads a TOML config file and merges it into cfg.
// Missing files are silently skipped.
func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // skip missing files
		}
		return err
	}
	return toml.Unmarshal(data, cfg)
}

// SDK returns the cache directory for the Go SDK.
func (l *Loader) SDK() string {
	if l.cfg.Compiler.SDKCacheDir != "" {
		return l.cfg.Compiler.SDKCacheDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mage_makefile")
}

// GoSDKPath returns the expected path for a cached Go SDK.
func (l *Loader) GoSDKPath(version string) string {
	sdkDir := l.SDK()
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return filepath.Join(sdkDir, "go-sdk", version, fmt.Sprintf("%s-%s", goos, goarch), "go")
}
