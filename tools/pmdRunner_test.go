package tools

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunPmdToFile(t *testing.T) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err.Error())
	}
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get absolute paths for test files
	testDirectory := filepath.Join(currentDirectory, "testdata/repositories/pmd")
	tempResultFile := filepath.Join(os.TempDir(), "pmd.sarif")
	defer os.Remove(tempResultFile)

	// Use absolute paths for repository and ruleset
	repositoryToAnalyze := testDirectory
	rulesetFile := filepath.Join(testDirectory, "pmd-ruleset.xml")

	pmdBinary := filepath.Join(homeDirectory, ".cache/codacy/tools/pmd@7.12.0/pmd-bin-7.12.0/bin/pmd")

	err = RunPmd(repositoryToAnalyze, pmdBinary, nil, tempResultFile, "sarif", rulesetFile)
	// PMD returns exit status 4 when violations are found, which is expected in our test
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 4 {
			t.Fatalf("Failed to run PMD: %v", err)
		}
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
	actualSarif = strings.TrimSpace(actualSarif)

	// Read the expected SARIF
	expectedSarifFile := filepath.Join(testDirectory, "expected.sarif")
	expectedSarifBytes, err := os.ReadFile(expectedSarifFile)
	if err != nil {
		log.Fatal(err)
	}
	expectedSarif := strings.TrimSpace(string(expectedSarifBytes))

	assert.Equal(t, expectedSarif, actualSarif, "output did not match expected")
}
