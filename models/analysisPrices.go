package models

type hasPrices interface {
	MinPrice() int
	GuaranteedPrice() int
	MaxPrice() int
}

type prices struct {
	// Price info
	minPrice int
	guaranteedPrice int
	maxPrice        int

	// chance info
	minChance float64
	maxChance float64
	midChance float64
}

// The absolute minimum price that may occur.
func (prices *prices) MinPrice() int {
	return prices.minPrice
}

// The highest guaranteed to happen minimum price that may occur. On
// PotentialPricePeriod objects, this is the *lowest possible price* for the given
// period, but on Week, Pattern, and Prediction object this is the minimum guaranteed
// price, or the highest price we can guarantee will occur this week.
func (prices *prices) GuaranteedPrice() int {
	return prices.guaranteedPrice
}

// The potential maximum price for this period / week / pattern / prediction.
func (prices *prices) MaxPrice() int {
	return prices.maxPrice
}

// Returns the chance of this price range resulting in this price.
func (prices *prices) PriceChance(price int) float64 {
	switch {
	case price == prices.maxPrice:
		return prices.maxChance
	case price == prices.guaranteedPrice:
		return prices.minChance
	default:
		return prices.midChance
	}
}

func (prices *prices) updateMin(value int) (updated bool) {
	updated = prices.minPrice == 0 || value < prices.minPrice

	if updated {
		prices.minPrice = value
	}

	return updated
}

func (prices *prices) updateGuaranteed(value int, useHigher bool) (updated bool) {
	updated = ((useHigher && value > prices.guaranteedPrice) ||
		(!useHigher && value < prices.guaranteedPrice) ||
		prices.guaranteedPrice == 0) &&
		value != 0

	if updated {
		prices.guaranteedPrice = value
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
	otherPrices hasPrices, useHigherGuaranteed bool,
) (guaranteedUpdated bool, maxUpdated bool, minUpdated bool) {
	minUpdated = prices.updateMin(otherPrices.MinPrice())
	guaranteedUpdated = prices.updateGuaranteed(
		otherPrices.GuaranteedPrice(), useHigherGuaranteed,
	)
	maxUpdated = prices.updateMax(otherPrices.MaxPrice())
	return guaranteedUpdated, maxUpdated, minUpdated
}
