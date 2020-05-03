package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ProgressionFunc = func(ticker *PriceTicker) []PatternPhase

func testPhaseNames(
	t *testing.T,
	pattern Pattern,
	expectedNames []string,
) {
	var name string
	var phase PatternPhase
	var i int

	priceTicker := new(PriceTicker)

	test := func(t *testing.T) {
		assert.Equal(t, name, phase.Name(), "phase name")
	}

	phaseList := pattern.PhaseProgression(priceTicker)

	for i, phase = range phaseList {
		name = expectedNames[i]
		t.Run(name, test)
	}

	assert.Equal(t, len(expectedNames), len(phaseList), "phase  count")
}

func TestPatternPhaseProgression(t *testing.T) {
	type testCase struct {
		Pattern       Pattern
		ExpectedNames []string
	}

	cases := []*testCase{
		{
			FLUCTUATING,
			[]string{
				"mild increase",
				"mild decrease",
				"mild increase",
				"mild decrease",
				"mild increase",
			},
		},
		{
			BIGSPIKE,
			[]string{
				"steady decrease",
				"sharp increase",
				"sharp decrease",
				"random low",
			},
		},
		{
			DECREASING,
			[]string{
				"whomp whomp",
			},
		},
		{
			SMALLSPIKE,
			[]string{
				"steady decrease",
				"slight spike",
				"steady decrease",
			},
		},
	}

	var thisCase *testCase

	test := func(t *testing.T) {
		testPhaseNames(t, thisCase.Pattern, thisCase.ExpectedNames)
	}

	for _, thisCase = range cases {
		t.Run(thisCase.Pattern.String(), test)
	}

}

// We're going to use this interface to test that we get panics when we ask a phase
// for a base price multiplier outside of a possible sub period
type hasBasePriceMultiplier interface {
	Name() string
	BasePriceMultiplier(subPeriod int) (min float64, max float64)
}

func TestBasePriceMultiplierPanics(t *testing.T) {
	type testCase struct {
		phase    hasBasePriceMultiplier
		panicsOn int
		message  string
	}

	testCases := []*testCase{
		{
			&sharpIncrease{},
			4,
			"sharp increase only has 3 price periods",
		},
		{
			&sharpDecrease{},
			2,
			"sharp decrease only has 2 price periods",
		},
	}

	var thisCase *testCase

	test := func(t *testing.T) {
		defer func() {
			recovered := recover()
			err, ok := recovered.(error)
			assert.True(t, ok, "panic is err")
			assert.EqualError(t, err, thisCase.message, "panic message")
		}()
		thisCase.phase.BasePriceMultiplier(thisCase.panicsOn)
	}

	for _, thisCase = range testCases {
		t.Run(thisCase.phase.Name(), test)
	}
}
