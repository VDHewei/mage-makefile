# Phase 4: Makefile Compatibility Detection / 第四阶段：Makefile 兼容性检测

## Status / 状态: Completed / 已完成

## Overview / 概述

Phase 4 extends the existing compatibility checking infrastructure to support cross-platform Makefile analysis, structured reports, JSON output, fills in missing command mappings with functional descriptions, and adds an interactive installation guidance system for missing commands.

第四阶段扩展了现有的兼容性检查基础设施，支持跨平台 Makefile 分析、结构化报告、JSON 输出，填补缺失的命令映射并添加功能描述，同时为缺失命令添加交互式安装引导系统。

## Tasks / 任务

| Task | EN Description | CN Description | Status |
|------|---------------|---------------|--------|
| 4.1 | Extend `CommandMap` with `Description`, `InstallURL`, `InstallMethod` fields | 扩展 `CommandMap` 结构体，添加功能描述、安装 URL 和安装方式字段 | Done / 完成 |
| 4.2 | Fill all command mappings with functional descriptions and Windows alternatives | 为所有命令映射补充功能描述和 Windows 替代命令 | Done / 完成 |
| 4.3 | Add `NewCompatCheckerForOS(targetOS)` and `CheckMakefileCompatibilityFor(mf, targetOS)` for target-OS checking | 添加目标操作系统检查方法，支持指定非当前平台进行兼容性检查 | Done / 完成 |
| 4.4 | Add `CompatReport` struct with text and JSON serialization | 添加 `CompatReport` 报告结构体，支持文本和 JSON 序列化 | Done / 完成 |
| 4.5 | Add `--os` flag to detect command | 为 detect 命令添加 `--os` 标志 | Done / 完成 |
| 4.6 | Wire up `--report` flag for detailed output | 连接 `--report` 标志，输出结构化详细报告 | Done / 完成 |
| 4.7 | Add `--json` flag for machine-readable output | 添加 `--json` 标志，输出 JSON 格式报告 | Done / 完成 |
| 4.8 | Add `Installer` system: interactive missing-command installation via URL/shell/bat/PowerShell/Go | 添加 `Installer` 安装引导系统：通过 URL/shell/bat/PowerShell/Go 实现交互式安装缺失命令 | Done / 完成 |
| 4.9 | Integrate installer into detect command with `--interactive` / `--install` flags | 将安装引导集成到 detect 命令中，添加 `--interactive` / `--install` 标志 | Done / 完成 |
| 4.10 | Add runtime tests for all new methods | 为所有新方法添加运行时测试 | Done / 完成 |
| 4.11 | Add CLI tests for detect with Makefile, --os, --report, --json, --interactive | 为 detect 命令添加全面的 CLI 测试 | Done / 完成 |
| 4.12 | Full test suite verification | 全量测试验证 | Done / 完成 |

---

## Files to Modify / 待修改文件

### 1. `pkg/runtime/mapping.go` — 命令映射数据库扩展

EN: Extend `CommandMap` with description and installation info fields.

CN: 扩展 `CommandMap`，添加功能描述和安装信息字段。

**Structural changes / 结构体变更:**
```go
// InstallMethod describes how to install a missing command.
type InstallMethod struct {
    Type    string // "url", "sh", "bat", "ps1", "go"
    Content string // URL or script content
    Label   string // Human-readable label e.g. "Download from official site"
}

// CommandMap maps a Unix command to its platform equivalents.
type CommandMap struct {
    UnixCmd      string
    WindowsCmd   string
    MacOSCmd     string
    Category     string   // file, process, network, build, shell, system, text
    Description  string   // NEW: what the command does (e.g. "copy files and directories")
    InstallURL   string   // NEW: download URL for official installer
    InstallCmds  []InstallMethod // NEW: installation methods
}
```

**Description additions / 描述补充示例:**
| Command / 命令 | Description (EN) | Description (CN) |
|---------------|------------------|------------------|
| `cp` | copy files and directories | 复制文件和目录 |
| `mv` | move or rename files | 移动或重命名文件 |
| `rm` | remove files or directories | 删除文件或目录 |
| `cat` | concatenate and display file contents | 连接并显示文件内容 |
| `ls` | list directory contents | 列出目录内容 |
| `grep` | search text using patterns | 使用模式搜索文本 |
| `sed` | stream editor for text transformation | 流式文本编辑器 |
| `awk` | pattern scanning and processing language | 模式扫描和处理语言 |
| ... | (all 51+ commands get descriptions) | (所有 51+ 个命令都获得描述) |

**Install methods examples / 安装方式示例:**

For `grep` (via findstr on Windows is built-in):
```go
{UnixCmd: "grep", WindowsCmd: "findstr", MacOSCmd: "grep", Category: CatText,
    Description: "search text using patterns / 使用模式搜索文本",
    InstallURL: "https://www.gnu.org/software/grep/",
}
```

For `make` (not built-in, needs download):
```go
{UnixCmd: "make", WindowsCmd: "mingw32-make", MacOSCmd: "make", Category: CatBuild,
    Description: "build automation tool / 构建自动化工具",
    InstallURL: "https://www.gnu.org/software/make/",
    InstallCmds: []InstallMethod{
        {Type: "url", Content: "https://gnuwin32.sourceforge.net/packages/make.htm", Label: "Download GnuWin32 make"},
        {Type: "sh", Content: "apt-get install make", Label: "apt-get install (Linux)"},
        {Type: "ps1", Content: "choco install make", Label: "Chocolatey install (Windows)"},
    },
}
```

**Missing Windows alternatives to fill / 需要填补的 Windows 替代命令:**

All commands with empty `WindowsCmd`:

| UnixCmd | Windows Alternative | Description |
|---------|-------------------|-------------|
| `basename` | `powershell -Command "Split-Path -Leaf %~f1"` | Extract filename from path |
| `dirname` | `powershell -Command "Split-Path -Parent %~f1"` | Extract directory from path |
| `awk` | `powershell -Command` | Pattern scanning language |
| `sed` | `powershell -Command` | Stream text editor |
| `uniq` | `powershell -Command "Get-Content ... \| Get-Unique"` | Remove duplicate lines |
| `wc` | `powershell -Command "Measure-Object -Line -Word -Character"` | Count lines/words/chars |
| `cut` | `powershell -Command "ForEach-Object { $_.Split(' ')[0] }"` | Extract columns from text |
| `tr` | `powershell -Command` | Translate/delete characters |
| `tee` | `powershell -Command "Tee-Object -FilePath"` | Output to file and stdout |
| `xargs` | `powershell -Command "ForEach-Object"` | Build and execute command lines |
| `scp` | `powershell -Command "Copy-Item"` | Secure file copy |
| `nc` | `powershell -Command "Test-Connection"` | Network utility |

---

### 2. `pkg/runtime/compat.go` — 兼容性检查扩展

EN: Add new methods, `CompatReport` struct, and integrate with `CommandMap` descriptions.

CN: 添加新方法、`CompatReport` 结构体，并与 `CommandMap` 描述集成。

**a) `NewCompatCheckerForOS(targetOS string)`**
- Pre-initializes with target OS (no reliance on `runtime.GOOS`)
- 预初始化为指定目标 OS（不依赖 `runtime.GOOS`）

**b) `CheckMakefileCompatibilityFor(mf, targetOS)`**
- Convenience: one call creates target-OS checker + scans Makefile
- 便捷方法：一次调用创建目标 OS 检查器 + 扫描 Makefile

**c) `CompatReport` struct — structured report / 结构化报告**
```go
type CompatReport struct {
    TargetOS   string         `json:"target_os"`
    CurrentOS  string         `json:"current_os"`
    Total      int            `json:"total"`
    Available  int            `json:"available"`
    Missing    int            `json:"missing"`
    Results    []*CompatResult `json:"results"`
}
```
- `String() string` — text format with summary line / 文本格式，带摘要行
- `MarshalJSON() ([]byte, error)` / `UnmarshalJSON()` — JSON serialization / JSON 序列化

**d) Enhanced `CompatResult` with description and install info / 增强的 `CompatResult`:**
```go
type CompatResult struct {
    Command        string           `json:"command"`
    Description    string           `json:"description,omitempty"`     // NEW
    IsAvailable    bool             `json:"is_available"`
    IsShellBuiltin bool             `json:"is_shell_builtin,omitempty"`
    Alternative    string           `json:"alternative,omitempty"`
    Platforms      []string         `json:"platforms,omitempty"`
    Notes          string           `json:"notes,omitempty"`
    InstallURL     string           `json:"install_url,omitempty"`    // NEW
    InstallMethods []InstallMethod  `json:"install_methods,omitempty"` // NEW
}
```

---

### 3. `pkg/runtime/installer.go` — **新建文件：安装引导系统**

EN: A new package file that provides interactive installation guidance for missing commands.

CN: 新的包文件，为缺失命令提供交互式安装引导。

```go
// InstallMethod describes how to install a missing command.
type InstallMethod struct {
    Type    string // "url", "sh", "bat", "ps1", "go"
    Content string // URL or script content
    Label   string // Human-readable label
}

// Installer provides installation guidance for missing commands.
type Installer struct {
    interactive bool   // enable user prompts
    verbose     bool
}

// NewInstaller creates a new Installer.
func NewInstaller(interactive bool) *Installer

// SuggestInstall returns installation guidance for a missing command.
func (inst *Installer) SuggestInstall(cmd string) *InstallSuggestion

// InstallSuggestion contains all ways to install a command.
type InstallSuggestion struct {
    Command     string
    Description string
    Methods     []InstallMethod
}

// InteractiveInstall prompts user to choose installation method and executes it.
// Returns true if installation succeeded or user chose to skip.
func (inst *Installer) InteractiveInstall(suggestion *InstallSuggestion) (bool, error)

// InstallViaURL downloads and runs an installer from URL.
func (inst *Installer) InstallViaURL(url string) error

// InstallViaScript executes an install script (sh/bat/ps1).
func (inst *Installer) InstallViaScript(scriptType, content string) error

// InstallViaGo uses Go implementation (os/exec) to install.
func (inst *Installer) InstallViaGo(cmd string) error
```

**Install method types / 安装方式类型:**

| Type / 类型 | Description / 描述 | Example / 示例 |
|-------------|-------------------|----------------|
| `url` | Download from external URL | Download from https://go.dev/dl/ |
| `sh` | Shell install script | `apt-get install make` |
| `bat` | Windows batch script | `@echo off && choco install make` |
| `ps1` | PowerShell script | `Install-PackageProvider -Name chocolatey` |
| `go` | Go implementation | Use Go's `os/exec` to invoke package manager |

**Interaction flow / 交互流程:**

```
$ makego detect Makefile --interactive

System: windows/amd64
Shell:  C:\Program Files\Git\bin\bash.exe
Go:     go1.25.0

Checking: Makefile

Makefile compatibility report (Makefile):
  gcc                   [MISSING] — GNU C Compiler / GNU C 编译器
  → try: gcc (via MinGW-w64)

Missing command: gcc
  [1] Download from: https://sourceforge.net/projects/mingw-w64/
  [2] Install via: choco install mingw
  [3] Install via: winget install mingw
  [4] Install via Go: go install ... (cross-compiler)
  [5] Skip this command
Choose installation method (1-5, or 's' to skip all): _
```

---

### 4. `cmd/makego/main.go` — CLI 入口点扩展

EN: Add `--os`, `--json`, `--interactive` flags to detect command.

CN: 为 detect 命令添加 `--os`、`--json`、`--interactive` 标志。

**New variables / 新变量:**
```go
var (
    detectOS          string   // --os flag
    detectJSON        bool     // --json flag
    detectInteractive bool     // --interactive flag (reuses interactiveMode or new var)
    detectInstall     bool     // --install flag: auto-install missing commands
)
```

**Flag registration / 标志注册:**
```go
// In init():
detectCmd.Flags().StringVar(&detectOS, "os", "", "Target OS for compatibility check (linux/darwin/windows)")
detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output compatibility report as JSON")
detectCmd.Flags().BoolVarP(&detectInteractive, "interactive", "i", false, "Interactive mode: prompt to install missing commands")
detectCmd.Flags().BoolVar(&detectInstall, "install", false, "Auto-install missing commands using default method")
```

**Updated `runDetect()` flow / 更新后的 `runDetect()` 流程:**

```
1. Detect system info (always)
2. If args[0] is a file:
   2a. Parse Makefile
   2b. Create CompatChecker (for OS if --os set)
   2c. Run CheckMakefileCompatibility
   2d. If --json: marshal CompatReport to JSON, output, return
   2e. Print text report (compact or --report verbose)
   2f. If --interactive or --install:
       - For each MISSING command:
         - Show installation options
         - Ask user (interactive) or auto-install (--install)
         - Execute chosen installation method
3. If args[0] is a command name:
   3a. Check availability
   3b. Show install options if missing
4. Print available commands list
```

---

### 5. New test files / 新建测试文件

**`pkg/runtime/installer_test.go`:**
- `TestInstaller_SuggestInstall_Existing` — existing command returns nil suggestion
- `TestInstaller_SuggestInstall_Missing` — missing command returns suggestion with methods
- `TestInstaller_SuggestInstall_WithURL` — command with InstallURL returns URL method
- `TestInstaller_SuggestInstall_WithScripts` — command with InstallCmds returns script methods
- `TestInstallSuggestion_String` — formatted output

**`pkg/runtime/compat_test.go`** or add to `runtime_test.go`:
- `TestNewCompatCheckerForOS` — verify target OS
- `TestCompatChecker_CheckMakefileCompatibilityFor` — scan for specific OS
- `TestCompatReport_String` — text report format
- `TestCompatReport_MarshalJSON` / `TestCompatReport_UnmarshalJSON` — JSON round-trip
- `TestEnhancedCompatResult_Description` — command description populated
- `TestEnhancedCompatResult_InstallURL` — install URL populated when missing

**`cmd/makego/main_test.go`** additions:
- `TestCLI_DetectWithMakefile` — basic Makefile compatibility check
- `TestCLI_DetectWithOS` — `--os linux` target OS in output
- `TestCLI_DetectWithReport` — `--report` detailed output
- `TestCLI_DetectWithJSON` — `--json` valid JSON
- `TestCLI_DetectWithInteractive` — `--interactive` mode prompt (limited test: just verify it doesn't crash)
- `TestCLI_DetectCommandName` — detect a specific command name

---

## Execution Order / 执行顺序

| Step / 步骤 | EN Action / 操作 | CN Action / 操作 | Verify / 验证 |
|-------------|------------------|------------------|--------------|
| 1 | Edit `pkg/runtime/mapping.go`: add `Description`, `InstallURL`, `InstallMethod` fields to `CommandMap`; fill all 51+ entries | 编辑 mapping.go：扩展 CommandMap 结构体，填充所有命令映射 | — |
| 2 | Edit `pkg/runtime/compat.go`: add `NewCompatCheckerForOS`, `CheckMakefileCompatibilityFor`, `CompatReport`, enhanced `CompatResult` | 编辑 compat.go：添加新方法和结构体 | — |
| 3 | Create `pkg/runtime/installer.go`: `Installer` struct, `SuggestInstall`, `InteractiveInstall` methods | 新建 installer.go：安装引导系统 | — |
| 4 | Run `go test ./pkg/runtime/` | 验证运行时测试 | ✅ Ensure existing tests pass |
| 5 | Edit `cmd/makego/main.go`: add `--os`, `--json`, `--interactive`, `--install` flags; update `runDetect()` | 编辑 main.go：添加标志，更新 detect 逻辑 | — |
| 6 | Run `go build ./cmd/makego/` | 编译 CLI | ✅ Build succeeds |
| 7 | Edit `pkg/runtime/runtime_test.go`: add new runtime tests | 编辑 runtime_test.go：添加新测试 | — |
| 8 | Create `pkg/runtime/installer_test.go`: installer unit tests | 新建 installer_test.go：安装器测试 | — |
| 9 | Edit `cmd/makego/main_test.go`: add new CLI tests | 编辑 main_test.go：添加 CLI 测试 | — |
| 10 | Run `go test ./...` | 全量测试 | ✅ All tests pass |
| 11 | Update plan status to Completed | 更新计划状态 | — |

---

## Key Design Decisions / 关键设计决策

| # | EN Decision | CN Decision | Rationale |
|---|-------------|-------------|-----------|
| 1 | `InstallMethod` uses script type enum (`url/sh/bat/ps1/go`) | 安装方式使用枚举类型 | Flexible — any installation mechanism fits one of these types / 灵活，任何安装机制都适合其中一种类型 |
| 2 | Installer is in `pkg/runtime` (not separate package) | 安装器在 `pkg/runtime` 内 | No circular deps; already has detector and command map / 无循环依赖；已有检测器和命令映射 |
| 3 | `--interactive` mode uses `os.Stdin` for user input | 交互模式使用 `os.Stdin` 获取用户输入 | Simpler than separate TUI; works on all platforms / 比独立 TUI 更简单；全平台兼容 |
| 4 | `--install` auto-installs without prompting | `--install` 自动安装无需提示 | For CI/CD pipelines / 适用于 CI/CD 流水线 |
| 5 | `detectOS` is separate from `compile --os` flag | `detectOS` 与 `compile --os` 标志分开 | Different commands, different contexts / 不同命令，不同上下文 |
| 6 | Descriptions are bilingual (EN + CN) in code comments | 描述在代码注释中提供中英双语 | Useful for CLI help text and report output / 对 CLI 帮助文本和报告输出有用 |

---

## Test Strategy / 测试策略

**Runtime unit tests / 运行时单元测试:**
- Verify `NewCompatCheckerForOS` sets correct target OS / 验证目标 OS 设置
- Verify `CheckMakefileCompatibilityFor` produces different results for different OS / 验证不同 OS 产生不同结果
- Verify `CompatReport.String()` contains summary and per-command details / 验证文本报告含摘要和详情
- Verify `CompatReport.MarshalJSON()` produces valid JSON with all fields / 验证 JSON 输出完整
- Verify `Installer.SuggestInstall()` returns correct methods for known commands / 验证安装建议正确
- Verify `Installer.SuggestInstall()` returns nil for unknown commands / 验证未知命令返回 nil

**CLI integration tests / CLI 集成测试:**
- Create temp dirs with test Makefiles / 创建临时目录和测试 Makefile
- Use `rootCmd.SetArgs()` + stdout capture / 使用 SetArgs + stdout 捕获
- Verify JSON output can be unmarshalled back / 验证 JSON 输出可反序列化
- Test `--interactive` with stdin pipe / 测试交互模式（使用 stdin 管道）

**Test data / 测试数据:**
- Inline Makefile strings (following `runtime_test.go` pattern) / 内联 Makefile 字符串
- No external fixture files needed / 无需外部 fixture 文件

---

## Dependencies / 依赖

- **Phase 2 (Parser)**: `parser.Makefile` struct
- **Phase 3 (CLI Integration)**: detect command structure, cobra patterns
- **Existing CompatChecker**: extended, not replaced
- **External deps**: none (`encoding/json` is stdlib)

---

## Final Test Results / 最终测试结果

```
=== pkg/runtime ===
PASS (36/36) — 25.269s
  ✓ TestNewInstaller
  ✓ TestInstaller_SuggestInstall_Existing
  ✓ TestInstaller_SuggestInstall_Missing
  ✓ TestInstaller_SuggestInstall_WithURL
  ✓ TestInstaller_SuggestInstall_WithScripts
  ✓ TestInstallSuggestion_String
  ✓ TestAutoInstall_SkipAvailable
  ✓ TestNewDetector
  ✓ TestDetect_ReturnsSystemInfo
  ✓ TestDetect_ShellDetection
  ✓ TestDetect_HomeDir
  ✓ TestIsCommandAvailable_Go
  ✓ TestBuiltinCommandMaps_NotEmpty
  ✓ TestBuiltinCommandMaps_HasAllCategories
  ✓ TestShellBuiltins_NotEmpty
  ✓ TestIsShellBuiltin_Echo / NonBuiltin
  ✓ TestLookupCommandMap_Found / NotFound
  ✓ TestGetAlternative_Windows / Linux / MacOS / UnknownOS
  ✓ TestGetAlternative_MakeOnWindows / ShellBuiltinExport
  ✓ TestGetPlatforms_Returns / UnixOnly / CrossPlatform / Unknown
  ✓ TestNewCompatChecker / Init
  ✓ TestCompatChecker_CheckCommand_Available / NotAvailable / ShellBuiltin
  ✓ TestCompatChecker_FindAlternative / NotFound
  ✓ TestCompatChecker_CheckMakefileCompatibility / Empty / WithoutInit
  ✓ TestExtractRecipeCommand_SimpleCommand (6 subtests)
  ✓ TestExtractRecipeCommand_WithPrefix / WithDashPrefix / Empty
  ✓ TestCommandMappings_CrossPlatformCoverage / ShellBuiltinsHaveMappings
  ✓ TestExtractRecipeCommand_VariablePrefix / BraceVariablePrefix / ComplexCommand
  ✓ TestNewCompatCheckerForOS
  ✓ TestCheckMakefileCompatibilityFor_DifferentOS
  ✓ TestNewCompatReport
  ✓ TestCompatReport_String
  ✓ TestCompatReport_MarshalJSON (+ UnmarshalJSON roundtrip)
  ✓ TestCommandMap_DescriptionsFilled
  ✓ TestCommandMap_WindowsAlternativesFilled
  ✓ TestCommandMap_InstallInfoForBuildTools
  ✓ TestCompatResult_DescriptionPopulated / IsShellBuiltin

=== cmd/makego (CLI) ===
PASS (10/10) — 19.521s
  ✓ TestCLI_Version
  ✓ TestCLI_ConvertBasic
  ✓ TestCLI_ConvertWithVars
  ✓ TestCLI_ConvertWithDeps
  ✓ TestCLI_DetectCommand
  ✓ TestCLI_DetectWithMakefile
  ✓ TestCLI_DetectWithOS
  ✓ TestCLI_DetectWithJSON
  ✓ TestCLI_DetectCommandName
  ✓ TestCLI_DetectWithReport

=== All packages ===
  ok  github.com/VDHewei/mage-makefile/cmd/makego
  ok  github.com/VDHewei/mage-makefile/pkg/compiler
  ok  github.com/VDHewei/mage-makefile/pkg/config
  ok  github.com/VDHewei/mage-makefile/pkg/converter/generator
  ok  github.com/VDHewei/mage-makefile/pkg/converter/interactive
  ok  github.com/VDHewei/mage-makefile/pkg/converter/parser
  ok  github.com/VDHewei/mage-makefile/pkg/converter/transformer
  ok  github.com/VDHewei/mage-makefile/pkg/runtime
  ok  github.com/VDHewei/mage-makefile/pkg/script

### Phase 4 Regression Check / 回归检查 ###
All Phase 4 new features verified:
  ✓ CommandMap with Description, InstallURL, InstallMethod
  ✓ All 51+ command mappings with bilingual descriptions
  ✓ NewCompatCheckerForOS + CheckMakefileCompatibilityFor
  ✓ CompatReport with text and JSON (roundtrip)
  ✓ --os / --report / --json flags on detect command
  ✓ Installer system with SuggestInstall / InteractiveInstall / AutoInstall
  ✓ Integrated into detect command with --interactive / --install flags
  ✓ All runtime tests pass (36/36)
  ✓ All CLI detect tests pass (5 new + 1 existing)
  ✓ Full test suite: 9/9 packages pass
