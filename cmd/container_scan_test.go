package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTrivyArgs(t *testing.T) {
	tests := []struct {
		name              string
		imageName         string
		severity          string
		pkgTypes          string
		ignoreUnfixed     bool
		expectedArgs      []string
		expectedContains  []string
		expectedNotContains []string
	}{
		{
			name:          "default flags",
			imageName:     "myapp:latest",
			severity:      "",
			pkgTypes:      "",
			ignoreUnfixed: true,
			expectedArgs: []string{
				"image",
				"--scanners", "vuln",
				"--ignore-unfixed",
				"--severity", "HIGH,CRITICAL",
				"--pkg-types", "os",
				"--exit-code", "1",
				"myapp:latest",
			},
		},
		{
			name:          "custom severity only",
			imageName:     "codacy/engine:1.0.0",
			severity:      "CRITICAL",
			pkgTypes:      "",
			ignoreUnfixed: true,
			expectedContains: []string{
				"--severity", "CRITICAL",
				"--pkg-types", "os",
				"--ignore-unfixed",
				"codacy/engine:1.0.0",
			},
			expectedNotContains: []string{
				"HIGH,CRITICAL",
			},
		},
		{
			name:          "custom pkg-types only",
			imageName:     "nginx:alpine",
			severity:      "",
			pkgTypes:      "os,library",
			ignoreUnfixed: true,
			expectedContains: []string{
				"--severity", "HIGH,CRITICAL",
				"--pkg-types", "os,library",
				"nginx:alpine",
			},
		},
		{
			name:          "all custom flags",
			imageName:     "ubuntu:22.04",
			severity:      "LOW,MEDIUM,HIGH,CRITICAL",
			pkgTypes:      "os,library",
			ignoreUnfixed: true,
			expectedContains: []string{
				"--severity", "LOW,MEDIUM,HIGH,CRITICAL",
				"--pkg-types", "os,library",
				"--ignore-unfixed",
				"ubuntu:22.04",
			},
		},
		{
			name:          "ignore-unfixed disabled",
			imageName:     "alpine:latest",
			severity:      "",
			pkgTypes:      "",
			ignoreUnfixed: false,
			expectedContains: []string{
				"--severity", "HIGH,CRITICAL",
				"--pkg-types", "os",
				"alpine:latest",
			},
			expectedNotContains: []string{
				"--ignore-unfixed",
			},
		},
		{
			name:          "exit-code always present",
			imageName:     "test:v1",
			severity:      "MEDIUM",
			pkgTypes:      "library",
			ignoreUnfixed: false,
			expectedContains: []string{
				"--exit-code", "1",
			},
		},
		{
			name:          "image with registry prefix",
			imageName:     "ghcr.io/codacy/codacy-cli:latest",
			severity:      "",
			pkgTypes:      "",
			ignoreUnfixed: true,
			expectedContains: []string{
				"ghcr.io/codacy/codacy-cli:latest",
			},
		},
		{
			name:          "image with digest",
			imageName:     "nginx@sha256:abc123",
			severity:      "",
			pkgTypes:      "",
			ignoreUnfixed: true,
			expectedContains: []string{
				"nginx@sha256:abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the global flag variables
			severityFlag = tt.severity
			pkgTypesFlag = tt.pkgTypes
			ignoreUnfixedFlag = tt.ignoreUnfixed

			// Build the args
			args := buildTrivyArgs(tt.imageName)

			// Check exact match if expectedArgs is provided
			if tt.expectedArgs != nil {
				assert.Equal(t, tt.expectedArgs, args, "Args should match exactly")
			}

			// Check that expected strings are present
			for _, exp := range tt.expectedContains {
				assert.Contains(t, args, exp, "Args should contain %s", exp)
			}

			// Check that not-expected strings are absent
			for _, notExp := range tt.expectedNotContains {
				assert.NotContains(t, args, notExp, "Args should not contain %s", notExp)
			}

			// Always verify base requirements
			assert.Contains(t, args, "image", "First arg should be 'image'")
			assert.Contains(t, args, "--scanners", "Should contain --scanners")
			assert.Contains(t, args, "vuln", "Should contain 'vuln' scanner")
			assert.Contains(t, args, "--exit-code", "Should always contain --exit-code")
			assert.Contains(t, args, "1", "Exit code should be 1")

			// Verify image name is always last
			assert.Equal(t, tt.imageName, args[len(args)-1], "Image name should be the last argument")
		})
	}
}

func TestBuildTrivyArgsOrder(t *testing.T) {
	// Reset flags to defaults
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	// Verify the order of arguments
	// image should be first
	assert.Equal(t, "image", args[0], "First arg should be 'image'")

	// image name should be last
	assert.Equal(t, "test:latest", args[len(args)-1], "Image name should be last")

	// --exit-code and 1 should be consecutive and before image name
	exitCodeIdx := -1
	for i, arg := range args {
		if arg == "--exit-code" {
			exitCodeIdx = i
			break
		}
	}
	assert.NotEqual(t, -1, exitCodeIdx, "--exit-code should be present")
	assert.Equal(t, "1", args[exitCodeIdx+1], "1 should follow --exit-code")
}

func TestContainerScanCommandSkipsValidation(t *testing.T) {
	// Test that container-scan is in the skip validation list
	result := shouldSkipValidation("container-scan")
	assert.True(t, result, "container-scan should skip validation")
}

func TestContainerScanCommandRequiresArg(t *testing.T) {
	// Test that the command requires exactly one argument
	assert.Equal(t, "container-scan [FLAGS] <IMAGE_NAME>", containerScanCmd.Use, "Command use should match expected format")

	// Verify Args is set to ExactArgs(1)
	err := containerScanCmd.Args(containerScanCmd, []string{})
	assert.Error(t, err, "Should error when no args provided")

	err = containerScanCmd.Args(containerScanCmd, []string{"image1", "image2"})
	assert.Error(t, err, "Should error when too many args provided")

	err = containerScanCmd.Args(containerScanCmd, []string{"myapp:latest"})
	assert.NoError(t, err, "Should not error when exactly one arg provided")
}

func TestContainerScanFlagDefaults(t *testing.T) {
	// Get the flags from the command
	severityFlagDef := containerScanCmd.Flags().Lookup("severity")
	pkgTypesFlagDef := containerScanCmd.Flags().Lookup("pkg-types")
	ignoreUnfixedFlagDef := containerScanCmd.Flags().Lookup("ignore-unfixed")

	// Verify flags exist
	assert.NotNil(t, severityFlagDef, "severity flag should exist")
	assert.NotNil(t, pkgTypesFlagDef, "pkg-types flag should exist")
	assert.NotNil(t, ignoreUnfixedFlagDef, "ignore-unfixed flag should exist")

	// Verify default values
	assert.Equal(t, "", severityFlagDef.DefValue, "severity default should be empty (uses HIGH,CRITICAL in buildTrivyArgs)")
	assert.Equal(t, "", pkgTypesFlagDef.DefValue, "pkg-types default should be empty (uses 'os' in buildTrivyArgs)")
	assert.Equal(t, "true", ignoreUnfixedFlagDef.DefValue, "ignore-unfixed default should be true")
}

func TestValidateImageName(t *testing.T) {
	tests := []struct {
		name        string
		imageName   string
		expectError bool
		errorMsg    string
	}{
		// Valid image names
		{
			name:        "simple image name",
			imageName:   "nginx",
			expectError: false,
		},
		{
			name:        "image with tag",
			imageName:   "nginx:latest",
			expectError: false,
		},
		{
			name:        "image with version tag",
			imageName:   "nginx:1.21.0",
			expectError: false,
		},
		{
			name:        "image with registry",
			imageName:   "docker.io/library/nginx:latest",
			expectError: false,
		},
		{
			name:        "image with private registry",
			imageName:   "ghcr.io/codacy/codacy-cli:v1.0.0",
			expectError: false,
		},
		{
			name:        "image with digest",
			imageName:   "nginx@sha256:abc123def456",
			expectError: false,
		},
		{
			name:        "image with underscore",
			imageName:   "my_app:latest",
			expectError: false,
		},
		{
			name:        "image with hyphen",
			imageName:   "my-app:latest",
			expectError: false,
		},
		{
			name:        "image with dots",
			imageName:   "my.app:v1.0.0",
			expectError: false,
		},
		// Invalid image names - command injection attempts
		{
			name:        "command injection with semicolon",
			imageName:   "nginx; rm -rf /",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with pipe",
			imageName:   "nginx | cat /etc/passwd",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with ampersand",
			imageName:   "nginx && malicious",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with backticks",
			imageName:   "nginx`whoami`",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with dollar",
			imageName:   "nginx$(whoami)",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with newline",
			imageName:   "nginx\nmalicious",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with quotes",
			imageName:   "nginx'malicious'",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with double quotes",
			imageName:   "nginx\"malicious\"",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with redirect",
			imageName:   "nginx > /tmp/output",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		{
			name:        "command injection with backslash",
			imageName:   "nginx\\malicious",
			expectError: true,
			errorMsg:    "disallowed character",
		},
		// Invalid format
		{
			name:        "empty image name",
			imageName:   "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "image name too long",
			imageName:   string(make([]byte, 300)),
			expectError: true,
			errorMsg:    "too long",
		},
		{
			name:        "image starting with hyphen",
			imageName:   "-nginx",
			expectError: true,
			errorMsg:    "invalid image name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageName(tt.imageName)

			if tt.expectError {
				assert.Error(t, err, "Expected error for image name: %s", tt.imageName)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain: %s", tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Did not expect error for image name: %s", tt.imageName)
			}
		})
	}
}

func TestBuildTrivyArgsDefaultsApplied(t *testing.T) {
	// Test that when flags are empty, defaults are applied
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	// Find severity value
	severityIdx := -1
	for i, arg := range args {
		if arg == "--severity" {
			severityIdx = i
			break
		}
	}
	assert.NotEqual(t, -1, severityIdx, "--severity should be present")
	assert.Equal(t, "HIGH,CRITICAL", args[severityIdx+1], "Default severity should be HIGH,CRITICAL")

	// Find pkg-types value
	pkgTypesIdx := -1
	for i, arg := range args {
		if arg == "--pkg-types" {
			pkgTypesIdx = i
			break
		}
	}
	assert.NotEqual(t, -1, pkgTypesIdx, "--pkg-types should be present")
	assert.Equal(t, "os", args[pkgTypesIdx+1], "Default pkg-types should be 'os'")

	// Verify --ignore-unfixed is present
	assert.Contains(t, args, "--ignore-unfixed", "--ignore-unfixed should be present when enabled")
}
