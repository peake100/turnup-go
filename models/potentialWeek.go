package models

import (
	"github.com/peake100/turnup-go/models/timeofday"
	"time"
)

// This will be implemented as a slice as we will not always hit all 12 price periods
// when evaluating if a potential week needs to be thrown out
type PotentialPricePeriods []*PotentialPricePeriod

// Return the potential price period for a given Weekday + time of day
func (prices PotentialPricePeriods) ForDay(
	weekday time.Weekday, tod timeofday.ToD,
) (period *PotentialPricePeriod, err error) {
	// We are already checking for sunday, so we can suppress the error
	pricePeriod, err := PricePeriodFromDay(weekday, tod)
	if err != nil {
		return nil, err
	}
	return prices[pricePeriod], nil
}

// Return the potential price period for a given time. The ticker does not contain any
// information about dates, so it is assumed that the time passed in to priceTime is for
// the week that the ticker describes.
func (prices PotentialPricePeriods) ForTime(
	priceTime time.Time,
) (period *PotentialPricePeriod, err error) {
	pricePeriod, err := PricePeriodFromTime(priceTime)
	if err != nil {
		return nil, err
	}
	return prices[pricePeriod], nil
}

type PotentialWeek struct {
	// Holds chance and price information
	*Analysis

	// Details about if and when a price spike could occur for this week.
	Spikes *SpikeRangeAll

	// Holds the details of the potential price periods.
	Prices PotentialPricePeriods
}
