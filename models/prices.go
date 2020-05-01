package models

type hasPrices interface {
	MinPrice() int
	MaxPrice() int
}

type prices struct {
	min int
	max int
}

func (prices *prices) MinPrice() int {
	return prices.min
}

func (prices *prices) MaxPrice() int {
	return prices.max
}

func (prices *prices) UpdateMin(value int, useHigher bool) {
	update := (useHigher && value > prices.min) ||
		(!useHigher && value < prices.min) ||
		prices.min == 0

	if update {
		prices.min = value
	}
}

func (prices *prices) UpdateMax(value int) {
	if value > prices.max {
		prices.max = value
	}
}

func (prices *prices) Update(otherPrices hasPrices, useHigherMin bool) {
	prices.UpdateMin(otherPrices.MinPrice(), useHigherMin)
	prices.UpdateMax(otherPrices.MaxPrice())
}
