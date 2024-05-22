package config

import (
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"path"
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
		"installDir": path.Join(Config.RuntimesDirectory(), nodeFileName),
		"node": path.Join(Config.RuntimesDirectory(), nodeFileName, "bin", "node"),
		"npm": path.Join(Config.RuntimesDirectory(), nodeFileName, "bin", "npm"),
	}
}

func getNodeDownloadURL(nodeRuntime *Runtime) string {
	// Detect the OS and architecture
	goos := runtime.GOOS

	// Construct the Node.js download URL
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/%s.%s", nodeRuntime.Version(), getNodeFileName(nodeRuntime), extension)
	return downloadURL
}

func InstallNode(r *Runtime) error {
	// TODO should delete downloaded archive
	// TODO check for deflated archive
	log.Println("Fetching node...")
	downloadNodeURL := getNodeDownloadURL(r)
	nodeTar, err := utils.DownloadFile(downloadNodeURL, Config.RuntimesDirectory())
	if err != nil {
		return err
	}

	// deflate node archive
	t, err := os.Open(nodeTar)
	defer t.Close()
	if err != nil {
		return err
	}
	err = utils.ExtractTarGz(t, Config.RuntimesDirectory())
	if err != nil {
		return err
	}

	return nil
}

