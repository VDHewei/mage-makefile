package transformer

import (
	"os/exec"
	"runtime"
	"strings"
)

// execShellCommand executes a shell command and returns its trimmed stdout.
// On error, returns empty string (matching GNU Make behavior).
func execShellCommand(command string) string {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("/bin/sh", "-c", command)
	}
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(out), "\n\r\t ")
}
