package main

import (
    "archive/tar"
    "codacy/cli-v2/cmd"
    "compress/gzip"
    "fmt"
    "gopkg.in/yaml.v3"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "runtime"
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
    nodeTar, err := downloadFile(downloadURL, ".codacy")
    fmt.Println(nodeTar)

    t, _ := os.Open(nodeTar)
    defer t.Close()
    uncompressedStream, _ := gzip.NewReader(t)

    tr := tar.NewReader(uncompressedStream)
    for {
        hdr, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Contents of %s:\n", hdr.Name)

        outFile, err := os.Create(".codacy/node/" + hdr.Name)
        if err != nil {
            break
        }

        if _, err := io.Copy(outFile, tr); err != nil {
            log.Fatal(err)
        }
        fmt.Println()
    }

    cmd.Execute()
}
