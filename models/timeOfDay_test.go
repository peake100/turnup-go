package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAMOffset(t *testing.T) {
	assert.Equal(t, 0, AM.PhaseOffset())
}

func TestPMOffset(t *testing.T) {
	assert.Equal(t, 1, PM.PhaseOffset())
}
