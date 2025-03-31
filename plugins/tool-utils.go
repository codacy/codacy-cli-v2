package plugins

import (
	"fmt"
)

// LoadToolPlugin stub implementation until the plugin system is complete
func LoadToolPlugin(toolName string) (interface{}, error) {
	return nil, fmt.Errorf("plugin system for tools is under development, tool %s cannot be loaded yet", toolName)
}
