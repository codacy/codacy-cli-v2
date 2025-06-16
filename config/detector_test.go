package config

import (
	"os"
	"path/filepath"
	"testing"

	"codacy/cli-v2/domain"

	"github.com/stretchr/testify/assert"
)

func TestDetectFileExtensions_SingleFile(t *testing.T) {
	// Create a temporary file
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.py")
	err = os.WriteFile(testFile, []byte("print('hello')"), 0644)
	assert.NoError(t, err)

	// Test detecting extensions for a single file
	extCount, err := DetectFileExtensions(testFile)
	assert.NoError(t, err)
	assert.Equal(t, map[string]int{".py": 1}, extCount)
}

func TestDetectFileExtensions_Directory(t *testing.T) {
	// Create a temporary directory with multiple files
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create files with different extensions
	files := map[string]string{
		"main.py":     "print('hello')",
		"script.js":   "console.log('hello')",
		"test.py":     "def test(): pass",
		"style.css":   "body { color: red; }",
		"data.json":   "{}",
		"README.md":   "# Title",
		"config.yaml": "key: value",
	}

	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Test detecting extensions for directory
	extCount, err := DetectFileExtensions(tempDir)
	assert.NoError(t, err)

	expected := map[string]int{
		".py":   2, // main.py and test.py
		".js":   1, // script.js
		".css":  1, // style.css
		".json": 1, // data.json
		".md":   1, // README.md
		".yaml": 1, // config.yaml
	}
	assert.Equal(t, expected, extCount)
}

func TestDetectFileExtensions_NonExistentPath(t *testing.T) {
	_, err := DetectFileExtensions("/non/existent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to stat path")
}

func TestDetectFileExtensions_EmptyDirectory(t *testing.T) {
	// Create empty temporary directory
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	extCount, err := DetectFileExtensions(tempDir)
	assert.NoError(t, err)
	assert.Empty(t, extCount)
}

func TestDetectRelevantTools_PythonFile(t *testing.T) {
	// Create a temporary Python file
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.py")
	err = os.WriteFile(testFile, []byte("print('hello')"), 0644)
	assert.NoError(t, err)

	// Create test tool language map
	toolLangMap := map[string]domain.ToolLanguageInfo{
		"pylint": {
			Name:       "pylint",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
		"eslint": {
			Name:       "eslint",
			Languages:  []string{"JavaScript", "TypeScript"},
			Extensions: []string{".js", ".jsx", ".ts", ".tsx"},
		},
		"lizard": {
			Name:       "lizard",
			Languages:  []string{"Python", "JavaScript", "Java", "C"},
			Extensions: []string{".py", ".js", ".java", ".c"},
		},
	}

	// Test detecting relevant tools
	relevantTools, err := DetectRelevantTools(testFile, toolLangMap)
	assert.NoError(t, err)

	// Should detect pylint and lizard (both support .py)
	expected := map[string]struct{}{
		"pylint": {},
		"lizard": {},
	}
	assert.Equal(t, expected, relevantTools)
}

func TestDetectRelevantTools_JavaScriptFile(t *testing.T) {
	// Create a temporary JavaScript file
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.js")
	err = os.WriteFile(testFile, []byte("console.log('hello')"), 0644)
	assert.NoError(t, err)

	// Create test tool language map
	toolLangMap := map[string]domain.ToolLanguageInfo{
		"pylint": {
			Name:       "pylint",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
		"eslint": {
			Name:       "eslint",
			Languages:  []string{"JavaScript", "TypeScript"},
			Extensions: []string{".js", ".jsx", ".ts", ".tsx"},
		},
		"lizard": {
			Name:       "lizard",
			Languages:  []string{"Python", "JavaScript", "Java", "C"},
			Extensions: []string{".py", ".js", ".java", ".c"},
		},
	}

	// Test detecting relevant tools
	relevantTools, err := DetectRelevantTools(testFile, toolLangMap)
	assert.NoError(t, err)

	// Should detect eslint and lizard (both support .js)
	expected := map[string]struct{}{
		"eslint": {},
		"lizard": {},
	}
	assert.Equal(t, expected, relevantTools)
}

func TestDetectRelevantTools_MultipleFiles(t *testing.T) {
	// Create a temporary directory with multiple files
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create files with different extensions
	files := map[string]string{
		"main.py":   "print('hello')",
		"script.js": "console.log('hello')",
		"App.java":  "public class App {}",
	}

	for filename, content := range files {
		filePath := filepath.Join(tempDir, filename)
		err = os.WriteFile(filePath, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Create test tool language map
	toolLangMap := map[string]domain.ToolLanguageInfo{
		"pylint": {
			Name:       "pylint",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
		"eslint": {
			Name:       "eslint",
			Languages:  []string{"JavaScript", "TypeScript"},
			Extensions: []string{".js", ".jsx", ".ts", ".tsx"},
		},
		"pmd": {
			Name:       "pmd",
			Languages:  []string{"Java", "JavaScript"},
			Extensions: []string{".java", ".js"},
		},
		"lizard": {
			Name:       "lizard",
			Languages:  []string{"Python", "JavaScript", "Java", "C"},
			Extensions: []string{".py", ".js", ".java", ".c"},
		},
	}

	// Test detecting relevant tools
	relevantTools, err := DetectRelevantTools(tempDir, toolLangMap)
	assert.NoError(t, err)

	// Should detect all tools since we have .py, .js, and .java files
	expected := map[string]struct{}{
		"pylint": {}, // supports .py
		"eslint": {}, // supports .js
		"pmd":    {}, // supports .java and .js
		"lizard": {}, // supports .py, .js, and .java
	}
	assert.Equal(t, expected, relevantTools)
}

func TestDetectRelevantTools_NoMatchingTools(t *testing.T) {
	// Create a temporary file with unsupported extension
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "README.txt")
	err = os.WriteFile(testFile, []byte("Some text"), 0644)
	assert.NoError(t, err)

	// Create test tool language map (none support .txt)
	toolLangMap := map[string]domain.ToolLanguageInfo{
		"pylint": {
			Name:       "pylint",
			Languages:  []string{"Python"},
			Extensions: []string{".py"},
		},
		"eslint": {
			Name:       "eslint",
			Languages:  []string{"JavaScript", "TypeScript"},
			Extensions: []string{".js", ".jsx", ".ts", ".tsx"},
		},
	}

	// Test detecting relevant tools
	relevantTools, err := DetectRelevantTools(testFile, toolLangMap)
	assert.NoError(t, err)
	assert.Empty(t, relevantTools)
}

func TestDetectRelevantTools_EmptyToolMap(t *testing.T) {
	// Create a temporary Python file
	tempDir, err := os.MkdirTemp("", "codacy-detector-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.py")
	err = os.WriteFile(testFile, []byte("print('hello')"), 0644)
	assert.NoError(t, err)

	// Test with empty tool map
	toolLangMap := map[string]domain.ToolLanguageInfo{}

	relevantTools, err := DetectRelevantTools(testFile, toolLangMap)
	assert.NoError(t, err)
	assert.Empty(t, relevantTools)
}

func TestGetSortedKeys(t *testing.T) {
	// Test with sample map
	testMap := map[string]struct{}{
		"zebra":  {},
		"apple":  {},
		"banana": {},
	}

	result := GetSortedKeys(testMap)
	expected := []string{"apple", "banana", "zebra"}
	assert.Equal(t, expected, result)
}

func TestGetSortedKeys_EmptyMap(t *testing.T) {
	testMap := map[string]struct{}{}
	result := GetSortedKeys(testMap)
	assert.Empty(t, result)
}

func TestGetRecognizableExtensions(t *testing.T) {
	// Create sample extension count and tool map
	extCount := map[string]int{
		".py":   3,
		".js":   2,
		".java": 1,
		".txt":  5, // Not supported by any tool
	}

	toolLangMap := map[string]domain.ToolLanguageInfo{
		"pylint": {
			Extensions: []string{".py"},
		},
		"eslint": {
			Extensions: []string{".js", ".jsx"},
		},
		"pmd": {
			Extensions: []string{".java", ".js"},
		},
	}

	result := GetRecognizableExtensions(extCount, toolLangMap)

	// Should return extensions that are supported by tools, sorted by count (desc) then name
	expected := []string{
		".py (3 files)",   // highest count
		".js (2 files)",   // second highest count
		".java (1 files)", // lowest count
		// .txt should not appear as it's not supported by any tool
	}
	assert.Equal(t, expected, result)
}
