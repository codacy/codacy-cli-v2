package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url string, destDir string) (string, error) {
	log.Printf("Attempting to download from URL: %s", url)

	// Get the file name from the URL
	fileName := filepath.Base(url)
	log.Printf("Target filename: %s", fileName)

	// Create the destination file path
	destPath := filepath.Join(destDir, fileName)
	log.Printf("Destination path: %s", destPath)

	_, errInfo := os.Stat(destPath)
	if errInfo != nil && os.IsExist(errInfo) {
		log.Printf("File already exists at destination, skipping download")
		return destPath, nil
	}

	// Create the destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Make the HTTP GET request
	log.Printf("Making HTTP GET request...")
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Codacy-CLI")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make GET request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to download file: status code %d, URL: %s, Response: %s", resp.StatusCode, url, string(body))
	}

	// Copy the response body to the destination file
	log.Printf("Downloading file content...")
	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}
	log.Printf("Downloaded %d bytes", written)

	if written == 0 {
		return "", fmt.Errorf("downloaded file is empty (0 bytes)")
	}

	return destPath, nil
}
