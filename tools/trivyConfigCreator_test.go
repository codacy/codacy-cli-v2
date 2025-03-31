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
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParamenterConfigurations: []PatternParameterConfiguration{
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
					ParamenterConfigurations: []PatternParameterConfiguration{
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
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_secret",
					ParamenterConfigurations: []PatternParameterConfiguration{
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
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParamenterConfigurations: []PatternParameterConfiguration{
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
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "true",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability_medium",
					ParamenterConfigurations: []PatternParameterConfiguration{
						{
							Name:  "enabled",
							Value: "false",
						},
					},
				},
				{
					PatternId: "Trivy_vulnerability",
					ParamenterConfigurations: []PatternParameterConfiguration{
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
