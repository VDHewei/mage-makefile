package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
)

// CompatResult represents the compatibility check result for a single command.
// CompatResult 表示单个命令的兼容性检查结果。
type CompatResult struct {
	Command        string           `json:"command"`
	Description    string           `json:"description,omitempty"`
	IsAvailable    bool             `json:"is_available"`
	IsShellBuiltin bool             `json:"is_shell_builtin,omitempty"`
	Alternative    string           `json:"alternative,omitempty"`
	Platforms      []string         `json:"platforms,omitempty"`
	Notes          string           `json:"notes,omitempty"`
	InstallURL     string           `json:"install_url,omitempty"`
	InstallMethods []InstallMethod  `json:"install_methods,omitempty"`
}

// CompatReport is a structured compatibility report for a Makefile.
// CompatReport 是 Makefile 的结构化兼容性报告。
type CompatReport struct {
	TargetOS  string          `json:"target_os"`
	CurrentOS string          `json:"current_os"`
	Total     int             `json:"total"`
	Available int             `json:"available"`
	Missing   int             `json:"missing"`
	Results   []*CompatResult `json:"results"`
}

// String returns a human-readable text format of the report.
// String 返回人类可读的文本格式报告。
func (r *CompatReport) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Makefile Compatibility Report / Makefile 兼容性报告\n"))
	b.WriteString(fmt.Sprintf("  Target OS / 目标系统:  %s\n", r.TargetOS))
	b.WriteString(fmt.Sprintf("  Current OS / 当前系统: %s\n", r.CurrentOS))
	b.WriteString(fmt.Sprintf("  Total / 总数:  %d  Available / 可用: %d  Missing / 缺失: %d\n\n", r.Total, r.Available, r.Missing))

	for _, result := range r.Results {
		status := "[OK]"
		if !result.IsAvailable {
			status = "[MISSING]"
		}
		b.WriteString(fmt.Sprintf("  %-20s %s", result.Command, status))
		if result.Description != "" {
			b.WriteString(fmt.Sprintf("  — %s", result.Description))
		}
		b.WriteString("\n")
		if !result.IsAvailable {
			if result.Alternative != "" {
				b.WriteString(fmt.Sprintf("    → try / 尝试: %s\n", result.Alternative))
			}
			if result.IsShellBuiltin {
				b.WriteString("    → shell built-in / Shell 内置命令\n")
			}
			if result.InstallURL != "" {
				b.WriteString(fmt.Sprintf("    → download / 下载: %s\n", result.InstallURL))
			}
		}
	}

	return b.String()
}

// MarshalJSON implements json.Marshaler for CompatReport.
// MarshalJSON 实现了 CompatReport 的 json.Marshaler 接口。
func (r *CompatReport) MarshalJSON() ([]byte, error) {
	type Alias CompatReport
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements json.Unmarshaler for CompatReport.
// UnmarshalJSON 实现了 CompatReport 的 json.Unmarshaler 接口。
func (r *CompatReport) UnmarshalJSON(data []byte) error {
	type Alias CompatReport
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, aux)
}

// CompatChecker checks command compatibility across platforms.
// CompatChecker 检查跨平台命令兼容性。
type CompatChecker struct {
	detector *Detector
	sysInfo  *SystemInfo
	targetOS string // empty = current OS / 空值表示当前 OS
}

// NewCompatChecker creates a new compatibility checker for the current OS.
// NewCompatChecker 为当前操作系统创建新的兼容性检查器。
func NewCompatChecker() *CompatChecker {
	return &CompatChecker{
		detector: NewDetector(),
	}
}

// NewCompatCheckerForOS creates a new compatibility checker for a specific target OS.
// targetOS can be "linux", "darwin", or "windows".
// The checker will use this target OS for alternative suggestions.
// NewCompatCheckerForOS 为特定目标操作系统创建新的兼容性检查器。
// targetOS 可以是 "linux"、"darwin" 或 "windows"。
func NewCompatCheckerForOS(targetOS string) *CompatChecker {
	return &CompatChecker{
		detector: NewDetector(),
		targetOS: targetOS,
	}
}

// Init initializes the checker by detecting the current system.
// Init 通过检测当前系统来初始化检查器。
func (c *CompatChecker) Init() error {
	info, err := c.detector.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect system: %w", err)
	}
	c.sysInfo = info
	return nil
}

// CurrentOS returns the detected current OS, or "unknown" if not initialized.
// CurrentOS 返回检测到的当前操作系统，如果未初始化则返回 "unknown"。
func (c *CompatChecker) CurrentOS() string {
	if c.sysInfo != nil {
		return c.sysInfo.OS
	}
	if c.detector != nil {
		info, err := c.detector.Detect()
		if err == nil {
			return info.OS
		}
	}
	return "unknown"
}

// CheckCommand checks if a single command is available and compatible on the current platform.
// CheckCommand 检查单个命令在当前平台上是否可用且兼容。
func (c *CompatChecker) CheckCommand(cmd string) (*CompatResult, error) {
	if c.sysInfo == nil {
		if err := c.Init(); err != nil {
			return nil, err
		}
	}

	checkOS := c.targetOS
	if checkOS == "" {
		checkOS = c.sysInfo.OS
	}

	cmdMap := LookupCommandMap(cmd)

	result := &CompatResult{
		Command:  cmd,
		Platforms: GetPlatforms(cmd),
	}

	// Populate description from CommandMap
	if cmdMap != nil {
		result.Description = cmdMap.Description
		result.InstallURL = cmdMap.InstallURL
		result.InstallMethods = cmdMap.InstallCmds
	}

	// Check if the command is available on the CURRENT system
	available, _ := c.detector.IsCommandAvailable(cmd)
	result.IsAvailable = available

	// Check if it's a shell built-in
	if IsShellBuiltin(cmd) {
		result.IsShellBuiltin = true
	}

	// If not available, try to find an alternative
	if !available {
		alt, found := GetAlternative(cmd, checkOS)
		if found {
			result.Alternative = alt
		}
	}

	// Build notes
	var notesParts []string
	if result.IsShellBuiltin {
		notesParts = append(notesParts, "shell built-in / Shell 内置命令")
	}
	if !result.IsAvailable && result.Alternative != "" {
		notesParts = append(notesParts, fmt.Sprintf("use '%s' on %s / 在 %s 上使用 '%s'", result.Alternative, checkOS, checkOS, result.Alternative))
	} else if !result.IsAvailable {
		notesParts = append(notesParts, fmt.Sprintf("not available on %s / 在 %s 上不可用", checkOS, checkOS))
	}
	if len(result.Platforms) > 0 {
		notesParts = append(notesParts, fmt.Sprintf("platforms / 平台: %s", strings.Join(result.Platforms, ", ")))
	}
	result.Notes = strings.Join(notesParts, "; ")
	return result, nil
}

// FindAlternative finds a platform-specific alternative for a command.
// FindAlternative 查找命令的平台特定替代方案。
func (c *CompatChecker) FindAlternative(cmd string, targetOS string) (string, bool) {
	return GetAlternative(cmd, targetOS)
}

// CheckMakefileCompatibility checks all recipes in a parsed Makefile for compatibility.
// CheckMakefileCompatibility 检查已解析 Makefile 中的所有配方。
func (c *CompatChecker) CheckMakefileCompatibility(mf *parser.Makefile) ([]*CompatResult, error) {
	return c.checkMakefileCompatibilityForOS(mf, c.targetOS)
}

// CheckMakefileCompatibilityFor checks all recipes in a Makefile for compatibility with a specific OS.
// This is a convenience method that creates a checker for the specified OS and runs the check.
// CheckMakefileCompatibilityFor 检查 Makefile 中的所有配方与特定 OS 的兼容性。
// 这是一个便捷方法，为指定 OS 创建检查器并运行检查。
func CheckMakefileCompatibilityFor(mf *parser.Makefile, targetOS string) ([]*CompatResult, error) {
	checker := NewCompatCheckerForOS(targetOS)
	return checker.checkMakefileCompatibilityForOS(mf, targetOS)
}

// checkMakefileCompatibilityForOS is the internal implementation that uses the checker's state.
// checkMakefileCompatibilityForOS 是使用检查器状态的内部实现。
func (c *CompatChecker) checkMakefileCompatibilityForOS(mf *parser.Makefile, targetOS string) ([]*CompatResult, error) {
	if c.sysInfo == nil {
		if err := c.Init(); err != nil {
			return nil, err
		}
	}

	var results []*CompatResult
	seen := make(map[string]bool)

	for _, target := range mf.Targets {
		for _, recipe := range target.Recipes {
			cmd := extractRecipeCommand(recipe)
			if cmd == "" {
				continue
			}
			if seen[cmd] {
				continue
			}
			seen[cmd] = true

			result, err := c.CheckCommand(cmd)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// NewCompatReport creates a structured report from compatibility check results.
// NewCompatReport 从兼容性检查结果创建结构化报告。
func NewCompatReport(targetOS string, results []*CompatResult) *CompatReport {
	total := len(results)
	available := 0
	missing := 0
	for _, r := range results {
		if r.IsAvailable {
			available++
		} else {
			missing++
		}
	}

	currentOS := "unknown"
	// Try to detect current OS
	d := NewDetector()
	if info, err := d.Detect(); err == nil {
		currentOS = info.OS
	}

	return &CompatReport{
		TargetOS:  targetOS,
		CurrentOS: currentOS,
		Total:     total,
		Available: available,
		Missing:   missing,
		Results:   results,
	}
}

// extractRecipeCommand extracts the first command name from a recipe line.
// It handles variable references, pipes, and command prefixes.
// extractRecipeCommand 从配方行中提取第一个命令名称。
func extractRecipeCommand(recipe string) string {
	// Trim leading whitespace and @ prefix
	s := strings.TrimLeft(recipe, " \t-@")

	// If the recipe starts with $(VAR) or ${VAR}, extract the variable name
	if strings.HasPrefix(s, "$(") || strings.HasPrefix(s, "${") {
		closeChar := byte(')')
		if s[1] == '{' {
			closeChar = '}'
		}
		closeIdx := strings.IndexByte(s[2:], closeChar)
		if closeIdx >= 0 {
			varName := strings.ToLower(s[2 : 2+closeIdx])
			// Check if there's a subcommand after the variable reference
			rest := strings.TrimSpace(s[2+closeIdx+1:])
			// If the rest starts with a path-like character, strip it and check subcommand
			rest = strings.TrimLeft(rest, "/\\")
			restParts := strings.Fields(rest)
			if len(restParts) > 0 {
				// If the variable name is a known build tool, return it
				if isKnownBuildTool(varName) {
					return varName
				}
				// Otherwise return the subcommand
				return restParts[0]
			}
			// No subcommand, return the variable name
			return varName
		}
	}

	// Split on the first whitespace to get the command name
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}

	// The first part is typically the command
	return parts[0]
}

// isKnownBuildTool checks if a lowercase name matches a known build tool or command.
// isKnownBuildTool 检查小写名称是否与已知的构建工具或命令匹配。
func isKnownBuildTool(name string) bool {
	known := map[string]bool{
		"go":      true,
		"cargo":   true,
		"python":  true,
		"python3": true,
		"pip":     true,
		"pip3":    true,
		"node":    true,
		"npm":     true,
		"npx":     true,
		"yarn":    true,
		"make":    true,
		"cmake":   true,
		"docker":  true,
		"gcc":     true,
		"g++":     true,
		"rustc":   true,
		"dotnet":  true,
		"java":    true,
		"javac":   true,
		"mvn":     true,
		"gradle":  true,
	}
	return known[name]
}
