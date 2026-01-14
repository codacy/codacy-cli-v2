package plugins

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessRuntimes(t *testing.T) {
	// Create a list of runtime configs for testing
	configs := []RuntimeConfig{
		{
			Name:    "node",
			Version: "18.17.1",
		},
		{
			Name:    "flutter",
			Version: "3.35.7",
		},
	}

	// Define a test runtime directory
	runtimesDir := "/test/runtimes"

	// Process the runtimes
	runtimeInfos, err := ProcessRuntimes(configs, runtimesDir)

	// Assert no errors occurred
	assert.NoError(t, err, "ProcessRuntimes should not return an error")

	// Assert we have the expected runtime in the results
	assert.Contains(t, runtimeInfos, "node")
	assert.Contains(t, runtimeInfos, "flutter")

	// Get the node runtime info
	nodeInfo := runtimeInfos["node"]
	flutterInfo := runtimeInfos["flutter"]

	// Basic assertions for flutter
	assert.Equal(t, "flutter", flutterInfo.Name)
	assert.Equal(t, "3.35.7", flutterInfo.Version)

	// Assert the basic runtime info is correct
	assert.Equal(t, "node", nodeInfo.Name)
	assert.Equal(t, "18.17.1", nodeInfo.Version)

	// Get the expected architecture for node
	var expectedArch string
	switch runtime.GOARCH {
	case "386":
		expectedArch = "x86"
	case "amd64":
		expectedArch = "x64"
	case "arm":
		expectedArch = "armv7l"
	case "arm64":
		expectedArch = "arm64"
	default:
		expectedArch = runtime.GOARCH
	}

	// Get the expected extension
	expectedExtension := "tar.gz"
	if runtime.GOOS == "windows" {
		expectedExtension = "zip"
	}

	// Assert the filename is correctly formatted
	expectedFileName := "node-v18.17.1-" + runtime.GOOS + "-" + expectedArch
	assert.Equal(t, expectedFileName, nodeInfo.FileName)

	// Assert the extension is correct
	assert.Equal(t, expectedExtension, nodeInfo.Extension)

	// Assert the install directory is correct
	assert.Equal(t, runtimesDir+"/"+expectedFileName, nodeInfo.InstallDir)

	// Assert the download URL is correctly formatted
	expectedDownloadURL := "https://nodejs.org/dist/v18.17.1/" + expectedFileName + "." + expectedExtension
	assert.Equal(t, expectedDownloadURL, nodeInfo.DownloadURL)

	// Assert binary paths are correctly set
	assert.NotNil(t, nodeInfo.Binaries)
	assert.Greater(t, len(nodeInfo.Binaries), 0)

	// Check if node and npm binaries are present
	nodeBinary := nodeInfo.InstallDir + "/bin/node"
	npmBinary := nodeInfo.InstallDir + "/bin/npm"

	// Add .exe extension for Windows
	if runtime.GOOS == "windows" {
		nodeBinary += ".exe"
		npmBinary += ".exe"
	}

	assert.Equal(t, nodeBinary, nodeInfo.Binaries["node"])
	assert.Equal(t, npmBinary, nodeInfo.Binaries["npm"])
}
