package tools

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunEnigma_NoConfig_NoOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	fakeBinary := "/bin/echo"
	err := RunEnigma(tempDir, tempDir, fakeBinary, []string{"foo.go"}, "", "text")
	assert.NoError(t, err)
}

func TestRunEnigma_WithConfig_NoOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "enigma.yaml")
	ioutil.WriteFile(configPath, []byte("config: true"), 0644)
	fakeBinary := "/bin/echo"
	err := RunEnigma(tempDir, tempDir, fakeBinary, []string{"foo.go"}, "", "text")
	assert.NoError(t, err)
}

func TestRunEnigma_NoConfig_WithOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")
	fakeBinary := "/bin/echo"
	err := RunEnigma(tempDir, tempDir, fakeBinary, []string{"foo.go"}, outputFile, "text")
	assert.NoError(t, err)
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)
}

func TestRunEnigma_WithConfig_WithOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "enigma.yaml")
	ioutil.WriteFile(configPath, []byte("config: true"), 0644)
	outputFile := filepath.Join(tempDir, "output.txt")
	fakeBinary := "/bin/echo"
	err := RunEnigma(tempDir, tempDir, fakeBinary, []string{"foo.go"}, outputFile, "text")
	assert.NoError(t, err)
	_, err = os.Stat(outputFile)
	assert.NoError(t, err)
}

func TestRunEnigma_CreateOutputFileError(t *testing.T) {
	tempDir := t.TempDir()
	fakeBinary := "/bin/echo"
	// Use a directory as output file to force error
	outputFile := tempDir
	err := RunEnigma(tempDir, tempDir, fakeBinary, []string{"foo.go"}, outputFile, "text")
	assert.Error(t, err)
}
