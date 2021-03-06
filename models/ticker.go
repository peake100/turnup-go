package models

import (
	"github.com/peake100/turnup-go/values"
)

type PriceTicker struct {
	// The previous week's price pattern
	PreviousPattern PricePattern

	// The purchase price on sunday for this week
	PurchasePrice int

	// The current price period. We need to support not knowing what the current
	// price is if we are charting data for someone else's island, but need to give
	// accurate future price ranges, so we will need to explicitly know from the
	// user what price period the island is currently in.
	CurrentPeriod PricePeriod

	// There are 12 buy-price periods in a week, we are going to store the 12 buy prices
	// in a 12-int array. A price of 'zero' will stand for 'not available'
	//
	// Because PricePeriod is an extension of int, we can access the array with
	// PricePeriod objects.
	Prices NookPriceArray
}

func NewTicker(
	purchasePrice int,
	previousPattern PricePattern,
	currentPeriod PricePeriod,
) *PriceTicker {
	return &PriceTicker{
		PreviousPattern: previousPattern,
		PurchasePrice:   purchasePrice,
		CurrentPeriod:   currentPeriod,
		Prices:          [values.PricePeriodCount]int{},
	}
}
