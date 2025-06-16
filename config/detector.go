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

// DetectLanguages detects languages based on file extensions found in the path
func DetectLanguages(rootPath string, toolLangMap map[string]domain.ToolLanguageInfo) (map[string]struct{}, error) {
	detectedLangs := make(map[string]struct{})
	extToLang := make(map[string][]string)

	// Build extension to language mapping based on file extensions, not tool capabilities
	// Each extension should map to its actual language, not all languages that tools supporting it can handle
	extensionLanguageMap := map[string][]string{
		".java":      {"Java"},
		".py":        {"Python"},
		".js":        {"JavaScript"},
		".jsx":       {"JSX"},
		".ts":        {"TypeScript"},
		".tsx":       {"TSX"},
		".go":        {"Go", "Golang"}, // Both Go and Golang are used in tool definitions
		".dart":      {"Dart"},
		".c":         {"C"},
		".cpp":       {"CPP"},
		".cc":        {"CPP"},
		".h":         {"C", "CPP"},
		".hpp":       {"CPP"},
		".cs":        {"C#"},
		".rb":        {"Ruby"},
		".php":       {"PHP"},
		".scala":     {"Scala"},
		".swift":     {"Swift"},
		".kt":        {"Kotlin"},
		".rs":        {"Rust"},
		".lua":       {"Lua"},
		".pl":        {"Perl"},
		".f":         {"Fortran"},
		".f90":       {"Fortran"},
		".erl":       {"Erlang"},
		".sol":       {"Solidity"},
		".zig":       {"Zig"},
		".m":         {"Objective-C"},
		".vue":       {"VueJS"},
		".ttcn":      {"TTCN-3"},
		".gd":        {"GDScript"},
		".json":      {"JSON"},
		".xml":       {"XML"},
		".jsp":       {"JSP"},
		".vm":        {"Velocity"},
		".cls":       {"Apex"},
		".trigger":   {"Apex"},
		".page":      {"VisualForce"},
		".component": {"VisualForce"},
		".tf":        {"Terraform"},
		".tfvars":    {"Terraform"},
	}

	// Use the precise extension mapping instead of tool-based mapping
	for ext, langs := range extensionLanguageMap {
		extToLang[ext] = langs
	}

	// Get file extensions from the path
	extCount, err := DetectFileExtensions(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file extensions in path %s: %w", rootPath, err)
	}

	// Map only found extensions to languages
	for ext := range extCount {
		if langs, ok := extToLang[ext]; ok {
			// Log which extensions map to which languages for debugging
			logger.Debug("Found files with extension", logrus.Fields{
				"extension": ext,
				"count":     extCount[ext],
				"languages": langs,
			})
			for _, lang := range langs {
				detectedLangs[lang] = struct{}{}
			}
		}
	}

	// Log the final set of detected languages with their corresponding extensions
	if len(detectedLangs) > 0 {
		langToExts := make(map[string][]string)
		for ext, count := range extCount {
			if langs, ok := extToLang[ext]; ok {
				for _, lang := range langs {
					langToExts[lang] = append(langToExts[lang], fmt.Sprintf("%s (%d files)", ext, count))
				}
			}
		}

		logger.Debug("Detected languages in path", logrus.Fields{
			"languages_with_files": langToExts,
			"path":                 rootPath,
		})
	}

	return detectedLangs, nil
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
