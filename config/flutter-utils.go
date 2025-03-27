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

func genInfoFlutter(r *Runtime) map[string]string {
	flutterFolder := "flutter"
	installDir := path.Join(Config.RuntimesDirectory(), flutterFolder)

	return map[string]string{
		"installDir": installDir,
		"flutter":    path.Join(installDir, "bin", "dart"),
	}
}

func InstallFlutter(flutterRuntime *Runtime) error {

	log.Println("Fetching flutter...")
	downloadFlutterURL := getFlutterDownloadURL(flutterRuntime)
	// Extract filename from URL
	fileName := filepath.Base(downloadFlutterURL)
	localPath := filepath.Join(Config.RuntimesDirectory(), fileName)

	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		log.Printf("File %s already exists, skipping download", fileName)
		return nil
	}
	dartTar, err := utils.DownloadFile(downloadFlutterURL, Config.RuntimesDirectory())
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

func InstallFlutterDartAnalyzer(flutterRuntime *Runtime, dartAnalyzer *Runtime, registry string) error {
	log.Println("Installing Dart Analyzer")
	cmd := exec.Command(flutterRuntime.Info()["flutter"], "pub", "add", "flutter_lints", "--dev")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Error installing Dart Analyzer:", err)
		fmt.Println(string(stdout))
	}
	// Print the output
	fmt.Println(string(stdout))
	return err
}

func getFlutterDownloadURL(flutterRuntime *Runtime) string {
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

	downloadURL := fmt.Sprintf("https://storage.googleapis.com/flutter_infra_release/releases/stable/%s/flutter_%s_%s_%s-stable.zip", os, os, arch, flutterRuntime.Version())
	fmt.Println("Downloading Flutter from:", downloadURL)
	return downloadURL
}
