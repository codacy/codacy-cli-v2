package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"codacy/cli-v2/utils"
)

// RuntimeManager handles the installation and management of different runtimes
type RuntimeManager struct {
	runtimesDir string
}

// NewRuntimeManager creates a new runtime manager
func NewRuntimeManager(runtimesDir string) *RuntimeManager {
	return &RuntimeManager{
		runtimesDir: runtimesDir,
	}
}

// matchesOSAndArch checks if a download configuration matches the current OS and architecture
func matchesOSAndArch(download Download, goos, goarch string) bool {
	// Handle OS matching
	switch osValue := download.OS.(type) {
	case string:
		// Direct string match
		if osValue == goos || (goos == "darwin" && osValue == "macos") {
			return true
		}
	case map[string]interface{}:
		// Map of OS mappings
		if mappedOS, ok := osValue["macos"].(string); ok && goos == "darwin" {
			return mappedOS == "apple-darwin"
		}
		if mappedOS, ok := osValue[goos].(string); ok {
			return mappedOS == getOSString(goos)
		}
	}

	// Handle CPU architecture matching if specified
	if download.CPU != nil {
		switch cpuValue := download.CPU.(type) {
		case string:
			// Direct string match
			if cpuValue != getArchString(goarch) {
				return false
			}
		case map[string]interface{}:
			// Map of CPU mappings
			if mappedArch, ok := cpuValue[goarch].(string); ok {
				return mappedArch == getArchString(goarch)
			}
			return false
		}
	}

	return false
}

// getOSString converts Go OS names to runtime-specific OS names
func getOSString(goos string) string {
	switch goos {
	case "darwin":
		return "darwin"
	case "windows":
		return "win"
	default:
		return goos
	}
}

// getArchString converts Go architecture names to runtime-specific architecture names
func getArchString(goarch string) string {
	switch goarch {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm64":
		return "arm64"
	default:
		return goarch
	}
}

// InstallRuntime installs a runtime based on its plugin.yaml definition
func (rm *RuntimeManager) InstallRuntime(runtimeName, version string) error {
	// Load the plugin configuration
	pluginPath := filepath.Join("definitions/runtimes", runtimeName, "plugin.yaml")
	plugin, err := LoadPluginConfig(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin config for %s: %w", runtimeName, err)
	}

	// Get current OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Find the runtime download configuration
	var runtimeDownload *Download
	for _, rt := range plugin.Downloads {
		if rt.Name == runtimeName {
			// Find the download configuration that matches our OS, architecture, and version
			for _, download := range rt.Downloads {
				// Check if this download matches our OS and architecture
				if matchesOSAndArch(download, goos, goarch) {
					// Check if the version constraint matches
					if download.Version != "" {
						// For now, we'll just check if the version is less than or equal to the constraint
						// This is a simple check and should be improved with proper semver comparison
						if strings.HasPrefix(download.Version, "<=") {
							constraint := strings.TrimPrefix(download.Version, "<=")
							// Split versions into major.minor.patch
							versionParts := strings.Split(version, ".")
							constraintParts := strings.Split(constraint, ".")

							// Convert parts to integers for comparison
							versionMajor, _ := strconv.Atoi(versionParts[0])
							versionMinor, _ := strconv.Atoi(versionParts[1])
							versionPatch, _ := strconv.Atoi(versionParts[2])
							constraintMajor, _ := strconv.Atoi(constraintParts[0])
							constraintMinor, _ := strconv.Atoi(constraintParts[1])
							constraintPatch, _ := strconv.Atoi(constraintParts[2])

							// Compare major version
							if versionMajor < constraintMajor {
								runtimeDownload = &download
								break
							} else if versionMajor == constraintMajor {
								// Compare minor version
								if versionMinor < constraintMinor {
									runtimeDownload = &download
									break
								} else if versionMinor == constraintMinor {
									// Compare patch version
									if versionPatch <= constraintPatch {
										runtimeDownload = &download
										break
									}
								}
							}
						}
					} else {
						runtimeDownload = &download
						break
					}
				}
			}
			if runtimeDownload != nil {
				break
			}
		}
	}

	if runtimeDownload == nil {
		return fmt.Errorf("no matching download found for %s runtime version %s on OS %s and architecture %s", runtimeName, version, goos, goarch)
	}

	// Get the mapped OS and CPU values from the download configuration
	var mappedOS, mappedCPU string
	if osMap, ok := runtimeDownload.OS.(map[string]interface{}); ok {
		if goos == "darwin" {
			if mapped, ok := osMap["macos"].(string); ok {
				mappedOS = mapped
			}
		} else if mapped, ok := osMap[goos].(string); ok {
			mappedOS = mapped
		}
	}

	if cpuMap, ok := runtimeDownload.CPU.(map[string]interface{}); ok {
		if mapped, ok := cpuMap[goarch].(string); ok {
			mappedCPU = mapped
		}
	}

	// Extract the date from the URL (e.g., "20221106" from the example URL)
	urlParts := strings.Split(runtimeDownload.URL, "/")
	var date string
	for _, part := range urlParts {
		if len(part) == 8 && strings.HasPrefix(part, "20") {
			date = part
			break
		}
	}

	// Construct the URL in the correct format
	url := fmt.Sprintf("https://github.com/indygreg/python-build-standalone/releases/download/%s/cpython-%s+%s-%s-%s-install_only.tar.gz",
		date,
		version,
		date,
		mappedCPU,
		mappedOS)

	// Log the constructed URL for debugging
	log.Printf("Download URL: %s", url)

	// Download the runtime
	log.Printf("Fetching %s %s...\n", runtimeName, version)
	archivePath, err := utils.DownloadFile(url, rm.runtimesDir)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", runtimeName, err)
	}

	// Extract the archive
	archive, err := os.Open(archivePath)
	defer archive.Close()
	if err != nil {
		return fmt.Errorf("failed to open downloaded archive: %w", err)
	}

	err = utils.ExtractTarGz(archive, rm.runtimesDir)
	if err != nil {
		return fmt.Errorf("failed to extract %s archive: %w", runtimeName, err)
	}

	return nil
}

// GetRuntimeInfo returns information about the installed runtime
func (rm *RuntimeManager) GetRuntimeInfo(runtimeName, version string) map[string]string {
	info := make(map[string]string)

	switch runtimeName {
	case "node":
		// Node.js installation directory structure
		nodeDir := path.Join(rm.runtimesDir, fmt.Sprintf("node-v%s-%s-%s", version, runtime.GOOS, getArchString(runtime.GOARCH)))
		info["installDir"] = nodeDir
		info["node"] = path.Join(nodeDir, "bin", "node")
		info["npm"] = path.Join(nodeDir, "bin", "npm")
		info["npx"] = path.Join(nodeDir, "bin", "npx")
		info["corepack"] = path.Join(nodeDir, "bin", "corepack")
	case "python":
		// Python installation directory structure
		pythonDir := path.Join(rm.runtimesDir, fmt.Sprintf("python-v%s", version))
		info["installDir"] = pythonDir
		info["python"] = path.Join(pythonDir, "bin", "python")
		info["pip"] = path.Join(pythonDir, "bin", "pip")
	default:
		log.Printf("Warning: Unknown runtime type %s, using default directory structure", runtimeName)
		runtimeDir := path.Join(rm.runtimesDir, fmt.Sprintf("%s-v%s", runtimeName, version))
		info["installDir"] = runtimeDir
	}

	return info
}

// IsRuntimeInstalled checks if a specific version of a runtime is installed
func (rm *RuntimeManager) IsRuntimeInstalled(runtimeName, version string) bool {
	var runtimeDir string
	switch runtimeName {
	case "node":
		runtimeDir = path.Join(rm.runtimesDir, fmt.Sprintf("node-v%s-%s-%s", version, runtime.GOOS, getArchString(runtime.GOARCH)))
	case "python":
		runtimeDir = path.Join(rm.runtimesDir, fmt.Sprintf("python-v%s", version))
	default:
		runtimeDir = path.Join(rm.runtimesDir, fmt.Sprintf("%s-v%s", runtimeName, version))
	}

	// Check if the directory exists and contains the expected binaries
	_, err := os.Stat(runtimeDir)
	if err != nil {
		return false
	}

	// For Node.js, check for the presence of node and npm
	if runtimeName == "node" {
		nodePath := path.Join(runtimeDir, "bin", "node")
		npmPath := path.Join(runtimeDir, "bin", "npm")
		_, nodeErr := os.Stat(nodePath)
		_, npmErr := os.Stat(npmPath)
		return nodeErr == nil && npmErr == nil
	}

	// For Python, check for the presence of python and pip
	if runtimeName == "python" {
		pythonPath := path.Join(runtimeDir, "bin", "python")
		pipPath := path.Join(runtimeDir, "bin", "pip")
		_, pythonErr := os.Stat(pythonPath)
		_, pipErr := os.Stat(pipPath)
		return pythonErr == nil && pipErr == nil
	}

	return true
}
