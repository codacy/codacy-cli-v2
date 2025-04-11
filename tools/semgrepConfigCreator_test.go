package tools

import (
	"codacy/cli-v2/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testSemgrepConfig(t *testing.T, configuration []domain.PatternConfiguration, expected string) {
	actual := CreateSemgrepConfig(configuration)
	assert.Equal(t, expected, actual)
}

func TestCreateSemgrepConfigEmptyConfig(t *testing.T) {
	testSemgrepConfig(t,
		[]domain.PatternConfiguration{},
		`rules:
  - id: all
    pattern: |
      $X
    message: "Semgrep analysis"
    languages: [generic]
    severity: INFO
`)
}

func TestCreateSemgrepConfigWithRules(t *testing.T) {
	testSemgrepConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Semgrep_bash_curl_security_curl-pipe-bash",
				},
			},
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Semgrep_c_buffer_rule-gets",
				},
			},
		},
		`rules:
  - id: curl_security_curl-pipe-bash
    pattern: |
      $X
    message: "Semgrep rule: curl_security_curl-pipe-bash"
    languages: [bash]
    severity: INFO
  - id: buffer_rule-gets
    pattern: |
      $X
    message: "Semgrep rule: buffer_rule-gets"
    languages: [c]
    severity: INFO
`)
}

func TestCreateSemgrepConfigAllEnabled(t *testing.T) {
	testSemgrepConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Semgrep_rule1",
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
					Id: "Semgrep_rule2",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "true",
					},
				},
			},
		},
		`rules:
  - id: all
    pattern: |
      $X
    message: "Semgrep analysis"
    languages: [generic]
    severity: INFO
`)
}

func TestCreateSemgrepConfigSomeDisabled(t *testing.T) {
	testSemgrepConfig(t,
		[]domain.PatternConfiguration{
			{
				PatternDefinition: domain.PatternDefinition{
					Id: "Semgrep_rule1",
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
					Id: "Semgrep_rule2",
				},
				Parameters: []domain.ParameterConfiguration{
					{
						Name:  "enabled",
						Value: "false",
					},
				},
			},
		},
		`rules:
  - id: rule1
    pattern: |
      $X
    message: "Semgrep rule: rule1"
    languages: [generic]
    severity: INFO
`)
}
