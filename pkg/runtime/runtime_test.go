package runtime

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/VDHewei/mage-makefile/pkg/converter/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Detector Tests
// =============================================================================

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	assert.NotNil(t, d)
}

func TestDetect_ReturnsSystemInfo(t *testing.T) {
	d := NewDetector()
	info, err := d.Detect()

	require.NoError(t, err)
	require.NotNil(t, info)

	assert.NotEmpty(t, info.OS, "OS should not be empty")
	assert.NotEmpty(t, info.Arch, "Arch should not be empty")
	assert.NotEmpty(t, info.GoVersion, "GoVersion should not be empty")
	assert.NotNil(t, info.AvailableCommands, "AvailableCommands map should be initialized")

	// OS should be a known value
	assert.Contains(t, []string{"linux", "darwin", "windows"}, info.OS,
		"OS should be a known platform")
}

func TestDetect_ShellDetection(t *testing.T) {
	d := NewDetector()
	info, err := d.Detect()

	require.NoError(t, err)

	// Shell should be set to something non-empty
	assert.NotEmpty(t, info.Shell, "Shell should not be empty")
}

func TestDetect_HomeDir(t *testing.T) {
	d := NewDetector()
	info, err := d.Detect()

	require.NoError(t, err)

	expectedHome, _ := os.UserHomeDir()
	assert.Equal(t, expectedHome, info.HomeDir)
}

func TestIsCommandAvailable_Go(t *testing.T) {
	d := NewDetector()

	// Go should be available since we're running Go tests
	available, path := d.IsCommandAvailable("go")
	assert.True(t, available, "go command should be available")
	assert.NotEmpty(t, path, "go path should not be empty")
}

func TestIsCommandAvailable_Nonexistent(t *testing.T) {
	d := NewDetector()

	// A command that almost certainly doesn't exist
	available, path := d.IsCommandAvailable("nonexistent_command_xyz_12345")
	assert.False(t, available, "nonexistent command should not be available")
	assert.Empty(t, path, "path should be empty for nonexistent command")
}

// =============================================================================
// Mapping Tests
// =============================================================================

func TestBuiltinCommandMaps_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, BuiltinCommandMaps, "command maps should be populated")
}

func TestBuiltinCommandMaps_HasAllCategories(t *testing.T) {
	categories := make(map[string]bool)
	for _, cm := range BuiltinCommandMaps {
		categories[cm.Category] = true
	}

	expectedCategories := []string{CatFile, CatProcess, CatNetwork, CatBuild, CatShell, CatSystem, CatText}
	for _, cat := range expectedCategories {
		assert.True(t, categories[cat], "expected category %s to be present", cat)
	}
}

func TestShellBuiltins_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, ShellBuiltins, "shell builtins should be populated")
}

func TestIsShellBuiltin_Echo(t *testing.T) {
	assert.True(t, IsShellBuiltin("echo"), "echo should be a shell built-in")
	assert.True(t, IsShellBuiltin("cd"), "cd should be a shell built-in")
}

func TestIsShellBuiltin_NonBuiltin(t *testing.T) {
	assert.False(t, IsShellBuiltin("go"), "go should not be a shell built-in")
	assert.False(t, IsShellBuiltin("git"), "git should not be a shell built-in")
}

func TestLookupCommandMap_Found(t *testing.T) {
	result := LookupCommandMap("cp")
	require.NotNil(t, result, "cp should be in the command map")
	assert.Equal(t, "cp", result.UnixCmd)
	assert.Equal(t, "copy", result.WindowsCmd)
}

func TestLookupCommandMap_NotFound(t *testing.T) {
	result := LookupCommandMap("nonexistent_command_abc")
	assert.Nil(t, result, "nonexistent command should not be found")
}

func TestGetAlternative_Windows(t *testing.T) {
	alt, found := GetAlternative("cp", "windows")
	assert.True(t, found, "cp should have a Windows alternative")
	assert.Equal(t, "copy", alt)
}

func TestGetAlternative_Linux(t *testing.T) {
	alt, found := GetAlternative("cp", "linux")
	assert.False(t, found, "cp on linux should not have an alternative (it's native)")
	assert.Empty(t, alt)
}

func TestGetAlternative_MacOS(t *testing.T) {
	alt, found := GetAlternative("cp", "darwin")
	assert.False(t, found, "cp on darwin should not have an alternative (it's native)")
	assert.Empty(t, alt)
}

func TestGetAlternative_UnknownOS(t *testing.T) {
	alt, found := GetAlternative("grep", "freebsd")
	assert.False(t, found, "unknown OS should not yield alternative")
	assert.Empty(t, alt)
}

func TestGetAlternative_MakeOnWindows(t *testing.T) {
	alt, found := GetAlternative("make", "windows")
	assert.True(t, found, "make should have a Windows alternative")
	assert.NotEmpty(t, alt)
}

func TestGetAlternative_ShellBuiltinExport(t *testing.T) {
	alt, found := GetAlternative("export", "windows")
	assert.True(t, found, "export should have a Windows alternative")
	assert.Equal(t, "set", alt)
}

func TestGetPlatforms_ReturnsPlatforms(t *testing.T) {
	platforms := GetPlatforms("cp")
	assert.NotEmpty(t, platforms)
	assert.Contains(t, platforms, "linux")
	assert.Contains(t, platforms, "windows")
}

func TestGetPlatforms_UnixOnly(t *testing.T) {
	platforms := GetPlatforms("grep")
	assert.NotEmpty(t, platforms)
	assert.Contains(t, platforms, "linux")
	// grep has findstr on windows
	assert.Contains(t, platforms, "windows")
}

func TestGetPlatforms_CrossPlatform(t *testing.T) {
	platforms := GetPlatforms("go")
	assert.NotEmpty(t, platforms)
	assert.Contains(t, platforms, "linux")
	assert.Contains(t, platforms, "darwin")
	assert.Contains(t, platforms, "windows")
}

func TestGetPlatforms_Unknown(t *testing.T) {
	platforms := GetPlatforms("unknown_command_xyz")
	assert.Nil(t, platforms)
}

// =============================================================================
// CompatChecker Tests
// =============================================================================

func TestNewCompatChecker(t *testing.T) {
	cc := NewCompatChecker()
	assert.NotNil(t, cc)
	assert.NotNil(t, cc.detector)
}

func TestCompatChecker_Init(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)
	require.NotNil(t, cc.sysInfo)
	assert.Equal(t, runtime.GOOS, cc.sysInfo.OS)
}

func TestCompatChecker_CheckCommand_Available(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	result, err := cc.CheckCommand("go")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "go", result.Command)
	assert.True(t, result.IsAvailable, "go should be available")
	assert.NotEmpty(t, result.Platforms)
	assert.NotEmpty(t, result.Notes)
}

func TestCompatChecker_CheckCommand_NotAvailable(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	result, err := cc.CheckCommand("nonexistent_command_xyz")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.False(t, result.IsAvailable)
	assert.NotEmpty(t, result.Notes)
}

func TestCompatChecker_CheckCommand_ShellBuiltin(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	// echo is a shell built-in that is also available as a command
	// The notes should mention it's a shell built-in
	result, err := cc.CheckCommand("echo")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Notes)
}

func TestCompatChecker_FindAlternative(t *testing.T) {
	cc := NewCompatChecker()

	alt, found := cc.FindAlternative("cp", "windows")
	assert.True(t, found)
	assert.Equal(t, "copy", alt)
}

func TestCompatChecker_FindAlternative_NotFound(t *testing.T) {
	cc := NewCompatChecker()

	alt, found := cc.FindAlternative("cp", "linux")
	assert.False(t, found)
	assert.Empty(t, alt)
}

func TestCompatChecker_CheckMakefileCompatibility(t *testing.T) {
	cc := NewCompatChecker()

	mf := &parser.Makefile{
		Targets: []*parser.Target{
			{
				Name: "build",
				Recipes: []string{
					"\tgo build -v ./...",
					"\techo Build complete",
				},
			},
			{
				Name: "test",
				Recipes: []string{
					"\tgo test ./...",
				},
			},
		},
	}

	results, err := cc.CheckMakefileCompatibility(mf)
	require.NoError(t, err)
	assert.NotEmpty(t, results, "should have results for recipes")

	// Each unique command should have one result
	seen := make(map[string]bool)
	for _, r := range results {
		assert.False(t, seen[r.Command], "duplicate command in results: %s", r.Command)
		seen[r.Command] = true
	}

	// go should be available and echo should be checked
	commandsFound := make(map[string]bool)
	for _, r := range results {
		commandsFound[r.Command] = true
	}
	assert.True(t, commandsFound["go"], "go should be in results")
	assert.True(t, commandsFound["echo"], "echo should be in results")
}

func TestCompatChecker_CheckMakefileCompatibility_Empty(t *testing.T) {
	cc := NewCompatChecker()

	mf := &parser.Makefile{
		Targets: []*parser.Target{
			{
				Name:    "phony-target",
				Recipes: nil,
			},
		},
	}

	results, err := cc.CheckMakefileCompatibility(mf)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestCompatChecker_CheckMakefileCompatibility_WithoutInit(t *testing.T) {
	cc := NewCompatChecker()

	mf := &parser.Makefile{
		Targets: []*parser.Target{
			{
				Name: "hello",
				Recipes: []string{
					"\techo hello",
				},
			},
		},
	}

	// Should auto-init
	results, err := cc.CheckMakefileCompatibility(mf)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

// =============================================================================
// extractRecipeCommand Tests
// =============================================================================

func TestExtractRecipeCommand_SimpleCommand(t *testing.T) {
	tests := []struct {
		recipe   string
		expected string
	}{
		{"\tgo build ./...", "go"},
		{"\techo hello", "echo"},
		{"\tmkdir -p dir", "mkdir"},
		{"\trm -rf temp", "rm"},
		{"\tcp src dst", "cp"},
		{"\tgit commit -m msg", "git"},
	}

	for _, tt := range tests {
		t.Run(tt.recipe, func(t *testing.T) {
			result := extractRecipeCommand(tt.recipe)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRecipeCommand_WithPrefix(t *testing.T) {
	result := extractRecipeCommand("\t@echo silent")
	assert.Equal(t, "echo", result)
}

func TestExtractRecipeCommand_WithDashPrefix(t *testing.T) {
	result := extractRecipeCommand("\t-go build")
	assert.Equal(t, "go", result)
}

func TestExtractRecipeCommand_Empty(t *testing.T) {
	assert.Empty(t, extractRecipeCommand(""))
	assert.Empty(t, extractRecipeCommand("\t"))
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestCommandMappings_CrossPlatformCoverage(t *testing.T) {
	// Verify that core build commands have cross-platform mappings
	coreBuildCommands := []string{"go", "gcc", "npm", "cargo", "python", "docker"}
	for _, cmd := range coreBuildCommands {
		platforms := GetPlatforms(cmd)
		assert.NotEmpty(t, platforms, "Command %s should have platform entries", cmd)
	}
}

func TestCommandMappings_ShellBuiltinsHaveMappings(t *testing.T) {
	// Shell built-ins should have cross-platform mappings
	shellCmds := []string{"echo", "export", "cd"}
	for _, cmd := range shellCmds {
		m := LookupCommandMap(cmd)
		assert.NotNil(t, m, "Shell built-in %s should have a mapping", cmd)
	}
}

func TestExtractRecipeCommand_VariablePrefix(t *testing.T) {
	// Recipe that starts with a variable reference
	result := extractRecipeCommand("\t$(GO) build ./...")
	assert.Equal(t, "go", result, "should extract go after $(GO)")
}

func TestExtractRecipeCommand_BraceVariablePrefix(t *testing.T) {
	result := extractRecipeCommand("\t${CARGO} build --release")
	assert.Equal(t, "cargo", result, "should extract cargo after ${CARGO}")
}

func TestExtractRecipeCommand_ComplexCommand(t *testing.T) {
	// Commands with special characters in arguments shouldn't confuse extraction
	result := extractRecipeCommand("\tgo build -ldflags \"-X main.version=1.0\" ./...")
	assert.Equal(t, "go", result)
}

func TestCompatChecker_CheckCommand_WithAlternative(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	// Test a command that exists in the mapping but might not be installed
	result, err := cc.CheckCommand("awk")
	require.NoError(t, err)
	require.NotNil(t, result)

	if strings.ToLower(runtime.GOOS) == "windows" {
		// On Windows, awk may or may not be available (e.g., through Git Bash)
		// If not available, it should have an alternative from the mapping
		if !result.IsAvailable {
			assert.NotEmpty(t, result.Alternative, "awk should have alternative on Windows if not available")
		}
	}
}

// =============================================================================
// Phase 4: New CompatCheckerForOS Tests
// =============================================================================

func TestNewCompatCheckerForOS(t *testing.T) {
	cc := NewCompatCheckerForOS("linux")
	require.NotNil(t, cc)
	assert.Equal(t, "linux", cc.targetOS, "targetOS should be linux")

	cc2 := NewCompatCheckerForOS("windows")
	require.NotNil(t, cc2)
	assert.Equal(t, "windows", cc2.targetOS, "targetOS should be windows")
}

func TestCheckMakefileCompatibilityFor_DifferentOS(t *testing.T) {
	mf := &parser.Makefile{
		Targets: []*parser.Target{
			{Name: "build", Recipes: []string{"\tgcc -o program main.c"}},
		},
	}

	// Check for linux (gcc should be the same)
	results, err := CheckMakefileCompatibilityFor(mf, "linux")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "gcc", results[0].Command)
	assert.NotEmpty(t, results[0].Description, "gcc should have description")
}

// =============================================================================
// Phase 4: CompatReport Tests
// =============================================================================

func TestNewCompatReport(t *testing.T) {
	results := []*CompatResult{
		{Command: "gcc", IsAvailable: true},
		{Command: "make", IsAvailable: false, Alternative: "mingw32-make"},
	}

	report := NewCompatReport("windows", results)
	require.NotNil(t, report)
	assert.Equal(t, "windows", report.TargetOS)
	assert.Equal(t, 2, report.Total)
	assert.Equal(t, 1, report.Available)
	assert.Equal(t, 1, report.Missing)
}

func TestCompatReport_String(t *testing.T) {
	results := []*CompatResult{
		{
			Command:     "gcc",
			IsAvailable: true,
			Description: "GNU C Compiler / GNU C 编译器",
		},
		{
			Command:     "make",
			IsAvailable: false,
			Description: "build automation tool / 构建自动化工具",
			Alternative: "mingw32-make",
			InstallURL:  "https://www.gnu.org/software/make/",
		},
	}

	report := NewCompatReport("linux", results)
	s := report.String()
	assert.Contains(t, s, "gcc")
	assert.Contains(t, s, "make")
	assert.Contains(t, s, "[OK]")
	assert.Contains(t, s, "[MISSING]")
	assert.Contains(t, s, "mingw32-make")
	assert.Contains(t, s, "2")
	assert.Contains(t, s, "1")
}

func TestCompatReport_MarshalJSON(t *testing.T) {
	results := []*CompatResult{
		{
			Command:     "gcc",
			IsAvailable: true,
			Description: "GNU C Compiler / GNU C 编译器",
			Platforms:   []string{"linux", "darwin", "windows"},
		},
		{
			Command:     "make",
			IsAvailable: false,
			Description: "build automation tool / 构建自动化工具",
			Alternative: "mingw32-make",
			InstallURL:  "https://www.gnu.org/software/make/",
			Platforms:   []string{"linux", "darwin", "windows"},
		},
	}

	report := NewCompatReport("linux", results)
	data, err := report.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "target_os")
	assert.Contains(t, jsonStr, "linux")
	assert.Contains(t, jsonStr, "is_available")
	assert.Contains(t, jsonStr, "true")
	assert.Contains(t, jsonStr, "description")
	assert.Contains(t, jsonStr, "install_url")

	// Verify unmarshal works
	var report2 CompatReport
	err = report2.UnmarshalJSON(data)
	require.NoError(t, err)
	assert.Equal(t, report.Total, report2.Total)
	assert.Equal(t, report.Available, report2.Available)
	assert.Equal(t, report.Missing, report2.Missing)
}

// =============================================================================
// Phase 4: CommandMap Description Tests
// =============================================================================

func TestCommandMap_DescriptionsFilled(t *testing.T) {
	for _, cm := range BuiltinCommandMaps {
		assert.NotEmpty(t, cm.Description,
			"Command %s should have a description / 命令 %s 应该有描述", cm.UnixCmd, cm.UnixCmd)
	}
}

func TestCommandMap_WindowsAlternativesFilled(t *testing.T) {
	for _, cm := range BuiltinCommandMaps {
		assert.NotEmpty(t, cm.WindowsCmd,
			"Command %s should have a Windows alternative / 命令 %s 应该有 Windows 替代", cm.UnixCmd, cm.UnixCmd)
	}
}

func TestCommandMap_InstallInfoForBuildTools(t *testing.T) {
	// Build tools like make, gcc, docker should have install URLs
	buildToolsWithInstallURL := []string{"make", "gcc", "docker", "go", "cmake", "npm"}

	for _, cmd := range buildToolsWithInstallURL {
		cm := LookupCommandMap(cmd)
		require.NotNil(t, cm, "Command %s should exist in BuiltinCommandMaps", cmd)
		assert.NotEmpty(t, cm.InstallURL,
			"Command %s should have InstallURL / 命令 %s 应该有安装链接", cmd, cmd)
	}
}

func TestCompatResult_DescriptionPopulated(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	result, err := cc.CheckCommand("gcc")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Description should be populated from CommandMap
	assert.NotEmpty(t, result.Description, "gcc should have description from CommandMap")
}

func TestCompatResult_IsShellBuiltin(t *testing.T) {
	cc := NewCompatChecker()
	err := cc.Init()
	require.NoError(t, err)

	result, err := cc.CheckCommand("echo")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsShellBuiltin, "echo should be detected as shell built-in")

	// gcc should not be a shell built-in
	result2, err := cc.CheckCommand("gcc")
	require.NoError(t, err)
	require.NotNil(t, result2)
	assert.False(t, result2.IsShellBuiltin, "gcc should not be a shell built-in")
}
