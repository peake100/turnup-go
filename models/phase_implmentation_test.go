package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testPhaseNames(
	t *testing.T,
	pattern PricePattern,
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
		Pattern       PricePattern
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
				"small hasSpikeAny",
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

// Tests that we panic if we try to duplicate the decreasing pattern phase
func TestPhaseDecreasingDuplicatePanics(t *testing.T) {
	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		assert.True(t, ok, "recovered error")
		assert.EqualError(
			t, err, "decreasing phase should never be duplicated",
		)
	}()

	new(decreasingPattern).Duplicate()
}
