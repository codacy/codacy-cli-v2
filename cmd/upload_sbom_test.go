package cmd

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sbomTestState struct {
	apiToken string
	provider string
	org      string
	repoName string
	env      string
	tag    string
	format string
}

func saveSBOMState() sbomTestState {
	return sbomTestState{
		apiToken: sbomAPIToken,
		provider: sbomProvider,
		org:      sbomOrg,
		repoName: sbomRepoName,
		env:      sbomEnv,
		tag:    sbomTag,
		format:   sbomFormat,
	}
}

func (s sbomTestState) restore() {
	sbomAPIToken = s.apiToken
	sbomProvider = s.provider
	sbomOrg = s.org
	sbomRepoName = s.repoName
	sbomEnv = s.env
	sbomTag = s.tag
	sbomFormat = s.format
}

// setSBOMDefaults sets the minimum required SBOM globals for tests
func setSBOMDefaults() {
	sbomProvider = "gh"
	sbomOrg = "test-org"
	sbomAPIToken = "test-token"
	sbomRepoName = ""
	sbomEnv = ""
	sbomTag = ""
	sbomFormat = "cyclonedx"
}

func TestParseImageRef(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantTag  string
	}{
		{"myapp:latest", "myapp", "latest"},
		{"myapp:v1.0.0", "myapp", "v1.0.0"},
		{"myapp", "myapp", "latest"},
		{"ghcr.io/codacy/app:v2", "ghcr.io/codacy/app", "v2"},
		{"registry.example.com:5000/myapp:tag", "registry.example.com:5000/myapp", "tag"},
		{"nginx@sha256:abc123", "nginx", "sha256:abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, tag := parseImageRef(tt.input)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantTag, tag)
		})
	}
}

func TestExecuteUploadSBOM_InvalidImage(t *testing.T) {
	state := saveState()
	defer state.restore()

	exitCode := executeUploadSBOM("nginx;rm -rf /")
	assert.Equal(t, 2, exitCode)
}

func TestExecuteUploadSBOM_InvalidFormat(t *testing.T) {
	state := saveState()
	defer state.restore()
	ss := saveSBOMState()
	defer ss.restore()

	setSBOMDefaults()
	sbomFormat = "invalid-format"

	exitCode := executeUploadSBOM("alpine:latest")
	assert.Equal(t, 2, exitCode)
}

func TestExecuteUploadSBOM_TrivyNotFound(t *testing.T) {
	state := saveState()
	defer state.restore()
	ss := saveSBOMState()
	defer ss.restore()

	var capturedExitCode int
	exitFunc = func(code int) {
		capturedExitCode = code
	}

	getTrivyPathResolver = func() (string, error) {
		return "", errors.New("trivy not found")
	}
	setSBOMDefaults()

	exitCode := executeUploadSBOM("alpine:latest")
	assert.Equal(t, 2, capturedExitCode)
	assert.Equal(t, 2, exitCode)
}

func TestExecuteUploadSBOM_TrivyGenerationFails(t *testing.T) {
	state := saveState()
	defer state.restore()
	ss := saveSBOMState()
	defer ss.restore()

	getTrivyPathResolver = func() (string, error) {
		return "/usr/local/bin/trivy", nil
	}

	mockRunner := &MockCommandRunner{
		RunWithStderrFunc: func(_ string, _ []string, stderr io.Writer) error {
			if stderr != nil {
				_, _ = stderr.Write([]byte("FATAL   Fatal error"))
			}
			return &mockExitError{code: 1}
		},
	}
	commandRunner = mockRunner
	setSBOMDefaults()

	exitCode := executeUploadSBOM("alpine:latest")
	assert.Equal(t, 2, exitCode)
}


func TestExecuteUploadSBOM_TrivyCalledWithCorrectFormat(t *testing.T) {
	formats := []string{"cyclonedx", "spdx-json"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			state := saveState()
			defer state.restore()
			ss := saveSBOMState()
			defer ss.restore()

			getTrivyPathResolver = func() (string, error) {
				return "/usr/local/bin/trivy", nil
			}

			mockRunner := &MockCommandRunner{
				RunWithStderrFunc: func(_ string, args []string, _ io.Writer) error {
					for i, arg := range args {
						if arg == "-o" && i+1 < len(args) {
							os.WriteFile(args[i+1], []byte(`{}`), 0644)
							break
						}
					}
					return nil
				},
			}
			commandRunner = mockRunner
			setSBOMDefaults()
			sbomFormat = format

			// Will fail at upload (no real API), but we can verify Trivy args
			_ = executeUploadSBOM("alpine:latest")

			assert.Len(t, mockRunner.Calls, 1)
			assert.Contains(t, mockRunner.Calls[0].Args, "--format")
			assert.Contains(t, mockRunner.Calls[0].Args, format)
		})
	}
}

func TestUploadSBOMToCodacy_FileNotFound(t *testing.T) {
	err := uploadSBOMToCodacy("/nonexistent/file.json", "myapp", "latest")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open SBOM file")
}

func TestUploadSBOMSkipsValidation(t *testing.T) {
	result := shouldSkipValidation("upload-sbom")
	assert.True(t, result, "upload-sbom should skip validation")
}

func TestUploadSBOMCommandRequiresArg(t *testing.T) {
	err := uploadSBOMCmd.Args(uploadSBOMCmd, []string{})
	assert.Error(t, err, "Should error when no args provided")

	err = uploadSBOMCmd.Args(uploadSBOMCmd, []string{"myapp:latest"})
	assert.NoError(t, err, "Should accept single image")

	err = uploadSBOMCmd.Args(uploadSBOMCmd, []string{"img1", "img2"})
	assert.Error(t, err, "Should error when multiple args provided")
}
