package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTools(t *testing.T) {
	obtained, err := getTools()

	assert.Nil(t, err)

	assert.Contains(t, obtained, Tool{
		Uuid:    "f8b29663-2cb2-498d-b923-a10c6a8c05cd",
		Name:    "ESLint",
		Version: "8.57.0",
	})
}
