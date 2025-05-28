package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLanguageDetector(t *testing.T) {
	detector := NewLanguageDetector()

	// Test that the detector is initialized with known languages
	expectedLanguages := []string{
		"JavaScript", "TypeScript", "Python", "Java", "Go",
		"Ruby", "PHP", "C", "C++", "C#", "Dart", "Kotlin",
		"Swift", "Scala", "Rust", "Shell", "HTML", "CSS",
		"XML", "JSON", "YAML", "Markdown", "Terraform",
	}

	for _, lang := range expectedLanguages {
		_, exists := detector.languages[lang]
		assert.True(t, exists, "Expected language %s to be initialized", lang)
	}

	// Test that extension mappings are correctly set up
	testCases := []struct {
		ext      string
		language string
	}{
		{".js", "JavaScript"},
		{".py", "Python"},
		{".java", "Java"},
		{".go", "Go"},
		{".rb", "Ruby"},
		{".php", "PHP"},
		{".ts", "TypeScript"},
		{".tsx", "TypeScript"},
		{".jsx", "JavaScript"},
		{".cpp", "C++"},
		{".cs", "C#"},
		{".dart", "Dart"},
		{".kt", "Kotlin"},
		{".swift", "Swift"},
		{".scala", "Scala"},
		{".rs", "Rust"},
		{".tf", "Terraform"},
	}

	for _, tc := range testCases {
		lang, exists := detector.extensionMap[tc.ext]
		assert.True(t, exists, "Expected extension %s to be mapped", tc.ext)
		assert.Equal(t, tc.language, lang, "Expected extension %s to map to language %s", tc.ext, tc.language)
	}
}

func TestDetectLanguages(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"src/main.go":         "package main",
		"src/app.js":          "console.log('hello')",
		"src/lib.py":          "print('hello')",
		"src/Main.java":       "class Main {}",
		"src/styles.css":      "body { margin: 0; }",
		"src/config.json":     "{}",
		"src/Dockerfile":      "FROM ubuntu",
		"src/app.dart":        "void main() {}",
		"src/test.rs":         "fn main() {}",
		"vendor/ignore.js":    "// should be ignored",
		"node_modules/pkg.js": "// should be ignored",
		".git/config":         "// should be ignored",
	}

	// Create the files in the temporary directory
	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Create and run the detector
	detector := NewLanguageDetector()
	languages, err := detector.DetectLanguages(tempDir)
	assert.NoError(t, err)

	// Verify detected languages
	expectedLanguages := map[string][]string{
		"Go":         {"src/main.go"},
		"JavaScript": {"src/app.js"},
		"Python":     {"src/lib.py"},
		"Java":       {"src/Main.java"},
		"CSS":        {"src/styles.css"},
		"JSON":       {"src/config.json"},
		"Dart":       {"src/app.dart"},
		"Rust":       {"src/test.rs"},
	}

	// Check that we found all expected languages
	assert.Equal(t, len(expectedLanguages), len(languages))

	// Check each language's files
	for langName, expectedFiles := range expectedLanguages {
		lang, exists := languages[langName]
		assert.True(t, exists, "Language %s should be detected", langName)
		if exists {
			assert.ElementsMatch(t, expectedFiles, lang.Files, "Files for language %s should match", langName)
		}
	}

	// Verify that ignored directories were actually ignored
	for langName, lang := range languages {
		for _, file := range lang.Files {
			assert.NotContains(t, file, "vendor/", "Language %s should not contain files from vendor/", langName)
			assert.NotContains(t, file, "node_modules/", "Language %s should not contain files from node_modules/", langName)
			assert.NotContains(t, file, ".git/", "Language %s should not contain files from .git/", langName)
		}
	}
}

func TestDetectLanguages_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	detector := NewLanguageDetector()
	languages, err := detector.DetectLanguages(tempDir)
	assert.NoError(t, err)
	assert.Empty(t, languages, "Empty directory should not detect any languages")
}

func TestDetectLanguages_NonExistentDirectory(t *testing.T) {
	detector := NewLanguageDetector()
	languages, err := detector.DetectLanguages("/path/that/does/not/exist")
	assert.Error(t, err)
	assert.Nil(t, languages)
}

func TestDetectLanguages_WithGitignore(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a .gitignore file
	gitignoreContent := `
# Ignore build directory
build/
# Ignore all .log files
*.log
# Ignore specific file
ignored.js
# Ignore test files
**/*.test.js
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644)
	assert.NoError(t, err)

	// Create test files
	testFiles := map[string]string{
		"src/app.js":           "console.log('hello')",
		"build/output.js":      "// should be ignored",
		"debug.log":            "// should be ignored",
		"ignored.js":           "// should be ignored",
		"src/not-ignored.js":   "// should be included",
		"src/test/app.test.js": "// should be ignored",
		"src/lib.py":           "print('hello')",
		"src/Main.java":        "class Main {}",
	}

	// Create the files in the temporary directory
	for path, content := range testFiles {
		fullPath := filepath.Join(tempDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		assert.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Create and run the detector
	detector := NewLanguageDetector()
	languages, err := detector.DetectLanguages(tempDir)
	assert.NoError(t, err)

	// Verify detected languages and files
	expectedLanguages := map[string][]string{
		"JavaScript": {"src/app.js", "src/not-ignored.js"},
		"Python":     {"src/lib.py"},
		"Java":       {"src/Main.java"},
	}

	// Check that we found all expected languages
	assert.Equal(t, len(expectedLanguages), len(languages))

	// Check each language's files
	for langName, expectedFiles := range expectedLanguages {
		lang, exists := languages[langName]
		assert.True(t, exists, "Language %s should be detected", langName)
		if exists {
			assert.ElementsMatch(t, expectedFiles, lang.Files, "Files for language %s should match", langName)
		}
	}

	// Verify that ignored files were actually ignored
	for _, lang := range languages {
		for _, file := range lang.Files {
			assert.NotContains(t, file, "build/", "Should not contain files from build/")
			assert.NotContains(t, file, ".log", "Should not contain .log files")
			assert.NotEqual(t, file, "ignored.js", "Should not contain the ignored.js file")
			assert.NotContains(t, file, ".test.js", "Should not contain test files")
		}
	}
}
