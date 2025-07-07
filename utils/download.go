package utils

import (
	"codacy/cli-v2/utils/logger"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func DownloadFile(url string, destDir string) (string, error) {
	logger.Debug("Attempting to download from URL", logrus.Fields{
		"url": url,
	})

	// Get the file name from the URL
	fileName := filepath.Base(url)
	logger.Debug("Target filename determined", logrus.Fields{
		"fileName": fileName,
	})

	// Create the destination file path
	destPath := filepath.Join(destDir, fileName)
	logger.Debug("Destination path set", logrus.Fields{
		"destPath": destPath,
	})

	_, errInfo := os.Stat(destPath)
	if errInfo != nil && os.IsExist(errInfo) {
		logger.Debug("File already exists at destination, skipping download", logrus.Fields{
			"destPath": destPath,
		})
		return destPath, nil
	}

	// Create the destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Make the HTTP GET request
	logger.Debug("Making HTTP GET request", logrus.Fields{
		"url": url,
	})
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
	logger.Debug("Downloading file content")
	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}
	logger.Debug("Downloaded file successfully", logrus.Fields{
		"bytes": written,
	})

	if written == 0 {
		return "", fmt.Errorf("downloaded file is empty (0 bytes)")
	}

	return destPath, nil
}
