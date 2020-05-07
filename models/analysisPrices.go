package models

type hasPrices interface {
	MinPrice() int
	MaxPrice() int
}

type prices struct {
	// Price info
	minPrice int
	maxPrice int

	// chance info
	minChance float64
	maxChance float64
	midChance float64
}

func (prices *prices) MinPrice() int {
	return prices.minPrice
}

func (prices *prices) MaxPrice() int {
	return prices.maxPrice
}

// Returns the chance of this price range resulting in this price
func (prices *prices) PriceChance(price int) float64 {
	switch {
	case price == prices.maxPrice:
		return prices.maxChance
	case price == prices.minPrice:
		return prices.minChance
	default:
		return prices.midChance
	}
}

func (prices *prices) updateMin(value int, useHigher bool) (updated bool) {
	updated = ((useHigher && value > prices.minPrice) ||
		(!useHigher && value < prices.minPrice) ||
		prices.minPrice == 0) &&
		value != 0

	if updated {
		prices.minPrice = value
	}

	return updated
}

func (prices *prices) updateMax(value int) (updated bool) {
	updated = value > prices.maxPrice

	if updated {
		prices.maxPrice = value
	}

	return updated
}

func (prices *prices) updatePrices(
	otherPrices hasPrices, useHigherMin bool,
) (minUpdated bool, maxUpdated bool) {
	minUpdated = prices.updateMin(otherPrices.MinPrice(), useHigherMin)
	maxUpdated = prices.updateMax(otherPrices.MaxPrice())
	return minUpdated, maxUpdated
}
