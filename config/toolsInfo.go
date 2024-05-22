package config

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

func GetToolRunInfo(name string) (map[string]string, error) {
	// Let's assume that for each tool we can get a map[string]string
	// with the tools parameters needed to run it
	switch name {
	case "eslint":
		m := make(map[string]string)

		node := Config.runtimes["node"]
		eslint := node.GetTool("eslint")
		eslintFolder := fmt.Sprintf("%s@%s", eslint.name, eslint.version)


		m["eslintInstallationDirectory"] = filepath.Join(Config.ToolsDirectory(), eslintFolder)
		m["nodeBinary"] = filepath.Join(Config.RuntimesDirectory(), GetNodeFileName(node), "bin", "node")

		return m, nil
	default:
		log.Fatal("eslint is the only supported tool")
		// This never gets called
		// TODO return an error type instead of nil
		return nil, nil
	}

}

func GetNodeFileName(nodeRuntime *Runtime) string {
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

func GetNodeDownloadURL(nodeRuntime *Runtime) string {
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