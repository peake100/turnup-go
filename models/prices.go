package models

type hasPrices interface {
	MinPrice() int
	MaxPrice() int
}

type prices struct {
	// Price info
	min int
	max int

	// Chance info
	minChance float64
	maxChance float64
	midChance float64
}

func (prices *prices) MinPrice() int {
	return prices.min
}

func (prices *prices) MaxPrice() int {
	return prices.max
}

// Returns the chance of this price range resulting in this price
func (prices *prices) PriceChance(price int) float64 {
	switch {
	case price == prices.max:
		return prices.maxChance
	case price == prices.min:
		return prices.minChance
	default:
		return prices.midChance
	}
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
