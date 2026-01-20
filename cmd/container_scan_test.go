package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test cases for buildTrivyArgs
var trivyArgsCases = []struct {
	name            string
	imageName       string
	severity        string
	pkgTypes        string
	ignoreUnfixed   bool
	expectedArgs    []string
	expectedContains []string
	notContains     []string
}{
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
		name: "custom severity only", imageName: "codacy/engine:1.0.0",
		severity: "CRITICAL", pkgTypes: "", ignoreUnfixed: true,
		expectedContains: []string{"--severity", "CRITICAL", "--pkg-types", "os"},
		notContains:      []string{"HIGH,CRITICAL"},
	},
	{
		name: "custom pkg-types only", imageName: "nginx:alpine",
		severity: "", pkgTypes: "os,library", ignoreUnfixed: true,
		expectedContains: []string{"--severity", "HIGH,CRITICAL", "--pkg-types", "os,library"},
	},
	{
		name: "all custom flags", imageName: "ubuntu:22.04",
		severity: "LOW,MEDIUM,HIGH,CRITICAL", pkgTypes: "os,library", ignoreUnfixed: true,
		expectedContains: []string{"--severity", "LOW,MEDIUM,HIGH,CRITICAL", "--pkg-types", "os,library"},
	},
	{
		name: "ignore-unfixed disabled", imageName: "alpine:latest",
		severity: "", pkgTypes: "", ignoreUnfixed: false,
		expectedContains: []string{"--severity", "HIGH,CRITICAL", "--pkg-types", "os"},
		notContains:      []string{"--ignore-unfixed"},
	},
	{
		name: "exit-code always present", imageName: "test:v1",
		severity: "MEDIUM", pkgTypes: "library", ignoreUnfixed: false,
		expectedContains: []string{"--exit-code", "1"},
	},
	{
		name: "image with registry prefix", imageName: "ghcr.io/codacy/codacy-cli:latest",
		severity: "", pkgTypes: "", ignoreUnfixed: true,
		expectedContains: []string{"ghcr.io/codacy/codacy-cli:latest"},
	},
	{
		name: "image with digest", imageName: "nginx@sha256:abc123",
		severity: "", pkgTypes: "", ignoreUnfixed: true,
		expectedContains: []string{"nginx@sha256:abc123"},
	},
}

func TestBuildTrivyArgs(t *testing.T) {
	for _, tc := range trivyArgsCases {
		t.Run(tc.name, func(t *testing.T) {
			severityFlag = tc.severity
			pkgTypesFlag = tc.pkgTypes
			ignoreUnfixedFlag = tc.ignoreUnfixed

			args := buildTrivyArgs(tc.imageName)

			if tc.expectedArgs != nil {
				assert.Equal(t, tc.expectedArgs, args, "Args should match exactly")
			}
			for _, exp := range tc.expectedContains {
				assert.Contains(t, args, exp, "Args should contain %s", exp)
			}
			for _, notExp := range tc.notContains {
				assert.NotContains(t, args, notExp, "Args should not contain %s", notExp)
			}
			assert.Equal(t, tc.imageName, args[len(args)-1], "Image name should be last")
		})
	}
}

func TestBuildTrivyArgsOrder(t *testing.T) {
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	assert.Equal(t, "image", args[0], "First arg should be 'image'")
	assert.Equal(t, "test:latest", args[len(args)-1], "Image name should be last")

	exitCodeIdx := findIndex(args, "--exit-code")
	assert.NotEqual(t, -1, exitCodeIdx, "--exit-code should be present")
	assert.Equal(t, "1", args[exitCodeIdx+1], "1 should follow --exit-code")
}

func TestBuildTrivyArgsDefaultsApplied(t *testing.T) {
	severityFlag = ""
	pkgTypesFlag = ""
	ignoreUnfixedFlag = true

	args := buildTrivyArgs("test:latest")

	severityIdx := findIndex(args, "--severity")
	assert.NotEqual(t, -1, severityIdx)
	assert.Equal(t, "HIGH,CRITICAL", args[severityIdx+1])

	pkgTypesIdx := findIndex(args, "--pkg-types")
	assert.NotEqual(t, -1, pkgTypesIdx)
	assert.Equal(t, "os", args[pkgTypesIdx+1])

	assert.Contains(t, args, "--ignore-unfixed")
}

func TestContainerScanCommandSkipsValidation(t *testing.T) {
	result := shouldSkipValidation("container-scan")
	assert.True(t, result, "container-scan should skip validation")
}

func TestContainerScanCommandArgs(t *testing.T) {
	assert.Equal(t, "container-scan [FLAGS] [IMAGE_NAME]", containerScanCmd.Use)

	// Verify Args allows 0 args (auto-detection)
	err := containerScanCmd.Args(containerScanCmd, []string{})
	assert.NoError(t, err)

	// Verify Args allows 1 arg
	err = containerScanCmd.Args(containerScanCmd, []string{"myapp:latest"})
	assert.NoError(t, err)

	// Verify Args rejects 2+ args
	err = containerScanCmd.Args(containerScanCmd, []string{"image1", "image2"})
	assert.Error(t, err)
}

func TestContainerScanFlagDefaults(t *testing.T) {
	flags := map[string]string{
		"severity":       "",
		"pkg-types":      "",
		"ignore-unfixed": "true",
		"dockerfile":     "",
		"compose-file":   "",
	}

	for name, expected := range flags {
		flag := containerScanCmd.Flags().Lookup(name)
		assert.NotNil(t, flag, "%s flag should exist", name)
		assert.Equal(t, expected, flag.DefValue, "%s default should be %s", name, expected)
	}
}

// Helper function to find index of element in slice
func findIndex(slice []string, target string) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}
