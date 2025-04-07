package domain

type ToolName string

const (
	ESLint ToolName = "eslint"
	Trivy  ToolName = "trivy"
	PMD    ToolName = "pmd"
)
