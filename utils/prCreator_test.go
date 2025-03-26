package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePr(t *testing.T) {

	result := CreatePr(true)

	assert.Equal(t, result, false)
}
