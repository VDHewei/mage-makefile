// Package main is the CLI entry point for makego.
// makego converts Makefiles to magefile.go, compiles them to native binaries,
// and interacts with the magego.hub.io API service.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/VDHewei/mage-makefile/pkg/compiler"
	"github.com/VDHewei/mage-makefile/pkg/config"
	"github.com/VDHewei/mage-makefile/pkg/converter/generator"
	"github.com/VDHewei/mage-makefile/pkg/converter/interactive"
	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/VDHewei/mage-makefile/pkg/converter/transformer"
	"github.com/VDHewei/mage-makefile/pkg/hub"
	"github.com/VDHewei/mage-makefile/pkg/runtime"
	"github.com/VDHewei/mage-makefile/pkg/script"
)

var (
	cfgLoader *config.Loader
	cfg       *config.Config
	hubMgr    *hub.HubManager

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
		hubMgr = hub.NewHubManager(cfg)
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

// parseMagefileMetadata parses magefile.go file metadata.
func parseMagefileMetadata(filePath string) magefileMetadata {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return magefileMetadata{
			Name:        "unknown",
			Description: "magefile snippet",
			Tags:        []string{},
			Author:      "unknown",
			Platform:    "",
			Engine:      "",
			Metadata:    map[string]string{},
			CreatedAt:   time.Now(),
		}
	}

	// Extract metadata from file header comments
	meta := magefileMetadata{
		Name:        "unknown",
		Description: "magefile snippet",
		Tags:        []string{},
		Author:      "unknown",
		Platform:    "",
		Engine:      "",
		Metadata:    map[string]string{},
		CreatedAt:   time.Now(),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		// Look for // @hub lines
		if strings.HasPrefix(line, "// @hub") {
			parts := strings.Split(line[7:], " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if strings.HasPrefix(part, "name=") {
					meta.Name = strings.Trim(strings.TrimPrefix(part, "name="), q)
				} else if strings.HasPrefix(part, "description=") {
					meta.Description = strings.Trim(strings.TrimPrefix(part, "description="), q)
				} else if strings.HasPrefix(part, "author=") {
					meta.Author = strings.Trim(strings.TrimPrefix(part, "author="), q)
				} else if strings.HasPrefix(part, "platform=") {
					meta.Platform = strings.Trim(strings.TrimPrefix(part, "platform="), q)
				} else if strings.HasPrefix(part, "engine=") {
					meta.Engine = strings.Trim(strings.TrimPrefix(part, "engine="), q)
				} else if strings.HasPrefix(part, "tags=") {
					tagsStr := strings.Trim(strings.TrimPrefix(part, "tags="), q)
					meta.Tags = strings.Split(tagsStr, ",")
				} else if strings.HasPrefix(part, "created=") {
					createdAt := strings.Trim(strings.TrimPrefix(part, "created="), q)
					if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
						meta.CreatedAt = t
					}
				} else if strings.HasPrefix(part, "metadata=") {
					m := strings.Trim(strings.TrimPrefix(part, "metadata="), q)
					meta.Metadata = parseMetadataMap(m)
				}
			}
		}
	}

	// If no metadata found, use filename
	if meta.Name == "unknown" {
		meta.Name = filepath.Base(filePath)
	}

	return meta
}

// q is a single quote character for trimming.
var q = "\""

// parseMetadataMap 简单解析 metadata 字段。
func parseMetadataMap(jsonStr string) map[string]string {
	meta := map[string]string{}
	// Simple parsing, extracting key-value pairs
	pairs := strings.Split(jsonStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, `:`, 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(strings.Trim(kv[0], q))
			value := strings.TrimSpace(strings.Trim(kv[1], q))
			if key != "" && value != "" {
				meta[key] = value
			}
		}
	}
	return meta
}

// magefileMetadata represents metadata for a magefile.go file.
type magefileMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Platform    string            `json:"platform"`
	Engine      string            `json:"engine"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
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
	Use:   "login [username] [password]",
	Short: "Login to the hub with username/password or API key",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runHubLogin,
}

var hubSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search snippets on the hub",
	Args:  cobra.ExactArgs(1),
	RunE:  runHubSearch,
}

// Add hub list and version commands
var hubListCmd = &cobra.Command{
	Use:   "list [--page=N] [--size=M]",
	Short: "List all snippets on the hub (paginated)",
	Args:  cobra.NoArgs,
	RunE:  runHubList,
}

var hubVersionCmd = &cobra.Command{
	Use:   "versions [name]",
	Short: "Get version history for a snippet",
	Args:  cobra.ExactArgs(1),
	RunE:  runHubVersions,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print makego version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("makego v0.1.0")
	},
}

// serveCmd provides a built-in hub development server for local testing
var serveCmd = &cobra.Command{
	Use:   "serve [path]",
	Short: "Start Hub development server",
	Long: `Run a local Hub development server with embedded static files.

This provides a self-contained Hub server for testing and local development.
The server embeds all static assets (HTML, CSS, JS) and provides a simple
file-based snippet storage for local usage.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runServe,
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
	rootCmd.AddCommand(serveCmd)

	hubCmd.AddCommand(hubPullCmd)
	hubCmd.AddCommand(hubPushCmd)
	hubCmd.AddCommand(hubLoginCmd)
	hubCmd.AddCommand(hubSearchCmd)
	hubCmd.AddCommand(hubListCmd)
	hubCmd.AddCommand(hubVersionCmd)
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

	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", inputFile, err)
	}

	// Parse Makefile
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

	// Select target platform based on config or CLI flag
	engine := newScriptEngine(scriptEngine)
	var tr *transformer.Transformer
	if targetPlatform != "" {
		tr = transformer.NewTransformerWithEngine(targetPlatform, engine)
	} else if cfg != nil && cfg.Convert.DefaultPlatform != "" {
		tr = transformer.NewTransformerWithEngine(cfg.Convert.DefaultPlatform, engine)
	} else {
		tr = transformer.NewTransformerWithEngine("", engine)
	}
	ir := tr.Transform(makefileAST)

	// --list-targets flag: list targets and exit
	if listTargets {
		fmt.Printf("Targets in %s / 目标列表:\n", inputFile)
		cat := interactive.NewCategorizer()
		categorized := cat.Categorize(ir)
		cat.DisplayCategories(categorized)
		fmt.Printf("\nTotal / 总计：%d targets\n", len(ir.Targets))
		return nil
	}

	// Config-driven non-interactive filtering
	if cfg != nil && !interactiveMode {
		include := cfg.Convert.IncludeTargets
		exclude := cfg.Convert.ExcludeTargets
		if len(include) > 0 || len(exclude) > 0 {
			ir = filterTargets(ir, include, exclude)
			fmt.Printf("Filtered: %d targets remaining (config-driven / 配置驱动)\n", len(ir.Targets))
		}
	}

	// Interactive mode
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

	// Prepare generator options from config
	genOpts := generator.DefaultGeneratorOptions()
	if cfg != nil {
		genOpts.ApplyConfig(&cfg.Convert)
	}
	gen := generator.NewGeneratorWithConfig(ir, genOpts)

	// Generate magefile.go
	code, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// Write output file
	if err := os.WriteFile(convertOutput, []byte(code), 0644); err != nil {
		return fmt.Errorf("write %s: %w", convertOutput, err)
	}

	absPath, _ := filepath.Abs(convertOutput)
	fmt.Printf("Generated / 生成：%s\n", absPath)
	if interactiveMode {
		fmt.Printf("Targets converted / 已转换目标：%d\n", len(ir.Targets))
	}
	return nil
}

// filterTargets filters targets in IR based on include/exclude lists.
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

	c := compiler.NewNativeCompiler()
	if sdkCache != "" {
		c.SetSDKCacheDir(sdkCache)
	}

	// Bootstrap: ensure Go SDK is available
	if bootstrap {
		if err := c.Bootstrap(); err != nil {
			return fmt.Errorf("bootstrap: %w", err)
		}
		fmt.Println("Go SDK verified")
	} else {
		status, err := c.Detect()
		if err != nil {
			return fmt.Errorf("detect Go: %w", err)
		}
		if !status.Installed && !status.Cached {
			fmt.Println("Warning: Go not found. Use --bootstrap to download SDK.")
		}
	}

	// Perform compilation
	if targetOS != "" || targetArch != "" {
		fmt.Printf("Cross-compiling %s -> %s (%s/%s)\n", magefile, output, targetOS, targetArch)
		return c.Cross(magefile, targetOS, targetArch, output)
	}
	fmt.Printf("Compiling %s -> %s (native)\n", magefile, output)
	return c.Native(magefile, output)
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
		fmt.Printf("  Target OS / 目标系统：%s\n", targetOS)
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
					fmt.Printf("    → try / 尝试：%s\n", r.Alternative)
				}
				if r.InstallURL != "" {
					fmt.Printf("    → download / 下载：%s\n", r.InstallURL)
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
							fmt.Printf("  Install error / 安装错误：%v\n", err)
						}
						if !ok {
							// User chose to skip all
							skipAll = true
						}
					} else if detectInstall {
						_, err := installer.AutoInstall(suggestion)
						if err != nil {
							fmt.Printf("  Auto-install error / 自动安装错误：%v\n", err)
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
		fmt.Printf("  Available at / 路径：%s\n", path)
		if cmdMap != nil && cmdMap.Description != "" {
			fmt.Printf("  Description / 描述：%s\n", cmdMap.Description)
		}
	} else {
		fmt.Printf("  NOT available on this platform / 当前平台不可用\n")
		if cmdMap != nil {
			if cmdMap.Description != "" {
				fmt.Printf("  Description / 描述：%s\n", cmdMap.Description)
			}

			// Show alternatives for all platforms
			if alt, found := runtime.GetAlternative(target, "windows"); found {
				fmt.Printf("  Windows alternative / 替代：%s\n", alt)
			}
			if alt, found := runtime.GetAlternative(target, "darwin"); found {
				fmt.Printf("  macOS alternative / 替代：%s\n", alt)
			}
			if cmdMap.InstallURL != "" {
				fmt.Printf("  Download / 下载：%s\n", cmdMap.InstallURL)
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
					fmt.Printf("  Install error / 安装错误：%v\n", err)
				}
			} else if detectInstall {
				_, err := installer.AutoInstall(suggestion)
				if err != nil {
					fmt.Printf("  Auto-install error / 自动安装错误：%v\n", err)
				}
			}
		}
	}

	return nil
}

// runHubPull handles hub pull.
func runHubPull(cmd *cobra.Command, args []string) error {
	// If not logged in, prompt user to login
	if hubMgr == nil || !hubMgr.IsAuthenticated() {
		fmt.Println("Not authenticated. Please login first with: makego hub login [user]")
		return nil
	}

	name := args[0]
	version := "latest"
	if len(args) > 1 {
		version = args[1]
	}

	client := hubMgr.GetClient()
	snippet, err := client.Pull(name, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error pulling %s: %v\n", name, err)
		return err
	}

	fmt.Printf("\nSnippet: %s\n", snippet.Name)
	fmt.Printf("Description: %s\n", snippet.Description)
	fmt.Printf("Tags: %v\n", snippet.Tags)
	fmt.Printf("Author: %s\n", snippet.Author)
	fmt.Printf("Platform: %s\n", snippet.Platform)
	fmt.Printf("Engine: %s\n", snippet.Engine)
	fmt.Println("\nCode:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(snippet.Code)
	fmt.Println("--------------------------------------------------")
	return nil
}

// runHubPush handles hub push.
func runHubPush(cmd *cobra.Command, args []string) error {
	// If not logged in, prompt user to login
	if hubMgr == nil || !hubMgr.IsAuthenticated() {
		fmt.Println("Not authenticated. Please login first with: makego hub login [user]")
		return nil
	}

	filePath := args[0]
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	// Parse magefile.go to get metadata
	meta := parseMagefileMetadata(filePath)

	snippet := hub.Snippet{
		Name:        fmt.Sprintf("%s-latest", meta.Name),
		Code:        string(data),
		Description: meta.Description,
		Tags:        meta.Tags,
		Author:      meta.Author,
		Platform:    meta.Platform,
		Engine:      meta.Engine,
		Metadata:    meta.Metadata,
		CreatedAt:   meta.CreatedAt,
	}

	client := hubMgr.GetClient()
	resp, err := client.Push(snippet)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error pushing snippet: %v\n", err)
		return err
	}

	fmt.Printf("\nUpload successful!\n")
	fmt.Printf("Name:        %s\n", resp.Name)
	fmt.Printf("Version:     %s\n", resp.Version)
	fmt.Printf("Snippet URL: %s\n", resp.URL)
	return nil
}

// runHubLogin handles hub login.
func runHubLogin(cmd *cobra.Command, args []string) error {
	username := ""
	if len(args) > 0 {
		username = args[0]
	}

	hm := hubMgr
	if !hm.IsAuthenticated() {
		// Need to login
		_ = hm.URL()

		// Check for API key or password
		apiKey := os.Getenv("MAAGEHUB_API_KEY")
		password := os.Getenv("MAAGEHUB_PASSWORD")

		if apiKey != "" {
			fmt.Printf("\nLogging in with API key...\n")
			req := hub.LoginRequest{APIKey: apiKey}
			if _, err := hm.Login(req); err != nil {
				fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
				return err
			}
		} else if username != "" && password != "" {
			fmt.Printf("\nLogging in as %s...\n", username)
			req := hub.LoginRequest{Username: username, Password: password}
			if _, err := hm.Login(req); err != nil {
				fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
				return err
			}
		} else {
			fmt.Println("Please provide login credentials:")
			fmt.Println("  Export MAAGEHUB_API_KEY='your-api-key'")
			fmt.Println("  or Export MAAGEHUB_PASSWORD='your-password'")
			fmt.Println("  or run: makego hub login [username] [password]")
			fmt.Println()
			fmt.Println("Current hub server:", hm.URL())
			return nil
		}
	}

	fmt.Println("\nAuthentication status:")
	fmt.Printf("  Server: %s\n", hm.URL())
	fmt.Printf("  Authenticated: %v\n", hm.IsAuthenticated())
	fmt.Println("  Token: ***cached***")
	return nil
}

// runHubSearch handles hub search.
func runHubSearch(cmd *cobra.Command, args []string) error {
	// If not logged in, prompt user to login
	if hubMgr == nil || !hubMgr.IsAuthenticated() {
		fmt.Println("Not authenticated. Please login first with: makego hub login [user]")
		return nil
	}

	query := args[0]
	hm := hubMgr
	client := hm.GetClient()

	req := hub.SearchRequest{
		Query: query,
	}

	result, err := client.Search(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
		return err
	}

	fmt.Printf("\nSearch results for '%s' (%d total):\n", query, result.Total)
	fmt.Printf("Page: %d / %d\n", result.Page, (result.Total+result.PageSize-1)/result.PageSize)
	fmt.Println("--------------------------------------------------")
	for _, snippet := range result.Snippets {
		fmt.Printf("\nSnippet: %s\n", snippet.Name)
		fmt.Printf("  Description: %s\n", snippet.Description)
		fmt.Printf("  Tags: %v\n", snippet.Tags)
		fmt.Printf("  Author: %s\n", snippet.Author)
		fmt.Printf("  Platform: %s, Engine: %s\n", snippet.Platform, snippet.Engine)
		fmt.Println("  Created: ", snippet.CreatedAt)
	}
	fmt.Println("--------------------------------------------------")
	return nil
}

// runHubList handles listing all snippets.
func runHubList(cmd *cobra.Command, args []string) error {
	if hubMgr == nil || !hubMgr.IsAuthenticated() {
		fmt.Println("Not authenticated. Please login first with: makego hub login [user]")
		return nil
	}

	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("size")

	client := hubMgr.GetClient()

	result, err := client.List(page, pageSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "List failed: %v\n", err)
		return err
	}

	fmt.Printf("\nAll snippets (page %d of %d, total: %d):\n", result.Page, (result.Total+pageSize-1)/pageSize, result.Total)
	fmt.Println("--------------------------------------------------")
	for _, snippet := range result.Snippets {
		fmt.Printf("\nSnippet: %s\n", snippet.Name)
		fmt.Printf("  Description: %s\n", snippet.Description)
		fmt.Printf("  Tags: %v\n", snippet.Tags)
		fmt.Printf("  Author: %s\n", snippet.Author)
		fmt.Printf("  Platform: %s, Engine: %s\n", snippet.Platform, snippet.Engine)
	}
	fmt.Println("--------------------------------------------------")
	return nil
}

// runHubVersions handles getting version history.
func runHubVersions(cmd *cobra.Command, args []string) error {
	if hubMgr == nil || !hubMgr.IsAuthenticated() {
		fmt.Println("Not authenticated. Please login first with: makego hub login [user]")
		return nil
	}

	name := args[0]
	hm := hubMgr
	client := hm.GetClient()

	versions, err := client.Versions(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Version lookup failed: %v\n", err)
		return err
	}

	fmt.Printf("\nVersion history for %s:\n", name)
	fmt.Println("--------------------------------------------------")
	for _, v := range versions {
		fmt.Printf("  %s (%s) - %s\n", v.Version, v.CreatedAt, v.Snippet.Description)
	}
	fmt.Println("--------------------------------------------------")
	return nil
}

// orDefault returns val if non-empty, otherwise returns defaultVal.
func orDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

// newScriptEngine creates the corresponding script engine based on --script flag.
// Returns nil to indicate no script engine is used (use default shell execution mode).
func newScriptEngine(engineType string) script.ScriptEngine {
	switch engineType {
	case "lua":
		return script.NewLuaEngine(0)
	case "js":
		return script.NewJSEngine(0)
	case "go":
		return script.NewGoEngine()
	default:
		return nil
	}
}

// serveServer is a built-in Hub development server.
type serveServer struct {
	storage    *SnippetStorage
	mgr        *hub.HubManager
	serverURL  string
	listenAddr string
	verbose    bool
}

// SnippetStorage provides a simple file-based storage for local snippet development.
type SnippetStorage struct {
	data    map[string][]SnippetData
	index   map[string]map[string]string // name -> version -> filepath
	mu      sync.RWMutex
	baseDir string
}

// SnippetData represents a stored snippet.
type SnippetData struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Code        string            `json:"code"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Platform    string            `json:"platform"`
	Engine      string            `json:"engine"`
	Version     string            `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
}

// NewSnippetStorage creates a new in-memory snippet storage.
func NewSnippetStorage() *SnippetStorage {
	return &SnippetStorage{
		data:  make(map[string][]SnippetData),
		index: make(map[string]map[string]string),
	}
}

// SetBaseDir sets the base directory for persistent storage.
func (s *SnippetStorage) SetBaseDir(baseDir string) {
	s.baseDir = baseDir
}

// Save saves a snippet.
func (s *SnippetStorage) Save(snippet SnippetData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[snippet.Name] = append(s.data[snippet.Name], snippet)

	if _, ok := s.index[snippet.Name]; !ok {
		s.index[snippet.Name] = make(map[string]string)
	}
	s.index[snippet.Name][snippet.Version] = ""

	return nil
}

// Get retrieves a snippet by name and version.
func (s *SnippetStorage) Get(name, version string) (*SnippetData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snippets, ok := s.data[name]
	if !ok {
		return nil, fmt.Errorf("snippet %q not found", name)
	}

	for _, snip := range snippets {
		if snip.Version == version {
			return &snip, nil
		}
	}

	return nil, fmt.Errorf("version %q not found", version)
}

// List returns all snippets.
func (s *SnippetStorage) List() []SnippetData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []SnippetData
	for _, snippets := range s.data {
		result = append(result, snippets...)
	}
	return result
}

// search returns a search storage with query and tag support.
func (s *SnippetStorage) Search(query string, tags []string) []SnippetData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []SnippetData
	for _, snippets := range s.data {
		for _, snip := range snippets {
			matches := true
			if query != "" {
				descMatches := strings.Contains(snip.Description, query)
				codeMatches := strings.Contains(snip.Code, query)
				if !descMatches && !codeMatches {
					matches = false
				}
			}
			if len(tags) > 0 {
				for _, tag := range tags {
					if !contains(snip.Tags, tag) {
						matches = false
						break
					}
				}
			}
			if matches {
				results = append(results, snip)
			}
		}
	}
	return results
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// runServe handles the serve subcommand.
func runServe(cmd *cobra.Command, args []string) error {
	storage := NewSnippetStorage()

	serve := &serveServer{
		storage:    storage,
		serverURL:  "http://localhost:8080",
		listenAddr: ":8080",
		verbose:    verbose,
	}

	fmt.Println("Starting Hub development server...")
	fmt.Println("Server running at http://localhost:8080")

	// Serve the hub web server
	// For now, we'll use a simple static file server with embedded HTML
	go func() {
		// Start a simple Go HTTP server
		fs := http.FileServer(http.Dir("."))
		http.Handle("/", http.StripPrefix("/", fs))

		fmt.Printf("Serving static files from current directory on %s\n", serve.listenAddr)
		if err := http.ListenAndServe(serve.listenAddr, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Wait for interrupt
	select {}
}

// DefaultSnippetStorage provides a zero-value storage.
var DefaultSnippetStorage = NewSnippetStorage()
