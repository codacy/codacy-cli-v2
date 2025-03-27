package plugins

import (
	"codacy/cli-v2/config"
	"codacy/cli-v2/utils"
	"embed"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

//go:embed runtimes/*/plugin.yaml
var runtimePlugins embed.FS

type PluginSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Download    struct {
		URLTemplate      string            `yaml:"url_template"`
		FileNameTemplate string            `yaml:"file_name_template"`
		Extension        map[string]string `yaml:"extension"`
		ArchMapping      map[string]string `yaml:"arch_mapping"`
	} `yaml:"download"`
	Binaries []struct {
		Name string `yaml:"name"`
		Path string `yaml:"path"`
	} `yaml:"binaries"`
}

// getPluginSpec loads a runtime plugin specification by name
func getPluginSpec(runtimeName string) (*PluginSpec, error) {
	pluginPath := fmt.Sprintf("runtimes/%s/plugin.yaml", runtimeName)

	data, err := runtimePlugins.ReadFile(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin specification for runtime %s: %v", runtimeName, err)
	}

	var spec PluginSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse plugin specification for runtime %s: %v", runtimeName, err)
	}

	return &spec, nil
}

// getExtension returns the appropriate file extension based on OS
func getExtension(spec *PluginSpec, goos string) string {
	if ext, ok := spec.Download.Extension[goos]; ok {
		return ext
	}
	return spec.Download.Extension["default"]
}

// getArchMapping maps Go architecture to runtime architecture
func getArchMapping(spec *PluginSpec, goarch string) string {
	if arch, ok := spec.Download.ArchMapping[goarch]; ok {
		return arch
	}
	return goarch
}

// executeTemplate processes a template string with provided data
func executeTemplate(tmplString string, data map[string]string) (string, error) {
	tmpl, err := template.New("template").Parse(tmplString)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// generateRuntimeInfo creates runtime info based on the plugin specification
func generateRuntimeInfo(r *config.Runtime, spec *PluginSpec, goos, goarch string) (map[string]string, error) {
	// Prepare template data
	data := map[string]string{
		"Version": r.Version(),
		"OS":      goos,
		"Arch":    getArchMapping(spec, goarch),
	}

	// Process file name template
	fileName, err := executeTemplate(spec.Download.FileNameTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file name: %v", err)
	}

	// Add fileName to template data for URL generation
	data["FileName"] = fileName
	data["Extension"] = getExtension(spec, goos)

	info := map[string]string{
		"fileName":   fileName,
		"installDir": path.Join(config.Config.RuntimesDirectory(), fileName),
	}

	// Add binary paths
	for _, binary := range spec.Binaries {
		info[binary.Name] = path.Join(config.Config.RuntimesDirectory(), fileName, binary.Path)
	}

	return info, nil
}

// getDownloadURL generates the download URL for a runtime
func getDownloadURL(r *config.Runtime, spec *PluginSpec, goos, goarch string) (string, error) {
	// Generate file name and prepare template data
	info, err := generateRuntimeInfo(r, spec, goos, goarch)
	if err != nil {
		return "", err
	}

	data := map[string]string{
		"Version":   r.Version(),
		"OS":        goos,
		"Arch":      getArchMapping(spec, goarch),
		"FileName":  info["fileName"],
		"Extension": getExtension(spec, goos),
	}

	// Process URL template
	url, err := executeTemplate(spec.Download.URLTemplate, data)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %v", err)
	}

	return url, nil
}

// InstallRuntime installs a single runtime using its plugin specification
func InstallRuntime(r *config.Runtime) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	spec, err := getPluginSpec(r.Name())
	if err != nil {
		return err
	}

	downloadURL, err := getDownloadURL(r, spec, goos, goarch)
	if err != nil {
		return err
	}

	fileName := filepath.Base(downloadURL)
	filePath := filepath.Join(config.Config.RuntimesDirectory(), fileName)

	// Check if runtime archive already exists
	t, err := os.Open(filePath)
	if err == nil {
		defer t.Close()
		log.Printf("%s is already downloaded, skipping download...", r.Name())
	} else {
		// Download the runtime archive
		log.Printf("Downloading %s version %s...", r.Name(), r.Version())
		downloadedFile, err := utils.DownloadFile(downloadURL, config.Config.RuntimesDirectory())
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", r.Name(), err)
		}

		t, err = os.Open(downloadedFile)
		if err != nil {
			return fmt.Errorf("failed to open downloaded file: %v", err)
		}
		defer t.Close()
	}

	log.Printf("Extracting %s...", r.Name())

	// Extract based on file extension
	extension := getExtension(spec, goos)
	if extension == "zip" {
		err = utils.ExtractZip(filePath, config.Config.RuntimesDirectory())
	} else {
		err = utils.ExtractTarGz(t, config.Config.RuntimesDirectory())
	}

	if err != nil {
		return fmt.Errorf("failed to extract %s: %v", r.Name(), err)
	}

	log.Printf("Successfully installed %s version %s", r.Name(), r.Version())
	return nil
}

// InstallRuntimes installs multiple runtimes based on the provided map
func InstallRuntimes(runtimes map[string]*config.Runtime) error {
	for _, runtime := range runtimes {
		if err := InstallRuntime(runtime); err != nil {
			return fmt.Errorf("failed to install runtime %s: %v", runtime.Name(), err)
		}
	}
	return nil
}
