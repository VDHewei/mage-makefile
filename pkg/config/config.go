// Package config provides configuration management for mage-makefile.
// Supports hierarchy loading: ./.mage_makefile.toml -> ENV -> $HOME/.mage_makefile.toml -> embedded base.toml
package config

import (
	"time"
)

// Config represents the full application configuration.
type Config struct {
	Hub       HubConfig       `toml:"hub"`
	Compiler  CompilerConfig  `toml:"compiler"`
	Script    ScriptConfig    `toml:"script"`
	Runtime   RuntimeConfig   `toml:"runtime"`
	Log       LogConfig       `toml:"log"`
	Convert    ConvertConfig   `toml:"convert"`
	// HubManager will be set by the CLI (not in config file)
	HubManager any `toml:"-"`
}

// HubConfig holds settings for the magego.hub.io API client.
type HubConfig struct {
	// Default hub server URL
	ServerURL string `toml:"server_url"`
	// HTTP request timeout
	Timeout time.Duration `toml:"timeout"`
	// Max retry attempts for failed requests
	MaxRetries int `toml:"max_retries"`
}

// CompilerConfig holds settings for the magefile.go compiler.
type CompilerConfig struct {
	// Default Go SDK version to use when downloading
	GoVersion string `toml:"go_version"`
	// Custom Go SDK cache directory
	SDKCacheDir string `toml:"sdk_cache_dir"`
	// Go module proxy URL
	GoProxy string `toml:"go_proxy"`
}

// ScriptConfig holds settings for the script engines (Lua/JS/Go).
type ScriptConfig struct {
	// Default script engine (go, lua, js)
	DefaultEngine string `toml:"default_engine"`
	// Lua VM max execution time
	LuaTimeout time.Duration `toml:"lua_timeout"`
	// JS VM max execution time
	JSTimeout time.Duration `toml:"js_timeout"`
}

// RuntimeConfig holds settings for platform detection.
type RuntimeConfig struct {
	// Comma-separated list of extra PATH directories to search
	ExtraPaths []string `toml:"extra_paths"`
	// Additional command mapping files
	CustomMappings []string `toml:"custom_mappings"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	// Log level: debug, info, warn, error
	Level string `toml:"level"`
	// Output format: text, json
	Format string `toml:"format"`
}

// ConvertConfig holds settings for Makefile-to-magefile.go conversion.
type ConvertConfig struct {
	// 目标过滤：空值表示所有目标
	IncludeTargets []string `toml:"include_targets"`
	// 排除目标列表
	ExcludeTargets []string `toml:"exclude_targets"`
	// 默认目标
	DefaultTarget string `toml:"default_target"`

	// 自定义命令映射覆盖
	CommandMap []CommandMapOverride `toml:"command_map"`

	// 输出样式：standard, verbose, minimal
	OutputStyle string `toml:"output_style"`
	// 添加 Makefile 源码注释
	AddComments bool `toml:"add_comments"`
	// 包含原始 shell 命令作为注释
	AddOriginal bool `toml:"add_original"`
	// 按类别分组函数
	GroupByCategory bool `toml:"group_by_category"`

	// 自定义别名
	Aliases []AliasOverride `toml:"aliases"`

	// 目标平台覆盖（空值=当前系统）
	DefaultPlatform string `toml:"default_platform"`
}

// CommandMapOverride 允许用户覆盖平台命令映射。
type CommandMapOverride struct {
	Unix    string `toml:"unix"`
	Windows string `toml:"windows"`
	MacOS   string `toml:"macos"`
}

// AliasOverride 允许用户自定义目标简写别名。
type AliasOverride struct {
	Alias  string `toml:"alias"`
	Target string `toml:"target"`
}

// DefaultConfig returns the default configuration (matching base.toml).
func DefaultConfig() *Config {
	return &Config{
		Hub: HubConfig{
			ServerURL:  "https://magego.hub.io",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		},
		Compiler: CompilerConfig{
			GoVersion: "1.24.3",
			GoProxy:   "https://proxy.golang.org,direct",
		},
		Script: ScriptConfig{
			DefaultEngine: "go",
			LuaTimeout:    10 * time.Second,
			JSTimeout:     10 * time.Second,
		},
		Runtime: RuntimeConfig{
			ExtraPaths: []string{},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Convert: ConvertConfig{
			ExcludeTargets:  []string{".PHONY"},
			OutputStyle:     "standard",
			AddComments:     true,
			AddOriginal:     false,
			GroupByCategory: false,
		},
	}
}
