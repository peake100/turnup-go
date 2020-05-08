package models

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"fmt"
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models/timeofday"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func testSpikeInfoPeriodData(
	t *testing.T,
	thisCase *pricePeriodTestCase,
	spikeChanceBreakdown *SpikeChanceBreakdown,
) {

	expectedChance := 0.0 + float64(thisCase.ExpectedPeriod)

	testGetByDay := func(t *testing.T) {
		assert := assert.New(t)

		chance, err := spikeChanceBreakdown.ForDay(thisCase.Weekday, thisCase.ToD)
		assert.NoError(err)

		assert.Equal(expectedChance, chance, "chance by day")
	}

	t.Run("by day", testGetByDay)

	testGetByTime := func(t *testing.T) {
		assert := assert.New(t)

		chance, err := spikeChanceBreakdown.ForTime(thisCase.Time)
		assert.NoError(err)

		assert.Equal(expectedChance, chance, "chance by day")
	}

	t.Run("by time", testGetByTime)
}

func TestSpikeDensity(t *testing.T) {
	var thisCase *pricePeriodTestCase

	spikeChanceBreakdown := &SpikeChanceBreakdown{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	}

	testPeriod := func(t *testing.T) {
		testSpikeInfoPeriodData(t, thisCase, spikeChanceBreakdown)
	}

	for _, thisCase = range pricePeriodTestCases {
		name := fmt.Sprintf("%v %v", thisCase.Weekday, thisCase.ToD)
		t.Run(name, testPeriod)
	}

	testGetByTimeSunday := func(t *testing.T) {
		assert := assert.New(t)

		chance, err := spikeChanceBreakdown.ForTime(sunday)
		assert.Equal(0.0, chance, "chance should be 0.0")
		assert.EqualError(err, errs.ErrNoSundayPricePeriod.Error())
	}

	t.Run("by time sunday error", testGetByTimeSunday)

	testGetByDaySunday := func(t *testing.T) {
		assert := assert.New(t)

		chance, err := spikeChanceBreakdown.ForDay(time.Sunday, timeofday.AM)
		assert.Equal(0.0, chance, "chance should be 0.0")
		assert.EqualError(err, errs.ErrNoSundayPricePeriod.Error())
	}

	t.Run("by day sunday error", testGetByDaySunday)
}
