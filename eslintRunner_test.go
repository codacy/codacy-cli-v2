package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunEslint(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}
	testDirectory := "testdata/repositories/test1"
	repositoryToAnalyze := filepath.Join(testDirectory, "src")
	sarifOutputFile := filepath.Join(testDirectory, "sarif.json")
	eslintInstallationDirectory := filepath.Join(homeDirectory, ".cache/codacy-cli-v2/tools/eslint")
	nodeBinary := "node"

	eslintOutput, err := runEslint(repositoryToAnalyze, eslintInstallationDirectory, nodeBinary)
	if err != nil {
		log.Fatal(err.Error())
	}

	expectedSarifBytes, err := os.ReadFile(sarifOutputFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	filePrefix := "file://" + currentDirectory + "/"
	actualSarif := strings.ReplaceAll(eslintOutput, filePrefix, "")

	expectedSarif := string(expectedSarifBytes)

	assert.Equal(t, expectedSarif, actualSarif, "output did not match expected")
}
