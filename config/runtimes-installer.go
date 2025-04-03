package config

import (
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// InstallRuntimes installs all runtimes defined in the configuration
func InstallRuntimes() error {
	for name, runtimeInfo := range Config.Runtimes() {
		err := InstallRuntime(name, runtimeInfo)
		if err != nil {
			return fmt.Errorf("failed to install runtime %s: %w", name, err)
		}
	}
	return nil
}

// InstallRuntime installs a specific runtime
func InstallRuntime(name string, runtimeInfo *plugins.RuntimeInfo) error {
	// Check if the runtime is already installed
	if isRuntimeInstalled(runtimeInfo) {
		fmt.Printf("Runtime %s v%s is already installed\n", name, runtimeInfo.Version)
		return nil
	}

	// Download and extract the runtime
	err := downloadAndExtractRuntime(runtimeInfo)
	if err != nil {
		return fmt.Errorf("failed to download and extract runtime %s: %w", name, err)
	}

	return nil
}

// isRuntimeInstalled checks if a runtime is already installed by checking for the binary
func isRuntimeInstalled(runtimeInfo *plugins.RuntimeInfo) bool {
	// If there are no binaries, check the install directory
	if len(runtimeInfo.Binaries) == 0 {
		_, err := os.Stat(runtimeInfo.InstallDir)
		return err == nil
	}

	// Check if at least one binary exists
	for _, binaryPath := range runtimeInfo.Binaries {
		_, err := os.Stat(binaryPath)
		if err == nil {
			return true
		}
	}

	return false
}

// downloadAndExtractRuntime downloads and extracts a runtime
func downloadAndExtractRuntime(runtimeInfo *plugins.RuntimeInfo) error {
	// Create a file name for the downloaded archive
	fileName := filepath.Base(runtimeInfo.DownloadURL)
	downloadPath := filepath.Join(Config.RuntimesDirectory(), fileName)

	// Check if the file already exists
	_, err := os.Stat(downloadPath)
	if os.IsNotExist(err) {
		// Download the file
		// log.Printf("Downloading %s v%s...\n", runtimeInfo.Name, runtimeInfo.Version)
		downloadPath, err = utils.DownloadFile(runtimeInfo.DownloadURL, Config.RuntimesDirectory())
		if err != nil {
			return fmt.Errorf("failed to download runtime: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for existing download: %w", err)
	} else {
		log.Printf("Using existing download for %s v%s\n", runtimeInfo.Name, runtimeInfo.Version)
	}

	// Open the downloaded file
	file, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer file.Close()

	// Extract based on file extension
	// log.Printf("Extracting %s v%s...\n", runtimeInfo.Name, runtimeInfo.Version)
	if strings.HasSuffix(fileName, ".zip") {
		err = utils.ExtractZip(file.Name(), Config.RuntimesDirectory())
	} else {
		err = utils.ExtractTarGz(file, Config.RuntimesDirectory())
	}

	if err != nil {
		return fmt.Errorf("failed to extract runtime: %w", err)
	}

	log.Printf("Successfully installed %s v%s\n", runtimeInfo.Name, runtimeInfo.Version)
	return nil
}
