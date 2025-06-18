package constants

// Tool configuration file names - shared constants to avoid duplication
const (
	// Language and project configuration files
	LanguagesConfigFileName = "languages-config.yaml"
	GitIgnoreFileName       = ".gitignore"

	// Tool-specific configuration files
	ESLintConfigFileName       = "eslint.config.mjs"
	TrivyConfigFileName        = "trivy.yaml"
	PMDConfigFileName          = "ruleset.xml"
	PylintConfigFileName       = "pylint.rc"
	DartAnalyzerConfigFileName = "analysis_options.yaml"
	SemgrepConfigFileName      = "semgrep.yaml"
	ReviveConfigFileName       = "revive.toml"
	LizardConfigFileName       = "lizard.yaml"
)

// ToolConfigFileNames maps tool names to their configuration filenames
var ToolConfigFileNames = map[string]string{
	"eslint":       ESLintConfigFileName,
	"trivy":        TrivyConfigFileName,
	"pmd":          PMDConfigFileName,
	"pylint":       PylintConfigFileName,
	"dartanalyzer": DartAnalyzerConfigFileName,
	"semgrep":      SemgrepConfigFileName,
	"revive":       ReviveConfigFileName,
	"lizard":       LizardConfigFileName,
}
