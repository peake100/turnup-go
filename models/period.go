package models

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models/timeofday"
	"time"
)

type PricePeriod int

func (period PricePeriod) Weekday() time.Weekday {
	return time.Weekday(period/2 + 1)
}

func (period PricePeriod) ToD() timeofday.ToD {
	if period%2 == 0 {
		return timeofday.AM
	}
	return timeofday.PM
}

// Converts price Weekday and time of day to price period. Returns error if sunday is
// passed
func PricePeriodFromDay(weekday time.Weekday, tod timeofday.ToD) (PricePeriod, error) {
	if weekday == time.Sunday {
		return -1, errs.ErrNoSundayPricePeriod
	}
	pricePeriod := ((int(weekday) - 1) * 2) + tod.PhaseOffset()
	return PricePeriod(pricePeriod), nil
}

func PricePeriodFromTime(priceTime time.Time) (PricePeriod, error) {
	tod := timeofday.PM
	if priceTime.Hour() < 12 {
		tod = timeofday.AM
	}

	return PricePeriodFromDay(priceTime.Weekday(), tod)
}
