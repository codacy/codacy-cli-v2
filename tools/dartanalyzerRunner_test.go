package tools

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunDartAnalyzerToFile(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	testDirectory := "testdata/repositories/dartanalizer"
	tempResultFile := filepath.Join(os.TempDir(), "eslint.sarif")
	defer os.Remove(tempResultFile)

	repositoryToAnalyze := filepath.Join(testDirectory, "src")
	expectedSarifFile := filepath.Join(testDirectory, "expected.sarif")
	dartInstallationDirectory := filepath.Join(homeDirectory, ".cache/codacy/runtimes/dart-sdk")
	dartBinary := "dart"

	RunDartAnalyzer(repositoryToAnalyze, dartInstallationDirectory, dartBinary, nil, tempResultFile, "sarif")

	expectedSarifBytes, err := os.ReadFile(expectedSarifFile)
	if err != nil {
		log.Fatal(err)
	}

	obtainedSarifBytes, err := os.ReadFile(tempResultFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	obtainedSarif := string(obtainedSarifBytes)

	filePrefix := currentDirectory + "/"
	actualSarif := strings.ReplaceAll(obtainedSarif, filePrefix, "")

	expectedSarif := strings.TrimSpace(string(expectedSarifBytes))

	assert.Equal(t, expectedSarif, actualSarif, "output did not match expected")
}
