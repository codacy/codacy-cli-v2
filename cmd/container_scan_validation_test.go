package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"codacy/cli-v2/constants"

	"github.com/stretchr/testify/assert"
)

// Test cases for image name validation
var validImageNameCases = []struct {
	name      string
	imageName string
}{
	{"simple image name", "nginx"},
	{"image with tag", "nginx:latest"},
	{"image with version tag", "nginx:1.21.0"},
	{"image with registry", "docker.io/library/nginx:latest"},
	{"image with private registry", "ghcr.io/codacy/codacy-cli:v1.0.0"},
	{"image with digest", "nginx@sha256:abc123def456"},
	{"image with underscore", "my_app:latest"},
	{"image with hyphen", "my-app:latest"},
	{"image with dots", "my.app:v1.0.0"},
}

// Test cases for command injection attempts
var invalidImageNameCases = []struct {
	name      string
	imageName string
	errorMsg  string
}{
	{"command injection with semicolon", "nginx; rm -rf /", "disallowed character"},
	{"command injection with pipe", "nginx | cat /etc/passwd", "disallowed character"},
	{"command injection with ampersand", "nginx && malicious", "disallowed character"},
	{"command injection with backticks", "nginx`whoami`", "disallowed character"},
	{"command injection with dollar", "nginx$(whoami)", "disallowed character"},
	{"command injection with newline", "nginx\nmalicious", "disallowed character"},
	{"command injection with quotes", "nginx'malicious'", "disallowed character"},
	{"command injection with double quotes", "nginx\"malicious\"", "disallowed character"},
	{"command injection with redirect", "nginx > /tmp/output", "disallowed character"},
	{"command injection with backslash", "nginx\\malicious", "disallowed character"},
	{"empty image name", "", "cannot be empty"},
	{"image name too long", string(make([]byte, 300)), "too long"},
	{"image starting with hyphen", "-nginx", "invalid image name format"},
}

func TestValidImageNames(t *testing.T) {
	for _, tc := range validImageNameCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateImageName(tc.imageName)
			assert.NoError(t, err, "Did not expect error for image name: %s", tc.imageName)
		})
	}
}

func TestInvalidImageNames(t *testing.T) {
	for _, tc := range invalidImageNameCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateImageName(tc.imageName)
			assert.Error(t, err, "Expected error for image name: %s", tc.imageName)
			if tc.errorMsg != "" {
				assert.Contains(t, err.Error(), tc.errorMsg, "Error should contain: %s", tc.errorMsg)
			}
		})
	}
}

// Test cases for Dockerfile parsing
var dockerfileParseCases = []struct {
	name           string
	content        string
	expectedImages []string
}{
	{"simple FROM", "FROM alpine:3.16\nRUN echo hello", []string{"alpine:3.16"}},
	{"FROM with AS", "FROM golang:1.21 AS builder\nRUN go build\nFROM alpine:latest\nCOPY --from=builder /app /app", []string{"golang:1.21", "alpine:latest"}},
	{"multiple FROM stages", "FROM node:18 AS build\nRUN npm install\nFROM nginx:alpine\nCOPY --from=build /app /usr/share/nginx/html", []string{"node:18", "nginx:alpine"}},
	{"FROM with registry", "FROM ghcr.io/codacy/base:1.0.0\nRUN echo test", []string{"ghcr.io/codacy/base:1.0.0"}},
	{"skip scratch", "FROM golang:1.21 AS builder\nRUN go build\nFROM scratch\nCOPY --from=builder /app /app", []string{"golang:1.21"}},
	{"case insensitive FROM", "from ubuntu:22.04\nrun apt-get update", []string{"ubuntu:22.04"}},
	{"empty dockerfile", "", nil},
	{"no FROM instruction", "# Just a comment\nRUN echo hello", nil},
	{"duplicate images", "FROM alpine:3.16\nRUN echo 1\nFROM alpine:3.16\nRUN echo 2", []string{"alpine:3.16"}},
}

func TestParseDockerfileContent(t *testing.T) {
	for _, tc := range dockerfileParseCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
			err := os.WriteFile(dockerfilePath, []byte(tc.content), constants.DefaultFilePerms)
			assert.NoError(t, err)

			images := parseDockerfile(dockerfilePath)
			assert.Equal(t, tc.expectedImages, images)
		})
	}
}

func TestParseDockerfileNotFound(t *testing.T) {
	images := parseDockerfile("/nonexistent/Dockerfile")
	assert.Nil(t, images, "Should return nil for nonexistent file")
}

// Test cases for docker-compose parsing
var dockerComposeParseCases = []struct {
	name           string
	content        string
	expectedImages []string
}{
	{
		"simple service with image",
		"services:\n  web:\n    image: nginx:latest",
		[]string{"nginx:latest"},
	},
	{
		"multiple services with images",
		"services:\n  web:\n    image: nginx:alpine\n  db:\n    image: postgres:15\n  cache:\n    image: redis:7",
		[]string{"nginx:alpine", "postgres:15", "redis:7"},
	},
	{
		"service without image",
		"services:\n  app:\n    build: .",
		nil,
	},
	{
		"mixed services",
		"services:\n  web:\n    image: nginx:latest\n  app:\n    build:\n      context: .\n      dockerfile: Dockerfile",
		[]string{"nginx:latest"},
	},
	{"empty compose", "", nil},
	{
		"duplicate images",
		"services:\n  web1:\n    image: nginx:latest\n  web2:\n    image: nginx:latest",
		[]string{"nginx:latest"},
	},
}

func TestParseDockerComposeContent(t *testing.T) {
	for _, tc := range dockerComposeParseCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			composePath := filepath.Join(tmpDir, "docker-compose.yml")
			err := os.WriteFile(composePath, []byte(tc.content), constants.DefaultFilePerms)
			assert.NoError(t, err)

			images := parseDockerCompose(composePath)

			if tc.expectedImages == nil {
				assert.Nil(t, images)
			} else {
				assert.ElementsMatch(t, tc.expectedImages, images)
			}
		})
	}
}

func TestParseDockerComposeNotFound(t *testing.T) {
	images := parseDockerCompose("/nonexistent/docker-compose.yml")
	assert.Nil(t, images, "Should return nil for nonexistent file")
}

func TestParseDockerComposeWithBuildDockerfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tmpDir)

	// Create a Dockerfile in a subdirectory
	appDir := filepath.Join(tmpDir, "app")
	err := os.MkdirAll(appDir, constants.DefaultDirPerms)
	assert.NoError(t, err)

	dockerfileContent := "FROM python:3.11\nRUN pip install flask"
	err = os.WriteFile(filepath.Join(appDir, "Dockerfile"), []byte(dockerfileContent), constants.DefaultFilePerms)
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
	err = os.WriteFile(composePath, []byte(composeContent), constants.DefaultFilePerms)
	assert.NoError(t, err)

	images := parseDockerCompose(composePath)

	assert.Contains(t, images, "nginx:alpine", "Should include direct image reference")
	assert.Contains(t, images, "python:3.11", "Should include base image from Dockerfile")
}
