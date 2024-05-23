package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCreatePr(t *testing.T) {

	result := CreatePr(true)

	assert.Equal(t, result, false)
}
