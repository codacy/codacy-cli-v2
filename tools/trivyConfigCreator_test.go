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

scan:
  scanners:
    - vuln
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
					Id: "Trivy_vulnerability_high",
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
					Id: "Trivy_vulnerability_critical",
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
					Id: "Trivy_vulnerability_high",
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
					Id: "Trivy_vulnerability_critical",
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
  - MEDIUM
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigOnlyHighAndCritical(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{
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
		},
		`severity:
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
`)
}

func TestCreateTrivyConfigNoVulnerabilitiesWithSecret(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{
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

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigOnlyLowWithSecrets(t *testing.T) {
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
					Id: "Trivy_secret",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "enabled",
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

func TestCreateTrivyConfigOnlyHigh(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_high",
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
  - HIGH

scan:
  scanners:
    - vuln
`)
}

func TestCreateTrivyConfigOnlyCriticalWithSecrets(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Trivy_vulnerability_critical",
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
  - CRITICAL

scan:
  scanners:
    - vuln
    - secret
`)
}

func TestCreateTrivyConfigOnlyHighAndCriticalEventIfPatternsOverlap(t *testing.T) {
	testTrivyConfig(t,
		[]domain.PatternConfiguration{
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
		},
		`severity:
  - HIGH
  - CRITICAL

scan:
  scanners:
    - vuln
`)
}
