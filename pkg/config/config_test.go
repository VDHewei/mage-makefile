package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "https://magego.hub.io", cfg.Hub.ServerURL)
	assert.Equal(t, "1.24.3", cfg.Compiler.GoVersion)
	assert.Equal(t, "go", cfg.Script.DefaultEngine)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
}

func TestNewLoader_Defaults(t *testing.T) {
	l, err := NewLoader()
	require.NoError(t, err)
	require.NotNil(t, l)

	cfg := l.Config()
	assert.Equal(t, "https://magego.hub.io", cfg.Hub.ServerURL)
}

func TestLoadFile_Success(t *testing.T) {
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "test.toml")
	content := `
[hub]
server_url = "https://custom.hub.io"
timeout = "60s"

[log]
level = "debug"
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	cfg := DefaultConfig()
	err = loadFile(tmpFile, cfg)
	require.NoError(t, err)

	assert.Equal(t, "https://custom.hub.io", cfg.Hub.ServerURL)
	assert.Equal(t, "debug", cfg.Log.Level)
	// Other fields should retain defaults
	assert.Equal(t, "go", cfg.Script.DefaultEngine)
}

func TestLoadFile_NotExist(t *testing.T) {
	cfg := DefaultConfig()
	err := loadFile("/nonexistent/path/config.toml", cfg)
	assert.NoError(t, err)
}

func TestGoSDKPath(t *testing.T) {
	l := &Loader{cfg: DefaultConfig()}
	path := l.GoSDKPath("1.24.3")
	assert.Contains(t, path, "go-sdk")
	assert.Contains(t, path, "1.24.3")
	assert.Contains(t, path, "go")
}

func TestSDK_CustomDir(t *testing.T) {
	l := &Loader{cfg: DefaultConfig()}
	l.cfg.Compiler.SDKCacheDir = "/custom/cache"
	assert.Equal(t, "/custom/cache", l.SDK())
}

func TestSDK_DefaultDir(t *testing.T) {
	l := &Loader{cfg: DefaultConfig()}
	sdk := l.SDK()
	assert.Contains(t, sdk, ".mage_makefile")
}

func TestConfigPaths_Order(t *testing.T) {
	paths := configPaths()
	// Should contain at least 3 entries in correct order
	assert.GreaterOrEqual(t, len(paths), 3)
	// Local config should be last (highest priority)
	assert.Equal(t, ".mage_makefile.toml", paths[len(paths)-1])
}
