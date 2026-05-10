package runtime

// InstallMethod describes how to install a missing command.
// 安装方式：描述如何安装缺失的命令。
type InstallMethod struct {
	Type    string // "url", "sh", "bat", "ps1", "go"
	Content string // URL or script content
	Label   string // Human-readable label / 人类可读的标签
}

// CommandMap maps a Unix command to its Windows and macOS equivalents.
// 命令映射：将 Unix 命令映射到 Windows 和 macOS 对应命令。
type CommandMap struct {
	UnixCmd     string
	WindowsCmd  string
	MacOSCmd    string
	Category    string          // file, process, network, build, shell, system, text
	Description string          // Bilingual description / 中英双语描述
	InstallURL  string          // Official download URL / 官方下载链接
	InstallCmds []InstallMethod // Installation methods / 安装方式列表
}

// Category constants for command classification.
const (
	CatFile    = "file"
	CatProcess = "process"
	CatNetwork = "network"
	CatBuild   = "build"
	CatShell   = "shell"
	CatSystem  = "system"
	CatText    = "text"
)

// BuiltinCommandMaps is the database of common bash commands and their platform equivalents.
// 内置命令映射表：常用 bash 命令及其平台对应命令的数据库。
var BuiltinCommandMaps = []CommandMap{
	// File operations / 文件操作
	{UnixCmd: "cp", WindowsCmd: "copy", MacOSCmd: "cp", Category: CatFile,
		Description: "copy files and directories / 复制文件和目录"},
	{UnixCmd: "mv", WindowsCmd: "move", MacOSCmd: "mv", Category: CatFile,
		Description: "move or rename files / 移动或重命名文件"},
	{UnixCmd: "rm", WindowsCmd: "del", MacOSCmd: "rm", Category: CatFile,
		Description: "remove files or directories / 删除文件或目录"},
	{UnixCmd: "rmdir", WindowsCmd: "rmdir", MacOSCmd: "rmdir", Category: CatFile,
		Description: "remove empty directories / 删除空目录"},
	{UnixCmd: "mkdir", WindowsCmd: "mkdir", MacOSCmd: "mkdir", Category: CatFile,
		Description: "create directories / 创建目录"},
	{UnixCmd: "touch", WindowsCmd: "type nul >", MacOSCmd: "touch", Category: CatFile,
		Description: "create empty file or update timestamp / 创建空文件或更新时间戳"},
	{UnixCmd: "ln", WindowsCmd: "mklink", MacOSCmd: "ln", Category: CatFile,
		Description: "create hard and symbolic links / 创建硬链接和符号链接"},
	{UnixCmd: "chmod", WindowsCmd: "icacls", MacOSCmd: "chmod", Category: CatFile,
		Description: "change file permissions / 修改文件权限"},
	{UnixCmd: "chown", WindowsCmd: "icacls", MacOSCmd: "chown", Category: CatFile,
		Description: "change file owner / 修改文件所有者"},
	{UnixCmd: "ls", WindowsCmd: "dir", MacOSCmd: "ls", Category: CatFile,
		Description: "list directory contents / 列出目录内容"},
	{UnixCmd: "cat", WindowsCmd: "type", MacOSCmd: "cat", Category: CatFile,
		Description: "concatenate and display file contents / 连接并显示文件内容"},
	{UnixCmd: "tail", WindowsCmd: `powershell -Command "Get-Content -Tail"`, MacOSCmd: "tail", Category: CatFile,
		Description: "display last part of a file / 显示文件尾部内容"},
	{UnixCmd: "head", WindowsCmd: `powershell -Command "Get-Content -Head"`, MacOSCmd: "head", Category: CatFile,
		Description: "display first part of a file / 显示文件头部内容"},
	{UnixCmd: "find", WindowsCmd: "dir /s", MacOSCmd: "find", Category: CatFile,
		Description: "search for files in directory hierarchy / 在目录层次结构中搜索文件"},
	{UnixCmd: "which", WindowsCmd: "where", MacOSCmd: "which", Category: CatFile,
		Description: "locate a command's executable path / 定位命令的可执行文件路径"},
	{UnixCmd: "pwd", WindowsCmd: "cd", MacOSCmd: "pwd", Category: CatFile,
		Description: "print current working directory / 显示当前工作目录"},
	{UnixCmd: "basename", WindowsCmd: `powershell -Command "Split-Path -Leaf"`, MacOSCmd: "basename", Category: CatFile,
		Description: "strip directory and suffix from filenames / 从文件名中去除目录和后缀"},
	{UnixCmd: "dirname", WindowsCmd: `powershell -Command "Split-Path -Parent"`, MacOSCmd: "dirname", Category: CatFile,
		Description: "strip last component from file path / 从文件路径中去除最后一个组成部分"},

	// Process operations / 进程操作
	{UnixCmd: "ps", WindowsCmd: "tasklist", MacOSCmd: "ps", Category: CatProcess,
		Description: "report process status / 报告进程状态"},
	{UnixCmd: "kill", WindowsCmd: "taskkill", MacOSCmd: "kill", Category: CatProcess,
		Description: "terminate a process / 终止进程"},
	{UnixCmd: "top", WindowsCmd: "tasklist", MacOSCmd: "top", Category: CatProcess,
		Description: "display system tasks / 显示系统任务"},
	{UnixCmd: "nice", WindowsCmd: "start /low", MacOSCmd: "nice", Category: CatProcess,
		Description: "run a program with modified scheduling priority / 以调整后的调度优先级运行程序"},

	// Network operations / 网络操作
	{UnixCmd: "curl", WindowsCmd: "curl", MacOSCmd: "curl", Category: CatNetwork,
		Description: "transfer data from/to URLs / 从/向 URL 传输数据",
		InstallURL:  "https://curl.se/download.html"},
	{UnixCmd: "wget", WindowsCmd: "curl", MacOSCmd: "wget", Category: CatNetwork,
		Description: "non-interactive network downloader / 非交互式网络下载器",
		InstallURL:  "https://www.gnu.org/software/wget/"},
	{UnixCmd: "ping", WindowsCmd: "ping", MacOSCmd: "ping", Category: CatNetwork,
		Description: "test network connectivity / 测试网络连通性"},
	{UnixCmd: "ssh", WindowsCmd: "ssh", MacOSCmd: "ssh", Category: CatNetwork,
		Description: "secure shell remote login / 安全 Shell 远程登录",
		InstallURL:  "https://www.openssh.com/"},
	{UnixCmd: "scp", WindowsCmd: `powershell -Command "Copy-Item"`, MacOSCmd: "scp", Category: CatNetwork,
		Description: "secure file copy over SSH / 通过 SSH 安全复制文件",
		InstallURL:  "https://www.openssh.com/"},
	{UnixCmd: "nc", WindowsCmd: `powershell -Command "Test-Connection"`, MacOSCmd: "nc", Category: CatNetwork,
		Description: "TCP/UDP network utility / TCP/UDP 网络工具",
		InstallURL:  "https://nmap.org/ncat/"},
	{UnixCmd: "netstat", WindowsCmd: "netstat", MacOSCmd: "netstat", Category: CatNetwork,
		Description: "display network connections and listening ports / 显示网络连接和监听端口"},

	// Build tools / 构建工具
	{UnixCmd: "make", WindowsCmd: "mingw32-make", MacOSCmd: "make", Category: CatBuild,
		Description: "build automation tool / 构建自动化工具",
		InstallURL:  "https://www.gnu.org/software/make/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://gnuwin32.sourceforge.net/packages/make.htm", Label: "Download GnuWin32 make / 下载 GnuWin32 make"},
			{Type: "ps1", Content: "choco install make", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
			{Type: "ps1", Content: "winget install GnuWin32.Make", Label: "Install via winget / 通过 winget 安装"},
		}},
	{UnixCmd: "gcc", WindowsCmd: "gcc", MacOSCmd: "gcc", Category: CatBuild,
		Description: "GNU C Compiler / GNU C 编译器",
		InstallURL:  "https://gcc.gnu.org/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://sourceforge.net/projects/mingw-w64/", Label: "Download MinGW-w64 / 下载 MinGW-w64"},
			{Type: "ps1", Content: "choco install mingw", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "g++", WindowsCmd: "g++", MacOSCmd: "g++", Category: CatBuild,
		Description: "GNU C++ Compiler / GNU C++ 编译器",
		InstallURL:  "https://gcc.gnu.org/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://sourceforge.net/projects/mingw-w64/", Label: "Download MinGW-w64 / 下载 MinGW-w64"},
			{Type: "ps1", Content: "choco install mingw", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "cmake", WindowsCmd: "cmake", MacOSCmd: "cmake", Category: CatBuild,
		Description: "cross-platform build system generator / 跨平台构建系统生成器",
		InstallURL:  "https://cmake.org/download/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://cmake.org/download/", Label: "Download CMake / 下载 CMake"},
			{Type: "ps1", Content: "choco install cmake", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "docker", WindowsCmd: "docker", MacOSCmd: "docker", Category: CatBuild,
		Description: "containerization platform / 容器化平台",
		InstallURL:  "https://docker.com/get-started/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://desktop.docker.com/win/stable/Docker%20Desktop%20Installer.exe", Label: "Download Docker Desktop / 下载 Docker Desktop"},
			{Type: "ps1", Content: "choco install docker-desktop", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "go", WindowsCmd: "go", MacOSCmd: "go", Category: CatBuild,
		Description: "Go programming language toolchain / Go 编程语言工具链",
		InstallURL:  "https://go.dev/dl/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://go.dev/dl/", Label: "Download Go / 下载 Go"},
			{Type: "ps1", Content: "choco install golang", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "cargo", WindowsCmd: "cargo", MacOSCmd: "cargo", Category: CatBuild,
		Description: "Rust package manager and build tool / Rust 包管理器和构建工具",
		InstallURL:  "https://rustup.rs/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://rustup.rs/", Label: "Install via rustup / 通过 rustup 安装"},
			{Type: "ps1", Content: "choco install rust", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "npm", WindowsCmd: "npm", MacOSCmd: "npm", Category: CatBuild,
		Description: "Node.js package manager / Node.js 包管理器",
		InstallURL:  "https://nodejs.org/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://nodejs.org/download/", Label: "Download Node.js / 下载 Node.js"},
			{Type: "ps1", Content: "choco install nodejs", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "node", WindowsCmd: "node", MacOSCmd: "node", Category: CatBuild,
		Description: "JavaScript runtime environment / JavaScript 运行时环境",
		InstallURL:  "https://nodejs.org/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://nodejs.org/download/", Label: "Download Node.js / 下载 Node.js"},
			{Type: "ps1", Content: "choco install nodejs", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "python", WindowsCmd: "python", MacOSCmd: "python", Category: CatBuild,
		Description: "Python programming language / Python 编程语言",
		InstallURL:  "https://www.python.org/downloads/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://www.python.org/downloads/", Label: "Download Python / 下载 Python"},
			{Type: "ps1", Content: "choco install python", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "python3", WindowsCmd: "python", MacOSCmd: "python3", Category: CatBuild,
		Description: "Python 3 programming language / Python 3 编程语言",
		InstallURL:  "https://www.python.org/downloads/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://www.python.org/downloads/", Label: "Download Python 3 / 下载 Python 3"},
			{Type: "ps1", Content: "choco install python", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "pip", WindowsCmd: "pip", MacOSCmd: "pip", Category: CatBuild,
		Description: "Python package installer / Python 包安装器",
		InstallURL:  "https://pip.pypa.io/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://bootstrap.pypa.io/get-pip.py", Label: "Download get-pip.py / 下载 get-pip.py"},
			{Type: "ps1", Content: "python -m ensurepip", Label: "Install via ensurepip / 通过 ensurepip 安装"},
		}},
	{UnixCmd: "pip3", WindowsCmd: "pip", MacOSCmd: "pip3", Category: CatBuild,
		Description: "Python 3 package installer / Python 3 包安装器",
		InstallURL:  "https://pip.pypa.io/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://bootstrap.pypa.io/get-pip.py", Label: "Download get-pip.py / 下载 get-pip.py"},
			{Type: "ps1", Content: "python3 -m ensurepip", Label: "Install via ensurepip / 通过 ensurepip 安装"},
		}},

	// Text processing / 文本处理
	{UnixCmd: "grep", WindowsCmd: "findstr", MacOSCmd: "grep", Category: CatText,
		Description: "search text using patterns / 使用模式搜索文本",
		InstallURL:  "https://www.gnu.org/software/grep/"},
	{UnixCmd: "egrep", WindowsCmd: "findstr", MacOSCmd: "egrep", Category: CatText,
		Description: "extended grep with regex / 使用扩展正则表达式的 grep",
		InstallURL:  "https://www.gnu.org/software/grep/"},
	{UnixCmd: "awk", WindowsCmd: `powershell -Command "Select-Object"`, MacOSCmd: "awk", Category: CatText,
		Description: "pattern scanning and processing language / 模式扫描和处理语言",
		InstallURL:  "https://www.gnu.org/software/gawk/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://gnuwin32.sourceforge.net/packages/gawk.htm", Label: "Download Gawk for Windows / 下载 Windows 版 Gawk"},
			{Type: "ps1", Content: "choco install gawk", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "sed", WindowsCmd: `powershell -Command "(Get-Content) -replace"`, MacOSCmd: "sed", Category: CatText,
		Description: "stream editor for text transformation / 流式文本编辑器",
		InstallURL:  "https://www.gnu.org/software/sed/",
		InstallCmds: []InstallMethod{
			{Type: "url", Content: "https://gnuwin32.sourceforge.net/packages/sed.htm", Label: "Download Sed for Windows / 下载 Windows 版 Sed"},
			{Type: "ps1", Content: "choco install sed", Label: "Install via Chocolatey / 通过 Chocolatey 安装"},
		}},
	{UnixCmd: "sort", WindowsCmd: "sort", MacOSCmd: "sort", Category: CatText,
		Description: "sort lines of text / 对文本行排序"},
	{UnixCmd: "uniq", WindowsCmd: `powershell -Command "Get-Content | Get-Unique"`, MacOSCmd: "uniq", Category: CatText,
		Description: "report or omit duplicate lines / 报告或忽略重复行",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "wc", WindowsCmd: `powershell -Command "Measure-Object -Line -Word -Character"`, MacOSCmd: "wc", Category: CatText,
		Description: "count lines, words, and characters / 统计行数、单词数和字符数",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "cut", WindowsCmd: `powershell -Command "ForEach-Object { $_.Split()[0] }"`, MacOSCmd: "cut", Category: CatText,
		Description: "extract columns from text files / 从文本文件中提取列",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "tr", WindowsCmd: `powershell -Command "(Get-Content) -replace"`, MacOSCmd: "tr", Category: CatText,
		Description: "translate or delete characters / 转换或删除字符",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "tee", WindowsCmd: `powershell -Command "Tee-Object -FilePath"`, MacOSCmd: "tee", Category: CatText,
		Description: "output to file and stdout simultaneously / 同时输出到文件和控制台",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "xargs", WindowsCmd: `powershell -Command "ForEach-Object"`, MacOSCmd: "xargs", Category: CatText,
		Description: "build and execute command lines from stdin / 从标准输入构建并执行命令行",
		InstallURL:  "https://www.gnu.org/software/findutils/"},
	{UnixCmd: "diff", WindowsCmd: "fc", MacOSCmd: "diff", Category: CatText,
		Description: "compare files line by line / 逐行比较文件",
		InstallURL:  "https://www.gnu.org/software/diffutils/"},

	// Shell built-ins / Shell 内置命令
	{UnixCmd: "echo", WindowsCmd: "echo", MacOSCmd: "echo", Category: CatShell,
		Description: "display a line of text / 显示一行文本"},
	{UnixCmd: "export", WindowsCmd: "set", MacOSCmd: "export", Category: CatShell,
		Description: "set environment variable / 设置环境变量"},
	{UnixCmd: "source", WindowsCmd: "call", MacOSCmd: "source", Category: CatShell,
		Description: "execute commands from a file / 从文件执行命令"},
	{UnixCmd: "cd", WindowsCmd: "cd", MacOSCmd: "cd", Category: CatShell,
		Description: "change current directory / 切换当前目录"},
	{UnixCmd: "exit", WindowsCmd: "exit", MacOSCmd: "exit", Category: CatShell,
		Description: "exit the shell / 退出 Shell"},
	{UnixCmd: "true", WindowsCmd: `cmd /c "exit 0"`, MacOSCmd: "true", Category: CatShell,
		Description: "return zero exit status / 返回零退出状态"},
	{UnixCmd: "false", WindowsCmd: `cmd /c "exit 1"`, MacOSCmd: "false", Category: CatShell,
		Description: "return non-zero exit status / 返回非零退出状态"},
	{UnixCmd: "alias", WindowsCmd: "doskey", MacOSCmd: "alias", Category: CatShell,
		Description: "create command alias / 创建命令别名"},
	{UnixCmd: "set", WindowsCmd: "set", MacOSCmd: "set", Category: CatShell,
		Description: "set shell option or positional parameter / 设置 Shell 选项或位置参数"},
	{UnixCmd: "unset", WindowsCmd: "set =", MacOSCmd: "unset", Category: CatShell,
		Description: "unset shell variable / 取消设置 Shell 变量"},
	{UnixCmd: "printf", WindowsCmd: "echo", MacOSCmd: "printf", Category: CatShell,
		Description: "format and print data / 格式化并打印数据"},

	// System / 系统命令
	{UnixCmd: "date", WindowsCmd: "date", MacOSCmd: "date", Category: CatSystem,
		Description: "display or set system date/time / 显示或设置系统日期/时间"},
	{UnixCmd: "sleep", WindowsCmd: "timeout", MacOSCmd: "sleep", Category: CatSystem,
		Description: "delay execution for specified time / 延迟指定时间后继续执行"},
	{UnixCmd: "env", WindowsCmd: "set", MacOSCmd: "env", Category: CatSystem,
		Description: "display or modify environment variables / 显示或修改环境变量"},
	{UnixCmd: "uname", WindowsCmd: "ver", MacOSCmd: "uname", Category: CatSystem,
		Description: "print system information / 打印系统信息"},
	{UnixCmd: "hostname", WindowsCmd: "hostname", MacOSCmd: "hostname", Category: CatSystem,
		Description: "display or set system hostname / 显示或设置系统主机名"},
	{UnixCmd: "df", WindowsCmd: "wmic logicaldisk get size,freespace,caption", MacOSCmd: "df", Category: CatSystem,
		Description: "report file system disk space usage / 报告文件系统磁盘空间使用情况",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
	{UnixCmd: "du", WindowsCmd: `powershell -Command "Get-ChildItem -Recurse | Measure-Object -Property Length -Sum"`, MacOSCmd: "du", Category: CatSystem,
		Description: "estimate file/directory space usage / 估算文件/目录空间使用情况",
		InstallURL:  "https://www.gnu.org/software/coreutils/"},
}

// ShellBuiltins is the set of commands that are typically shell built-ins.
// ShellBuiltins 是通常为 Shell 内置命令的命令集合。
// Cross-platform shell built-in conversion / 跨平台 Shell 内置命令转换:
// | Unix Shell | Windows (cmd) | PowerShell Equivalent |
// |------------|---------------|----------------------|
// | echo       | echo          | Write-Host / Write-Output |
// | cd         | cd            | Set-Location |
// | exit       | exit          | exit |
// | export     | set           | $env:VAR = "value" |
// | source     | call          | . script.ps1 |
// | set        | set           | Set-Variable / $VAR=val |
// | unset      | set ="        | Remove-Variable / Remove-Item env:VAR |
// | alias      | doskey        | Set-Alias |
// | true       | cmd /c "exit 0" | $true |
// | false      | cmd /c "exit 1" | $false |
// | pwd        | cd            | Get-Location |
// | read       | set /p        | Read-Host |
// | printf     | echo          | Write-Host -NoNewline -f format |
// | type       | type          | Get-Content |
// | return     | exit /b       | return |
// | eval       | cmd /c        | Invoke-Expression |
// | exec       | start         | Start-Process |
// | shift      | shift         | (no direct equivalent / 无直接对应) |
// | trap       | (no direct eq) | trap |
var ShellBuiltins = map[string]bool{
	"echo":   true,
	"cd":     true,
	"exit":   true,
	"export": true,
	"source": true,
	"set":    true,
	"unset":  true,
	"alias":  true,
	"true":   true,
	"false":  true,
	"pwd":    true,
	"read":   true,
	"printf": true,
	"type":   true,
	"return": true,
	"eval":   true,
	"exec":   true,
	"shift":  true,
	"trap":   true,
}

// IsShellBuiltin returns true if the command is typically a shell built-in.
// IsShellBuiltin 如果命令通常是 Shell 内置命令则返回 true。
func IsShellBuiltin(cmd string) bool {
	return ShellBuiltins[cmd]
}

// LookupCommandMap finds the CommandMap entry for a given Unix command name.
// Returns nil if not found.
// LookupCommandMap 查找给定 Unix 命令名的 CommandMap 条目，找不到则返回 nil。
func LookupCommandMap(unixCmd string) *CommandMap {
	for i := range BuiltinCommandMaps {
		if BuiltinCommandMaps[i].UnixCmd == unixCmd {
			return &BuiltinCommandMaps[i]
		}
	}
	return nil
}

// GetAlternative returns the platform-specific alternative for a command.
// The platform parameter should be "windows" or "darwin".
// Returns ("", false) if no alternative exists or the command is native.
// GetAlternative 返回命令的平台特定替代方案。
func GetAlternative(unixCmd string, platform string) (string, bool) {
	m := LookupCommandMap(unixCmd)
	if m == nil {
		return "", false
	}
	switch platform {
	case "windows":
		if m.WindowsCmd != "" && m.WindowsCmd != unixCmd {
			return m.WindowsCmd, true
		}
	case "darwin":
		if m.MacOSCmd != "" && m.MacOSCmd != unixCmd {
			return m.MacOSCmd, true
		}
	}
	return "", false
}

// GetPlatforms returns the list of platforms that support a given command.
// GetPlatforms 返回支持给定命令的平台列表。
func GetPlatforms(unixCmd string) []string {
	m := LookupCommandMap(unixCmd)
	if m == nil {
		return nil
	}
	var platforms []string
	if m.UnixCmd != "" {
		platforms = append(platforms, "linux")
	}
	if m.MacOSCmd != "" {
		platforms = append(platforms, "darwin")
	}
	if m.WindowsCmd != "" {
		platforms = append(platforms, "windows")
	}
	return platforms
}
