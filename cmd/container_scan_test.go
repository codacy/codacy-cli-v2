package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type trivyArgsTestCase struct {
	name                string
	imageName           string
	severity            string
	pkgTypes            string
	ignoreUnfixed       bool
	expectedArgs        []string
	expectedContains    []string
	expectedNotContains []string
}

var trivyArgsTestCases = []trivyArgsTestCase{
	{
		name:          "default flags",
		imageName:     "myapp:latest",
		severity:      "",
		pkgTypes:      "",
		ignoreUnfixed: true,
		expectedArgs: []string{
			"image", "--scanners", "vuln", "--ignore-unfixed",
			"--severity", "HIGH,CRITICAL", "--pkg-types", "os",
			"--exit-code", "1", "myapp:latest",
		},
	},
	{
		name:                "custom severity only",
		imageName:           "codacy/engine:1.0.0",
		severity:            "CRITICAL",
		pkgTypes:            "",
		ignoreUnfixed:       true,
		expectedContains:    []string{"--severity", "CRITICAL", "--pkg-types", "os", "--ignore-unfixed", "codacy/engine:1.0.0"},
		expectedNotContains: []string{"HIGH,CRITICAL"},
	},
	{
		name:             "custom pkg-types only",
		imageName:        "nginx:alpine",
		severity:         "",
		pkgTypes:         "os,library",
		ignoreUnfixed:    true,
		expectedContains: []string{"--severity", "HIGH,CRITICAL", "--pkg-types", "os,library", "nginx:alpine"},
	},
	{
		name:             "all custom flags",
		imageName:        "ubuntu:22.04",
		severity:         "LOW,MEDIUM,HIGH,CRITICAL",
		pkgTypes:         "os,library",
		ignoreUnfixed:    true,
		expectedContains: []string{"--severity", "LOW,MEDIUM,HIGH,CRITICAL", "--pkg-types", "os,library", "--ignore-unfixed", "ubuntu:22.04"},
	},
	{
		name:                "ignore-unfixed disabled",
		imageName:           "alpine:latest",
		severity:            "",
		pkgTypes:            "",
		ignoreUnfixed:       false,
		expectedContains:    []string{"--severity", "HIGH,CRITICAL", "--pkg-types", "os", "alpine:latest"},
		expectedNotContains: []string{"--ignore-unfixed"},
	},
	{
		name:             "exit-code always present",
		imageName:        "test:v1",
		severity:         "MEDIUM",
		pkgTypes:         "library",
		ignoreUnfixed:    false,
		expectedContains: []string{"--exit-code", "1"},
	},
	{
		name:             "image with registry prefix",
		imageName:        "ghcr.io/codacy/codacy-cli:latest",
		severity:         "",
		pkgTypes:         "",
		ignoreUnfixed:    true,
		expectedContains: []string{"ghcr.io/codacy/codacy-cli:latest"},
	},
	{
		name:             "image with digest",
		imageName:        "nginx@sha256:abc123",
		severity:         "",
		pkgTypes:         "",
		ignoreUnfixed:    true,
		expectedContains: []string{"nginx@sha256:abc123"},
	},
}

func TestBuildTrivyArgs(t *testing.T) {
	for _, tt := range trivyArgsTestCases {
		t.Run(tt.name, func(t *testing.T) {
			severityFlag = tt.severity
			pkgTypesFlag = tt.pkgTypes
			ignoreUnfixedFlag = tt.ignoreUnfixed

			args := buildTrivyArgs(tt.imageName)

			if tt.expectedArgs != nil {
				assert.Equal(t, tt.expectedArgs, args, "Args should match exactly")
			}
			for _, exp := range tt.expectedContains {
				assert.Contains(t, args, exp, "Args should contain %s", exp)
			}
			for _, notExp := range tt.expectedNotContains {
				assert.NotContains(t, args, notExp, "Args should not contain %s", notExp)
			}
			assertTrivyArgsBaseRequirements(t, args, tt.imageName)
		})
	}
}

func assertTrivyArgsBaseRequirements(t *testing.T, args []string, imageName string) {
	t.Helper()
	assert.Contains(t, args, "image", "First arg should be 'image'")
	assert.Contains(t, args, "--scanners", "Should contain --scanners")
	assert.Contains(t, args, "vuln", "Should contain 'vuln' scanner")
	assert.Contains(t, args, "--exit-code", "Should always contain --exit-code")
	assert.Contains(t, args, "1", "Exit code should be 1")
	assert.Equal(t, imageName, args[len(args)-1], "Image name should be the last argument")
}

func TestBuildTrivyArgsOrder(t *testing.T) {
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	assert.Equal(t, "image", args[0], "First arg should be 'image'")
	assert.Equal(t, "test:latest", args[len(args)-1], "Image name should be last")

	exitCodeIdx := findArgIndex(args, "--exit-code")
	assert.NotEqual(t, -1, exitCodeIdx, "--exit-code should be present")
	assert.Equal(t, "1", args[exitCodeIdx+1], "1 should follow --exit-code")
}

func findArgIndex(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

func TestContainerScanCommandSkipsValidation(t *testing.T) {
	result := shouldSkipValidation("container-scan")
	assert.True(t, result, "container-scan should skip validation")
}

func TestContainerScanCommandRequiresArg(t *testing.T) {
	assert.Equal(t, "container-scan <IMAGE_NAME> [IMAGE_NAME...]", containerScanCmd.Use, "Command use should match expected format")

	err := containerScanCmd.Args(containerScanCmd, []string{})
	assert.Error(t, err, "Should error when no args provided")

	err = containerScanCmd.Args(containerScanCmd, []string{"myapp:latest"})
	assert.NoError(t, err, "Should not error when one arg provided")

	err = containerScanCmd.Args(containerScanCmd, []string{"image1", "image2"})
	assert.NoError(t, err, "Should not error when multiple args provided")

	err = containerScanCmd.Args(containerScanCmd, []string{"image1", "image2", "image3"})
	assert.NoError(t, err, "Should not error when many args provided")
}

func TestContainerScanFlagDefaults(t *testing.T) {
	severityFlagDef := containerScanCmd.Flags().Lookup("severity")
	pkgTypesFlagDef := containerScanCmd.Flags().Lookup("pkg-types")
	ignoreUnfixedFlagDef := containerScanCmd.Flags().Lookup("ignore-unfixed")

	assert.NotNil(t, severityFlagDef, "severity flag should exist")
	assert.NotNil(t, pkgTypesFlagDef, "pkg-types flag should exist")
	assert.NotNil(t, ignoreUnfixedFlagDef, "ignore-unfixed flag should exist")

	assert.Equal(t, "", severityFlagDef.DefValue, "severity default should be empty (uses HIGH,CRITICAL in buildTrivyArgs)")
	assert.Equal(t, "", pkgTypesFlagDef.DefValue, "pkg-types default should be empty (uses 'os' in buildTrivyArgs)")
	assert.Equal(t, "true", ignoreUnfixedFlagDef.DefValue, "ignore-unfixed default should be true")
}

type imageNameTestCase struct {
	name        string
	imageName   string
	expectError bool
	errorMsg    string
}

var validImageNameTestCases = []imageNameTestCase{
	{name: "simple image name", imageName: "nginx", expectError: false},
	{name: "image with tag", imageName: "nginx:latest", expectError: false},
	{name: "image with version tag", imageName: "nginx:1.21.0", expectError: false},
	{name: "image with registry", imageName: "docker.io/library/nginx:latest", expectError: false},
	{name: "image with private registry", imageName: "ghcr.io/codacy/codacy-cli:v1.0.0", expectError: false},
	{name: "image with digest", imageName: "nginx@sha256:abc123def456", expectError: false},
	{name: "image with underscore", imageName: "my_app:latest", expectError: false},
	{name: "image with hyphen", imageName: "my-app:latest", expectError: false},
	{name: "image with dots", imageName: "my.app:v1.0.0", expectError: false},
}

var invalidImageNameTestCases = []imageNameTestCase{
	{name: "command injection with semicolon", imageName: "nginx; rm -rf /", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with pipe", imageName: "nginx | cat /etc/passwd", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with ampersand", imageName: "nginx && malicious", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with backticks", imageName: "nginx`whoami`", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with dollar", imageName: "nginx$(whoami)", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with newline", imageName: "nginx\nmalicious", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with quotes", imageName: "nginx'malicious'", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with double quotes", imageName: "nginx\"malicious\"", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with redirect", imageName: "nginx > /tmp/output", expectError: true, errorMsg: "disallowed character"},
	{name: "command injection with backslash", imageName: "nginx\\malicious", expectError: true, errorMsg: "disallowed character"},
	{name: "empty image name", imageName: "", expectError: true, errorMsg: "cannot be empty"},
	{name: "image name too long", imageName: string(make([]byte, 300)), expectError: true, errorMsg: "too long"},
	{name: "image starting with hyphen", imageName: "-nginx", expectError: true, errorMsg: "invalid image name format"},
}

func TestValidateImageNameValid(t *testing.T) {
	for _, tt := range validImageNameTestCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageName(tt.imageName)
			assert.NoError(t, err, "Did not expect error for image name: %s", tt.imageName)
		})
	}
}

func TestValidateImageNameInvalid(t *testing.T) {
	for _, tt := range invalidImageNameTestCases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImageName(tt.imageName)
			assert.Error(t, err, "Expected error for image name: %s", tt.imageName)
			if tt.errorMsg != "" {
				assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain: %s", tt.errorMsg)
			}
		})
	}
}

func TestBuildTrivyArgsDefaultsApplied(t *testing.T) {
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	severityIdx := findArgIndex(args, "--severity")
	assert.NotEqual(t, -1, severityIdx, "--severity should be present")
	assert.Equal(t, "HIGH,CRITICAL", args[severityIdx+1], "Default severity should be HIGH,CRITICAL")

	pkgTypesIdx := findArgIndex(args, "--pkg-types")
	assert.NotEqual(t, -1, pkgTypesIdx, "--pkg-types should be present")
	assert.Equal(t, "os", args[pkgTypesIdx+1], "Default pkg-types should be 'os'")

	assert.Contains(t, args, "--ignore-unfixed", "--ignore-unfixed should be present when enabled")
}

// Tests for multiple image support

func TestValidateMultipleImages(t *testing.T) {
	// All valid images should pass
	validImages := []string{"alpine:latest", "nginx:1.21", "redis:7"}
	for _, img := range validImages {
		err := validateImageName(img)
		assert.NoError(t, err, "Valid image %s should not error", img)
	}
}

func TestValidateMultipleImagesFailsOnInvalid(t *testing.T) {
	// Test that validation catches invalid images in a list
	images := []string{"alpine:latest", "nginx;malicious", "redis:7"}

	var firstError error
	for _, img := range images {
		if err := validateImageName(img); err != nil {
			firstError = err
			break
		}
	}

	assert.Error(t, firstError, "Should catch invalid image in list")
	assert.Contains(t, firstError.Error(), "disallowed character", "Should report specific error")
}

func TestBuildTrivyArgsForMultipleImages(t *testing.T) {
	severityFlag = "CRITICAL"
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	images := []string{"alpine:latest", "nginx:1.21", "redis:7"}

	// Verify each image gets correct args with same flags
	for _, img := range images {
		args := buildTrivyArgs(img)

		assert.Equal(t, img, args[len(args)-1], "Image name should be last argument")
		assert.Contains(t, args, "--severity", "Should contain severity flag")
		assert.Contains(t, args, "CRITICAL", "Should use configured severity")
	}
}

func TestContainerScanCommandAcceptsMultipleImages(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		errMsg string
	}{
		{
			name: "single image",
			args: []string{"alpine:latest"},
		},
		{
			name: "two images",
			args: []string{"alpine:latest", "nginx:1.21"},
		},
		{
			name: "three images",
			args: []string{"alpine:latest", "nginx:1.21", "redis:7"},
		},
		{
			name: "many images",
			args: []string{"img1:v1", "img2:v2", "img3:v3", "img4:v4", "img5:v5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := containerScanCmd.Args(containerScanCmd, tt.args)
			assert.NoError(t, err, "Command should accept %d image(s)", len(tt.args))
		})
	}
}

func TestContainerScanCommandRejectsNoImages(t *testing.T) {
	err := containerScanCmd.Args(containerScanCmd, []string{})
	assert.Error(t, err, "Command should reject empty image list")
}
