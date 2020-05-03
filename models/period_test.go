package models

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"fmt"
	"github.com/illuscio-dev/turnup-go/errs"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPeriod(t *testing.T) {

	weekdays := []time.Weekday{
		time.Monday, time.Monday,
		time.Tuesday, time.Tuesday,
		time.Wednesday, time.Wednesday,
		time.Thursday, time.Thursday,
		time.Friday, time.Friday,
		time.Saturday, time.Saturday,
	}

	timeOfDays := []ToD{
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
	}

	for i := 0; i < 12; i++ {

		period := PricePeriod(i)

		testWeekday := func(t *testing.T) {
			assert.Equal(t, weekdays[i], period.Weekday(), "Weekday")
		}

		testTimeOfDay := func(t *testing.T) {
			assert.Equal(t, timeOfDays[i], period.ToD(), "time of day")
		}

		testPeriod := func(t *testing.T) {
			t.Run("Weekday", testWeekday)
			t.Run("Time of Day", testTimeOfDay)
		}

		t.Run(fmt.Sprint(int(period)), testPeriod)
	}

}

func TestPeriodFromWeekday(t *testing.T) {
	type testCase struct {
		Weekday time.Weekday
		ToD     ToD
	}

	cases := []*testCase{
		{
			time.Monday,
			AM,
		},
		{
			time.Monday,
			PM,
		},
		{
			time.Tuesday,
			AM,
		},
		{
			time.Tuesday,
			PM,
		},
		{
			time.Wednesday,
			AM,
		},
		{
			time.Wednesday,
			PM,
		},
		{
			time.Thursday,
			AM,
		},
		{
			time.Thursday,
			PM,
		},
		{
			time.Friday,
			AM,
		},
		{
			time.Friday,
			PM,
		},
		{
			time.Saturday,
			AM,
		},
		{
			time.Saturday,
			PM,
		},
	}

	var i int
	var thisCase *testCase

	test := func(t *testing.T) {
		assert := assert.New(t)

		expectedPricePeriod := PricePeriod(i)
		parsedPeriod, err := PricePeriodFromDay(thisCase.Weekday, thisCase.ToD)

		assert.NoError(err)
		assert.Equal(expectedPricePeriod, parsedPeriod)
	}

	for i, thisCase = range cases {
		t.Run(fmt.Sprintf("%v %v", thisCase.Weekday.String(), thisCase.ToD), test)
	}
}

func TestPricePeriodFromWeekdaySundayErr(t *testing.T) {
	_, err := PricePeriodFromDay(time.Sunday, AM)
	assert.EqualError(t, err, errs.ErrNoSundayPricePeriod.Error())
}

func TestPricePeriodFromTime(t *testing.T) {
	type testCase struct {
		Expected PricePeriod
		TestTime time.Time
	}

	testCases := []*testCase{
		// Monday AM
		{
			0,
			time.Date(
				2020,
				4,
				6,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Monday PM
		{
			1,
			time.Date(
				2020,
				4,
				6,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},

		// Tuesday AM
		{
			2,
			time.Date(
				2020,
				4,
				7,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Tuesday PM
		{
			3,
			time.Date(
				2020,
				4,
				7,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},

		// Wednesday AM
		{
			4,
			time.Date(
				2020,
				4,
				8,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Wednesday PM
		{
			5,
			time.Date(
				2020,
				4,
				8,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},

		// Thursday AM
		{
			6,
			time.Date(
				2020,
				4,
				9,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Thursday PM
		{
			7,
			time.Date(
				2020,
				4,
				9,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},

		// Friday AM
		{
			8,
			time.Date(
				2020,
				4,
				10,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Friday PM
		{
			9,
			time.Date(
				2020,
				4,
				10,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},

		// Saturday AM
		{
			10,
			time.Date(
				2020,
				4,
				11,
				10,
				0,
				0,
				0,
				time.UTC,
			),
		},
		// Saturday PM
		{
			11,
			time.Date(
				2020,
				4,
				11,
				12,
				0,
				0,
				0,
				time.UTC,
			),
		},
	}

	var thisCase *testCase

	test := func(t *testing.T) {
		assert := assert.New(t)

		pricePeriod, err := PricePeriodFromTime(thisCase.TestTime)

		assert.NoError(err)
		assert.Equal(thisCase.Expected, pricePeriod)
	}

	for _, thisCase = range testCases {
		t.Run(
			fmt.Sprintf("from %v", thisCase.TestTime.String()),
			test,
		)
	}
}

func TestFromTimeErr(t *testing.T) {
	sunday := time.Date(
		2020,
		4,
		5,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := PricePeriodFromTime(sunday)
	assert.EqualError(t, err, errs.ErrNoSundayPricePeriod.Error())
}
