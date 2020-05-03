package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPriceTicker_PriceForDay(t *testing.T) {

	type testCase struct {
		Weekday time.Weekday
		ToD ToD
		ExpectedPeriod  PricePeriod
	}

	cases := []*testCase{
		// MONDAY
		{
			time.Monday,
			AM,
			0,
		},
		{
			time.Monday,
			PM,
			1,
		},

		// TUESDAY
		{
			time.Tuesday,
			AM,
			2,
		},
		{
			time.Tuesday,
			PM,
			3,
		},

		// WEDNESDAY
		{
			time.Wednesday,
			AM,
			4,
		},
		{
			time.Wednesday,
			PM,
			5,
		},

		// THURSDAY
		{
			time.Thursday,
			AM,
			6,
		},
		{
			time.Thursday,
			PM,
			7,
		},

		// FRIDAY
		{
			time.Friday,
			AM,
			8,
		},
		{
			time.Friday,
			PM,
			9,
		},

		// SATURDAY
		{
			time.Saturday,
			AM,
			10,
		},
		{
			time.Saturday,
			PM,
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
	ticker.SetPriceForDay(time.Sunday, AM, 100)

	assert.Equal(t, ticker.PurchasePrice, 100, "purchase price")
	assert.Equal(
		t,
		ticker.PriceForDay(time.Sunday, PM),
		100,
		"purchase price from weekday",
	)
}

func TestTickerPriceForTime(t *testing.T) {
	monday := time.Date(
		2020, 4, 6, 10,0,0,0,time.UTC,
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

	for i := 0 ; i < 6 ; i++ {
		testTime = monday.AddDate(0, 0, i)
		for _, tod := range []ToD{AM, PM} {
			if tod == PM {
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
		2020, 4, 5, 10,0,0,0,time.UTC,
	)

	ticker := new(PriceTicker)
	ticker.SetPriceForTime(sunday, 100)

	assert.Equal(100, ticker.PurchasePrice, "struct field")
	assert.Equal(100, ticker.PriceForTime(sunday), "method")
}
