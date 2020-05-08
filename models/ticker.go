package models

import (
	"github.com/peake100/turnup-go/values"
)

type PriceTicker struct {
	// The previous week's price pattern
	PreviousPattern PricePattern

	// The purchase price on sunday for this week
	PurchasePrice int

	// There are 12 buy-price periods in a week, we are going to store the 12 buy prices
	// in a 12-int array. A price of 'zero' will stand for 'not available'
	//
	// Because PricePeriod is an extension of int, we can access the array with
	// PricePeriod objects.
	Prices NookPriceArray
}

func NewTicker(purchasePrice int, previousPattern PricePattern) *PriceTicker {
	return &PriceTicker{
		PreviousPattern: previousPattern,
		PurchasePrice:   purchasePrice,
		Prices:          [values.PricePeriodCount]int{},
	}
}
