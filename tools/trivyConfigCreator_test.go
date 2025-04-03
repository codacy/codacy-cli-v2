package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTrivyConfig(t *testing.T, configuration ToolConfiguration, expected string) {
	actual := CreateTrivyConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateTrivyConfigEmptyConfig(t *testing.T) {
	testTrivyConfig(t,
		ToolConfiguration{},
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
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParameterConfigurations: []PatternParameterConfiguration{
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
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []PatternParameterConfiguration{
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
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParameterConfigurations: []PatternParameterConfiguration{
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
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []PatternParameterConfiguration{
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
		ToolConfiguration{
			PatternsConfiguration: []PatternConfiguration{
				{
					PatternId: "Trivy_vulnerability_minor",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParameterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParameterConfigurations: []PatternParameterConfiguration{
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
