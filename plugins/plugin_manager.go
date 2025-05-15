package plugins

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// runtimePlugin represents a runtime plugin with methods to interact with it
type runtimePlugin struct {
	Config     PluginConfig
	ConfigPath string
}

// PluginManager manages runtime and tool plugins
type PluginManager struct {
	runtimePlugins map[string]*runtimePlugin
}

var pluginManager *PluginManager

// GetPluginManager returns the singleton instance of PluginManager
func GetPluginManager() *PluginManager {
	if pluginManager == nil {
		pluginManager = &PluginManager{
			runtimePlugins: make(map[string]*runtimePlugin),
		}
	}
	return pluginManager
}

// GetRuntimeConfig returns the plugin configuration for a runtime
func (pm *PluginManager) GetRuntimeConfig(name string) (PluginConfig, error) {
	plugin, err := loadPlugin(name)
	if err != nil {
		return PluginConfig{}, err
	}
	return plugin.Config, nil
}

// GetToolConfig returns the plugin configuration for a tool
func (pm *PluginManager) GetToolConfig(name string) (ToolPluginConfig, error) {
	// Always use forward slashes for embedded filesystem paths (for windows support)
	pluginPath := fmt.Sprintf(toolPluginYamlPathTemplate, name)

	// Read from embedded filesystem
	data, err := toolsFS.ReadFile(pluginPath)
	if err != nil {
		return ToolPluginConfig{}, fmt.Errorf("error reading plugin.yaml for %s: %w", name, err)
	}

	var config ToolPluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return ToolPluginConfig{}, fmt.Errorf("error parsing plugin.yaml for %s: %w", name, err)
	}

	return config, nil
}

// GetRuntimeVersions returns a map of runtime names to their default versions
func (pm *PluginManager) GetRuntimeVersions() map[string]string {
	versions := make(map[string]string)
	entries, err := pluginsFS.ReadDir("runtimes")
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		plugin, err := loadPlugin(name)
		if err != nil {
			continue
		}

		versions[name] = plugin.Config.DefaultVersion
	}

	return versions
}

// GetToolVersions returns a map of tool names to their default versions
func (pm *PluginManager) GetToolVersions() map[string]string {
	versions := make(map[string]string)
	entries, err := toolsFS.ReadDir("tools")
	if err != nil {
		return versions
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		pluginPath := fmt.Sprintf(toolPluginYamlPathTemplate, name)
		data, err := toolsFS.ReadFile(pluginPath)
		if err != nil {
			continue
		}

		var config ToolPluginConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			continue
		}

		versions[name] = config.DefaultVersion
	}

	return versions
}

// GetToolRuntimeDependencies returns a map of tool names to their runtime dependencies
func (pm *PluginManager) GetToolRuntimeDependencies() map[string]string {
	dependencies := make(map[string]string)
	entries, err := toolsFS.ReadDir("tools")
	if err != nil {
		return dependencies
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		pluginPath := fmt.Sprintf(toolPluginYamlPathTemplate, name)
		data, err := toolsFS.ReadFile(pluginPath)
		if err != nil {
			continue
		}

		var config ToolPluginConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			continue
		}

		if config.Runtime != "" {
			dependencies[name] = config.Runtime
		}
	}

	return dependencies
}
