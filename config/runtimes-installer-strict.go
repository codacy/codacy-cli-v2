package config

import (
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/utils"
	"codacy/cli-v2/utils/logger"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// InstallRuntimeStrict installs a runtime with strict path handling and validation
func InstallRuntimeStrict(name string, runtimeInfo *plugins.RuntimeInfo) error {
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

	// Update codacy.yaml with the new runtime
	if err := updateRuntimeInCodacyYaml(name, runtimeInfo.Version); err != nil {
		return fmt.Errorf("failed to update codacy.yaml with runtime %s: %w", name, err)
	}

	return nil
}

// updateRuntimeInCodacyYaml adds or updates a runtime entry in .codacy/codacy.yaml as a YAML list
func updateRuntimeInCodacyYaml(name string, version string) error {
	codacyPath := ".codacy/codacy.yaml"

	type CodacyConfig struct {
		Runtimes []string `yaml:"runtimes"`
		Tools    []string `yaml:"tools"`
	}

	// Read existing config
	var config CodacyConfig
	if data, err := os.ReadFile(codacyPath); err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse %s: %w", codacyPath, err)
		}
	}

	// Prepare the new runtime string
	runtimeEntry := name + "@" + version
	found := false
	for i, entry := range config.Runtimes {
		if strings.HasPrefix(entry, name+"@") {
			config.Runtimes[i] = runtimeEntry
			found = true
			break
		}
	}
	if !found {
		config.Runtimes = append(config.Runtimes, runtimeEntry)
	}

	// Write back to .codacy/codacy.yaml
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	if err := os.WriteFile(codacyPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", codacyPath, err)
	}

	logger.Info("Runtime entry updated in .codacy/codacy.yaml", logrus.Fields{
		"runtime": name,
		"version": version,
	})
	return nil
}
