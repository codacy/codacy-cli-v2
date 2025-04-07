package tools

import (
	"testing"

	"codacy/cli-v2/tools"
	"codacy/cli-v2/tools/trivy"

	"github.com/stretchr/testify/assert"
)

func testTrivyConfig(t *testing.T, configuration tools.ToolConfiguration, expected string) {
	actual := trivy.CreateTrivyConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateTrivyConfigEmptyConfig(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{},
		`severity:
  - LOW
  - MEDIUM
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigAllEnabled(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
			},
		},
		`severity:
  - LOW
  - MEDIUM
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigNoLow(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
			},
		},
		`severity:
  - MEDIUM
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigOnlyHigh(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
			},
		},
		`severity:
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
`)
}

func TestCreateTrivyConfigNoVulnerabilities(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
			},
		},
		`severity:

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigOnlySecretsLow(t *testing.T) {
	testTrivyConfig(t,
		tools.ToolConfiguration{
			PatternsConfiguration: []tools.PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []tools.PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
			},
		},
		`severity:
  - LOW

scan:
  scanners:
    - vuln
    - secret
`)
}
