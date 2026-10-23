package plugins

import (
	"path"
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

	flutterExpectedExtension := "zip"
	if runtime.GOOS == "linux" {
		flutterExpectedExtension = "tar.xz"
	}

	// Assert flutter extension
	assert.Equal(t, flutterExpectedExtension, flutterInfo.Extension)

	// Additional flutter assertions
	// Assert the filename is correctly set to a constant "flutter"
	assert.Equal(t, "flutter", flutterInfo.FileName)

	// Assert the install directory is correct for flutter
	assert.Equal(t, runtimesDir+"/"+flutterInfo.FileName, flutterInfo.InstallDir)

	// Compute expected OS mapping for flutter download URL
	var expectedFlutterOS string
	switch runtime.GOOS {
	case "darwin":
		expectedFlutterOS = "macos"
	case "linux":
		expectedFlutterOS = "linux"
	case "windows":
		expectedFlutterOS = "windows"
	default:
		expectedFlutterOS = runtime.GOOS
	}

	// Compute expected arch for flutter (only used on macOS/default template)
	var expectedFlutterArch string
	switch runtime.GOARCH {
	case "386":
		expectedFlutterArch = "ia32"
	case "amd64":
		expectedFlutterArch = "x64"
	case "arm":
		expectedFlutterArch = "arm"
	case "arm64":
		expectedFlutterArch = "arm64"
	default:
		expectedFlutterArch = runtime.GOARCH
	}

	// Build expected flutter download URL
	var expectedFlutterURL string
	if runtime.GOOS == "linux" {
		expectedFlutterURL = "https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_" + flutterInfo.Version + "-stable." + flutterExpectedExtension
	} else if runtime.GOOS == "windows" {
		expectedFlutterURL = "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_" + flutterInfo.Version + "-stable." + flutterExpectedExtension
	} else {
		// Default template includes arch and mapped OS (e.g., macos)
		expectedFlutterURL = "https://storage.googleapis.com/flutter_infra_release/releases/stable/" + expectedFlutterOS + "/flutter_" + expectedFlutterOS + "_" + expectedFlutterArch + "_" + flutterInfo.Version + "-stable." + flutterExpectedExtension
	}

	assert.Equal(t, expectedFlutterURL, flutterInfo.DownloadURL)

	// Assert flutter binaries map has expected entries
	assert.NotNil(t, flutterInfo.Binaries)
	assert.Greater(t, len(flutterInfo.Binaries), 0)

	// Check if dart binary is present and correctly mapped
	flutterDartBinary := flutterInfo.InstallDir + "/bin/dart"
	if runtime.GOOS == "windows" {
		flutterDartBinary += ".exe"
	}
	assert.Equal(t, flutterDartBinary, flutterInfo.Binaries["dart"])

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

func TestProcessRuntimesVersionsPythonAndGoInstallDir(t *testing.T) {
	runtimesDir := "/test/runtimes"

	for _, name := range []string{"python", "go"} {
		oldInfo, err := processRuntime(RuntimeConfig{Name: name, Version: "1.0.0"}, runtimesDir)
		assert.NoError(t, err)
		newInfo, err := processRuntime(RuntimeConfig{Name: name, Version: "2.0.0"}, runtimesDir)
		assert.NoError(t, err)

		assert.Equal(t, runtimesDir+"/"+newInfo.FileName+"/"+name, newInfo.InstallDir)
		assert.NotEqual(t, oldInfo.InstallDir, newInfo.InstallDir)
		// The installer downloads the archive next to the extract dir; same path breaks extraction
		assert.NotEqual(t, runtimesDir+"/"+path.Base(newInfo.DownloadURL), path.Dir(newInfo.InstallDir))
	}
}

func TestProcessRuntimesPythonReleaseVersionOverride(t *testing.T) {
	pinned, err := processRuntime(RuntimeConfig{Name: "python", Version: "3.11.11"}, "/test/runtimes")
	assert.NoError(t, err)
	assert.Contains(t, pinned.DownloadURL, "/20250317/cpython-3.11.11+20250317-")

	current, err := processRuntime(RuntimeConfig{Name: "python", Version: "3.11.17"}, "/test/runtimes")
	assert.NoError(t, err)
	assert.Contains(t, current.DownloadURL, "/20261021/cpython-3.11.17+20261021-")
	// A release-only bump must change the install dir too, or it never re-downloads
	assert.Contains(t, current.InstallDir, "/cpython-3.11.17+20261021-")
}

func TestGoDownloadURLUsesWindowsZip(t *testing.T) {
	cfg, err := GetPluginManager().GetRuntimeConfig("go")
	assert.NoError(t, err)
	url := GetDownloadURL(cfg.Download.URLTemplate, "", "1.26.10", "amd64", "windows", GetExtension(cfg.Download.Extension, "windows"), "")
	assert.Equal(t, "https://go.dev/dl/go1.26.10.windows-amd64.zip", url)
}
