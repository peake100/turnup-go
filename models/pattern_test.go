package models

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPanicOnPhaseProgressionUnknown(t *testing.T) {
	assert := assert.New(t)

	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		assert.True(ok, "recovered error conversion")
		assert.EqualError(err, errs.ErrUnknownPhasesInvalid.Error())
	}()

	UNKNOWN.PhaseProgression(nil)
}

func TestPanicOnPhaseProgressionInvalid(t *testing.T) {
	assert := assert.New(t)

	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		assert.True(ok, "recovered error conversion")
		assert.EqualError(err, errs.ErrBadPatternIndex.Error())
	}()

	PricePattern(10).PhaseProgression(nil)
}

func TestPatternFromString(t *testing.T) {
	type testCase struct {
		StringVal string
		Pattern   PricePattern
	}

	testCases := []*testCase{
		// Big Spike
		{
			"BIGSPIKE",
			BIGSPIKE,
		},
		{
			"BIG SPIKE",
			BIGSPIKE,
		},
		{
			"bigspike",
			BIGSPIKE,
		},
		{
			"big spike",
			BIGSPIKE,
		},

		// Small Spike
		{
			"small spike",
			SMALLSPIKE,
		},
		{
			"smallspike",
			SMALLSPIKE,
		},
		{
			"SMALL SPIKE",
			SMALLSPIKE,
		},
		{
			"SMALLSPIKE",
			SMALLSPIKE,
		},

		// Decreasing
		{
			"decreasing",
			DECREASING,
		},
		{
			"DECREASING",
			DECREASING,
		},

		// Fluctuating
		{
			"fluctuating",
			FLUCTUATING,
		},
		{
			"fluctuating",
			FLUCTUATING,
		},

		// Unknown
		{
			"unknown",
			UNKNOWN,
		},
		{
			"UNKNOWN",
			UNKNOWN,
		},
	}

	var thisCase *testCase

	test := func(t *testing.T) {
		assert := assert.New(t)

		pattern, err := PatternFromString(thisCase.StringVal)
		assert.NoError(err)
		assert.Equal(thisCase.Pattern, pattern)
	}

	for _, thisCase = range testCases {
		t.Run(thisCase.StringVal, test)
	}
}

func TestPatternFromStringErr(t *testing.T) {
	assert := assert.New(t)

	pattern, err := PatternFromString("blah")
	assert.Equal(PricePattern(5), pattern)
	assert.EqualError(err, errs.ErrPatternStringValue.Error())
}
