package cmd

import (
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Python file",
			filePath: "test.py",
			want:     ".py",
		},
		{
			name:     "C++ file",
			filePath: "test.cpp",
			want:     ".cpp",
		},
		{
			name:     "File with path",
			filePath: "/path/to/file.js",
			want:     ".js",
		},
		{
			name:     "File without extension",
			filePath: "noextension",
			want:     "",
		},
		{
			name:     "File with uppercase extension",
			filePath: "test.PY",
			want:     ".py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFileExtension(tt.filePath); got != tt.want {
				t.Errorf("GetFileExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsToolSupportedForFile(t *testing.T) {
	langConfig := &LanguagesConfig{
		Tools: []struct {
			Name       string   `yaml:"name" json:"name"`
			Languages  []string `yaml:"languages" json:"languages"`
			Extensions []string `yaml:"extensions" json:"extensions"`
		}{
			{
				Name:       "pylint",
				Languages:  []string{"Python"},
				Extensions: []string{".py"},
			},
			{
				Name:       "cppcheck",
				Languages:  []string{"C", "CPP"},
				Extensions: []string{".c", ".cpp", ".h", ".hpp"},
			},
			{
				Name:       "trivy",
				Languages:  []string{"Multiple"},
				Extensions: []string{},
			},
		},
	}

	tests := []struct {
		name     string
		toolName string
		filePath string
		config   *LanguagesConfig
		want     bool
	}{
		{
			name:     "Pylint with Python file",
			toolName: "pylint",
			filePath: "test.py",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Pylint with C++ file",
			toolName: "pylint",
			filePath: "test.cpp",
			config:   langConfig,
			want:     false,
		},
		{
			name:     "Cppcheck with C++ file",
			toolName: "cppcheck",
			filePath: "test.cpp",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Tool with no extensions specified",
			toolName: "trivy",
			filePath: "any.file",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Unknown tool",
			toolName: "unknown",
			filePath: "test.py",
			config:   langConfig,
			want:     true,
		},
		{
			name:     "Nil config",
			toolName: "pylint",
			filePath: "test.cpp",
			config:   nil,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsToolSupportedForFile(tt.toolName, tt.filePath, tt.config); got != tt.want {
				t.Errorf("IsToolSupportedForFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
