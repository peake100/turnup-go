package timeofday

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

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
