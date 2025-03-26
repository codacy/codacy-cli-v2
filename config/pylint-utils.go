package config

import (
	"fmt"
	"log"
	"os/exec"
	"path"
)

func getInfoPylint() map[string]string {
	pythonRuntime := Config.Runtimes()["python"]

	pythonFolder := fmt.Sprintf("%s@%s", pythonRuntime.Name(), pythonRuntime.Version())
	pythonInstallDir := path.Join(Config.RuntimesDirectory(), pythonFolder, "python")
	pylintPath := path.Join(pythonInstallDir, "bin", "pylint")

	return map[string]string{
		"installDir": pythonInstallDir,
		"pylint":     pylintPath,
	}

}

// installing in the python runtime because
// f you install Pylint in a different tools folder, it will not work properly because of the following reasons:
// Python Virtual Environment Isolation:
// When you install Pylint in the tools folder (separately from Python), you are essentially mixing environments.
// The python binary located in /Users/yasmin/.cache/codacy/runtimes/python@3.10.16/python/bin/python3 will not have
// access to packages installed elsewhere unless properly referenced.
// PYTHONPATH Limitation:
// You tried passing the tools directory via PYTHONPATH.
// While PYTHONPATH allows modules to be found, it doesn't register packages installed by pip properly.
// Pylint Binary (pylint):
// The pylint binary expects the pylint package to be installed within the same Python environment that is running it.
// If you run:
// /Users/.cache/codacy/runtimes/python@3.10.16/python/bin/python3 -m pylint
// It will look for the pylint module installed under its site-packages directory within:
// /Users/.cache/codacy/runtimes/python@3.10.16/python/lib/python3.10/site-packages
func InstallPylint(pythonRuntime *Runtime, pylint *Runtime) error {
	log.Println("Installing Pylint")

	pythonInfo := getInfoPython(pythonRuntime)

	pythonBinary := pythonInfo["python"]

	// to install pylint using oython binary
	cmd := exec.Command(pythonBinary, "-m", "pip", "install",
		fmt.Sprintf("pylint==%s", pylint.Version()),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error installing Pylint: %v\nOutput: %s", err, string(output))
	}

	log.Println("Pylint installed successfully")
	return nil
}
