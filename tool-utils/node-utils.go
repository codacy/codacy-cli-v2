package tool_utils

import (
	cfg "codacy/cli-v2/config"
	"fmt"
	"runtime"
)

func GetNodeFileName(nodeRuntime *cfg.Runtime) string {
	// Detect the OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture to Node.js architecture
	var nodeArch string
	switch goarch {
	case "386":
		nodeArch = "x86"
	case "amd64":
		nodeArch = "x64"
	case "arm":
		nodeArch = "armv7l"
	case "arm64":
		nodeArch = "arm64"
	default:
		nodeArch = goarch
	}

	return fmt.Sprintf("node-v%s-%s-%s", nodeRuntime.Version(), goos, nodeArch)
}

func GetNodeDownloadURL(nodeRuntime *cfg.Runtime) string {
	// Detect the OS and architecture
	goos := runtime.GOOS

	// Construct the Node.js download URL
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/%s.%s", nodeRuntime.Version(), GetNodeFileName(nodeRuntime), extension)
	return downloadURL
}
