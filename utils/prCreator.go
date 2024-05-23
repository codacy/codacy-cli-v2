package utils

import (
	"log"
	"os"

	"os/exec"

	"github.com/google/uuid"

	_ "embed"
)

//go:embed prCreator.sh
var prCreatorScriptContent string

func CreatePr(dryRun bool) bool {
	uuid := uuid.New().String()
	prBranchName := "codacy-cli-fix-" + uuid

	cmd := exec.Command("/bin/sh")
	cmd.Env = append(cmd.Env, "PR_BRANCH_NAME="+prBranchName)

	if dryRun {
		log.Println("Would create PR with branch name " + prBranchName)
		log.Println("Commands:")
		log.Println(prCreatorScriptContent)
		return false
	} else {
		tmpFile, err := os.CreateTemp(os.TempDir(), "prCreator*.sh")
		defer os.Remove(tmpFile.Name())
		if err != nil {
			log.Fatal("Error creating the temporary file for the script:", err)
		}

		_, err = tmpFile.WriteString(prCreatorScriptContent)
		if err != nil {
			log.Fatal("Error writing the script temporary file:", err)
		}

		// Execute the script
		cmd := exec.Command("/bin/sh", tmpFile.Name(), prBranchName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal("Error executing script:", err)
		}

		return true
	}
}
