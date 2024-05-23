package utils

import (
	"fmt"
	"io"
	"log"

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
	cmd.Env = append(cmd.Env, "PR_BRANCH_NAME=" + prBranchName)

	if dryRun {
		fmt.Println("Would create PR with branch name " + prBranchName)
		fmt.Println("Commands:")
		fmt.Println(prCreatorScriptContent)
		return false
	} else {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(stdin, prCreatorScriptContent)
		stdin.Close()

		stdout, err := cmd.Output()
		if err != nil {
			 log.Fatal(err, string(stdout))
			 return false
		 }

		// Print the output
		fmt.Println(string(stdout))
		return true
	}
}