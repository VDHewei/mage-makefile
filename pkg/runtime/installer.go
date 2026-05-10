package runtime

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// InstallSuggestion contains all ways to install a command.
// InstallSuggestion 包含安装命令的所有方式。
type InstallSuggestion struct {
	Command     string
	Description string
	Methods     []InstallMethod
}

// String returns a human-readable formatted string of all installation options.
// String 返回所有安装选项的人类可读格式化字符串。
func (s *InstallSuggestion) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Missing command / 缺失命令: %s\n", s.Command))
	if s.Description != "" {
		b.WriteString(fmt.Sprintf("  Description / 描述: %s\n", s.Description))
	}
	b.WriteString("  Installation methods / 安装方式:\n")
	for i, m := range s.Methods {
		b.WriteString(fmt.Sprintf("    [%d] %s\n", i+1, m.Label))
	}
	return b.String()
}

// Installer provides installation guidance for missing commands.
// Installer 为缺失命令提供安装引导。
type Installer struct {
	interactive bool // enable user prompts / 启用用户提示
	verbose     bool
}

// NewInstaller creates a new Installer.
// NewInstaller 创建一个新的 Installer。
func NewInstaller(interactive bool) *Installer {
	return &Installer{
		interactive: interactive,
		verbose:     false,
	}
}

// SetVerbose sets the verbose flag.
// SetVerbose 设置详细输出标志。
func (inst *Installer) SetVerbose(v bool) {
	inst.verbose = v
}

// SuggestInstall returns installation guidance for a missing command.
// Returns nil if the command is already available or no install info is known.
// SuggestInstall 返回缺失命令的安装引导。
// 如果命令已可用或没有安装信息则返回 nil。
func (inst *Installer) SuggestInstall(cmd string) *InstallSuggestion {
	cmdMap := LookupCommandMap(cmd)
	if cmdMap == nil {
		return nil
	}

	// Check if command is already available
	d := NewDetector()
	available, _ := d.IsCommandAvailable(cmd)
	if available {
		return nil
	}

	suggestion := &InstallSuggestion{
		Command:     cmd,
		Description: cmdMap.Description,
	}

	// Add install URL as first method if available
	if cmdMap.InstallURL != "" {
		suggestion.Methods = append(suggestion.Methods, InstallMethod{
			Type:    "url",
			Content: cmdMap.InstallURL,
			Label:   fmt.Sprintf("Download from / 从以下地址下载: %s", cmdMap.InstallURL),
		})
	}

	// Add all configured install methods
	suggestion.Methods = append(suggestion.Methods, cmdMap.InstallCmds...)

	// If no methods at all, add a generic suggestion
	if len(suggestion.Methods) == 0 {
		suggestion.Methods = append(suggestion.Methods, InstallMethod{
			Type:    "url",
			Content: fmt.Sprintf("https://google.com/search?q=install+%s", cmd),
			Label:   fmt.Sprintf("Search online for / 在线搜索 '%s' 的安装方式", cmd),
		})
	}

	return suggestion
}

// InteractiveInstall prompts user to choose installation method and executes it.
// Returns true if installation succeeded or user chose to skip.
// InteractiveInstall 提示用户选择安装方式并执行。
// 如果安装成功或用户选择跳过则返回 true。
func (inst *Installer) InteractiveInstall(suggestion *InstallSuggestion) (bool, error) {
	if !inst.interactive {
		return false, fmt.Errorf("interactive mode is disabled / 交互模式未启用")
	}

	if suggestion == nil || len(suggestion.Methods) == 0 {
		return true, nil
	}

	fmt.Printf("\nMissing command / 缺失命令: %s\n", suggestion.Command)
	if suggestion.Description != "" {
		fmt.Printf("  %s\n", suggestion.Description)
	}
	fmt.Println("  Installation methods / 安装方式:")

	for i, m := range suggestion.Methods {
		fmt.Printf("  [%d] %s\n", i+1, m.Label)
	}
	fmt.Printf("  [%d] Skip this command / 跳过此命令\n", len(suggestion.Methods)+1)
	fmt.Printf("  [s] Skip all remaining commands / 跳过所有剩余命令\n")
	fmt.Printf("Choose (1-%d, s): / 请选择 (1-%d, s): ", len(suggestion.Methods)+1, len(suggestion.Methods)+1)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input / 读取输入失败: %w", err)
	}
	input = strings.TrimSpace(input)

	if strings.ToLower(input) == "s" {
		return false, nil // signal: skip all
	}

	var choice int
	_, err = fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(suggestion.Methods)+1 {
		fmt.Println("Invalid choice, skipping / 无效选择，跳过")
		return true, nil
	}

	if choice == len(suggestion.Methods)+1 {
		// User chose to skip this command
		return true, nil
	}

	method := suggestion.Methods[choice-1]
	return true, inst.executeInstall(method)
}

// AutoInstall automatically installs a missing command using the first available method.
// Returns true if installation succeeded.
// AutoInstall 使用第一个可用方法自动安装缺失命令。
// 如果安装成功则返回 true。
func (inst *Installer) AutoInstall(suggestion *InstallSuggestion) (bool, error) {
	if suggestion == nil || len(suggestion.Methods) == 0 {
		return true, nil
	}

	// Use the first non-URL method first, otherwise use URL
	var method InstallMethod
	for _, m := range suggestion.Methods {
		if m.Type != "url" {
			method = m
			break
		}
	}
	if method.Type == "" {
		method = suggestion.Methods[0]
	}

	fmt.Printf("Auto-installing / 自动安装: %s...\n", suggestion.Command)
	return true, inst.executeInstall(method)
}

// executeInstall executes the chosen installation method.
// executeInstall 执行选定的安装方式。
func (inst *Installer) executeInstall(method InstallMethod) error {
	switch method.Type {
	case "url":
		return inst.installViaURL(method.Content)
	case "sh":
		return inst.installViaScript("sh", method.Content)
	case "bat":
		return inst.installViaScript("bat", method.Content)
	case "ps1":
		return inst.installViaScript("ps1", method.Content)
	case "go":
		return inst.installViaGo(method.Content)
	default:
		return fmt.Errorf("unknown installation method type / 未知安装方式类型: %s", method.Type)
	}
}

// installViaURL opens the URL in the default browser (informational only).
// installViaURL 在默认浏览器中打开 URL（仅信息提示）。
func (inst *Installer) installViaURL(url string) error {
	if inst.verbose {
		fmt.Printf("Opening URL / 打开链接: %s\n", url)
	}
	fmt.Printf("Please download and install from / 请从以下地址下载并安装:\n  %s\n", url)

	// Try to open browser
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start() // Best-effort, ignore errors
	return nil
}

// installViaScript executes an install script (sh/bat/ps1).
// installViaScript 执行安装脚本（sh/bat/ps1）。
func (inst *Installer) installViaScript(scriptType, content string) error {
	if inst.verbose {
		fmt.Printf("Executing install script / 执行安装脚本 (%s): %s\n", scriptType, content)
	}

	var cmd *exec.Cmd
	switch scriptType {
	case "sh":
		if runtime.GOOS == "windows" {
			// Try using Git Bash or WSL
			cmd = exec.Command("bash", "-c", content)
		} else {
			cmd = exec.Command("sh", "-c", content)
		}
	case "bat":
		cmd = exec.Command("cmd", "/c", content)
	case "ps1":
		cmd = exec.Command("powershell", "-NoProfile", "-Command", content)
	default:
		return fmt.Errorf("unsupported script type / 不支持的脚本类型: %s", scriptType)
	}

	// Stream output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if inst.verbose {
		fmt.Printf("Running / 运行: %s\n", cmd.String())
	}

	return cmd.Run()
}

// installViaGo uses a Go-based implementation for installation.
// installViaGo 使用基于 Go 的实现进行安装。
func (inst *Installer) installViaGo(cmd string) error {
	if inst.verbose {
		fmt.Printf("Installing via Go / 通过 Go 安装: %s\n", cmd)
	}

	// Try go install for Go-based tools
	execCmd := exec.Command("go", "install", cmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	return execCmd.Run()
}
