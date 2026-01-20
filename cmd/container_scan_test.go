package cmd

import (
	"os"
	"path/filepath"
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

func TestContainerScanCommandArgs(t *testing.T) {
	// Test that the command accepts 0 or 1 arguments (MaximumNArgs(1))
	assert.Equal(t, "container-scan [FLAGS] [IMAGE_NAME]", containerScanCmd.Use, "Command use should match expected format")

	// Verify Args allows 0 args (for auto-detection)
	err := containerScanCmd.Args(containerScanCmd, []string{})
	assert.NoError(t, err, "Should not error when no args provided (auto-detection mode)")

	// Verify Args allows 1 arg
	err = containerScanCmd.Args(containerScanCmd, []string{"myapp:latest"})
	assert.NoError(t, err, "Should not error when one arg provided")

	// Verify Args rejects 2+ args
	err = containerScanCmd.Args(containerScanCmd, []string{"image1", "image2"})
	assert.Error(t, err, "Should error when too many args provided")
}

func TestContainerScanFlagDefaults(t *testing.T) {
	// Get the flags from the command
	severityFlagDef := containerScanCmd.Flags().Lookup("severity")
	pkgTypesFlagDef := containerScanCmd.Flags().Lookup("pkg-types")
	ignoreUnfixedFlagDef := containerScanCmd.Flags().Lookup("ignore-unfixed")
	dockerfileFlagDef := containerScanCmd.Flags().Lookup("dockerfile")
	composeFileFlagDef := containerScanCmd.Flags().Lookup("compose-file")

	// Verify flags exist
	assert.NotNil(t, severityFlagDef, "severity flag should exist")
	assert.NotNil(t, pkgTypesFlagDef, "pkg-types flag should exist")
	assert.NotNil(t, ignoreUnfixedFlagDef, "ignore-unfixed flag should exist")
	assert.NotNil(t, dockerfileFlagDef, "dockerfile flag should exist")
	assert.NotNil(t, composeFileFlagDef, "compose-file flag should exist")

	// Verify default values
	assert.Equal(t, "", severityFlagDef.DefValue, "severity default should be empty (uses HIGH,CRITICAL in buildTrivyArgs)")
	assert.Equal(t, "", pkgTypesFlagDef.DefValue, "pkg-types default should be empty (uses 'os' in buildTrivyArgs)")
	assert.Equal(t, "true", ignoreUnfixedFlagDef.DefValue, "ignore-unfixed default should be true")
	assert.Equal(t, "", dockerfileFlagDef.DefValue, "dockerfile default should be empty")
	assert.Equal(t, "", composeFileFlagDef.DefValue, "compose-file default should be empty")
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

func TestParseDockerfile(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedImages []string
	}{
		{
			name:           "simple FROM",
			content:        "FROM alpine:3.16\nRUN echo hello",
			expectedImages: []string{"alpine:3.16"},
		},
		{
			name:           "FROM with AS",
			content:        "FROM golang:1.21 AS builder\nRUN go build\nFROM alpine:latest\nCOPY --from=builder /app /app",
			expectedImages: []string{"golang:1.21", "alpine:latest"},
		},
		{
			name:           "multiple FROM stages",
			content:        "FROM node:18 AS build\nRUN npm install\nFROM nginx:alpine\nCOPY --from=build /app /usr/share/nginx/html",
			expectedImages: []string{"node:18", "nginx:alpine"},
		},
		{
			name:           "FROM with registry",
			content:        "FROM ghcr.io/codacy/base:1.0.0\nRUN echo test",
			expectedImages: []string{"ghcr.io/codacy/base:1.0.0"},
		},
		{
			name:           "skip scratch",
			content:        "FROM golang:1.21 AS builder\nRUN go build\nFROM scratch\nCOPY --from=builder /app /app",
			expectedImages: []string{"golang:1.21"},
		},
		{
			name:           "case insensitive FROM",
			content:        "from ubuntu:22.04\nrun apt-get update",
			expectedImages: []string{"ubuntu:22.04"},
		},
		{
			name:           "empty dockerfile",
			content:        "",
			expectedImages: nil,
		},
		{
			name:           "no FROM instruction",
			content:        "# Just a comment\nRUN echo hello",
			expectedImages: nil,
		},
		{
			name:           "duplicate images",
			content:        "FROM alpine:3.16\nRUN echo 1\nFROM alpine:3.16\nRUN echo 2",
			expectedImages: []string{"alpine:3.16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and Dockerfile
			tmpDir := t.TempDir()
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
			err := os.WriteFile(dockerfilePath, []byte(tt.content), 0644)
			assert.NoError(t, err)

			images := parseDockerfile(dockerfilePath)
			assert.Equal(t, tt.expectedImages, images)
		})
	}
}

func TestParseDockerfileNotFound(t *testing.T) {
	images := parseDockerfile("/nonexistent/Dockerfile")
	assert.Nil(t, images, "Should return nil for nonexistent file")
}

func TestParseDockerCompose(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedImages []string
	}{
		{
			name: "simple service with image",
			content: `services:
  web:
    image: nginx:latest`,
			expectedImages: []string{"nginx:latest"},
		},
		{
			name: "multiple services with images",
			content: `services:
  web:
    image: nginx:alpine
  db:
    image: postgres:15
  cache:
    image: redis:7`,
			expectedImages: []string{"nginx:alpine", "postgres:15", "redis:7"},
		},
		{
			name: "service without image (build only)",
			content: `services:
  app:
    build: .`,
			expectedImages: nil,
		},
		{
			name: "mixed services",
			content: `services:
  web:
    image: nginx:latest
  app:
    build:
      context: .
      dockerfile: Dockerfile`,
			expectedImages: []string{"nginx:latest"},
		},
		{
			name:           "empty compose",
			content:        "",
			expectedImages: nil,
		},
		{
			name: "duplicate images",
			content: `services:
  web1:
    image: nginx:latest
  web2:
    image: nginx:latest`,
			expectedImages: []string{"nginx:latest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			composePath := filepath.Join(tmpDir, "docker-compose.yml")
			err := os.WriteFile(composePath, []byte(tt.content), 0644)
			assert.NoError(t, err)

			images := parseDockerCompose(composePath)

			// Sort both slices for comparison since map iteration order is random
			if tt.expectedImages == nil {
				assert.Nil(t, images)
			} else {
				assert.ElementsMatch(t, tt.expectedImages, images)
			}
		})
	}
}

func TestParseDockerComposeNotFound(t *testing.T) {
	images := parseDockerCompose("/nonexistent/docker-compose.yml")
	assert.Nil(t, images, "Should return nil for nonexistent file")
}

func TestParseDockerComposeWithBuildDockerfile(t *testing.T) {
	// Create temp directory and change to it for relative path resolution
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create a Dockerfile in a subdirectory
	appDir := filepath.Join(tmpDir, "app")
	err := os.MkdirAll(appDir, 0755)
	assert.NoError(t, err)

	dockerfileContent := "FROM python:3.11\nRUN pip install flask"
	err = os.WriteFile(filepath.Join(appDir, "Dockerfile"), []byte(dockerfileContent), 0644)
	assert.NoError(t, err)

	// Create docker-compose.yml that references the Dockerfile
	composeContent := `services:
  api:
    build:
      context: ./app
      dockerfile: Dockerfile
  web:
    image: nginx:alpine`

	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte(composeContent), 0644)
	assert.NoError(t, err)

	images := parseDockerCompose(composePath)

	// Should include both the direct image and the base image from Dockerfile
	assert.Contains(t, images, "nginx:alpine", "Should include direct image reference")
	assert.Contains(t, images, "python:3.11", "Should include base image from Dockerfile")
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
