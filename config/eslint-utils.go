package config

import (
	"fmt"
	"log"
	"os/exec"
	"path"
)

func genInfoEslint(r *Runtime) map[string]string {
	eslintFolder := fmt.Sprintf("%s@%s", r.Name(), r.Version())
	installDir := path.Join(Config.ToolsDirectory(), eslintFolder)

	return map[string]string{
		"installDir": installDir,
		"eslint": path.Join(installDir, "node_modules", ".bin", "eslint"),
	}
}

/*
 * This installs eslint using node's npm alongside its sarif extension
 */
func InstallEslint(nodeRuntime *Runtime, eslint *Runtime) error {
	log.Println("Installing ESLint")

	eslintInstallArg := fmt.Sprintf("%s@%s", eslint.Name(), eslint.Version())
	cmd := exec.Command(nodeRuntime.Info()["npm"], "install", "--prefix", eslint.Info()["installDir"],
		eslintInstallArg, "@microsoft/eslint-formatter-sarif")
	// to use the chdir command we needed to create the folder before, we can change this after
	// cmd.Dir = eslintInstallationFolder
	stdout, err := cmd.Output()
	// Print the output
	fmt.Println(string(stdout))
	return err
}
