package runtime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Installer Tests
// =============================================================================

func TestNewInstaller(t *testing.T) {
	inst := NewInstaller(false)
	require.NotNil(t, inst)
	assert.False(t, inst.interactive)

	inst2 := NewInstaller(true)
	require.NotNil(t, inst2)
	assert.True(t, inst2.interactive)
}

func TestInstaller_SuggestInstall_Existing(t *testing.T) {
	inst := NewInstaller(false)

	// Go should be available on the system (it's running this test)
	suggestion := inst.SuggestInstall("go")
	assert.Nil(t, suggestion, "existing command should return nil suggestion")
}

func TestInstaller_SuggestInstall_Missing(t *testing.T) {
	inst := NewInstaller(false)

	// Use a command that is unlikely to exist
	suggestion := inst.SuggestInstall("nonexistent_command_xyz_123")
	assert.Nil(t, suggestion, "unknown command should return nil suggestion")
}

func TestInstaller_SuggestInstall_WithURL(t *testing.T) {
	inst := NewInstaller(false)

	// make should have install info
	suggestion := inst.SuggestInstall("make")
	// make might be installed, so we check either way
	if suggestion != nil {
		assert.Equal(t, "make", suggestion.Command)
		assert.NotEmpty(t, suggestion.Description)

		// Should have at least a URL method
		hasURL := false
		for _, m := range suggestion.Methods {
			if m.Type == "url" {
				hasURL = true
				break
			}
		}
		assert.True(t, hasURL, "make should have URL installation method")
	}
}

func TestInstaller_SuggestInstall_WithScripts(t *testing.T) {
	inst := NewInstaller(false)

	// make has InstallCmds with sh scripts
	suggestion := inst.SuggestInstall("make")
	if suggestion != nil {
		// Should have some install methods
		assert.NotEmpty(t, suggestion.Methods)

		// Check if any methods have Type "sh" or "ps1"
		hasNonURL := false
		for _, m := range suggestion.Methods {
			if m.Type != "url" {
				hasNonURL = true
				break
			}
		}
		assert.True(t, hasNonURL, "make should have non-URL install methods")
	}
}

func TestInstallSuggestion_String(t *testing.T) {
	suggestion := &InstallSuggestion{
		Command:     "make",
		Description: "build automation tool / 构建自动化工具",
		Methods: []InstallMethod{
			{Type: "url", Content: "https://www.gnu.org/software/make/", Label: "Download GNU Make"},
			{Type: "ps1", Content: "choco install make", Label: "Install via Chocolatey"},
		},
	}

	s := suggestion.String()
	assert.Contains(t, s, "make")
	assert.Contains(t, s, "build automation tool")
	assert.Contains(t, s, "Download GNU Make")
	assert.Contains(t, s, "Install via Chocolatey")
	assert.Contains(t, s, "[1]")
	assert.Contains(t, s, "[2]")
}

func TestInstaller_SuggestInstall_KnownMissing(t *testing.T) {
	// Test that SuggestInstall handles a command in the map that isn't installed
	// We can't guarantee any command is missing, but we can check the logic
	inst := NewInstaller(false)

	// Pick a command that should have install info
	for _, cm := range BuiltinCommandMaps {
		if len(cm.InstallCmds) > 0 {
			suggestion := inst.SuggestInstall(cm.UnixCmd)
			if suggestion != nil {
				assert.NotEmpty(t, suggestion.Methods)
				assert.Equal(t, cm.UnixCmd, suggestion.Command)
			}
			// Only test one command with InstallCmds
			return
		}
	}
}

func TestAutoInstall_SkipAvailable(t *testing.T) {
	inst := NewInstaller(false)

	// For a non-existent command, AutoInstall should skip gracefully
	suggestion := inst.SuggestInstall("nonexistent_xyz")
	if suggestion == nil {
		// Already available (or unknown) - should be fine
		return
	}

	ok, err := inst.AutoInstall(suggestion)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestInstallSuggestion_Format(t *testing.T) {
	// Test the String() output format
	s := (&InstallSuggestion{
		Command:     "make",
		Description: "build automation tool",
		Methods: []InstallMethod{
			{Type: "url", Content: "https://example.com", Label: "Download"},
		},
	}).String()

	assert.True(t, strings.HasPrefix(s, "Missing command"))
	assert.True(t, strings.Contains(s, "Download"))
}
