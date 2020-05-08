package models

import (
	"fmt"
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models/timeofday"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

func newPotentialWeek() *PotentialWeek {
	week := &PotentialWeek{
		Analysis: nil,
		Spikes:   nil,
		Prices: PotentialPricePeriods{
			{
				prices: &prices{
					minPrice:  100,
					maxPrice:  100,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 0,
			},
			{
				prices: &prices{
					minPrice:  101,
					maxPrice:  101,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 1,
			},
			{
				prices: &prices{
					minPrice:  102,
					maxPrice:  102,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 2,
			},
			{
				prices: &prices{
					minPrice:  103,
					maxPrice:  103,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 3,
			},
			{
				prices: &prices{
					minPrice:  104,
					maxPrice:  104,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 4,
			},
			{
				prices: &prices{
					minPrice:  105,
					maxPrice:  105,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 5,
			},
			{
				prices: &prices{
					minPrice:  106,
					maxPrice:  106,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 6,
			},
			{
				prices: &prices{
					minPrice:  107,
					maxPrice:  107,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 7,
			},
			{
				prices: &prices{
					minPrice:  108,
					maxPrice:  108,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 8,
			},
			{
				prices: &prices{
					minPrice:  109,
					maxPrice:  109,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 9,
			},
			{
				prices: &prices{
					minPrice:  110,
					maxPrice:  110,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 10,
			},
			{
				prices: &prices{
					minPrice:  111,
					maxPrice:  111,
					minChance: 0,
					maxChance: 0,
					midChance: 0,
				},
				PricePeriod: 11,
			},
		},
	}

	return week
}

func testPotentialWeekPeriodData(
	t *testing.T, thisCase *pricePeriodTestCase, week *PotentialWeek,
) {
	testGetByDay := func(t *testing.T) {
		assert := assert.New(t)

		expectedPrice := 100 + int(thisCase.ExpectedPeriod)
		period, err := week.Prices.ForDay(thisCase.Weekday, thisCase.ToD)
		assert.NoError(err)

		assert.Equal(expectedPrice, period.MinPrice(), "period by day")
	}

	t.Run("by day", testGetByDay)

	testGetByTime := func(t *testing.T) {
		assert := assert.New(t)

		expectedPrice := 100 + int(thisCase.ExpectedPeriod)
		period, err := week.Prices.ForTime(thisCase.Time)
		assert.NoError(err)

		assert.Equal(expectedPrice, period.MinPrice(), "period by day")
	}

	t.Run("by time", testGetByTime)

}

func TestPotentialGetPeriod(t *testing.T) {
	var thisCase *pricePeriodTestCase
	week := newPotentialWeek()

	testPeriod := func(t *testing.T) {
		testPotentialWeekPeriodData(t, thisCase, week)
	}

	for _, thisCase = range pricePeriodTestCases {
		name := fmt.Sprintf("%v %v", thisCase.Weekday, thisCase.ToD)
		t.Run(name, testPeriod)
	}

	testGetByTimeSunday := func(t *testing.T) {
		assert := assert.New(t)

		period, err := week.Prices.ForTime(sunday)
		assert.Nil(period, "period should be nil")
		assert.EqualError(err, errs.ErrNoSundayPricePeriod.Error())
	}

	t.Run("by time sunday error", testGetByTimeSunday)

	testGetByDaySunday := func(t *testing.T) {
		assert := assert.New(t)

		period, err := week.Prices.ForDay(time.Sunday, timeofday.AM)
		assert.Nil(period, "period should be nil")
		assert.EqualError(err, errs.ErrNoSundayPricePeriod.Error())
	}

	t.Run("by day sunday error", testGetByDaySunday)
}
