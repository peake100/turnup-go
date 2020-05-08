package models

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models/timeofday"
	"time"
)

type PricePeriod int

// The weekday this price period occurs on. Sunday = 0.
func (period PricePeriod) Weekday() time.Weekday {
	return time.Weekday(period/2 + 1)
}

// The time of day (AM / PM) this price occurs on.
func (period PricePeriod) ToD() timeofday.ToD {
	if period%2 == 0 {
		return timeofday.AM
	}
	return timeofday.PM
}

// Get the price period for a given weekday and time of day (AM / PM).
func PricePeriodFromDay(weekday time.Weekday, tod timeofday.ToD) (PricePeriod, error) {
	if weekday == time.Sunday {
		return -1, errs.ErrNoSundayPricePeriod
	}
	pricePeriod := ((int(weekday) - 1) * 2) + tod.PhaseOffset()
	return PricePeriod(pricePeriod), nil
}

// Get the price period that would occur on a real-world time. Timezone information is
// ignored -- all times are treated as naive.
func PricePeriodFromTime(priceTime time.Time) (PricePeriod, error) {
	tod := timeofday.PM
	if priceTime.Hour() < 12 {
		tod = timeofday.AM
	}

	return PricePeriodFromDay(priceTime.Weekday(), tod)
}
