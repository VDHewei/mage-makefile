package transformer

import (
	"runtime"
	"strings"
)

// PlatformMapping manages command translations between Unix and Windows.
type PlatformMapping struct {
	TargetOS    string
	commandMap  map[string]string
}

// NewPlatformMapping creates a PlatformMapping for the given target OS.
// If targetOS is empty, the current runtime OS is used.
func NewPlatformMapping(targetOS string) *PlatformMapping {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	targetOS = strings.ToLower(targetOS)

	pm := &PlatformMapping{
		TargetOS: targetOS,
	}

	switch targetOS {
	case "windows":
		pm.commandMap = unixToWindows
	default:
		pm.commandMap = nil // identity mapping for Unix targets
	}

	return pm
}

// MapCommand translates a command to the target platform.
func (pm *PlatformMapping) MapCommand(cmd string) string {
	if pm.commandMap == nil {
		return cmd
	}

	trimmed := strings.TrimSpace(cmd)
	parts := shellSplit(trimmed)
	if len(parts) == 0 {
		return cmd
	}

	baseCmd := strings.ToLower(parts[0])
	if mapped, ok := pm.commandMap[baseCmd]; ok {
		if mapped == "" {
			// 无 Windows 等效命令，保留原命令（Windows 上会报错，但不会产生乱码）
			return cmd
		}
		parts[0] = mapped
		return strings.Join(parts, " ")
	}

	return cmd
}

// unixToWindows maps common Unix commands to their Windows equivalents.
// 命令映射：将 Unix 命令映射到 Windows 等效命令。
// 对于仅有词法差异的命令（rm→del, cp→copy）直接替换命令名。
// 对于参数语法完全不同的命令（sed, awk, head, tail 等）不在此处映射——
// 它们由 runtime.CompatChecker 在兼容性检测层处理，提供 PowerShell 替代方案。
var unixToWindows = map[string]string{
	// File operations
	"rm":     "del",
	"rmdir":  "rmdir",
	"cp":     "copy",
	"mv":     "move",
	"cat":    "type",
	"touch":  "copy nul",
	"mkdir":  "mkdir",
	"md":     "mkdir",
	"ln":     "mklink",
	"chmod":  "icacls",
	"which":  "where",
	"pwd":    "cd",
	"echo":   "echo",
	"true":   "cd .",
	"false":  "cd invalid_dir",

	// Process
	"ps":   "tasklist",
	"kill": "taskkill",
	"top":  "tasklist",

	// Network
	"wget":  "curl",
	"curl":  "curl",
	"ping":  "ping",

	// Text — commands with simple name-only mapping
	"grep":  "findstr",
	"sort":  "sort",
	"wc":    "find /c",

	// Shell builtins (approximations)
	"source": "call",
	"export": "set",
	"unset":  "set",
}

// AutoVarMap translates automatic Make variables to their descriptions.
var AutoVarMap = map[string]string{
	"$@": "target name",
	"$<": "first prerequisite",
	"$^": "all prerequisites (space-separated, no duplicates)",
	"$+": "all prerequisites (space-separated, with duplicates)",
	"$?": "prerequisites newer than target",
	"$*": "stem (pattern match)",
	"$%": "archive member name",
	"$(@D)": "directory part of target",
	"$(@F)": "file part of target",
	"$(<D)": "directory part of first prerequisite",
	"$(<F)": "file part of first prerequisite",
	"$(^D)": "directory parts of all prerequisites",
	"$(^F)": "file parts of all prerequisites",
}
