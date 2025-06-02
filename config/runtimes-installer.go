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
	if Config.IsRuntimeInstalled(name, runtimeInfo) {
		logger.Info("Runtime already installed", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
		})
		fmt.Printf("Runtime %s v%s is already installed\n", name, runtimeInfo.Version)
		return nil
	}

	// Install using the new implementation
	err := installRuntimeNew(name, runtimeInfo)
	if err != nil {
		logger.Error("Failed to install runtime", logrus.Fields{
			"runtime": name,
			"version": runtimeInfo.Version,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to install runtime %s: %w", name, err)
	}

	return nil
}

// installRuntimeWithRetry attempts to install a runtime with proper path handling
func installRuntimeWithRetry(name string, runtimeInfo *plugins.RuntimeInfo) error {
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

	// Ensure binaries have executable permissions with correct path handling
	for binaryName, binaryPath := range runtimeInfo.Binaries {
		// Use the full binary path directly from runtimeInfo
		if err := os.Chmod(binaryPath, utils.DefaultDirPerms); err != nil {
			if !os.IsNotExist(err) {
				logger.Debug("Failed to set binary permissions", logrus.Fields{
					"binary": binaryName,
					"path":   binaryPath,
					"error":  err,
				})
				return fmt.Errorf("failed to set binary permissions for %s: %w", binaryName, err)
			}
			// If file doesn't exist, continue to next binary
			logger.Debug("Binary not found, skipping", logrus.Fields{
				"binary": binaryName,
				"path":   binaryPath,
			})
		}
	}

	logger.Debug("Runtime extraction completed", logrus.Fields{
		"runtime": runtimeInfo.Name,
		"version": runtimeInfo.Version,
	})
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
	for _, binaryPath := range runtimeInfo.Binaries {
		fullPath := filepath.Join(Config.RuntimesDirectory(), filepath.Base(runtimeInfo.InstallDir), binaryPath)
		if err := os.Chmod(fullPath, utils.DefaultDirPerms); err != nil {
			logger.Debug("Failed to set binary permissions", logrus.Fields{
				"binary": binaryPath,
				"path":   fullPath,
				"error":  err,
			})
			return fmt.Errorf("failed to set binary permissions for %s: %w", binaryPath, err)
		}
	}

	logger.Debug("Runtime extraction completed", logrus.Fields{
		"runtime": runtimeInfo.Name,
		"version": runtimeInfo.Version,
	})
	return nil
}

// installRuntimeNew installs a runtime using a fresh implementation
func installRuntimeNew(name string, runtimeInfo *plugins.RuntimeInfo) error {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(runtimeInfo.InstallDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Download the runtime archive
	logger.Info("Downloading runtime", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
		"url":     runtimeInfo.DownloadURL,
	})

	archivePath, err := utils.DownloadFile(runtimeInfo.DownloadURL, Config.RuntimesDirectory())
	if err != nil {
		return fmt.Errorf("failed to download runtime archive: %w", err)
	}

	// Extract the archive
	logger.Info("Extracting runtime", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
		"archive": archivePath,
	})

	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer archive.Close()

	// Extract based on archive type
	if strings.HasSuffix(archivePath, ".zip") {
		err = utils.ExtractZip(archivePath, Config.RuntimesDirectory())
	} else {
		err = utils.ExtractTarGz(archive, Config.RuntimesDirectory())
	}
	if err != nil {
		return fmt.Errorf("failed to extract runtime archive: %w", err)
	}

	// Set executable permissions on binaries
	logger.Info("Setting binary permissions", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
	})

	for binaryName, binaryPath := range runtimeInfo.Binaries {
		// Skip if binary doesn't exist yet
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			logger.Debug("Binary not found, skipping", logrus.Fields{
				"binary": binaryName,
				"path":   binaryPath,
			})
			continue
		}

		// Set executable permissions
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions for binary %s: %w", binaryName, err)
		}
	}

	// Verify installation
	if !Config.IsRuntimeInstalled(name, runtimeInfo) {
		return fmt.Errorf("runtime %s was installed but binaries are not available", name)
	}

	logger.Info("Runtime installation completed", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
	})

	return nil
}

// installRuntimeStrict installs a runtime with strict path handling and validation
func installRuntimeStrict(name string, runtimeInfo *plugins.RuntimeInfo) error {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(runtimeInfo.InstallDir, utils.DefaultDirPerms); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	// Download the runtime archive
	logger.Info("Downloading runtime", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
		"url":     runtimeInfo.DownloadURL,
	})

	archivePath, err := utils.DownloadFile(runtimeInfo.DownloadURL, Config.RuntimesDirectory())
	if err != nil {
		return fmt.Errorf("failed to download runtime archive: %w", err)
	}

	// Extract the archive
	logger.Info("Extracting runtime", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
		"archive": archivePath,
	})

	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer archive.Close()

	// Extract based on archive type
	if strings.HasSuffix(archivePath, ".zip") {
		err = utils.ExtractZip(archivePath, Config.RuntimesDirectory())
	} else {
		err = utils.ExtractTarGz(archive, Config.RuntimesDirectory())
	}
	if err != nil {
		return fmt.Errorf("failed to extract runtime archive: %w", err)
	}

	// Set executable permissions on binaries
	logger.Info("Setting binary permissions", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
	})

	for binaryName, binaryPath := range runtimeInfo.Binaries {
		// Skip if binary doesn't exist yet
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			logger.Debug("Binary not found, skipping", logrus.Fields{
				"binary": binaryName,
				"path":   binaryPath,
			})
			continue
		}

		// Set executable permissions
		if err := os.Chmod(binaryPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions for binary %s: %w", binaryName, err)
		}
	}

	// Verify installation
	if !Config.IsRuntimeInstalled(name, runtimeInfo) {
		return fmt.Errorf("runtime %s was installed but binaries are not available", name)
	}

	logger.Info("Runtime installation completed", logrus.Fields{
		"runtime": name,
		"version": runtimeInfo.Version,
	})

	return nil
}
