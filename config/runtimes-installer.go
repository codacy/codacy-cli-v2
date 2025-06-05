package config

import (
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// InstallRuntimes installs all runtimes defined in the configuration
func InstallRuntimes(config *ConfigType) error {
	var failedRuntimes []string

	for name, runtimeInfo := range config.Runtimes() {
		logger.Info("Starting runtime installation", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
		})

		err := InstallRuntime(name, runtimeInfo)
		if err != nil {
			logger.Error("Failed to install runtime", logrus.Fields{
				"runtime": name,
				"version": runtimeInfo.Version,
				"error":   err.Error(),
			})
			failedRuntimes = append(failedRuntimes, name)
			continue
		}

		logger.Info("Successfully installed runtime", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
		})
	}

	if len(failedRuntimes) > 0 {
		return fmt.Errorf("failed to install the following runtimes: %v", failedRuntimes)
	}
	return nil
}

// InstallRuntime installs a specific runtime
func InstallRuntime(name string, runtimeInfo *plugins.RuntimeInfo) error {
	// Skip if already installed
	fmt.Printf("Checking if runtime %s v%s is installed\n", name, runtimeInfo.Version)
	if Config.IsRuntimeInstalled(name, runtimeInfo) {
		logger.Info("Runtime already installed", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
		})
		fmt.Printf("Runtime %s v%s is already installed\n", name, runtimeInfo.Version)
		return nil
	}

	// Download and extract the runtime
	err := downloadAndExtractRuntime(runtimeInfo)
	if err != nil {
		return fmt.Errorf("failed to download and extract runtime %s: %w", name, err)
	}

	// Verify that the runtime binaries are available
	if !Config.IsRuntimeInstalled(name, runtimeInfo) {
		logger.Error("Runtime binaries not found after extraction", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
		})
		return fmt.Errorf("runtime %s was extracted but binaries are not available", name)
	}

	return nil
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
		logger.Debug("Downloading runtime", logrus.Fields{
			"runtime":      runtimeInfo.Name,
			"version":      runtimeInfo.Version,
			"downloadURL":  runtimeInfo.DownloadURL,
			"downloadPath": downloadPath,
		})
		downloadPath, err = utils.DownloadFile(runtimeInfo.DownloadURL, Config.RuntimesDirectory())
		if err != nil {
			return fmt.Errorf("failed to download runtime: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking for existing download: %w", err)
	} else {
		logger.Debug("Using existing runtime download", logrus.Fields{
			"runtime":      runtimeInfo.Name,
			"version":      runtimeInfo.Version,
			"downloadPath": downloadPath,
		})
	}

	// Open the downloaded file
	file, err := os.Open(downloadPath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer file.Close()

	// Extract based on file extension
	logger.Debug("Extracting runtime", logrus.Fields{
		"runtime":          runtimeInfo.Name,
		"version":          runtimeInfo.Version,
		"fileName":         fileName,
		"extractDirectory": Config.RuntimesDirectory(),
	})

	if strings.HasSuffix(fileName, ".zip") {
		err = utils.ExtractZip(file.Name(), Config.RuntimesDirectory())
	} else {
		err = utils.ExtractTarGz(file, Config.RuntimesDirectory())
	}

	if err != nil {
		return fmt.Errorf("failed to extract runtime: %w", err)
	}

	// Ensure binaries have executable permissions
	for binaryName, fullPath := range runtimeInfo.Binaries {
		if err := os.Chmod(fullPath, utils.DefaultDirPerms); err != nil {
			logger.Debug("Failed to set binary permissions", logrus.Fields{
				"binary": binaryName,
				"path":   fullPath,
				"error":  err,
			})
			return fmt.Errorf("failed to set binary permissions for %s: %w", fullPath, err)
		}
	}

	logger.Debug("Runtime extraction completed", logrus.Fields{
		"runtime": runtimeInfo.Name,
		"version": runtimeInfo.Version,
	})
	return nil
}
