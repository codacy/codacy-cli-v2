package main

import (
	"codacy/cli-v2/cmd"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mholt/archiver/v4"
	"gopkg.in/yaml.v3"
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

func extract(t *os.File) {

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

		// fmt.Printf("isDir: %s\n", f.IsDir())

		f.Mode()

		switch f.IsDir() {
		case true:
			// create a directory
			fmt.Println("creating:   " + f.NameInArchive)
			err := os.MkdirAll(f.NameInArchive, 0777)
			if err != nil {
				log.Fatal(err)
			}

		case false:

			if f.LinkTarget != "" {
				err := os.Symlink(f.LinkTarget, f.NameInArchive)
				if err != nil {
					log.Fatal(err)
				}

				return nil
			}

			// write a file
			fmt.Println("extracting: " + f.NameInArchive)
			fmt.Println("targe link: " + f.LinkTarget)
			w, err := os.OpenFile(f.NameInArchive, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
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

func main() {
	content, err := os.ReadFile(".codacy/codacy.yaml")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Println(config)
	downloadURL := getNodeDownloadURL("v22.2.0")
	nodeTar, _ := downloadFile(downloadURL, ".codacy")
	fmt.Println(nodeTar)

	t, _ := os.Open(nodeTar)
	defer t.Close()

	extract(t)

	cmd.Execute()
}
