package config

import (
	"codacy/cli-v2/constants"
	"codacy/cli-v2/plugins"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRuntimeInstalled(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "runtime-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock RuntimeInfo with no binaries
	runtimeInfoNoBinaries := &plugins.RuntimeInfo{
		Name:       "test-runtime",
		Version:    "1.0.0",
		InstallDir: filepath.Join(tempDir, "test-runtime-1.0.0"),
	}

	// Test when the install directory doesn't exist
	assert.False(t, Config.IsRuntimeInstalled("test-runtime", runtimeInfoNoBinaries))

	// Create the install directory
	err = os.MkdirAll(runtimeInfoNoBinaries.InstallDir, constants.DefaultDirPerms)
	assert.NoError(t, err)

	// Test when the install directory exists
	assert.True(t, Config.IsRuntimeInstalled("test-runtime", runtimeInfoNoBinaries))

	// Create a mock RuntimeInfo with binaries
	binPath := filepath.Join(tempDir, "test-runtime-bin")
	runtimeInfoWithBinaries := &plugins.RuntimeInfo{
		Name:       "test-runtime",
		Version:    "1.0.0",
		InstallDir: filepath.Join(tempDir, "test-runtime-1.0.0"),
		Binaries: map[string]string{
			"test-bin": binPath,
		},
	}

	// Test when the binary doesn't exist
	assert.False(t, Config.IsRuntimeInstalled("test-runtime", runtimeInfoWithBinaries))

	// Create a mock binary file
	_, err = os.Create(binPath)
	assert.NoError(t, err)

	// Test when the binary exists
	assert.True(t, Config.IsRuntimeInstalled("test-runtime", runtimeInfoWithBinaries))
}
