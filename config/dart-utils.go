package config

import (
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
)

func genInfoDart(r *Runtime) map[string]string {
	dartSdkFolder := "dart-sdk"
	installDir := path.Join(Config.RuntimesDirectory(), dartSdkFolder)

	return map[string]string{
		"installDir": installDir,
		"dart":       path.Join(installDir, "bin", "dart"),
	}
}

func InstallDartAnalyzer(dartRuntime *Runtime, dartAnalyzer *Runtime, registry string) error {
	log.Println("Installing Dart Analyzer")
	cmd := exec.Command(dartRuntime.Info()["dart"], "pub", "add", "--dev flutter_lints")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Error installing Dart Analyzer:", err)
		fmt.Println(string(stdout))
	}
	// Print the output
	fmt.Println(string(stdout))
	return err
}

func InstallDart(dartRuntime *Runtime) error {

	log.Println("Fetching dart...")
	downloadDartURL := getDartDownloadURL(dartRuntime)
	// Extract filename from URL
	fileName := filepath.Base(downloadDartURL)
	localPath := filepath.Join(Config.RuntimesDirectory(), fileName)

	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		log.Printf("File %s already exists, skipping download", fileName)
		return nil
	}
	dartTar, err := utils.DownloadFile(downloadDartURL, Config.RuntimesDirectory())
	if err != nil {
		return err
	}

	// deflate node archive
	t, err := os.Open(dartTar)
	if err != nil {
		return err
	}
	defer t.Close()
	err = utils.ExtractZip(t, Config.RuntimesDirectory())
	if err != nil {
		return err
	}

	return nil
}

func getDartDownloadURL(dartRuntime *Runtime) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture to Dart architecture
	var arch string
	switch goarch {
	case "386":
		arch = "ia32"
	case "amd64":
		arch = "x64"
	case "arm":
		arch = "arm"
	case "arm64":
		arch = "arm64"
	default:
		arch = goarch
	}

	var os string
	switch goos {
	case "darwin":
		os = "macos"
	case "linux":
		os = "linux"
	case "windows":
		os = "windows"
	default:
		os = goos
	}

	downloadURL := fmt.Sprintf("https://storage.googleapis.com/dart-archive/channels/stable/release/%s/sdk/dartsdk-%s-%s-release.zip", dartRuntime.Version(), os, arch)
	return downloadURL
}
