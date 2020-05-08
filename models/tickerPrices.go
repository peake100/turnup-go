package models

import (
	"github.com/peake100/turnup-go/models/timeofday"
	"github.com/peake100/turnup-go/values"
	"time"
)

// Holds the Nook turnip purchase prices for a week in price-period order. Adds methods
// for setting and fetching via time package values.
type NookPriceArray [values.PricePeriodCount]int

// Return the price for a given Weekday + time of day
func (prices *NookPriceArray) ForDay(
	weekday time.Weekday, tod timeofday.ToD,
) (price int, err error) {
	pricePeriod, err := PricePeriodFromDay(weekday, tod)
	if err != nil {
		return 0, err
	}
	return prices[pricePeriod], nil
}

// Return the price for a given time. The ticker does not contain any information about
// dates, so it is assumed that the time passed in to priceTime is for the week that
// the ticker describes.
func (prices *NookPriceArray) ForTime(priceTime time.Time) (price int, err error) {
	pricePeriod, err := PricePeriodFromTime(priceTime)
	if err != nil {
		return 0, err
	}
	return prices[pricePeriod], nil
}

// Set the price with a Weekday / time of day for a little more ease in setting values.
func (prices *NookPriceArray) SetForDay(
	weekday time.Weekday, tod timeofday.ToD, price int,
) error {
	pricePeriod, err := PricePeriodFromDay(weekday, tod)
	if err != nil {
		return err
	}
	prices[pricePeriod] = price
	return nil
}

// Set a price period for a specific time. Timezone is not taken into account during
// this operation.
func (prices *NookPriceArray) SetForTime(priceTime time.Time, price int) error {
	pricePeriod, err := PricePeriodFromTime(priceTime)
	if err != nil {
		return err
	}
	prices[pricePeriod] = price
	return nil
}
