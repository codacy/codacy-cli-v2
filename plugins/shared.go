package plugins

import (
	"bytes"
	"strings"
	"text/template"
)

// ExtensionConfig defines the file extension based on OS
type ExtensionConfig struct {
	Linux   string `yaml:"linux"`
	Windows string `yaml:"windows"`
	Default string `yaml:"default"`
}

// DownloadConfig holds the download configuration from the plugin.yaml
type DownloadConfig struct {
	URLTemplate      string            `yaml:"url_template"`
	FileNameTemplate string            `yaml:"file_name_template"`
	Extension        ExtensionConfig   `yaml:"extension"`
	ArchMapping      map[string]string `yaml:"arch_mapping"`
	OSMapping        map[string]string `yaml:"os_mapping"`
	ReleaseVersion   string            `yaml:"release_version,omitempty"`
}

// Binary represents a binary executable provided by the runtime or tool
type Binary struct {
	Name string      `yaml:"name"`
	Path interface{} `yaml:"path"` // Can be either string or map[string]string
}

// PluginConfig holds the structure of the plugin.yaml file
type PluginConfig struct {
	Name           string         `yaml:"name"`
	Description    string         `yaml:"description"`
	Download       DownloadConfig `yaml:"download"`
	Binaries       []Binary       `yaml:"binaries"`
	DefaultVersion string         `yaml:"default_version"`
}

// RuntimeConfig represents configuration for a runtime
type RuntimeConfig struct {
	Name    string
	Version string
}

// RuntimeInfo contains all processed information about a runtime
type RuntimeInfo struct {
	Name        string
	Version     string
	InstallDir  string
	DownloadURL string
	FileName    string
	Extension   string
	Binaries    map[string]string // Map of binary name to full path
}

// templateData holds the data to be used in template substitution
type templateData struct {
	Version        string
	MajorVersion   string
	FileName       string
	OS             string
	Arch           string
	Extension      string
	ReleaseVersion string
}

// GetMappedArch returns the architecture mapping for the current system
func GetMappedArch(archMapping map[string]string, goarch string) string {
	// Check if there's a mapping for this architecture
	if mappedArch, ok := archMapping[goarch]; ok {
		return mappedArch
	}
	// Return the original architecture if no mapping exists
	return goarch
}

// GetExtension returns the appropriate file extension based on the OS
func GetExtension(extension ExtensionConfig, goos string) string {
	if goos == "windows" {
		return extension.Windows
	}
	if goos == "linux" && extension.Linux != "" {
		return extension.Linux
	}
	return extension.Default
}

// GetMajorVersion extracts the major version from a version string (e.g. "17.0.10" -> "17")
func GetMajorVersion(version string) string {
	if idx := strings.Index(version, "."); idx != -1 {
		return version[:idx]
	}
	return version
}

// GetFileName generates the filename based on the template
func GetFileName(fileNameTemplate string, version string, mappedArch string, goos string) string {
	// Prepare template data
	data := templateData{
		Version:      version,
		MajorVersion: GetMajorVersion(version),
		OS:           goos,
		Arch:         mappedArch,
	}

	// Execute template substitution for filename
	tmpl, err := template.New("filename").Parse(fileNameTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ""
	}

	return buf.String()
}

// GetDownloadURL generates the download URL based on the template
func GetDownloadURL(urlTemplate string, fileName string, version string, mappedArch string, mappedOS string, extension string, releaseVersion string) string {
	// Prepare template data
	data := templateData{
		Version:        version,
		MajorVersion:   GetMajorVersion(version),
		FileName:       fileName,
		OS:             mappedOS,
		Arch:           mappedArch,
		Extension:      extension,
		ReleaseVersion: releaseVersion,
	}

	// Execute template substitution for URL
	tmpl, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ""
	}

	url := buf.String()
	return url
}

// GetMappedOS returns the OS mapping for the current system
func GetMappedOS(osMapping map[string]string, goos string) string {
	// Check if there's a mapping for this OS
	if mappedOS, ok := osMapping[goos]; ok {
		return mappedOS
	}
	// Return the original OS if no mapping exists
	return goos
}
