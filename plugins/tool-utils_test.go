package plugins

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessTools(t *testing.T) {
	// Create a list of tool configs for testing
	configs := []ToolConfig{
		{
			Name:    "eslint",
			Version: "8.38.0",
		},
	}

	// Define a test tool directory
	toolDir := "/test/tools"

	// Process the tools
	toolInfos, err := ProcessTools(configs, toolDir, nil)

	// Assert no errors occurred
	assert.NoError(t, err, "ProcessTools should not return an error")

	// Assert we have the expected tool in the results
	assert.Contains(t, toolInfos, "eslint")

	// Get the eslint tool info
	eslintInfo := toolInfos["eslint"]

	// Assert the basic tool info is correct
	assert.Equal(t, "eslint", eslintInfo.Name)
	assert.Equal(t, "8.38.0", eslintInfo.Version)
	assert.Equal(t, "node", eslintInfo.Runtime)

	// Assert the install directory is correct
	expectedInstallDir := filepath.Join(toolDir, "eslint@8.38.0")
	assert.Equal(t, expectedInstallDir, eslintInfo.InstallDir)

	// Assert binary paths are correctly set
	assert.NotNil(t, eslintInfo.Binaries)
	assert.Greater(t, len(eslintInfo.Binaries), 0)

	// Check if eslint binary is present
	eslintBinary := filepath.Join(expectedInstallDir, "node_modules/.bin/eslint")
	assert.Equal(t, eslintBinary, eslintInfo.Binaries["eslint"])

	// Assert formatters are correctly set
	assert.NotNil(t, eslintInfo.Formatters)
	assert.Greater(t, len(eslintInfo.Formatters), 0)
	assert.Equal(t, "-f @microsoft/eslint-formatter-sarif", eslintInfo.Formatters["sarif"])

	// Assert output and analysis options are correctly set
	assert.Equal(t, "-o", eslintInfo.OutputFlag)
	assert.Equal(t, "--fix", eslintInfo.AutofixFlag)
	assert.Equal(t, ".", eslintInfo.DefaultPath)

	// Assert runtime binaries are correctly set
	assert.Equal(t, "npm", eslintInfo.PackageManager)
	assert.Equal(t, "node", eslintInfo.ExecutionBinary)

	// Assert installation command templates are correctly set
	assert.Equal(t, "install --prefix {{.InstallDir}} {{.PackageName}}@{{.Version}} @microsoft/eslint-formatter-sarif", eslintInfo.InstallCommand)
	assert.Equal(t, "config set registry {{.Registry}}", eslintInfo.RegistryCommand)
}

func TestProcessToolsWithDownload(t *testing.T) {
	// Create a list of tool configs for testing
	configs := []ToolConfig{
		{
			Name:    "trivy",
			Version: "0.37.3",
		},
	}

	// Define a test tool directory
	toolDir := "/test/tools"

	// Process the tools
	toolInfos, err := ProcessTools(configs, toolDir, nil)

	// Assert no errors occurred
	assert.NoError(t, err, "ProcessTools should not return an error")

	// Assert we have the expected tool in the results
	assert.Contains(t, toolInfos, "trivy")

	// Get the trivy tool info
	trivyInfo := toolInfos["trivy"]

	// Assert the basic tool info is correct
	assert.Equal(t, "trivy", trivyInfo.Name)
	assert.Equal(t, "0.37.3", trivyInfo.Version)

	// Assert the install directory is correct
	expectedInstallDir := filepath.Join(toolDir, "trivy@0.37.3")
	assert.Equal(t, expectedInstallDir, trivyInfo.InstallDir)

	// Assert download information is correctly set
	assert.NotEmpty(t, trivyInfo.DownloadURL)
	assert.NotEmpty(t, trivyInfo.FileName)
	assert.NotEmpty(t, trivyInfo.Extension)

	// Assert the correct file extension based on OS
	if runtime.GOOS == "windows" {
		assert.Equal(t, "zip", trivyInfo.Extension)
	} else {
		assert.Equal(t, "tar.gz", trivyInfo.Extension)
	}

	// Assert binary paths are correctly set
	assert.NotNil(t, trivyInfo.Binaries)
	assert.Greater(t, len(trivyInfo.Binaries), 0)

	// Check if trivy binary is present
	trivyBinary := filepath.Join(expectedInstallDir, "trivy")
	assert.Equal(t, trivyBinary, trivyInfo.Binaries["trivy"])

	// Verify URL components
	assert.Contains(t, trivyInfo.DownloadURL, "aquasecurity/trivy/releases/download")
	assert.Contains(t, trivyInfo.DownloadURL, trivyInfo.Version)

	// Test OS mapping
	var expectedOS string
	if runtime.GOOS == "darwin" {
		expectedOS = "macOS"
	} else if runtime.GOOS == "linux" {
		expectedOS = "Linux"
	} else if runtime.GOOS == "windows" {
		expectedOS = "Windows"
	} else {
		expectedOS = runtime.GOOS
	}
	assert.Contains(t, trivyInfo.DownloadURL, expectedOS)

	// Test architecture mapping
	var expectedArch string
	if runtime.GOARCH == "386" {
		expectedArch = "32bit"
	} else if runtime.GOARCH == "amd64" {
		expectedArch = "64bit"
	} else if runtime.GOARCH == "arm" {
		expectedArch = "ARM"
	} else if runtime.GOARCH == "arm64" {
		expectedArch = "ARM64"
	} else {
		expectedArch = runtime.GOARCH
	}
	assert.Contains(t, trivyInfo.DownloadURL, expectedArch)
}

func TestGetSupportedTools(t *testing.T) {
	tests := []struct {
		name          string
		expectedTools []string
		expectedError bool
	}{
		{
			name: "should return supported tools",
			expectedTools: []string{
				"eslint",
				"pmd",
				"pylint",
				"trivy",
				"dartanalyzer",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supportedTools, err := GetSupportedTools()

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, supportedTools)

			// Check that all expected tools are supported
			for _, expectedTool := range tt.expectedTools {
				_, exists := supportedTools[expectedTool]
				assert.True(t, exists, "tool %s should be supported", expectedTool)
			}

			// Check that we have exactly the expected number of tools
			assert.Equal(t, len(tt.expectedTools), len(supportedTools), "number of supported tools should match")
		})
	}
}
