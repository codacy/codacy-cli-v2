package domain

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// LanguageInfo represents information about a detected language
type LanguageInfo struct {
	Name       string
	Extensions []string
	Files      []string
}

// LanguageDetector handles language detection in a project
type LanguageDetector struct {
	// Map of language name to language info
	languages map[string]*LanguageInfo
	// Map of extension to language name
	extensionMap map[string]string
	// GitIgnore handler
	gitIgnore *gitignore.GitIgnore
}

// NewLanguageDetector creates a new language detector with predefined language mappings
func NewLanguageDetector() *LanguageDetector {
	detector := &LanguageDetector{
		languages:    make(map[string]*LanguageInfo),
		extensionMap: make(map[string]string),
	}

	// Initialize with known languages and their extensions
	detector.addLanguage("JavaScript", []string{".js", ".jsx", ".mjs"})
	detector.addLanguage("TypeScript", []string{".ts", ".tsx"})
	detector.addLanguage("Python", []string{".py", ".pyi", ".pyw"})
	detector.addLanguage("Java", []string{".java"})
	detector.addLanguage("Go", []string{".go"})
	detector.addLanguage("Ruby", []string{".rb", ".rake", ".gemspec"})
	detector.addLanguage("PHP", []string{".php"})
	detector.addLanguage("C", []string{".c", ".h"})
	detector.addLanguage("C++", []string{".cpp", ".hpp", ".cc", ".hh"})
	detector.addLanguage("C#", []string{".cs"})
	detector.addLanguage("Dart", []string{".dart"})
	detector.addLanguage("Kotlin", []string{".kt", ".kts"})
	detector.addLanguage("Swift", []string{".swift"})
	detector.addLanguage("Scala", []string{".scala", ".sc"})
	detector.addLanguage("Rust", []string{".rs"})
	detector.addLanguage("Shell", []string{".sh", ".bash"})
	detector.addLanguage("HTML", []string{".html", ".htm"})
	detector.addLanguage("CSS", []string{".css", ".scss", ".sass", ".less"})
	detector.addLanguage("XML", []string{".xml"})
	detector.addLanguage("JSON", []string{".json"})
	detector.addLanguage("YAML", []string{".yml", ".yaml"})
	detector.addLanguage("Markdown", []string{".md", ".markdown"})
	detector.addLanguage("Dockerfile", []string{"Dockerfile"})
	detector.addLanguage("Terraform", []string{".tf", ".tfvars"})

	return detector
}

// addLanguage adds a language and its extensions to the detector
func (d *LanguageDetector) addLanguage(name string, extensions []string) {
	d.languages[name] = &LanguageInfo{
		Name:       name,
		Extensions: extensions,
		Files:      make([]string, 0),
	}
	for _, ext := range extensions {
		d.extensionMap[ext] = name
	}
}

// DetectLanguages scans a directory and detects programming languages used
func (d *LanguageDetector) DetectLanguages(rootDir string) (map[string]*LanguageInfo, error) {
	// Initialize gitignore handler
	gitIgnorePath := filepath.Join(rootDir, ".gitignore")
	var err error
	d.gitIgnore, err = gitignore.CompileIgnoreFile(gitIgnorePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Reset files for each language
	for _, lang := range d.languages {
		lang.Files = make([]string, 0)
	}

	// Walk through the directory
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip vendor directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}

		// Get relative path for gitignore matching
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Skip files from .gitignore
		if d.gitIgnore != nil && d.gitIgnore.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			// For files without extension, use the filename (e.g., "Dockerfile")
			ext = strings.ToLower(info.Name())
		}

		// Check if extension is mapped to a language
		if langName, ok := d.extensionMap[ext]; ok {
			// Add file to language info
			if lang, exists := d.languages[langName]; exists {
				lang.Files = append(lang.Files, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Filter out languages with no files
	result := make(map[string]*LanguageInfo)
	for name, lang := range d.languages {
		if len(lang.Files) > 0 {
			result[name] = lang
		}
	}

	return result, nil
}
