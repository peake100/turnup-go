package models

import "time"

type PricePeriod int

func (period PricePeriod) Weekday() time.Weekday {
	return time.Weekday(period/2 + 1)
}

func (period PricePeriod) ToD() ToD {
	if period%2 == 0 {
		return AM
	}
	return PM
}
