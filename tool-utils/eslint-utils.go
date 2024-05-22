package tool_utils

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

func InstallESLint(npmExecutablePath string, ESLintversion string, toolsDirectory string) {
	log.Println("Installing ESLint")

	eslintInstallationFolder := filepath.Join(toolsDirectory, ESLintversion)

	cmd := exec.Command(npmExecutablePath, "install", "--prefix", eslintInstallationFolder, ESLintversion, "@microsoft/eslint-formatter-sarif")
	// to use the chdir command we needed to create the folder before, we can change this after
	// cmd.Dir = eslintInstallationFolder
	stdout, err := cmd.Output()

	// Print the output
	fmt.Println(string(stdout))

	if err != nil {
		log.Fatal(err)
	}
}
