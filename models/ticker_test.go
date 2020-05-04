package models

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"fmt"
	"github.com/peake100/turnup-go/models/timeofday"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPriceTicker_PriceForDay(t *testing.T) {

	type testCase struct {
		Weekday        time.Weekday
		ToD            timeofday.ToD
		ExpectedPeriod PricePeriod
	}

	cases := []*testCase{
		// MONDAY
		{
			time.Monday,
			timeofday.AM,
			0,
		},
		{
			time.Monday,
			timeofday.PM,
			1,
		},

		// TUESDAY
		{
			time.Tuesday,
			timeofday.AM,
			2,
		},
		{
			time.Tuesday,
			timeofday.PM,
			3,
		},

		// WEDNESDAY
		{
			time.Wednesday,
			timeofday.AM,
			4,
		},
		{
			time.Wednesday,
			timeofday.PM,
			5,
		},

		// THURSDAY
		{
			time.Thursday,
			timeofday.AM,
			6,
		},
		{
			time.Thursday,
			timeofday.PM,
			7,
		},

		// FRIDAY
		{
			time.Friday,
			timeofday.AM,
			8,
		},
		{
			time.Friday,
			timeofday.PM,
			9,
		},

		// SATURDAY
		{
			time.Saturday,
			timeofday.AM,
			10,
		},
		{
			time.Saturday,
			timeofday.PM,
			11,
		},
	}

	var thisCase *testCase

	test := func(t *testing.T) {
		assert := assert.New(t)
		ticker := new(PriceTicker)

		ticker.SetPriceForDay(thisCase.Weekday, thisCase.ToD, 100)

		assert.Equal(
			100, ticker.PriceForDay(thisCase.Weekday, thisCase.ToD),
			"method access",
		)
		assert.Equal(
			ticker.Prices[thisCase.ExpectedPeriod],
			100,
			"array access",
		)
	}

	for _, thisCase = range cases {
		t.Run(
			fmt.Sprintf("%v %v", thisCase.Weekday.String(), thisCase.ToD), test,
		)
	}

}

func TestTickerWeekdayPurchasePrice(t *testing.T) {
	ticker := new(PriceTicker)
	ticker.SetPriceForDay(time.Sunday, timeofday.AM, 100)

	assert.Equal(t, ticker.PurchasePrice, 100, "purchase price")
	assert.Equal(
		t,
		ticker.PriceForDay(time.Sunday, timeofday.PM),
		100,
		"purchase price from weekday",
	)
}

func TestTickerPriceForTime(t *testing.T) {
	monday := time.Date(
		2020, 4, 6, 10, 0, 0, 0, time.UTC,
	)
	var periodIndex int
	var testTime time.Time

	test := func(t *testing.T) {
		assert := assert.New(t)

		ticker := new(PriceTicker)
		ticker.SetPriceForTime(testTime, 100)

		assert.Equal(
			100, ticker.PriceForTime(testTime), "method",
		)
		assert.Equal(
			100, ticker.Prices[periodIndex], "array access",
		)
	}

	for i := 0; i < 6; i++ {
		testTime = monday.AddDate(0, 0, i)
		for _, tod := range []timeofday.ToD{timeofday.AM, timeofday.PM} {
			if tod == timeofday.PM {
				testTime = testTime.Add(time.Hour * 3)
			}

			t.Run(fmt.Sprintf("price_period_%v", periodIndex), test)

			periodIndex++

		}
	}
}

func TestTickerPriceForTimeSunday(t *testing.T) {
	assert := assert.New(t)

	sunday := time.Date(
		2020, 4, 5, 10, 0, 0, 0, time.UTC,
	)

	ticker := new(PriceTicker)
	ticker.SetPriceForTime(sunday, 100)

	assert.Equal(100, ticker.PurchasePrice, "struct field")
	assert.Equal(100, ticker.PriceForTime(sunday), "method")
}
