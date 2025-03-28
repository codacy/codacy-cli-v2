package config

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"codacy/cli-v2/plugins"
)

func genInfoEslint(r *Runtime) map[string]string {
	eslintFolder := fmt.Sprintf("%s@%s", r.Name(), r.Version())
	installDir := path.Join(Config.ToolsDirectory(), eslintFolder)

	return map[string]string{
		"installDir": installDir,
		"eslint":     path.Join(installDir, "node_modules", ".bin", "eslint"),
	}
}

/*
 * This installs eslint using node's npm alongside its sarif extension
 */
func InstallEslint(nodeRuntime *plugins.RuntimeInfo, eslint *Runtime, registry string) error {
	log.Println("Installing ESLint")

	eslintInstallArg := fmt.Sprintf("%s@%s", eslint.Name(), eslint.Version())
	if registry != "" {
		fmt.Println("Using registry:", registry)
		configCmd := exec.Command(nodeRuntime.Binaries["npm"], "config", "set", "registry", registry)
		if configOut, err := configCmd.Output(); err != nil {
			fmt.Println("Error setting npm registry:", err)
			fmt.Println(string(configOut))
			return err
		}
	}
	cmd := exec.Command(nodeRuntime.Binaries["npm"], "install", "--prefix", eslint.Info()["installDir"],
		eslintInstallArg, "@microsoft/eslint-formatter-sarif")
	fmt.Println(cmd.String())
	// to use the chdir command we needed to create the folder before, we can change this after
	// cmd.Dir = eslintInstallationFolder
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("Error installing ESLint:", err)
		fmt.Println(string(stdout))
	}
	// Print the output
	fmt.Println(string(stdout))
	return err
}
