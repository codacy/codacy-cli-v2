package test

import (
	"codacy/cli-v2/plugins"
	"codacy/cli-v2/tools/lizard"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunLizard(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Use the current directory containing our sample complex code
	testDirectory := currentDirectory
	globalCache := filepath.Join(homeDirectory, ".cache/codacy")

	// Setup tool info
	toolInfo := &plugins.ToolInfo{
		InstallDir: filepath.Join(globalCache, "tools/lizard@1.17.19"),
	}

	t.Run("Output to file", func(t *testing.T) {
		tempResultFile := filepath.Join(os.TempDir(), "lizard.csv")
		defer os.Remove(tempResultFile)

		err := lizard.RunLizard(testDirectory, toolInfo, []string{"complex.py"}, tempResultFile, "")
		assert.NoError(t, err)

		// Read actual output
		actualData, err := os.ReadFile(tempResultFile)
		assert.NoError(t, err)

		// Read expected output
		expectedData, err := os.ReadFile("expected.csv")
		assert.NoError(t, err)

		// Normalize and compare the files
		actualStr := strings.TrimSpace(string(actualData))
		expectedStr := strings.TrimSpace(string(expectedData))
		assert.Equal(t, expectedStr, actualStr, "Output does not match expected output")
	})

	t.Run("No output file specified", func(t *testing.T) {
		err := lizard.RunLizard(testDirectory, toolInfo, []string{"complex.py"}, "", "")
		assert.NoError(t, err)
	})

	t.Run("No files specified", func(t *testing.T) {
		err := lizard.RunLizard(testDirectory, toolInfo, nil, "", "")
		assert.NoError(t, err)
	})
}
