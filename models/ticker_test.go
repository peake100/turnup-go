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

type pricePeriodTestCase struct {
	Weekday        time.Weekday
	ToD            timeofday.ToD
	Time           time.Time
	ExpectedPeriod PricePeriod
}

var mondayAM = time.Date(
	2020, 4, 6, 10, 0, 0, 0, time.UTC,
)

var mondayPM = time.Date(
	2020, 4, 6, 13, 0, 0, 0, time.UTC,
)

const oneDay = time.Hour * 24

var sunday = mondayAM.Add(oneDay * -1)

var pricePeriodTestCases = []*pricePeriodTestCase{
	// MONDAY
	{
		time.Monday,
		timeofday.AM,
		mondayAM,
		0,
	},
	{
		time.Monday,
		timeofday.PM,
		mondayPM,
		1,
	},

	// TUESDAY
	{
		time.Tuesday,
		timeofday.AM,
		mondayAM.Add(oneDay),
		2,
	},
	{
		time.Tuesday,
		timeofday.PM,
		mondayPM.Add(oneDay),
		3,
	},

	// WEDNESDAY
	{
		time.Wednesday,
		timeofday.AM,
		mondayAM.Add(oneDay * 2),
		4,
	},
	{
		time.Wednesday,
		timeofday.PM,
		mondayPM.Add(oneDay * 2),
		5,
	},

	// THURSDAY
	{
		time.Thursday,
		timeofday.AM,
		mondayAM.Add(oneDay * 3),
		6,
	},
	{
		time.Thursday,
		timeofday.PM,
		mondayPM.Add(oneDay * 3),
		7,
	},

	// FRIDAY
	{
		time.Friday,
		timeofday.AM,
		mondayAM.Add(oneDay * 4),
		8,
	},
	{
		time.Friday,
		timeofday.PM,
		mondayPM.Add(oneDay * 4),
		9,
	},

	// SATURDAY
	{
		time.Saturday,
		timeofday.AM,
		mondayAM.Add(oneDay * 5),
		10,
	},
	{
		time.Saturday,
		timeofday.PM,
		mondayPM.Add(oneDay * 5),
		11,
	},
}

func TestPriceTicker_PriceForDay(t *testing.T) {

	var thisCase *pricePeriodTestCase

	test := func(t *testing.T) {
		assert := assert.New(t)
		ticker := new(PriceTicker)

		err := ticker.Prices.SetForDay(thisCase.Weekday, thisCase.ToD, 100)
		assert.NoError(err)

		price, err := ticker.Prices.ForDay(thisCase.Weekday, thisCase.ToD)
		assert.NoError(err)

		assert.Equal(
			100, price,
			"method access",
		)
		assert.Equal(
			ticker.Prices[thisCase.ExpectedPeriod],
			100,
			"array access",
		)
	}

	for _, thisCase = range pricePeriodTestCases {
		t.Run(
			fmt.Sprintf("%v %v", thisCase.Weekday.String(), thisCase.ToD), test,
		)
	}

}

func testSundayErr(t *testing.T, err error) {
	assert.EqualError(t, err, errs.ErrNoSundayPricePeriod.Error())
}

func TestTickerWeekdayPurchasePrice(t *testing.T) {
	ticker := new(PriceTicker)
	err := ticker.Prices.SetForDay(time.Sunday, timeofday.AM, 100)
	testSundayErr(t, err)

	price, err := ticker.Prices.ForDay(time.Sunday, timeofday.PM)
	testSundayErr(t, err)
	assert.Equal(t, price, 0)
}

func TestTickerPriceForTime(t *testing.T) {
	var thisCase *pricePeriodTestCase

	test := func(t *testing.T) {
		assert := assert.New(t)

		ticker := new(PriceTicker)
		err := ticker.Prices.SetForTime(thisCase.Time, 400)
		assert.NoError(err)

		price, err := ticker.Prices.ForTime(thisCase.Time)
		assert.NoError(err)

		assert.Equal(
			400, price, "method",
		)
		assert.Equal(
			400,
			ticker.Prices[thisCase.ExpectedPeriod],
			"array access",
		)
	}

	for _, thisCase = range pricePeriodTestCases {
		t.Run(fmt.Sprintf("price_period_%v", thisCase.ExpectedPeriod), test)
	}
}

func TestTickerPriceForTimeSunday(t *testing.T) {
	sunday := time.Date(
		2020, 4, 5, 10, 0, 0, 0, time.UTC,
	)

	ticker := new(PriceTicker)
	err := ticker.Prices.SetForTime(sunday, 100)
	testSundayErr(t, err)

	price, err := ticker.Prices.ForTime(sunday)
	testSundayErr(t, err)
	assert.Equal(t, price, 0)
}
