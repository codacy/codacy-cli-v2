package tools

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunEslintToFile(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	testDirectory := "testdata/repositories/test1"
	tempResultFile := filepath.Join(os.TempDir(), "eslint.sarif")
	defer os.Remove(tempResultFile)

	repositoryToAnalyze := filepath.Join(testDirectory, "src")
	expectedSarifFile := filepath.Join(testDirectory, "expected.sarif")
	eslintInstallationDirectory := filepath.Join(homeDirectory, ".cache/codacy/tools/eslint@8.57.0")
	nodeBinary := "node"

	RunEslint(repositoryToAnalyze, eslintInstallationDirectory, nodeBinary, nil, false, tempResultFile, "sarif")

	expectedSarifBytes, err := os.ReadFile(expectedSarifFile)
	if err != nil {
		log.Fatal(err)
	}

	obtainedSarifBytes, err := os.ReadFile(tempResultFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	obtainedSarif := string(obtainedSarifBytes)
	filePrefix := "file://" + currentDirectory + "/"
	actualSarif := strings.ReplaceAll(obtainedSarif, filePrefix, "")

	expectedSarif := strings.TrimSpace(string(expectedSarifBytes))

	assert.Equal(t, expectedSarif, actualSarif, "output did not match expected")
}
