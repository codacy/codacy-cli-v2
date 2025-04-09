package tools

import (
	"codacy/cli-v2/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTrivyConfig(t *testing.T, configuration []domain.PatternConfiguration, expected string) {
	actual := CreateTrivyConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateTrivyConfigEmptyConfig(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{},
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_minor",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_medium",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_secret",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_minor",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_minor",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_medium",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_secret",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_minor",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_medium",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
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
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_minor",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_medium",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
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
