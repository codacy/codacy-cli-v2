package config

import (
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

func getInfoPython(r *Runtime) map[string]string {
	pythonFolder := fmt.Sprintf("%s@%s", r.Name(), r.Version())
	installDir := path.Join(Config.RuntimesDirectory(), pythonFolder)

	var pythonBinary, pipBinary string

	//todo check windows dire,
	//had to add python subdir to path since tar extracts it there
	if runtime.GOOS == "windows" {
		pythonBinary = path.Join(installDir, "Scripts", "python.exe")
		pipBinary = path.Join(installDir, "Scripts", "pip.exe")
	} else {
		pythonBinary = path.Join(installDir, "python", "bin", "python3")
		pipBinary = path.Join(installDir, "python", "bin", "pip")
	}

	return map[string]string{
		"installDir": installDir,
		"python":     pythonBinary,
		"pip":        pipBinary,
	}
}

func getDownloadURL(pythonRuntime *Runtime) string {

	version := pythonRuntime.Version()
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var pyArch string
	switch goarch {
	case "386":
		pyArch = "x86"
	case "amd64":
		pyArch = "x86_64"
	case "arm":
		pyArch = "armv7l"
	case "arm64":
		pyArch = "aarch64"
	default:
		pyArch = goarch
	}

	var pyOS string
	switch goos {
	case "darwin":
		pyOS = "apple-darwin"
	case "linux":
		pyOS = "unknown-linux-gnu"
	case "windows":
		pyOS = "pc-windows-msvc"
	default:
		pyOS = goos
	}

	releaseVersion := "20250317"
	baseURL := "https://github.com/astral-sh/python-build-standalone/releases/download/"

	filename := fmt.Sprintf("cpython-%s+%s-%s-%s-install_only.tar.gz", version, releaseVersion, pyArch, pyOS)

	return fmt.Sprintf("%s%s/%s", baseURL, releaseVersion, filename)

}

func InstallPython(r *Runtime) error {

	pythonFolder := fmt.Sprintf("%s@%s", r.Name(), r.Version())
	installDir := path.Join(Config.RuntimesDirectory(), pythonFolder)
	log.Println("Fetching python...")
	downloadPythonURL := getDownloadURL(r)
	pythonTar, err := utils.DownloadFile(downloadPythonURL, Config.RuntimesDirectory())
	if err != nil {
		return err
	}

	// Make sure the installDir exists
	err = os.MkdirAll(installDir, 0777)
	if err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// Open the downloaded file
	t, err := os.Open(pythonTar)
	defer t.Close()
	if err != nil {
		return err
	}

	// Extract the archive to the desired directory without creating links yet
	err = utils.ExtractTarGz(t, installDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %v", err)
	}

	//remove tar after extraction
	err = os.Remove(pythonTar)
	if err != nil {
		return fmt.Errorf("failed to delete downloaded archive: %v", err)
	}

	log.Println("Python successfully installed at:", installDir)
	return nil
}
