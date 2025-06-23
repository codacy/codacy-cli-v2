package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codacy/cli-v2/domain"
	"codacy/cli-v2/utils/logger"

	"github.com/sirupsen/logrus"
)

// DetectFileExtensions walks the directory and collects all unique file extensions with their counts
func DetectFileExtensions(rootPath string) (map[string]int, error) {
	extCount := make(map[string]int)

	// Check if rootPath is a file or directory
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", rootPath, err)
	}

	if !info.IsDir() {
		// If it's a single file, only process that file
		ext := strings.ToLower(filepath.Ext(rootPath))
		if ext != "" {
			extCount[ext] = 1
		}
		return extCount, nil
	}

	// If it's a directory, walk through it
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext != "" {
				extCount[ext]++
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk path %s: %w", rootPath, err)
	}

	return extCount, nil
}

// GetRecognizableExtensions returns a sorted list of extensions that are mapped to languages
func GetRecognizableExtensions(extCount map[string]int, toolLangMap map[string]domain.ToolLanguageInfo) []string {
	// Build set of recognized extensions
	recognizedExts := make(map[string]struct{})
	for _, toolInfo := range toolLangMap {
		for _, ext := range toolInfo.Extensions {
			recognizedExts[ext] = struct{}{}
		}
	}

	// Filter and format recognized extensions with counts
	type extInfo struct {
		ext   string
		count int
	}
	var recognizedExtList []extInfo

	for ext, count := range extCount {
		if _, ok := recognizedExts[ext]; ok {
			recognizedExtList = append(recognizedExtList, extInfo{ext, count})
		}
	}

	// Sort by count (descending) and then by extension name
	sort.Slice(recognizedExtList, func(i, j int) bool {
		if recognizedExtList[i].count != recognizedExtList[j].count {
			return recognizedExtList[i].count > recognizedExtList[j].count
		}
		return recognizedExtList[i].ext < recognizedExtList[j].ext
	})

	// Format extensions with their counts
	result := make([]string, len(recognizedExtList))
	for i, info := range recognizedExtList {
		result[i] = fmt.Sprintf("%s (%d files)", info.ext, info.count)
	}

	return result
}

// DetectRelevantTools detects tools based on file extensions found in the path
func DetectRelevantTools(rootPath string, toolLangMap map[string]domain.ToolLanguageInfo) (map[string]struct{}, error) {
	// Get file extensions from the path
	extCount, err := DetectFileExtensions(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file extensions in path %s: %w", rootPath, err)
	}

	// Find tools that support these extensions
	relevantTools := make(map[string]struct{})
	for toolName, toolInfo := range toolLangMap {
		for _, ext := range toolInfo.Extensions {
			if _, found := extCount[ext]; found {
				logger.Debug("Found relevant tool for extension", logrus.Fields{
					"tool":      toolName,
					"extension": ext,
					"count":     extCount[ext],
				})
				relevantTools[toolName] = struct{}{}
				break
			}
		}
	}

	if len(relevantTools) > 0 {
		logger.Debug("Detected relevant tools for path", logrus.Fields{
			"tools": GetSortedKeys(relevantTools),
			"path":  rootPath,
		})
	}

	return relevantTools, nil
}

// GetSortedKeys returns a sorted slice of strings from a string set
func GetSortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
