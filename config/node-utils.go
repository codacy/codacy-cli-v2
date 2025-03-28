package config

import (
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func getNodeFileName(nodeRuntime *Runtime) string {
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

func genInfoNode(r *Runtime) map[string]string {
	nodeFileName := getNodeFileName(r)

	return map[string]string{
		"nodeFileName": nodeFileName,
		"installDir":   path.Join(Config.RuntimesDirectory(), nodeFileName),
		"node":         path.Join(Config.RuntimesDirectory(), nodeFileName, "bin", "node"),
		"npm":          path.Join(Config.RuntimesDirectory(), nodeFileName, "bin", "npm"),
	}
}

func getNodeDownloadURL(nodeRuntime *Runtime) string {
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

	// Map Go OS to Node.js OS name
	var nodeOS string
	switch goos {
	case "windows":
		nodeOS = "win"
	default:
		nodeOS = goos
	}

	// Construct the Node.js download URL
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	// Use a more reliable Node.js version if the requested one doesn't exist
	version := nodeRuntime.Version()

	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/node-v%s-%s-%s.%s", version, version, nodeOS, nodeArch, extension)
	return downloadURL
}

func InstallNode(r *Runtime) error {
	downloadNodeURL := getNodeDownloadURL(r)
	fileName := filepath.Base(downloadNodeURL)
	t, err := os.Open(filepath.Join(Config.RuntimesDirectory(), fileName))
	defer t.Close()
	if err != nil {
		log.Printf("Node is not present, fetching node from %s...\n", downloadNodeURL)
		nodeTar, err := utils.DownloadFile(downloadNodeURL, Config.RuntimesDirectory())
		if err != nil {
			return fmt.Errorf("failed to download Node.js: %w", err)
		}
		t, err = os.Open(nodeTar)
		defer t.Close()
		if err != nil {
			return fmt.Errorf("failed to open downloaded Node.js archive: %w", err)
		}
	} else {
		fmt.Println("Node is already present...")
	}
	fmt.Println("Extracting node...")
	// deflate node archive

	err = utils.ExtractTarGz(t, Config.RuntimesDirectory())
	if err != nil {
		return fmt.Errorf("failed to extract Node.js archive: %w", err)
	}
	return nil
}
