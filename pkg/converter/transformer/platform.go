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
		parts[0] = mapped
		return strings.Join(parts, " ")
	}

	return cmd
}

// unixToWindows maps common Unix commands to their Windows equivalents.
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
	"chown":  "", // no direct equivalent
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
	"nc":    "", // no direct equivalent

	// Text
	"grep":  "findstr",
	"sed":   "",   // no direct equivalent
	"awk":   "",   // no direct equivalent
	"sort":  "sort",
	"uniq":  "",   // no direct equivalent
	"wc":    "find /c",
	"head":  "",   // no direct equivalent
	"tail":  "",   // no direct equivalent
	"tr":    "",   // no direct equivalent
	"cut":   "",   // no direct equivalent

	// Shell builtins (approximations)
	"source": "call",
	"export": "set",
	"unset":  "set",

	// Package managers (approximations)
	"apt":   "",
	"apt-get": "",
	"yum":   "",
	"brew":  "",

	// Permissions
	"sudo":   "",
	"su":     "",
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
