package tools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunTrivyToFile(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	testDirectory := "testdata/repositories/trivy"
	tempResultFile := filepath.Join(os.TempDir(), "trivy.sarif")
	defer os.Remove(tempResultFile)

	repositoryToAnalyze := filepath.Join(testDirectory, "src")

	trivyBinary := filepath.Join(homeDirectory, ".cache/codacy/tools/trivy@0.59.1/trivy")

	err = RunTrivy(repositoryToAnalyze, trivyBinary, nil, tempResultFile, "sarif")
	if err != nil {
		t.Fatalf("Failed to run trivy: %v", err)
	}

	// Check if the output file was created
	obtainedSarifBytes, err := os.ReadFile(tempResultFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	obtainedSarif := string(obtainedSarifBytes)
	filePrefix := "file://" + currentDirectory + "/"
	fmt.Println(filePrefix)
	actualSarif := strings.ReplaceAll(obtainedSarif, filePrefix, "")

	// Read the expected SARIF
	expectedSarifFile := filepath.Join(testDirectory, "expected.sarif")
	expectedSarifBytes, err := os.ReadFile(expectedSarifFile)
	if err != nil {
		log.Fatal(err)
	}
	expectedSarif := strings.TrimSpace(string(expectedSarifBytes))

	assert.Equal(t, expectedSarif, actualSarif, "output did not match expected")
}
