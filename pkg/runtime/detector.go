package runtime

import (
	"os"
	"os/exec"
	"runtime"
)

// SystemInfo holds information about the current system environment.
type SystemInfo struct {
	OS                string
	Arch              string
	Shell             string
	GoVersion         string
	HomeDir           string
	AvailableCommands map[string]string // cmd -> path
}

// Detector detects system information and available commands.
type Detector struct{}

// NewDetector creates a new system detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect gathers system information including OS, architecture, shell, and available commands.
func (d *Detector) Detect() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		Shell:             detectShell(),
		GoVersion:         runtime.Version(),
		HomeDir:           detectHomeDir(),
		AvailableCommands: make(map[string]string),
	}

	// Detect available commands from the mapping database
	for _, cm := range BuiltinCommandMaps {
		// Try the unix command first
		if path, err := exec.LookPath(cm.UnixCmd); err == nil {
			info.AvailableCommands[cm.UnixCmd] = path
		}
		// Also check Windows-specific commands
		if cm.WindowsCmd != "" && cm.WindowsCmd != cm.UnixCmd {
			// Strip quotes and args for LookPath check
			winCmd := extractCommandName(cm.WindowsCmd)
			if winCmd != "" && winCmd != cm.UnixCmd {
				if path, err := exec.LookPath(winCmd); err == nil {
					info.AvailableCommands[cm.UnixCmd] = path
					if _, exists := info.AvailableCommands[cm.UnixCmd]; !exists {
						info.AvailableCommands[cm.UnixCmd] = path
					}
				}
			}
		}
	}

	return info, nil
}

// IsCommandAvailable checks if a specific command is available on the system.
// Returns (true, path) if found, or (false, "") if not.
func (d *Detector) IsCommandAvailable(cmd string) (bool, string) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return false, ""
	}
	return true, path
}

// detectShell attempts to detect the user's current shell.
func detectShell() string {
	// Check common shell environment variables
	for _, env := range []string{"SHELL", "COMSPEC"} {
		if shell := os.Getenv(env); shell != "" {
			return shell
		}
	}

	// Fall back to platform default
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	return "/bin/sh"
}

// detectHomeDir returns the user's home directory.
func detectHomeDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return ""
}

// IsWindows returns true if the current OS is Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// extractCommandName strips arguments and quoting from a command string,
// returning just the executable name suitable for LookPath.
func extractCommandName(cmdStr string) string {
	// Handle simple cases: the first token is the command name
	s := cmdStr

	// Skip leading whitespace
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}

	// Split on space to get the first token
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			s = s[:i]
			break
		}
	}

	return s
}
