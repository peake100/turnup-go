package models

type PotentialPricePeriod struct {
	prices
	PricePeriod  PricePeriod
	PatternPhase PatternPhase

	// We're going to store the chances of any particular price happening. Because of
	// the way prices are rounded, the upper and lower bounds will often have a lower
	// chance of happening than prices in the center.
}

// Returns ``true`` if ``price`` falls within the price range of this potential period.
// Used by the predictor to remove phase permutations that do not match the current
// price values of a user.
func (potential *PotentialPricePeriod) IsValidPrice(price int) bool {
	// if the price is zero, it means the price is unknown, so we pass it.
	if price == 0 {
		return true
	}

	return price >= potential.prices.MinPrice() &&
		price <= potential.prices.MaxPrice()
}
