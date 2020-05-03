package models

import (
	"github.com/illuscio-dev/turnup-go/errs"
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

	Pattern(10).PhaseProgression(nil)
}

func TestPatternFromString(t *testing.T) {
	type testCase struct {
		StringVal string
		Pattern Pattern
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
	assert.Equal(Pattern(5), pattern)
	assert.EqualError(err, errs.ErrPatternStringValue.Error())
}
