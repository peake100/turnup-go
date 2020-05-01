package patterns

import (
	"fmt"
	"github.com/illuscio-dev/turnup-go/errs"
	"github.com/illuscio-dev/turnup-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

// NAMES
func TestNameUnknown(t *testing.T) {
	assert.Equal(t, "UNKNOWN", UNKNOWN.String())
}

func TestNameFluctuating(t *testing.T) {
	assert.Equal(t, "FLUCTUATING", FLUCTUATING.String())
}

func TestNameBigSpike(t *testing.T) {
	assert.Equal(t, "BIG SPIKE", BIGSPIKE.String())
}

func TestNameDecreasing(t *testing.T) {
	assert.Equal(t, "DECREASING", DECREASING.String())
}

func TestNameSmallSpike(t *testing.T) {
	assert.Equal(t, "SMALL SPIKE", SMALLSPIKE.String())
}

// NAMES

func TestIntFluctuating(t *testing.T) {
	assert.Equal(t, 0, int(FLUCTUATING))
}

func TestIntBigSpike(t *testing.T) {
	assert.Equal(t, 1, int(BIGSPIKE))
}

func TestIntDecreasing(t *testing.T) {
	assert.Equal(t, 2, int(DECREASING))
}

func TestIntSmallSpike(t *testing.T) {
	assert.Equal(t, 3, int(SMALLSPIKE))
}

func TestIntUnknown(t *testing.T) {
	assert.Equal(t, 4, int(UNKNOWN))
}

// Base Chances
func TestBaseChances(t *testing.T) {
	type testCase struct {
		previousWeek   models.Pattern
		thisWeek       models.Pattern
		chanceExpected float64
	}

	testCases := []*testCase{
		// PREVIOUS: UNKNOWN
		{
			UNKNOWN,
			FLUCTUATING,
			0.35,
		},
		{
			UNKNOWN,
			BIGSPIKE,
			0.2625,
		},
		{
			UNKNOWN,
			DECREASING,
			0.1375,
		},
		{
			UNKNOWN,
			SMALLSPIKE,
			0.25,
		},

		// PREVIOUS: FLUCTUATING
		{
			FLUCTUATING,
			FLUCTUATING,
			0.20,
		},
		{
			FLUCTUATING,
			BIGSPIKE,
			0.30,
		},
		{
			FLUCTUATING,
			DECREASING,
			0.15,
		},
		{
			FLUCTUATING,
			SMALLSPIKE,
			0.35,
		},

		// PREVIOUS: BIG SPIKE
		{
			BIGSPIKE,
			FLUCTUATING,
			0.50,
		},
		{
			BIGSPIKE,
			BIGSPIKE,
			0.05,
		},
		{
			BIGSPIKE,
			DECREASING,
			0.20,
		},
		{
			BIGSPIKE,
			SMALLSPIKE,
			0.25,
		},

		// PREVIOUS: DECREASING
		{
			DECREASING,
			FLUCTUATING,
			0.25,
		},
		{
			DECREASING,
			BIGSPIKE,
			0.45,
		},
		{
			DECREASING,
			DECREASING,
			0.05,
		},
		{
			DECREASING,
			SMALLSPIKE,
			0.25,
		},

		// PREVIOUS: SMALL SPIKE
		{
			SMALLSPIKE,
			FLUCTUATING,
			0.45,
		},
		{
			SMALLSPIKE,
			BIGSPIKE,
			0.25,
		},
		{
			SMALLSPIKE,
			DECREASING,
			0.15,
		},
		{
			SMALLSPIKE,
			SMALLSPIKE,
			0.15,
		},
	}

	for _, thisCase := range testCases {

		test := func(t *testing.T) {
			assert.Equal(
				t,
				thisCase.chanceExpected,
				thisCase.thisWeek.BaseChance(thisCase.previousWeek))
		}

		t.Run(
			fmt.Sprintf("%v->%v", thisCase.previousWeek, thisCase.thisWeek),
			test,
		)
	}
}

func TestUnknownUnknownPanic(t *testing.T) {
	defer func() {
		recovered := recover()
		err := recovered.(error)
		assert.EqualError(t, err, errs.ErrUnknownBaseChanceInvalid.Error())
	}()

	UNKNOWN.BaseChance(FLUCTUATING)
}
