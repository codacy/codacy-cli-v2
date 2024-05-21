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

// https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-arm64.tar.xz
// https://nodejs.org/dist/v22.2.0/node-v22.2.0-darwin-x64.tar.gz
// https://nodejs.org/dist/v13.14.0/node-v13.14.0-win-x64.zip
// https://nodejs.org/dist/v22.2.0/node-v22.2.0-linux-armv7l.tar.xz
func getNodeDownloadURL(version string) string {
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

	// Construct the Node.js download URL
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}

	downloadURL := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-%s-%s.%s", version, version, goos, nodeArch, extension)
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

func extract(t *os.File, codacyDirectory string) {

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	// format, _, err := archiver.Identify(t.Name(), nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// the list of files we want out of the archive; any
	// directories will include all their contents unless
	// we return fs.SkipDir from our handler
	// (leave this nil to walk ALL files from the archive)
	// fileList := []string{"file1.txt", "subfolder"}

	handler := func(ctx context.Context, f archiver.File) error {

		fmt.Printf("Contents of %s:\n", f.NameInArchive)

		path := filepath.Join(codacyDirectory, "runtimes", f.NameInArchive)

		switch f.IsDir() {
		case true:
			// create a directory
			fmt.Println("creating:   " + f.NameInArchive)
			err := os.MkdirAll(path, 0777)
			if err != nil {
				log.Fatal(err)
			}

		case false:

			if f.LinkTarget != "" {
				os.Remove(path)
				err := os.Symlink(f.LinkTarget, path)
				if err != nil {
					log.Fatal(err)
				}

				return nil
			}

			// write a file
			fmt.Println("extracting: " + f.NameInArchive)
			fmt.Println("targe link: " + f.LinkTarget)
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

func installESLint(npmExecutablePath string, ESLintversion string, codacyPath string) {

	fmt.Println("Installing ESLint")

	eslintInstallationFolder := filepath.Join(codacyPath, "tools", ESLintversion)

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

func fetchRuntimes(runtimes map[string]*config.Runtime) {
	for _, runtime := range runtimes {
		switch runtime.Name() {
		case "node":
			fmt.Println("Fetching node...")
		default:
			fmt.Println("Unknown runtime:", runtime.Name())
		}
	}
}

func main() {
	_, configErr := config.ReadConfigFile(".codacy/codacy.yaml")
	if configErr != nil {
		log.Fatal(configErr)
		return
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	codacyDirectory := filepath.Join(homePath, ".cache", "codacy")
	runtimesDirectory := filepath.Join(codacyDirectory, "runtimes")
	toolsDirectory := filepath.Join(codacyDirectory, "tools")

	fmt.Println("creating:   " + codacyDirectory)
	err = os.MkdirAll(codacyDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating:   " + runtimesDirectory)
	err = os.MkdirAll(runtimesDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating:   " + toolsDirectory)
	err = os.MkdirAll(toolsDirectory, 0777)
	if err != nil {
		log.Fatal(err)
	}

	downloadNodeURL := getNodeDownloadURL("v22.2.0")

	nodeTar, err := downloadFile(downloadNodeURL, codacyDirectory)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Downloaded node: " + nodeTar)

	t, err := os.Open(nodeTar)
	defer t.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("About to extract node: " + t.Name())
	extract(t, codacyDirectory)

	npmPath := filepath.Join(codacyDirectory, "runtimes", "node-v22.2.0-darwin-x64", "bin", "npm")

	fmt.Println("About to install eslint")
	installESLint(npmPath, "eslint@9.3.0", codacyDirectory)

	cmd.Execute()
}
