// Package main is the CLI entry point for makego.
// makego converts Makefiles to magefile.go, compiles them to native binaries,
// and interacts with the magego.hub.io API service.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/VDHewei/mage-makefile/pkg/config"
	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/interactive"
	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
	"github.com/VDHewei/mage-makefile/pkg/runtime"
)

var (
	cfgLoader *config.Loader
	cfg       *config.Config

	// Global flags
	configFile string
	verbose    bool

	// Convert flags
	convertOutput   string
	targetPlatform  string
	scriptEngine    string
	interactiveMode bool
	listTargets     bool

	// Compile flags
	compileOutput string
	targetOS      string
	targetArch    string
	bootstrap     bool
	goVersion     string
	sdkCache      string

	// Detect flags
	detectReport      bool
	detectOS          string
	detectJSON        bool
	detectInteractive bool
	detectInstall     bool

	// Hub flags
	hubServerURL string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "makego",
	Short: "Convert Makefile to magefile.go and compile to native binary",
	Long: `makego is a comprehensive CLI tool that converts GNU Makefiles to Go mage (magefile.go),
handles cross-platform compatibility (Windows/macOS/Linux), supports Lua/JS/Go scripting
for bash function conversion, and provides Hub API integration for sharing snippets.

Example usage:
  makego convert Makefile -o magefile.go     # Convert a Makefile
  makego compile magefile.go -o makego.exe   # Compile to native binary
  makego detect Makefile                     # Check platform compatibility
  makego hub search "docker build"           # Search hub for snippets`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfgLoader, err = config.NewLoader()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg = cfgLoader.Config()
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var convertCmd = &cobra.Command{
	Use:   "convert [makefile]",
	Short: "Convert a Makefile to magefile.go",
	Long: `Convert parses a GNU Makefile and generates an equivalent magefile.go
that uses Go code patterns for shell commands, variable handling, and target dependencies.

The converter supports:
- Standard targets and prerequisites
- Variable assignments and expansion
- Shell commands (converted to Go os/exec calls)
- Conditional directives (ifeq/ifneq/ifdef/ifndef)
- Include directives
- Multi-line define/endef variables
- .PHONY targets

If no Makefile path is provided, looks for 'Makefile' in the current directory.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConvert,
}

var compileCmd = &cobra.Command{
	Use:   "compile [magefile.go]",
	Short: "Compile magefile.go to a native binary",
	Long: `Compile takes a magefile.go and produces a self-contained native binary.
If Go is not installed on the system, it automatically downloads and caches the Go SDK.

Without --os/--arch flags, compiles for the current platform (native mode).
With --os/--arch flags, cross-compiles for the specified target platform.`,
	Args: cobra.ExactArgs(1),
	RunE: runCompile,
}

var detectCmd = &cobra.Command{
	Use:   "detect [makefile|command]",
	Short: "Detect platform compatibility of a Makefile or command",
	Long: `Detect analyzes a Makefile or bash command and reports platform compatibility.

For a Makefile: scans all shell commands and checks if they're available on the current platform.
For a command: checks if the command exists and suggests alternatives for other platforms.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDetect,
}

var hubCmd = &cobra.Command{
	Use:   "hub",
	Short: "Interact with magego.hub.io API service",
	Long:  `Hub provides commands for the magego.hub.io API service for sharing and discovering code snippets.`,
}

var hubPullCmd = &cobra.Command{
	Use:   "pull [name] [version]",
	Short: "Pull a code snippet from the hub",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runHubPull,
}

var hubPushCmd = &cobra.Command{
	Use:   "push [file]",
	Short: "Push a code snippet to the hub",
	Args:  cobra.ExactArgs(1),
	RunE:  runHubPush,
}

var hubLoginCmd = &cobra.Command{
	Use:   "login [username]",
	Short: "Login to the hub",
	Args:  cobra.ExactArgs(1),
	RunE:  runHubLogin,
}

var hubSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search snippets on the hub",
	Args:  cobra.ExactArgs(1),
	RunE:  runHubSearch,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print makego version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("makego v0.1.0")
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Convert flags
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "magefile.go", "Output magefile.go path")
	convertCmd.Flags().StringVarP(&targetPlatform, "platform", "p", "", "Target platform for conversion (linux/darwin/windows)")
	convertCmd.Flags().StringVar(&scriptEngine, "script", "go", "Script engine for bash conversion (go/lua/js)")
	convertCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "Interactive mode: select targets and preview code / 交互模式：选择目标并预览代码")
	convertCmd.Flags().BoolVar(&listTargets, "list-targets", false, "List available targets and exit / 列出可用的目标并退出")

	// Compile flags
	compileCmd.Flags().StringVarP(&compileOutput, "output", "o", "", "Output binary path")
	compileCmd.Flags().StringVar(&targetOS, "os", "", "Target OS for cross-compile (linux/darwin/windows)")
	compileCmd.Flags().StringVar(&targetArch, "arch", "", "Target architecture for cross-compile (amd64/arm64)")
	compileCmd.Flags().BoolVar(&bootstrap, "bootstrap", false, "Force download Go SDK even if installed")
	compileCmd.Flags().StringVar(&goVersion, "go-version", "", "Go SDK version (default: from config)")
	compileCmd.Flags().StringVar(&sdkCache, "sdk-cache", "", "Go SDK cache directory")

	// Detect flags
	detectCmd.Flags().BoolVarP(&detectReport, "report", "r", false, "Generate detailed compatibility report")
	detectCmd.Flags().StringVar(&detectOS, "os", "", "Target OS for compatibility check (linux/darwin/windows)")
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output compatibility report as JSON")
	detectCmd.Flags().BoolVarP(&detectInteractive, "interactive", "i", false, "Interactive mode: prompt to install missing commands")
	detectCmd.Flags().BoolVar(&detectInstall, "install", false, "Auto-install missing commands using default method")

	// Hub flags
	hubCmd.PersistentFlags().StringVar(&hubServerURL, "server", "", "Hub server URL (overrides config)")

	// Add subcommands
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(hubCmd)
	rootCmd.AddCommand(versionCmd)

	hubCmd.AddCommand(hubPullCmd)
	hubCmd.AddCommand(hubPushCmd)
	hubCmd.AddCommand(hubLoginCmd)
	hubCmd.AddCommand(hubSearchCmd)
}

// runConvert handles the convert subcommand with interactive and config-driven modes.
func runConvert(cmd *cobra.Command, args []string) error {
	inputFile := "Makefile"
	if len(args) > 0 {
		inputFile = args[0]
	}

	fmt.Printf("Converting %s -> %s (platform: %s, engine: %s)\n",
		inputFile, convertOutput,
		orDefault(targetPlatform, "current"),
		scriptEngine)

	// 读取输入文件
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", inputFile, err)
	}

	// 解析 Makefile
	p := parser.NewParser(string(data))
	makefileAST, err := p.Parse()
	if err != nil {
		if interactiveMode {
			fmt.Printf("Parse error: %v\nSkip this error? (y/N): ", err)
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				return fmt.Errorf("user chose to stop")
			}
		} else {
			return fmt.Errorf("parse %s: %w", inputFile, err)
		}
	}

	// 根据配置或 CLI 标志选择目标平台
	var tr *transformer.Transformer
	if targetPlatform != "" {
		tr = transformer.NewTransformerWithPlatform(targetPlatform)
	} else if cfg != nil && cfg.Convert.DefaultPlatform != "" {
		tr = transformer.NewTransformerWithPlatform(cfg.Convert.DefaultPlatform)
	} else {
		tr = transformer.NewTransformerWithPlatform("")
	}
	ir := tr.Transform(makefileAST)

	// --list-targets 标志：列出目标并退出
	if listTargets {
		fmt.Printf("Targets in %s / 目标列表:\n", inputFile)
		cat := interactive.NewCategorizer()
		categorized := cat.Categorize(ir)
		cat.DisplayCategories(categorized)
		fmt.Printf("\nTotal / 总计: %d targets\n", len(ir.Targets))
		return nil
	}

	// 配置驱动的非交互式过滤
	if cfg != nil && !interactiveMode {
		include := cfg.Convert.IncludeTargets
		exclude := cfg.Convert.ExcludeTargets
		if len(include) > 0 || len(exclude) > 0 {
			ir = filterTargets(ir, include, exclude)
			fmt.Printf("Filtered: %d targets remaining (config-driven / 配置驱动)\n", len(ir.Targets))
		}
	}

	// 交互模式
	if interactiveMode {
		engine := interactive.NewInteractiveEngine(cfg)
		if err := engine.Run(makefileAST, ir); err != nil {
			return fmt.Errorf("interactive: %w", err)
		}
		if len(ir.Targets) == 0 {
			fmt.Println("No targets to convert. Exiting. / 没有目标可转换。退出。")
			return nil
		}
	}

	// 从配置准备生成选项
	genOpts := generator.DefaultGeneratorOptions()
	if cfg != nil {
		genOpts.ApplyConfig(&cfg.Convert)
	}
	gen := generator.NewGeneratorWithConfig(ir, genOpts)

	// 生成 magefile.go
	code, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// 写入输出文件
	if err := os.WriteFile(convertOutput, []byte(code), 0644); err != nil {
		return fmt.Errorf("write %s: %w", convertOutput, err)
	}

	absPath, _ := filepath.Abs(convertOutput)
	fmt.Printf("Generated / 生成: %s\n", absPath)
	if interactiveMode {
		fmt.Printf("Targets converted / 已转换目标: %d\n", len(ir.Targets))
	}
	return nil
}

// filterTargets 根据 include/exclude 列表过滤 IR 中的目标。
func filterTargets(ir *transformer.IR, include, exclude []string) *transformer.IR {
	excludeSet := make(map[string]bool)
	for _, name := range exclude {
		excludeSet[name] = true
	}
	includeSet := make(map[string]bool)
	for _, name := range include {
		includeSet[name] = true
	}

	var filtered []transformer.IRTarget
	for _, target := range ir.Targets {
		if excludeSet[target.Name] {
			continue
		}
		if len(include) > 0 && !includeSet[target.Name] {
			continue
		}
		filtered = append(filtered, target)
	}
	ir.Targets = filtered
	return ir
}

// runCompile handles the compile subcommand.
func runCompile(cmd *cobra.Command, args []string) error {
	magefile := args[0]
	output := compileOutput
	if output == "" {
		output = "magebuild"
		if runtime.IsWindows() {
			output += ".exe"
		}
	}

	if targetOS != "" || targetArch != "" {
		fmt.Printf("Cross-compiling %s -> %s (%s/%s)\n", magefile, output, targetOS, targetArch)
		fmt.Println("Cross-compilation: will be implemented in Phase 6")
	} else {
		fmt.Printf("Compiling %s -> %s (native)\n", magefile, output)
		fmt.Println("Native compilation: will be implemented in Phase 6")
	}

	return nil
}

// runDetect handles the detect subcommand.
func runDetect(cmd *cobra.Command, args []string) error {
	d := runtime.NewDetector()
	info, err := d.Detect()
	if err != nil {
		return fmt.Errorf("detect: %w", err)
	}

	fmt.Printf("System: %s/%s\n", info.OS, info.Arch)
	fmt.Printf("Shell:  %s\n", info.Shell)
	fmt.Printf("Go:     %s\n", info.GoVersion)

	targetOS := detectOS
	if targetOS == "" {
		targetOS = info.OS
	}

	if len(args) == 0 {
		// No arguments: just show system info and available commands
		if len(info.AvailableCommands) > 0 {
			fmt.Printf("\nAvailable commands (%d):\n", len(info.AvailableCommands))
			for cmdName, cmdPath := range info.AvailableCommands {
				fmt.Printf("  %-20s %s\n", cmdName, cmdPath)
			}
		}
		return nil
	}

	target := args[0]
	fmt.Printf("\nChecking: %s\n", target)

	// Try to check if it's a file or a command
	if _, statErr := os.Stat(target); statErr == nil {
		// It's a file, parse and check as Makefile
		return handleMakefileDetect(target, targetOS, info.OS)
	}

	// Treat as a command name
	return handleCommandDetect(target, d)
}

// handleMakefileDetect parses a Makefile and runs compatibility checks.
func handleMakefileDetect(target, targetOS, currentOS string) error {
	data, readErr := os.ReadFile(target)
	if readErr != nil {
		return fmt.Errorf("read %s: %w", target, readErr)
	}

	mf, parseErr := parser.NewParser(string(data)).Parse()
	if parseErr != nil {
		return fmt.Errorf("parse %s: %w", target, parseErr)
	}

	var results []*runtime.CompatResult
	var checkErr error

	if detectOS != "" {
		// Check for a specific target OS
		results, checkErr = runtime.CheckMakefileCompatibilityFor(mf, targetOS)
	} else {
		// Check for current OS
		checker := runtime.NewCompatChecker()
		results, checkErr = checker.CheckMakefileCompatibility(mf)
	}

	if checkErr != nil {
		return fmt.Errorf("compat check: %w", checkErr)
	}

	// Create structured report
	report := runtime.NewCompatReport(targetOS, results)

	if detectJSON {
		// JSON output
		jsonData, jsonErr := report.MarshalJSON()
		if jsonErr != nil {
			return fmt.Errorf("json marshal: %w", jsonErr)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	if detectReport {
		// Detailed report
		fmt.Println(report.String())
	} else {
		// Compact output
		fmt.Printf("\nMakefile compatibility report (%s):\n", target)
		fmt.Printf("  Target OS / 目标系统: %s\n", targetOS)
		fmt.Printf("  Total: %d   Available: %d   Missing: %d\n\n",
			report.Total, report.Available, report.Missing)
		for _, r := range results {
			status := "[OK]"
			if !r.IsAvailable {
				status = "[MISSING]"
			}
			fmt.Printf("  %-20s %s", r.Command, status)
			if r.Description != "" {
				fmt.Printf("  — %s", r.Description)
			}
			fmt.Println()
			if !r.IsAvailable {
				if r.Alternative != "" {
					fmt.Printf("    → try / 尝试: %s\n", r.Alternative)
				}
				if r.InstallURL != "" {
					fmt.Printf("    → download / 下载: %s\n", r.InstallURL)
				}
			}
		}
	}

	// Handle installation for missing commands
	if detectInteractive || detectInstall {
		missingCount := 0
		for _, r := range results {
			if !r.IsAvailable {
				missingCount++
			}
		}

		if missingCount > 0 {
			installer := runtime.NewInstaller(detectInteractive)
			installer.SetVerbose(verbose)

			fmt.Printf("\n--- Missing commands / 缺失命令 (%d) ---\n", missingCount)
			skipAll := false
			for _, r := range results {
				if !r.IsAvailable {
					suggestion := installer.SuggestInstall(r.Command)

					if detectInteractive && !skipAll {
						ok, err := installer.InteractiveInstall(suggestion)
						if err != nil {
							fmt.Printf("  Install error / 安装错误: %v\n", err)
						}
						if !ok {
							// User chose to skip all
							skipAll = true
						}
					} else if detectInstall {
						_, err := installer.AutoInstall(suggestion)
						if err != nil {
							fmt.Printf("  Auto-install error / 自动安装错误: %v\n", err)
						}
					}
				}
			}
		} else {
			fmt.Println("\nAll commands are available / 所有命令都可用")
		}
	}

	return nil
}

// handleCommandDetect checks compatibility of a single command.
func handleCommandDetect(target string, d *runtime.Detector) error {
	available, path := d.IsCommandAvailable(target)
	cmdMap := runtime.LookupCommandMap(target)

	if available {
		fmt.Printf("  Available at / 路径: %s\n", path)
		if cmdMap != nil && cmdMap.Description != "" {
			fmt.Printf("  Description / 描述: %s\n", cmdMap.Description)
		}
	} else {
		fmt.Printf("  NOT available on this platform / 当前平台不可用\n")
		if cmdMap != nil {
			if cmdMap.Description != "" {
				fmt.Printf("  Description / 描述: %s\n", cmdMap.Description)
			}

			// Show alternatives for all platforms
			if alt, found := runtime.GetAlternative(target, "windows"); found {
				fmt.Printf("  Windows alternative / 替代: %s\n", alt)
			}
			if alt, found := runtime.GetAlternative(target, "darwin"); found {
				fmt.Printf("  macOS alternative / 替代: %s\n", alt)
			}
			if cmdMap.InstallURL != "" {
				fmt.Printf("  Download / 下载: %s\n", cmdMap.InstallURL)
			}
		}

		// Offer installation if interactive or install mode
		if detectInteractive || detectInstall {
			installer := runtime.NewInstaller(detectInteractive)
			installer.SetVerbose(verbose)
			suggestion := installer.SuggestInstall(target)

			if detectInteractive {
				_, err := installer.InteractiveInstall(suggestion)
				if err != nil {
					fmt.Printf("  Install error / 安装错误: %v\n", err)
				}
			} else if detectInstall {
				_, err := installer.AutoInstall(suggestion)
				if err != nil {
					fmt.Printf("  Auto-install error / 自动安装错误: %v\n", err)
				}
			}
		}
	}

	return nil
}

// runHubPull handles hub pull.
func runHubPull(cmd *cobra.Command, args []string) error {
	name := args[0]
	version := "latest"
	if len(args) > 1 {
		version = args[1]
	}
	fmt.Printf("Pulling snippet: %s@%s from %s\n", name, version, cfg.Hub.ServerURL)
	fmt.Println("Hub pull: will be implemented in Phase 8")
	return nil
}

// runHubPush handles hub push.
func runHubPush(cmd *cobra.Command, args []string) error {
	fmt.Printf("Pushing snippet: %s to %s\n", args[0], cfg.Hub.ServerURL)
	fmt.Println("Hub push: will be implemented in Phase 8")
	return nil
}

// runHubLogin handles hub login.
func runHubLogin(cmd *cobra.Command, args []string) error {
	fmt.Printf("Logging in as %s to %s\n", args[0], cfg.Hub.ServerURL)
	fmt.Println("Hub login: will be implemented in Phase 8")
	return nil
}

// runHubSearch handles hub search.
func runHubSearch(cmd *cobra.Command, args []string) error {
	fmt.Printf("Searching for '%s' on %s\n", args[0], cfg.Hub.ServerURL)
	fmt.Println("Hub search: will be implemented in Phase 8")
	return nil
}

// orDefault returns val if non-empty, otherwise returns defaultVal.
func orDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
