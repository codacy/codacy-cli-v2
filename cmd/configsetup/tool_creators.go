package configsetup

import (
	"fmt"
	"os"
	"path/filepath"

	"codacy/cli-v2/constants"
	"codacy/cli-v2/domain"
	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/lizard"
	"codacy/cli-v2/tools/pylint"
	reviveTool "codacy/cli-v2/tools/revive"
)

// toolConfigRegistry maps tool UUIDs to their configuration creators
var toolConfigRegistry = map[string]ToolConfigCreator{
	domain.ESLint:       &eslintConfigCreator{},
	domain.ESLint9:      &eslintConfigCreator{},
	domain.Trivy:        &trivyConfigCreator{},
	domain.PMD:          &pmdConfigCreator{},
	domain.PMD7:         &pmd7ConfigCreator{},
	domain.PyLint:       &pylintConfigCreator{},
	domain.DartAnalyzer: &dartAnalyzerConfigCreator{},
	domain.Semgrep:      &semgrepConfigCreator{},
	domain.Lizard:       &lizardConfigCreator{},
	domain.Revive:       &reviveConfigCreator{},
}

// writeConfigFile is a helper function to write configuration files with consistent error handling
func writeConfigFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, constants.DefaultFilePerms)
}

// eslintConfigCreator implements ToolConfigCreator for ESLint
type eslintConfigCreator struct{}

func (e *eslintConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	err := tools.CreateEslintConfig(toolsConfigDir, patterns)
	if err == nil {
		fmt.Println("ESLint configuration created based on Codacy settings. Ignoring plugin rules. ESLint plugins are not supported yet.")
	}
	return err
}

func (e *eslintConfigCreator) GetConfigFileName() string { return "eslint.config.mjs" }
func (e *eslintConfigCreator) GetToolName() string       { return "ESLint" }

// trivyConfigCreator implements ToolConfigCreator for Trivy
type trivyConfigCreator struct{}

func (t *trivyConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreateTrivyConfig(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.TrivyConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Trivy configuration created based on Codacy settings")
	}
	return err
}

func (t *trivyConfigCreator) GetConfigFileName() string { return constants.TrivyConfigFileName }
func (t *trivyConfigCreator) GetToolName() string       { return "Trivy" }

// pmdConfigCreator implements ToolConfigCreator for PMD
type pmdConfigCreator struct{}

func (p *pmdConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreatePmd6Config(patterns)
	return writeConfigFile(filepath.Join(toolsConfigDir, constants.PMDConfigFileName), []byte(configString))
}

func (p *pmdConfigCreator) GetConfigFileName() string { return constants.PMDConfigFileName }
func (p *pmdConfigCreator) GetToolName() string       { return "PMD" }

// pmd7ConfigCreator implements ToolConfigCreator for PMD7
type pmd7ConfigCreator struct{}

func (p *pmd7ConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreatePmd7Config(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.PMDConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("PMD7 configuration created based on Codacy settings")
	}
	return err
}

func (p *pmd7ConfigCreator) GetConfigFileName() string { return constants.PMDConfigFileName }
func (p *pmd7ConfigCreator) GetToolName() string       { return "PMD7" }

// pylintConfigCreator implements ToolConfigCreator for Pylint
type pylintConfigCreator struct{}

func (p *pylintConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := pylint.GeneratePylintRC(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.PylintConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Pylint configuration created based on Codacy settings")
	}
	return err
}

func (p *pylintConfigCreator) GetConfigFileName() string { return constants.PylintConfigFileName }
func (p *pylintConfigCreator) GetToolName() string       { return "Pylint" }

// dartAnalyzerConfigCreator implements ToolConfigCreator for Dart Analyzer
type dartAnalyzerConfigCreator struct{}

func (d *dartAnalyzerConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configString := tools.CreateDartAnalyzerConfig(patterns)
	err := writeConfigFile(filepath.Join(toolsConfigDir, constants.DartAnalyzerConfigFileName), []byte(configString))
	if err == nil {
		fmt.Println("Dart configuration created based on Codacy settings")
	}
	return err
}

func (d *dartAnalyzerConfigCreator) GetConfigFileName() string {
	return constants.DartAnalyzerConfigFileName
}
func (d *dartAnalyzerConfigCreator) GetToolName() string { return "Dart Analyzer" }

// semgrepConfigCreator implements ToolConfigCreator for Semgrep
type semgrepConfigCreator struct{}

func (s *semgrepConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	configData, err := tools.GetSemgrepConfig(patterns)
	if err != nil {
		return fmt.Errorf("failed to create Semgrep config: %v", err)
	}
	err = writeConfigFile(filepath.Join(toolsConfigDir, constants.SemgrepConfigFileName), configData)
	if err == nil {
		fmt.Println("Semgrep configuration created based on Codacy settings")
	}
	return err
}

func (s *semgrepConfigCreator) GetConfigFileName() string { return constants.SemgrepConfigFileName }
func (s *semgrepConfigCreator) GetToolName() string       { return "Semgrep" }

// lizardConfigCreator implements ToolConfigCreator for Lizard
type lizardConfigCreator struct{}

func (l *lizardConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	// patternDefinitions := make([]domain.PatternDefinition, len(patterns))
	// for i, pattern := range patterns {
	// 	patternDefinitions[i] = pattern.PatternDefinition
	// }
	err := lizard.CreateLizardConfig(toolsConfigDir, patterns)
	if err != nil {
		return fmt.Errorf("failed to create Lizard configuration: %w", err)
	}
	fmt.Println("Lizard configuration created based on Codacy settings")
	return nil
}

func (l *lizardConfigCreator) GetConfigFileName() string { return "lizard.json" }
func (l *lizardConfigCreator) GetToolName() string       { return "Lizard" }

// reviveConfigCreator implements ToolConfigCreator for Revive
type reviveConfigCreator struct{}

func (r *reviveConfigCreator) CreateConfig(toolsConfigDir string, patterns []domain.PatternConfiguration) error {
	err := createReviveConfigFile(patterns, toolsConfigDir)
	if err == nil {
		fmt.Println("Revive configuration created based on Codacy settings")
	}
	return err
}

func (r *reviveConfigCreator) GetConfigFileName() string { return "revive.toml" }
func (r *reviveConfigCreator) GetToolName() string       { return "Revive" }

func createReviveConfigFile(config []domain.PatternConfiguration, toolsConfigDir string) error {
	reviveConfigurationString := reviveTool.GenerateReviveConfig(config)
	return os.WriteFile(filepath.Join(toolsConfigDir, "revive.toml"), []byte(reviveConfigurationString), constants.DefaultFilePerms)
}
