package domain

type Token interface {
	Value() string
}

type ProjectToken struct {
	value string
}

func (t ProjectToken) Value() string {
	return t.value
}

func NewProjectToken(value string) ProjectToken {
	return ProjectToken{value: value}
}

type ApiToken struct {
	value string
}

func (t ApiToken) Value() string {
	return t.value
}

func NewApiToken(value string) ApiToken {
	return ApiToken{value: value}
}
