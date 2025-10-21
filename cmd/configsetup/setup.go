// Package configsetup defines interfaces and shared types for building
// configuration files from Codacy settings.
package configsetup

import "codacy/cli-v2/domain"

// ToolConfigCreator defines the interface for tool configuration creators
type ToolConfigCreator interface {
	CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error
	GetConfigFileName() string
	GetToolName() string
}
