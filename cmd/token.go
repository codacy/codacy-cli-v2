package cmd

type Token interface {
	Value() string
}

type ProjectToken struct {
	value string
}

func (t ProjectToken) Value() string {
	return t.value
}

type ApiToken struct {
	value string
}

func (t ApiToken) Value() string {
	return t.value
}
