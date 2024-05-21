package main

import (
	"codacy/cli-v2/cmd"
	"codacy/cli-v2/config"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver/v4"
)

func getNodeFileName(nodeRuntime *config.Runtime) string {
	// Detect the OS and architecture
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go architecture to Node.js architecture
	var nodeArch string
	switch goarch {
	case "386":
		nodeArch = "x86"
	case "amd64":
		nodeArch = "x64"
	case "arm":
		nodeArch = "armv7l"
	case "arm64":
		nodeArch = "arm64"
	default:
		nodeArch = goarch
	}

	return fmt.Sprintf("node-v%s-%s-%s", nodeRuntime.Version(), goos, nodeArch)
}

func getNodeDownloadURL(nodeRuntime *config.Runtime) string {
	// Detect the OS and architecture
	goos := runtime.GOOS

	// Construct the Node.js download URL
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	downloadURL := fmt.Sprintf("https://nodejs.org/dist/v%s/%s.%s", nodeRuntime.Version(), getNodeFileName(nodeRuntime), extension)
	return downloadURL
}

func downloadFile(url string, destDir string) (string, error) {
	// Get the file name from the URL
	fileName := filepath.Base(url)

	// Create the destination file path
	destPath := filepath.Join(destDir, fileName)

	_, errInfo := os.Stat(destPath)
	if errInfo != nil && os.IsExist(errInfo) {
		return destPath, nil
	}
	// Create the destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}

	// Copy the response body to the destination file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	return destPath, nil
}

func extract(t *os.File, targetDir string) {
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	handler := func(ctx context.Context, f archiver.File) error {
		path := filepath.Join(targetDir, f.NameInArchive)

		switch f.IsDir() {
		case true:
			// create a directory
			fmt.Println("creating:   " + f.NameInArchive)
			err := os.MkdirAll(path, 0777)
			if err != nil {
				log.Fatal(err)
			}

		case false:
			log.Print("extracting: " + f.NameInArchive)

			// if is a symlink
			if f.LinkTarget != "" {
				os.Remove(path)
				err := os.Symlink(f.LinkTarget, path)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			}

			// write a file
			w, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				log.Fatal(err)
			}

			stream, _ := f.Open()
			defer stream.Close()

			_, err = io.Copy(w, stream)
			if err != nil {
				log.Fatal(err)
			}
			w.Close()
		}

		return nil
	}

	err := format.Extract(context.Background(), t, nil, handler)
	if err != nil {
		log.Fatal(err)
	}
}

func installESLint(npmExecutablePath string, ESLintversion string, toolsDirectory string) {
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

func fetchRuntimes(runtimes map[string]*config.Runtime, runtimesDirectory string) {
	for _, runtime := range runtimes {
		switch runtime.Name() {
		case "node":
			// TODO should delete downloaded archive
			// TODO check for deflated archive
			log.Println("Fetching node...")
			downloadNodeURL := getNodeDownloadURL(runtime)
			nodeTar, err := downloadFile(downloadNodeURL, runtimesDirectory)
			if err != nil {
				log.Fatal(err)
			}

			// deflate node archive
			t, err := os.Open(nodeTar)
			defer t.Close()
			if err != nil {
				log.Fatal(err)
			}
			extract(t, runtimesDirectory)
		default:
			log.Fatal("Unknown runtime:", runtime.Name())
		}
	}
}

func fetchTools(runtime *config.Runtime, runtimesDirectory string, toolsDirectory string) {
	for _, tool := range runtime.Tools() {
		switch tool.Name() {
		case "eslint":
			npmPath := filepath.Join(runtimesDirectory, getNodeFileName(runtime),
				"bin", "npm")
			installESLint(npmPath, "eslint@" + tool.Version(), toolsDirectory)
		default:
			log.Fatal("Unknown tool:", tool.Name())
		}
	}
}

func main() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	codacyDirectory := filepath.Join(homePath, ".cache", "codacy")
	runtimesDirectory := filepath.Join(codacyDirectory, "runtimes")
	toolsDirectory := filepath.Join(codacyDirectory, "tools")
	fmt.Println("creating: " + codacyDirectory)
	if os.MkdirAll(codacyDirectory, 0777) != nil {
		log.Fatal(err)
	}
	fmt.Println("creating: " + runtimesDirectory)
	if os.MkdirAll(runtimesDirectory, 0777) != nil {
		log.Fatal(err)
	}
	fmt.Println("creating: " + toolsDirectory)
	if os.MkdirAll(toolsDirectory, 0777) != nil {
		log.Fatal(err)
	}

	// TODO can use a variable to stored the "local" codacy dir
	runtimes, configErr := config.ReadConfigFile(filepath.Join(".codacy", "codacy.yaml"))
	if configErr != nil {
		log.Fatal(configErr)
	}

	// install runtimes
	fetchRuntimes(runtimes, runtimesDirectory)
	for _, r := range runtimes {
		fetchTools(r, runtimesDirectory, toolsDirectory)
	}

	cmd.Execute()
}
