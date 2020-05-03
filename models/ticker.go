package models

import "time"

type PriceTicker struct {
	// The previous week's price pattern
	PreviousPattern Pattern

	// The purchase price on sunday for this week
	PurchasePrice int

	// There are 12 buy-price periods in a week, we are going to store the 12 buy prices
	// in a 12-int array. A price of 'zero' will stand for 'not available'
	//
	// Because PricePeriod is an extension of int, we can access the array with
	// PricePeriod objects.
	Prices [12]int
}

// Return the price for a given Weekday + time of day
func (ticker *PriceTicker) PriceForDay(weekday time.Weekday, tod ToD) int {
	if weekday == 0 {
		return ticker.PurchasePrice
	}

	// We are already checking for sunday, so we can suppress the error
	pricePeriod, _ := PricePeriodFromDay(weekday, tod)
	return ticker.Prices[pricePeriod]
}

// Return the price for a given time. The ticker does not contain any information about
// dates, so it is assumed that the time passed in to priceTime is for the week that
// the ticker describes.
func (ticker *PriceTicker) PriceForTime(priceTime time.Time) int {
	weekday := priceTime.Weekday()
	if weekday == 0 {
		return ticker.PurchasePrice
	}
	pricePeriod, _ := PricePeriodFromTime(priceTime)
	return ticker.Prices[pricePeriod]
}

// Set the price with a Weekday / time of day for a little more ease in setting values.
func (ticker *PriceTicker) SetPriceForDay(weekday time.Weekday, tod ToD, price int) {
	if weekday == 0 {
		ticker.PurchasePrice = price
		return
	}

	// We are already checking for sunday, so we can suppress the error
	pricePeriod, _ := PricePeriodFromDay(weekday, tod)
	ticker.Prices[pricePeriod] = price
}

// Set a price period for a specific time. Timezone is not taken into account during
// this operation.
func (ticker *PriceTicker) SetPriceForTime(priceTime time.Time, price int) {
	weekday := priceTime.Weekday()
	if weekday == 0 {
		ticker.PurchasePrice = price
		return
	}
	pricePeriod, _ := PricePeriodFromTime(priceTime)
	ticker.Prices[pricePeriod] = price
}
