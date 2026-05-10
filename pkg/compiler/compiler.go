package compiler

// GoSDKStatus describes the state of the Go SDK on the system.
type GoSDKStatus struct {
	Installed bool
	Cached    bool
	Path      string
	Version   string
}

// Compiler is the interface for compiling magefiles to binaries.
type Compiler interface {
	Native(magefile string, output string) error
	Cross(magefile string, targetOS, targetArch, output string) error
	Bootstrap() error
	Detect() (*GoSDKStatus, error)
}
